package service

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"
)

type VariantService interface {
	CreateVariant(c context.Context, productPublicID string, item *domain.Variant) error
	UpdatePriceAndStock(c context.Context, sku string, newPrice float64, stock int) error
	GetBySKU(c context.Context, sku string) (*domain.Variant, error)
}

type productVariantService struct {
	variantRepo repository.VariantRepository
	productRepo repository.ProductRepository
}

func NewProductVariantService(variantRepo repository.VariantRepository, productRepo repository.ProductRepository) VariantService {
	return &productVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
	}
}

func (pv *productVariantService) CreateVariant(c context.Context, productPublicID string, item *domain.Variant) error {
	product, err := pv.productRepo.GetByPublicID(c, productPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to get product by public id: %w", err)
	} else if product == nil {
		return fmt.Errorf("service: no product exists")
	}
	item.ProductID = product.ID

	err = pv.variantRepo.Create(c, item)
	if err != nil {
		return fmt.Errorf("service: failed to create product variant: %w", err)
	}

	return nil
}

func (pv *productVariantService) UpdatePriceAndStock(c context.Context, sku string, newPrice float64, stock int) error {
	variant, err := pv.variantRepo.GetBySKU(c, sku)
	if err != nil {
		return fmt.Errorf("service: failed to get product variant by sku: %w", err)
	} else if variant == nil {
		return fmt.Errorf("service: no product variant")
	}
	if variant.Inventory+stock < 0 {
		return fmt.Errorf("service: invalid stock update, final quantity must be non negative")
	}
	err = pv.variantRepo.UpdateInventory(c, sku, stock)
	if err != nil {
		return fmt.Errorf("service: failed to update product variant inventory: %w", err)
	}

	if variant.Price != newPrice {
		err = pv.variantRepo.UpdatePrice(c, sku, newPrice)
		if err != nil {
			return fmt.Errorf("service: failed to update product variant price: %w", err)
		}
	}

	return nil
}

func (pv *productVariantService) GetBySKU(c context.Context, sku string) (*domain.Variant, error) {
	variant, err := pv.variantRepo.GetBySKU(c, sku)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get product variant: %w", err)
	} else if variant == nil {
		return nil, fmt.Errorf("service: no product variant")
	}
	return variant, nil
}
