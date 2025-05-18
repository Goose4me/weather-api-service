package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"weather-app/internal/database/models"
)

func getDSNFromEnv() (string, error) {
	dsn := os.Getenv("DBDSN")
	if dsn == "" {
		return "", fmt.Errorf("enviroment setup error: no DBDSN specified")
	}

	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")

	if db_user == "" {
		return "", fmt.Errorf("enviroment setup error: no DB_USER specified")
	}

	if db_password == "" {
		return "", fmt.Errorf("enviroment setup error: no DB_PASSWORD specified")
	}

	if db_name == "" {
		return "", fmt.Errorf("enviroment setup error: no DB_NAME specified")
	}

	dsn = fmt.Sprintf(dsn,
		db_user,
		db_password,
		db_name)

	return dsn, nil
}

func InitDB() (*gorm.DB, error) {
	dsn, err := getDSNFromEnv()

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DSN: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Subscription{}, &models.WeatherLog{}, &models.Token{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate DB: %w", err)
	}

	return db, nil
}

func GetUser(email string, db *gorm.DB) (*models.User, error) {
	var user models.User

	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByID(id uuid.UUID, db *gorm.DB) (*models.User, error) {
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteUser(user *models.User, db *gorm.DB) error {
	return db.Delete(user).Error
}

func DeleteUserWithTokensAndSubscription(userID uuid.UUID, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
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

func CreateNewUser(email string, db *gorm.DB) (*models.User, error) {
	user := models.User{
		Email:       email,
		IsConfirmed: false,
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateUserWithSubscriptionAndTokens(
	email, city, frequency string,
	tokenTypes []string,
	generateToken func() (string, error),
	db *gorm.DB,
) (*models.User, error) {

	var user models.User

	err := db.Transaction(func(tx *gorm.DB) error {
		createdTime := time.Now()

		// Create User
		user = models.User{
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
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(user *models.User, db *gorm.DB) error {

	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func UpdateUserConfirmed(user *models.User, isConfirmed bool, db *gorm.DB) error {
	if err := db.Model(&user).Update("is_confirmed", isConfirmed).Error; err != nil {
		return err
	}

	return nil
}

func GetSubscription(user *models.User, db *gorm.DB) (*models.Subscription, error) {
	var subscription models.Subscription

	err := db.Where("user_id = ?", user.ID).First(&subscription).Error
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func DeleteSubscription(sub *models.Subscription, db *gorm.DB) error {
	return db.Delete(sub).Error
}

func CreateSubscription(user *models.User, city, frequency string, db *gorm.DB) (*models.Subscription, error) {
	var existingSub models.Subscription
	err := db.Where("user_id = ? AND city = ? AND frequency = ?", user.ID, city, frequency).
		First(&existingSub).Error

	if err == nil {
		// Subscription already exists
		return nil, fmt.Errorf("subscription already exists")
	} else if err != gorm.ErrRecordNotFound {
		// Unexpected DB error
		return nil, err
	}

	// Add new subscription
	sub := models.Subscription{
		UserID:    user.ID,
		City:      city,
		Frequency: frequency,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&sub).Error; err != nil {
		return nil, err
	}

	return &sub, nil
}

func GetToken(value string, db *gorm.DB) (*models.Token, error) {
	var token models.Token

	err := db.Where("value = ?", value).First(&token).Error
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func DeleteToken(token *models.Token, db *gorm.DB) error {
	return db.Delete(token).Error
}

func DeleteTokensByUserID(db *gorm.DB, userID uuid.UUID) error {
	return db.Where("user_id = ?", userID).Delete(&models.Token{}).Error
}

func CreateToken(user *models.User, value, token_type string, db *gorm.DB) (*models.Token, error) {
	var existingToken models.Token
	err := db.Where("value = ? AND type = ? AND user_id = ?", value, token_type, user.ID).
		First(&existingToken).Error

	if err == nil {
		// Token already exists
		return nil, fmt.Errorf("token already exists")
	} else if err != gorm.ErrRecordNotFound {
		// Unexpected DB error
		return nil, err
	}

	// Add new token
	token := models.Token{
		Value:     value,
		Type:      token_type,
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}
