package domain

import "time"

type CustomerProfile struct {
	UserID string `gorm:"type:varchar(21);primaryKey" json:"user_id"`
	Name   string `gorm:"type:varchar(100);not null" json:"name"`
	Phone  string `gorm:"type:varchar(20)" json:"phone"`

	Addresses []Address `gorm:"foreignKey:UserID" json:"addresses"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Address struct {
	ID     string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID string `gorm:"type:varchar(50);not null;index" json:"-"`
	Title  string `gorm:"type:varchar(50)" json:"title"`

	AddressLine string `gorm:"type:text;not null" json:"address_line"`
	City        string `gorm:"type:varchar(50);not null" json:"city"`
	State       string `gorm:"type:varchar(50);not null" json:"state"`
	ZipCode     string `gorm:"type:varchar(20);not null" json:"zip_code"`

	IsDefault bool `gorm:"default:false" json:"is_default"`
}
