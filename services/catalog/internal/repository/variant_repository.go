package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type VariantRepository interface {
	Create(ctx context.Context, item *domain.Variant) error
	GetBySKU(ctx context.Context, sku string) (*domain.Variant, error)
	UpdateInventory(ctx context.Context, sku string, quantityChange int) error
	UpdatePrice(ctx context.Context, sku string, newPrice float64) error
}

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) VariantRepository {
	return &productVariantRepository{db: db}
}

func (pv *productVariantRepository) Create(ctx context.Context, item *domain.Variant) error {
	err := gorm.G[domain.Variant](pv.db).Create(ctx, item)
	if err != nil {
		return fmt.Errorf("repository: failed to create product variant: %w", err)
	}
	return nil
}

func (pv *productVariantRepository) GetBySKU(ctx context.Context, sku string) (*domain.Variant, error) {
	res, err := gorm.G[*domain.Variant](pv.db).Where("sku = ?", sku).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product variant by sku: %w", err)
	}

	return res, nil
}

func (pv *productVariantRepository) UpdateInventory(ctx context.Context, sku string, quantityChange int) error {
	rowsAffected, err := gorm.G[*domain.Variant](pv.db).
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

func (pv *productVariantRepository) UpdatePrice(ctx context.Context, sku string, newPrice float64) error {
	_, err := gorm.G[*domain.Variant](pv.db).
		Where("sku = ?", sku).
		Update(ctx, "price", newPrice) // Just update the column directly

	if err != nil {
		return fmt.Errorf("repository: failed to update price: %w", err)
	}
	return nil
}
