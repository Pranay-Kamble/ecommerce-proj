package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sixafter/nanoid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;" json:"-"`
	PublicID    string    `gorm:"type:varchar(25);uniqueIndex;not null" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Slug        string    `gorm:"type:varchar(50);uniqueIndex" json:"slug"`
	SellerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"sellerId"`
	CategoryID  uuid.UUID `gorm:"type:uuid;not null;index" json:"categoryId"`
	Price       float64   `gorm:"type:decimal(10,2);" json:"price"`
	Inventory   int       `gorm:"type:int;default:0" json:"inventory"`

	CreatedAt time.Time      `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"precision:6" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"precision:6" json:"deletedAt"`

	Seller   Seller   `gorm:"foreignKey:SellerID;references:ID" json:"seller,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

type Category struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;" json:"-"`
	PublicID string    `gorm:"type:varchar(25);uniqueIndex:idx_public_id" json:"id"`
	Name     string    `gorm:"type:varchar(50);uniqueIndex:idx_name" json:"name"`
	Path     string    `gorm:"type:ltree;uniqueIndex:idx_path" json:"path"`

	CreatedAt time.Time      `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"precision:6" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"precision:6" json:"deletedAt"`
}

type Seller struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;" json:"-"`
	PublicID string    `gorm:"type:varchar(25);uniqueIndex" json:"id"`
	UserID   string    `gorm:"type:varchar(21);uniqueIndex" json:"userId"`
	Name     string    `gorm:"type:varchar(50);" json:"name"`
	Slug     string    `gorm:"type:varchar(50)" json:"slug"`

	CreatedAt time.Time      `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"precision:6" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"precision:6" json:"deletedAt"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("domain: could not generate product ID: %w", err)
		}
		p.ID = newID
	}

	if p.PublicID == "" {
		publicID, err := nanoid.New()
		if err != nil {
			return fmt.Errorf("domain: could not generate product public ID: %w", err)
		}
		publicIDStr := "itm_" + publicID.String()
		p.PublicID = publicIDStr
	}

	return nil
}

func (s *Seller) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("domain: could not generate seller ID: %w", err)
		}
		s.ID = newID
	}

	if s.PublicID == "" {
		publicID, err := nanoid.New()
		if err != nil {
			return fmt.Errorf("domain: could not generate seller public ID: %w", err)
		}
		publicIDStr := "sel_" + publicID.String()
		s.PublicID = publicIDStr
	}

	return nil
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("domain: could not generate category ID: %w", err)
		}
		c.ID = newID
	}

	if c.PublicID == "" {
		publicID, err := nanoid.New()
		if err != nil {
			return fmt.Errorf("domain: could not generate category public ID: %w", err)
		}
		publicIDStr := "cat_" + publicID.String()
		c.PublicID = publicIDStr
	}

	return nil
}
