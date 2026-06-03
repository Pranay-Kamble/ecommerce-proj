package domain

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"-"`
	PublicID  string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"id"`
	OrderID   string         `gorm:"type:varchar(20);not null;index" json:"order_id"`
	ObjectKey string         `gorm:"type:varchar(255);not null" json:"-"`
	FileURL   string         `gorm:"type:text;not null" json:"file_url"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
