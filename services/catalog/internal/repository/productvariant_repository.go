package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ProductVariantRepository interface {
	GetBySKU(ctx context.Context, sku string) (*domain.ProductVariant, error)
	UpdateInventory(ctx context.Context, sku string, quantityChange int) error
}

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (pv *productVariantRepository) GetBySKU(ctx context.Context, sku string) (*domain.ProductVariant, error) {
	res, err := gorm.G[*domain.ProductVariant](pv.db).Where("sku = ?", sku).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product variant by sku: %w", err)
	}

	return res, nil
}

func (pv *productVariantRepository) UpdateInventory(ctx context.Context, sku string, quantityChange int) error {
	_, err := gorm.G[*domain.ProductVariant](pv.db).Where("sku = ?", sku).Update(ctx, "inventory", gorm.Expr("inventory + ?", quantityChange))
	if err != nil {
		return fmt.Errorf("repository: failed to update product variant quantity: %w", err)
	}

	return nil
}
