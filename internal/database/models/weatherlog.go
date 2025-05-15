package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
