package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TokenTypeConfirm     = "confirm"
	TokenTypeUnsubscribe = "unsubscribe"
)

type Token struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Value     string    `gorm:"uniqueIndex;not null"`
	Type      string    `gorm:"not null"` // "confirm", "unsubscribe"
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	CreatedAt time.Time
}

func (s *Token) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}
