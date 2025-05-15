package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email       string    `gorm:"uniqueIndex;not null"`
	IsConfirmed bool
	CreatedAt   time.Time

	Subscriptions Subscription `gorm:"constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
