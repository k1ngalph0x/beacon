package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Issue struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	Fingerprint string `gorm:"uniqueIndex:idx_project_fp"`
	ProjectID   string `gorm:"uniqueIndex:idx_project_fp"`
	Title       string
	Level       string
	Count       int
	FirstSeen   time.Time
	LastSeen    time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}


type Event struct {
	ProjectID  string     `json:"project_id"`
	Timestamp  time.Time  `json:"timestamp"`
	Level      string     `json:"level"`
	Message    string     `json:"message"`
	StackTrace *string    `json:"stack_trace"`
}

func (i *Issue) BeforeCreate(tx *gorm.DB) error{
	if i.ID == ""{
		i.ID = uuid.New().String()
	}

	return nil
}