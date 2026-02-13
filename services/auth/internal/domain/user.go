package domain

import (
	"ecommerce/pkg/logger"
	"time"

	"github.com/sixafter/nanoid"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"primaryKey;type:varchar(21);" json:"id"`
	Email    string `gorm:"uniqueIndex; not null" json:"email"`
	Password string `gorm:"not null; default:''" json:"-"`
	Role     string `gorm:"default: 'buyer'" json:"role"`

	Provider   string `gorm:"default:'email'" json:"provider"`
	ProviderID string `gorm:"index" json:"-"`

	IsVerified         bool       `gorm:"default:false" json:"isVerified"`
	VerificationCode   string     `gorm:"index" json:"-"`
	VerificationExpiry *time.Time `json:"-"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		id, err := nanoid.New()
		if err != nil {
			logger.Error("could not create nanoid " + err.Error())
			return err
		}

		u.ID = string(id)
	}
	return nil
}
