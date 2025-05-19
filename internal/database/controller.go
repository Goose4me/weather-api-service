package database

import (
	"fmt"
	"os"

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

	err = db.AutoMigrate(&models.User{}, &models.Subscription{}, &models.Token{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate DB: %w", err)
	}

	return db, nil
}

// TODO: Refactor controller to be less bulky and separate logic

func GetToken(value string, db *gorm.DB) (*models.Token, error) {
	var token models.Token

	err := db.Where("value = ?", value).First(&token).Error
	if err != nil {
		return nil, err
	}

	return &token, nil
}
