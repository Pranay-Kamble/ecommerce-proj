package domain

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID          string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"-"`
	PublicID    string  `gorm:"type:varchar(20);uniqueIndex;not null" json:"id"`
	UserID      string  `gorm:"type:varchar(21);not null;index" json:"user_id"`
	TotalAmount float64 `gorm:"not null" json:"total_amount"`
	Status      string  `gorm:"type:varchar(20);default:'pending'" json:"status"`

	ShippingName    string `gorm:"type:varchar(100);not null" json:"shipping_name"`
	ShippingPhone   string `gorm:"type:varchar(20);not null" json:"shipping_phone"`
	ShippingAddress string `gorm:"type:text;not null" json:"shipping_address"`
	ShippingCity    string `gorm:"type:varchar(50);not null" json:"shipping_city"`
	ShippingState   string `gorm:"type:varchar(50);not null" json:"shipping_state"`
	ShippingZip     string `gorm:"type:varchar(20);not null" json:"shipping_zip"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	ID      string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"-"`
	OrderID string `gorm:"type:uuid;not null;index" json:"-"`

	ProductID string  `gorm:"type:varchar(25);not null" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"not null" json:"price"`
}
