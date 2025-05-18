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

	sub, err := database.CreateSubscription(user, city, frequency, srv.DB)

	if err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
	}

	tokenValue, err := generateToken(32)

	if err != nil {
		return fmt.Errorf("error genertating token: %w", err)
	}

	token, err := database.CreateToken(sub, tokenValue, "confirm", srv.DB)

	if err != nil {
		return fmt.Errorf("error creating token: %w", err)
	}

	log.Printf("Created token: %s\n", token.Value)

	return nil
}
