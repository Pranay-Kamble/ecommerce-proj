package domain

import (
	"time"

	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	UserID    string    `gorm:"not null;index" json:"userID"`
	TokenHash string    `gorm:"not null;uniqueIndex" json:"tokenHash"`
	ExpiresOn time.Time `gorm:"not null" json:"expiresOn"`
	FamilyID  string    `gorm:"not null;index" json:"familyID"`
	IsUsed    bool      `gorm:"default:false" json:"isUsed"`
	IsRevoked bool      `gorm:"default:false" json:"isRevoked"`
}

func NewToken(userID, tokenHash, familyID string, expiry time.Time) *Token {
	return &Token{
		UserID:    userID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		ExpiresOn: expiry,
		IsRevoked: false,
		IsUsed:    false,
	}
}
