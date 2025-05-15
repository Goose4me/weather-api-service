package subscription

import (
	"fmt"
	"weather-app/internal/database"

	"gorm.io/gorm"
)

type SubscriptionService struct {
	DB *gorm.DB
}

func NewSubscriptionService(db *gorm.DB) *SubscriptionService {
	return &SubscriptionService{DB: db}
}

func (srv *SubscriptionService) Subscribe(email, city, frequency string) error {
	user, err := database.GetUser(email, srv.DB)

	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	err = database.CreateSubscription(&user, city, frequency, srv.DB)

	if err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
	}

	return nil
}
