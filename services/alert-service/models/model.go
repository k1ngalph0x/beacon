package models

import "time"

type AlertRule struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	ProjectID string `gorm:"index;not null"`
	Threshold int    `gorm:"not null"`
	Level     string `gorm:"not null"`
	IsActive  bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
