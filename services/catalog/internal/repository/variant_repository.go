package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VariantRepository interface {
	Create(ctx context.Context, item *domain.Variant) error
	GetBySKU(ctx context.Context, sku string) (*domain.Variant, error)
	GetByPublicID(ctx context.Context, publicID string) (*domain.Variant, error)
	UpdateInventory(ctx context.Context, sku string, quantityChange int) error
	UpdatePrice(ctx context.Context, sku string, newPrice float64) error
	UpdateByID(ctx context.Context, id *uuid.UUID, updated *domain.Variant) error

	Delete(ctx context.Context, id *uuid.UUID) error
}

type variantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) VariantRepository {
	return &variantRepository{db: db}
}

func (v *variantRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Variant, error) {
	res, err := gorm.G[*domain.Variant](v.db).Where("public_id = ?", publicID).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product variant by public id: %w", err)
	}

	return res, nil
}

func (v *variantRepository) Create(ctx context.Context, item *domain.Variant) error {
	err := gorm.G[domain.Variant](v.db).Create(ctx, item)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("repository: variant with the same SKU already exists")
	}
	if err != nil {
		return fmt.Errorf("repository: failed to create product variant: %w", err)
	}
	return nil
}

func (v *variantRepository) GetBySKU(ctx context.Context, sku string) (*domain.Variant, error) {
	res, err := gorm.G[*domain.Variant](v.db).Where("sku = ?", sku).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product variant by sku: %w", err)
	}

	return res, nil
}

func (v *variantRepository) UpdateInventory(ctx context.Context, sku string, quantityChange int) error {
	rowsAffected, err := gorm.G[*domain.Variant](v.db).
		Where("sku = ? AND inventory + ? >= 0", sku, quantityChange).
		Update(ctx, "inventory", gorm.Expr("inventory + ?", quantityChange))

	if err != nil {
		return fmt.Errorf("repository: failed to update inventory: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repository: insufficient stock or invalid SKU")
	}

	return nil
}

func (v *variantRepository) UpdatePrice(ctx context.Context, sku string, newPrice float64) error {
	_, err := gorm.G[*domain.Variant](v.db).
		Where("sku = ?", sku).
		Update(ctx, "price", newPrice) // Just update the column directly

	if err != nil {
		return fmt.Errorf("repository: failed to update price: %w", err)
	}
	return nil
}

func (v *variantRepository) UpdateByID(ctx context.Context, id *uuid.UUID, updated *domain.Variant) error {
	updates := &domain.Variant{
		Title:          updated.Title,
		Price:          updated.Price,
		SKU:            updated.SKU,
		Inventory:      updated.Inventory,
		Images:         updated.Images,
		Specifications: updated.Specifications,
	}
	_, err := gorm.G[*domain.Variant](v.db).Where("id = ?", id).Updates(ctx, updates)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("repository: variant with the same SKU already exists")
	}
	if err != nil {
		return fmt.Errorf("repository: failed to update product variant: %w", err)
	}
	return nil
}

func (v *variantRepository) Delete(ctx context.Context, id *uuid.UUID) error {
	_, err := gorm.G[*domain.Variant](v.db).Where("id = ?", id).Delete(ctx)

	if err != nil {
		return fmt.Errorf("repository: failed to delete product variant: %w", err)
	}
	return nil
}
