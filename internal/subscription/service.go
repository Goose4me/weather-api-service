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
	"weather-app/internal/database"
	"weather-app/internal/database/models"
	"weather-app/internal/mail"

	"gorm.io/gorm"
)

var (
	ErrTokenNotFound  = errors.New("token not found")
	ErrTokenWrongType = errors.New("invalid token type")
	ErrTokenEmpty     = errors.New("token is empty")
)

type SubscriptionService struct {
	DB *gorm.DB
	ms *mail.MailService
}

func NewSubscriptionService(db *gorm.DB, mailService *mail.MailService) *SubscriptionService {
	return &SubscriptionService{DB: db, ms: mailService}
}

var ErrUserAlreadyExists = errors.New("user already exists")

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

func buildConfirmURL(base, apiPath, token string) (string, error) {
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
	_, err := database.GetUser(email, srv.DB)

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

	result, err := database.CreateUserWithSubscriptionAndTokens(email, city, frequency, tokenTypes, generateTokenDefault, srv.DB)

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

	confirmUrl, err := buildConfirmURL(os.Getenv("BASE_URL"), "/api/confirm/", confirmationToken.Value)
	if err != nil {
		return fmt.Errorf("error building confirmation url: %w", err)
	}

	unsubscribeUrl, err := buildConfirmURL(os.Getenv("BASE_URL"), "/api/unsubscribe/", unsubscribeToken.Value)
	if err != nil {
		return fmt.Errorf("error building unsubscribe url: %w", err)
	}

	// TODO: Move to separate container and send in chunks. Not 1 by 1
	srv.ms.SendConfirmationMail(email, confirmUrl, unsubscribeUrl)

	return nil
}

func (srv *SubscriptionService) Confirm(tokenValue string) error {
	if tokenValue == "" {
		return ErrTokenEmpty
	}

	token, err := database.GetToken(tokenValue, srv.DB)

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

	user, err := database.GetUserByID(token.UserID, srv.DB)
	if err != nil {
		// database error

		return fmt.Errorf("error getting user: %w", err)
	}

	err = database.UpdateUserConfirmed(user, true, srv.DB)

	if err != nil {
		// database error

		return fmt.Errorf("error updating user confirm: %w", err)
	}

	err = database.DeleteToken(token, srv.DB)

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

	token, err := database.GetToken(tokenValue, srv.DB)

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

	user, err := database.GetUserByID(token.UserID, srv.DB)
	if err != nil {
		// database error

		return fmt.Errorf("error getting user: %w", err)
	}

	err = database.DeleteUserWithTokensAndSubscription(user.ID, srv.DB)

	if err != nil {
		// database error

		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}
