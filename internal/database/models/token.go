package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Value          string    `gorm:"uniqueIndex;not null"`
	Type           string    `gorm:"not null"` // "confirm", "unsubscribe"
	SubscriptionID uuid.UUID
	CreatedAt      time.Time
}
