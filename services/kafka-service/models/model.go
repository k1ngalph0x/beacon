package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Events struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID      string    `gorm:"type:text;not null;index:idx_beacon_events_project_id" json:"project_id"`
	Level          string    `gorm:"type:text;not null;index:idx_beacon_events_level" json:"level"`
	Message        string    `gorm:"type:text;not null" json:"message"`
	StackTrace     *string   `gorm:"type:text" json:"stack_trace,omitempty"` 
	EventTimestamp time.Time `gorm:"not null;index:idx_beacon_events_timestamp" json:"event_timestamp"`
	ReceivedAt     time.Time `gorm:"not null;autoCreateTime" json:"received_at"`
	KafkaPartition *int      `gorm:"type:int" json:"kafka_partition,omitempty"` 
	KafkaOffset    *int64    `gorm:"type:bigint" json:"kafka_offset,omitempty"` 
}

func (e *Events) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

