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

	Subscriptions []Subscription `gorm:"constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

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

type WeatherLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	City        string    `gorm:"not null"`
	Temperature float64   `gorm:"not null"`
	Humidity    int       `gorm:"not null"`
	Description string    `gorm:"not null"`
	FetchedAt   time.Time
}

func (w *WeatherLog) BeforeCreate(tx *gorm.DB) error {
	w.ID = uuid.New()
	return nil
}
