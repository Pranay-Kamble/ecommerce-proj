package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID       uuid.UUID `gorm:"primary_key" json:"-"`
	PublicID string    `gorm:"varchar(30);uniqueIndex" json:"id"`
	OrderID  string    `gorm:"varchar(30);not null; index" json:"order_id"`
	UserID   string    `gorm:"varchar(30);not null; index" json:"user_id"`

	StripeID string `gorm:"varchar(255);uniqueIndex" json:"stripe_id"`
	Amount   int64  `gorm:"not null,min=0" json:"amount"`
	Currency string `gorm:"varchar(10); default='inr'" json:"currency"`

	Status string `gorm:"type:varchar(20);default:'pending'" json:"status"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
