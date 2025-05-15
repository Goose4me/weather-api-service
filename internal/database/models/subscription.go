package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	City      string    `gorm:"not null"`
	Frequency string    `gorm:"not null"` // "hourly" or "daily"
	CreatedAt time.Time
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}
