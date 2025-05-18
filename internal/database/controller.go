package database

import (
	"fmt"
	"os"
	"time"

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
		if err == gorm.ErrRecordNotFound {
			return nil, err
		} else {
			// Real DB error
			return nil, err
		}
	}

	return &user, nil
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

func CreateToken(subscription *models.Subscription, value, token_type string, db *gorm.DB) (*models.Token, error) {
	var existingToken models.Token
	err := db.Where("value = ? AND type = ? AND subscription_id = ?", value, token_type, subscription.ID).
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
		Value:          value,
		Type:           token_type,
		SubscriptionID: subscription.ID,
		CreatedAt:      time.Now(),
	}

	if err := db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}
