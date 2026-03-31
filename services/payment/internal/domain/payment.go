package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sixafter/nanoid"
	"gorm.io/gorm"
)

type Payment struct {
	ID       uuid.UUID `gorm:"primary_key" json:"-"`
	PublicID string    `gorm:"varchar(30);uniqueIndex" json:"id"`
	OrderID  string    `gorm:"varchar(30);not null; index" json:"order_id"`
	UserID   string    `gorm:"varchar(30);not null; index" json:"user_id"`

	Provider         string `gorm:"type:varchar(20);default:'stripe'" json:"provider"`
	GatewaySessionID string `gorm:"varchar(255);uniqueIndex" json:"gateway_session_id"`

	Amount   int64  `gorm:"not null,min=0" json:"amount"`
	Currency string `gorm:"varchar(10); default='inr'" json:"currency"`

	Status string `gorm:"type:varchar(20);default:'pending'" json:"status"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (r *Payment) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("domain: could not generate payment ID: %w", err)
		}
		r.ID = newID
	}
	if r.PublicID == "" {
		publicID, err := nanoid.New()
		if err != nil {
			return fmt.Errorf("domain: could not generate payment public ID: %w", err)
		}
		r.PublicID = publicID.String()
	}

	return nil
}
