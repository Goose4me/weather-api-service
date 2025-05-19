package repository

import (
	"weather-app/internal/database/models"

	"gorm.io/gorm"
)

type TokenRepository struct {
	*BaseRepository
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *TokenRepository) GetToken(value string) (*models.Token, error) {
	var token models.Token

	err := r.db.Where("value = ?", value).First(&token).Error
	if err != nil {
		return nil, err
	}

	return &token, nil
}
