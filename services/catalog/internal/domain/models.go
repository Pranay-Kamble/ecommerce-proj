package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sixafter/nanoid"
	"gorm.io/gorm"
)

type Image struct {
	URL       string `json:"url"`
	AltText   string `json:"altText"`
	IsPrimary bool   `json:"isPrimary"`
}
type Product struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;" json:"-"`
	PublicID   string    `gorm:"type:varchar(25);uniqueIndex;not null" json:"id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"categoryId"`
	SellerID   uuid.UUID `gorm:"type:uuid;not null;index" json:"sellerId"`

	Title       string                 `gorm:"type:varchar(500);not null" json:"title"`
	Brand       string                 `gorm:"type:varchar(100)" json:"brand"`
	Description string                 `gorm:"type:text" json:"description"`
	Highlights  []string               `gorm:"type:jsonb;serializer:json" json:"highlights"`
	Dimensions  map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"dimensions"`

	Variants []ProductVariant `gorm:"foreignKey:ProductID;references:ID" json:"variants"`
	Images   []Image          `gorm:"type:jsonb;serializer:json" json:"images"`

	Seller   Seller   `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	CreatedAt time.Time      `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"precision:6" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"precision:6" json:"deletedAt"`
}

type ProductVariant struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;" json:"-"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`

	SKU       string  `gorm:"type:varchar(100);not null" json:"sku"`
	Price     float64 `gorm:"type:decimal(10,2);" json:"price"`
	Inventory int     `gorm:"type:int;default:0" json:"inventory"`

	Images []Image `gorm:"type:jsonb;serializer:json" json:"images"`

	Specifications map[string]interface{} `gorm:"type:jsonb;serializer:json;index:idx_specs,type:gin" json:"specifications"`

	CreatedAt time.Time      `gorm:"precision:6" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"precision:6" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"precision:6" json:"deletedAt"`
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

func (p *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("domain: could not generate product variant ID: %w", err)
		}
		p.ID = newID
	}
	return nil
}
