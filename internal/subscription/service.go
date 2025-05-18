package subscription

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"weather-app/internal/database"
	"weather-app/internal/database/models"

	"gorm.io/gorm"
)

var (
	ErrTokenNotFound  = errors.New("token not found")
	ErrTokenWrongType = errors.New("invalid token type")
	ErrTokenEmpty     = errors.New("token is empty")
)

type SubscriptionService struct {
	DB *gorm.DB
}

func NewSubscriptionService(db *gorm.DB) *SubscriptionService {
	return &SubscriptionService{DB: db}
}

var ErrUserAlreadyExists = errors.New("user already exists")

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (srv *SubscriptionService) Subscribe(email, city, frequency string) error {
	var user *models.User
	_, err := database.GetUser(email, srv.DB)

	if err == nil {
		// user already exists
		log.Printf("User %s already exists\n", email)

		return ErrUserAlreadyExists

	} else if err == gorm.ErrRecordNotFound {
		// user doesn't exists. Crearte one
		user, err = database.CreateNewUser(email, srv.DB)

		if err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}

	} else {
		// database error
		log.Printf("Database error: %s\n", err.Error())

		return fmt.Errorf("error getting user: %w", err)
	}

	_, err = database.CreateSubscription(user, city, frequency, srv.DB)

	if err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
	}

	confirmTokenValue, err := generateToken(32)

	if err != nil {
		return fmt.Errorf("error generating confirm token: %w", err)
	}

	confirmToken, err := database.CreateToken(user, confirmTokenValue, models.TokenTypeConfirm, srv.DB)

	if err != nil {
		return fmt.Errorf("error creating confirm token: %w", err)
	}

	log.Printf("Created confirm token %s \n", confirmToken.Value)

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
