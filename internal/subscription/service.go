package subscription

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"weather-app/internal/database/models"
	"weather-app/internal/database/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenNotFound  = errors.New("token not found")
	ErrTokenWrongType = errors.New("invalid token type")
	ErrTokenEmpty     = errors.New("token is empty")
)

type ConfirmationMailServiceInterface interface {
	SendConfirmationMail(email, confirmationUrl, unsubscribeUrl string) error
}

type UserRepositoryInterface interface {
	CreateUserWithSubscriptionAndTokens(
		email, city, frequency string,
		tokenTypes []string,
		generateToken func() (string, error),
	) (*repository.CreateUserWithSubscriptionAndTokensResult, error)

	GetByEmail(email string) (*models.User, error)
	UpdateUserConfirmationAndDeleteToken(userID uuid.UUID, tokenID uuid.UUID) error
	DeleteUserWithTokensAndSubscription(userID uuid.UUID) error
}

type TokenRepositoryInterface interface {
	GetToken(value string) (*models.Token, error)
}

type SubscriptionService struct {
	userRepo  UserRepositoryInterface
	tokenRepo TokenRepositoryInterface

	ms ConfirmationMailServiceInterface
}

func NewSubscriptionService(
	userRepo UserRepositoryInterface,
	tokenRepo TokenRepositoryInterface,
	mailService ConfirmationMailServiceInterface,
) *SubscriptionService {
	return &SubscriptionService{userRepo: userRepo, tokenRepo: tokenRepo, ms: mailService}
}

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrConfirmationMailError = errors.New("something went wrong with confirmation email")
)

func generateTokenDefault() (string, error) {
	return generateToken(32)
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// TODO: Move to other place. Should be common
func BuildTokenURL(base, apiPath, token string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Join the path and token properly
	u.Path = path.Join(u.Path, apiPath, token)

	return u.String(), nil
}

// TODO: Validate city
func (srv *SubscriptionService) Subscribe(email, city, frequency string) error {
	_, err := srv.userRepo.GetByEmail(email)

	if err == nil {
		// user already exists
		log.Printf("User %s already exists\n", email)

		return ErrUserAlreadyExists

	} else if err != gorm.ErrRecordNotFound {
		// database error
		log.Printf("Database error: %s\n", err.Error())

		return fmt.Errorf("error getting user: %w", err)
	}

	tokenTypes := []string{models.TokenTypeConfirm, models.TokenTypeUnsubscribe}

	result, err := srv.userRepo.CreateUserWithSubscriptionAndTokens(email, city, frequency, tokenTypes, generateTokenDefault)

	if err != nil {
		// database error

		return fmt.Errorf("error creating user: %w", err)
	}
	var confirmationToken, unsubscribeToken *models.Token

	confirmationToken, ok := result.Tokens[models.TokenTypeConfirm]

	if !ok {
		return fmt.Errorf("error getting confirmation token: %w", err)
	}

	unsubscribeToken, ok = result.Tokens[models.TokenTypeUnsubscribe]

	if !ok {
		return fmt.Errorf("error getting unsubscribe token: %w", err)
	}

	confirmUrl, err := BuildTokenURL(os.Getenv("BASE_URL"), "/api/confirm/", confirmationToken.Value)
	if err != nil {
		return fmt.Errorf("error building confirmation url: %w", err)
	}

	unsubscribeUrl, err := BuildTokenURL(os.Getenv("BASE_URL"), "/api/unsubscribe/", unsubscribeToken.Value)
	if err != nil {
		return fmt.Errorf("error building unsubscribe url: %w", err)
	}

	// TODO: Move to mail-sender container and send in chunks. Not one by one
	err = srv.ms.SendConfirmationMail(email, confirmUrl, unsubscribeUrl)

	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfirmationMailError, err)
	}

	return nil
}

func (srv *SubscriptionService) Confirm(tokenValue string) error {
	if tokenValue == "" {
		return ErrTokenEmpty
	}

	token, err := srv.tokenRepo.GetToken(tokenValue)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTokenNotFound
		} else {
			// database error

			return fmt.Errorf("error getting token: %w", err)
		}
	}

	if token.Type != models.TokenTypeConfirm {
		return ErrTokenWrongType
	}

	err = srv.userRepo.UpdateUserConfirmationAndDeleteToken(token.UserID, token.ID)

	if err != nil {
		// database error

		return fmt.Errorf("error deleting token: %w", err)
	}

	return nil
}

func (srv *SubscriptionService) Unsubscribe(tokenValue string) error {
	if tokenValue == "" {
		return ErrTokenEmpty
	}

	token, err := srv.tokenRepo.GetToken(tokenValue)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTokenNotFound
		} else {
			// database error

			return fmt.Errorf("error getting token: %w", err)
		}
	}

	if token.Type != models.TokenTypeUnsubscribe {
		return ErrTokenWrongType
	}

	err = srv.userRepo.DeleteUserWithTokensAndSubscription(token.UserID)

	if err != nil {
		// database error

		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}
