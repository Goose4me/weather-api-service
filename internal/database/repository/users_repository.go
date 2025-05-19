package repository

import (
	"fmt"
	"log"
	"time"
	"weather-app/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, HandleDBError(err, "user")
	}

	return &user, nil
}

type UserEmailInfo struct {
	Email      string
	City       string
	TokenValue string
}

// TODO: Need to separate this big transactional functions and use BaseRepository::WithTransaction
func (r *UserRepository) GetUserEmailInfoBatch(limit, offset int, subscriptionFrequency string) ([]UserEmailInfo, error) {
	var results []UserEmailInfo

	err := r.db.Table("users").
		Select("users.email, subscriptions.city, tokens.value AS token_value").
		Joins("JOIN subscriptions ON subscriptions.user_id = users.id AND subscriptions.frequency = ?", subscriptionFrequency).
		Joins("JOIN tokens ON tokens.user_id = users.id AND tokens.type = ?", "unsubscribe").
		Where("users.is_confirmed = true").
		Order("users.created_at ASC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return results, nil
}

type CreateUserWithSubscriptionAndTokensResult struct {
	User         *models.User
	Subscription *models.Subscription
	Tokens       map[string]*models.Token
}

func (r *UserRepository) CreateUserWithSubscriptionAndTokens(
	email, city, frequency string,
	tokenTypes []string,
	generateToken func() (string, error),
) (*CreateUserWithSubscriptionAndTokensResult, error) {

	result := CreateUserWithSubscriptionAndTokensResult{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		createdTime := time.Now()

		// Create User
		user := models.User{
			Email:       email,
			IsConfirmed: false,
			CreatedAt:   createdTime,
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Create Subscription
		sub := models.Subscription{
			UserID:    user.ID,
			City:      city,
			Frequency: frequency,
			CreatedAt: createdTime,
		}
		if err := tx.Create(&sub).Error; err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}

		tokensMap := make(map[string]*models.Token)

		// Create Tokens
		for _, tokenType := range tokenTypes {
			value, err := generateToken()
			if err != nil {
				return fmt.Errorf("failed to generate token: %w", err)
			}

			token := models.Token{
				Value:     value,
				Type:      tokenType,
				UserID:    user.ID,
				CreatedAt: createdTime,
			}
			if err := tx.Create(&token).Error; err != nil {
				return fmt.Errorf("failed to create %s token: %w", tokenType, err)
			}

			log.Printf("Created %s token: %s", tokenType, value)

			tokensMap[tokenType] = &token
		}

		result.User = &user
		result.Subscription = &sub
		result.Tokens = tokensMap

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *UserRepository) UpdateUserConfirmationAndDeleteToken(userID uuid.UUID, tokenID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete token
		if err := tx.Delete(&models.Token{}, "id = ?", tokenID).Error; err != nil {
			return fmt.Errorf("failed to delete token: %w", err)
		}

		// Update the user
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("is_confirmed", true).Error; err != nil {

			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

func (r *UserRepository) DeleteUserWithTokensAndSubscription(userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all tokens for the user
		if err := tx.Where("user_id = ?", userID).Delete(&models.Token{}).Error; err != nil {
			return fmt.Errorf("failed to delete tokens: %w", err)
		}

		// Delete subscription for the user
		if err := tx.Where("user_id = ?", userID).Delete(&models.Subscription{}).Error; err != nil {
			return fmt.Errorf("failed to delete subscription: %w", err)
		}

		// Delete the user
		if err := tx.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
}
