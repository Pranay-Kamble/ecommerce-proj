package domain

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"-"`
	EventType string    `gorm:"type:varchar(50);not null"`
	Payload   string    `gorm:"type:jsonb;not null"`
	Processed bool      `gorm:"default:false"`
	CreatedAt time.Time
}
