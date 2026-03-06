package service

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"
)

type VariantService interface {
	CreateVariant(c context.Context, productPublicID string, sellerPublicID string, item *domain.Variant) error
	UpdatePriceAndStock(c context.Context, sku string, newPrice float64, stock int) error
	UpdateVariant(c context.Context, productPublicID string, variantPublicID string, sellerPublicID string, newVariant *domain.Variant) error
	GetBySKU(c context.Context, sku string) (*domain.Variant, error)
}

type variantService struct {
	variantRepo repository.VariantRepository
	productRepo repository.ProductRepository
	sellerRepo  repository.SellerRepository
}

func NewVariantService(variantRepo repository.VariantRepository, productRepo repository.ProductRepository, sellerRepo repository.SellerRepository) VariantService {
	return &variantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
		sellerRepo:  sellerRepo,
	}
}

func (v *variantService) UpdateVariant(c context.Context, productPublicID string, variantPublicID string, sellerPublicID string, newVariant *domain.Variant) error {
	product, err := v.productRepo.GetByPublicID(c, productPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to get product by public id: %w", err)
	} else if product == nil {
		return fmt.Errorf("service: no product exists")
	}

	existingVariant, err := v.variantRepo.GetByPublicID(c, variantPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to get variant by public id: %w", err)
	} else if existingVariant == nil {
		return fmt.Errorf("service: no variant exists")
	}

	if existingVariant.ProductID != product.ID {
		return fmt.Errorf("service: product id does not match")
	}

	seller, err := v.sellerRepo.GetByPublicID(c, sellerPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to get seller by public id: %w", err)
	} else if seller == nil {
		return fmt.Errorf("service: no seller exists")
	}

	if seller.ID != product.SellerID {
		return fmt.Errorf("service: seller does not own the product")
	}

	if newVariant.Price < 0 || newVariant.Inventory < 0 {
		return fmt.Errorf("service: price and inventory must be non negative")
	}

	if existingVariant.SKU != newVariant.SKU {
		variantWithSameSKU, innerErr := v.variantRepo.GetBySKU(c, newVariant.SKU)
		if innerErr != nil {
			return fmt.Errorf("service: failed to get variant by sku: %w", innerErr)
		} else if variantWithSameSKU != nil {
			return fmt.Errorf("service: variant with the same SKU already exists")
		}
	}

	existingVariant.Title = newVariant.Title
	existingVariant.Price = newVariant.Price
	existingVariant.SKU = newVariant.SKU
	existingVariant.Inventory = newVariant.Inventory
	existingVariant.Images = newVariant.Images
	existingVariant.Specifications = newVariant.Specifications

	err = v.variantRepo.UpdateByID(c, &existingVariant.ID, newVariant)
	if err != nil {
		return fmt.Errorf("service: failed to update product variant: %w", err)
	}

	return nil
}

func (v *variantService) CreateVariant(c context.Context, productPublicID string, sellerPublicID string, item *domain.Variant) error {
	product, err := v.productRepo.GetByPublicID(c, productPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to get product by public id: %w", err)
	} else if product == nil {
		return fmt.Errorf("service: no product exists")
	}
	item.ProductID = product.ID

	seller, err := v.sellerRepo.GetByID(c, product.SellerID)
	if err != nil {
		return fmt.Errorf("service: failed to get seller by id: %w", err)
	} else if seller == nil {
		return fmt.Errorf("service: no seller exists")
	}

	if seller.PublicID != sellerPublicID {
		return fmt.Errorf("service: seller does not own the product")
	}

	err = v.variantRepo.Create(c, item)
	if err != nil {
		if err.Error() == "repository: variant with the same SKU already exists" {
			return fmt.Errorf("service: variant with the same SKU already exists")
		}
		return fmt.Errorf("service: failed to create product variant: %w", err)
	}

	return nil
}

func (v *variantService) UpdatePriceAndStock(c context.Context, sku string, newPrice float64, stock int) error {
	variant, err := v.variantRepo.GetBySKU(c, sku)
	if err != nil {
		return fmt.Errorf("service: failed to get product variant by sku: %w", err)
	} else if variant == nil {
		return fmt.Errorf("service: no product variant")
	}
	if variant.Inventory+stock < 0 {
		return fmt.Errorf("service: invalid stock update, final quantity must be non negative")
	}
	err = v.variantRepo.UpdateInventory(c, sku, stock)
	if err != nil {
		return fmt.Errorf("service: failed to update product variant inventory: %w", err)
	}

	if variant.Price != newPrice {
		err = v.variantRepo.UpdatePrice(c, sku, newPrice)
		if err != nil {
			return fmt.Errorf("service: failed to update product variant price: %w", err)
		}
	}

	return nil
}

func (v *variantService) GetBySKU(c context.Context, sku string) (*domain.Variant, error) {
	variant, err := v.variantRepo.GetBySKU(c, sku)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get product variant: %w", err)
	} else if variant == nil {
		return nil, fmt.Errorf("service: no product variant")
	}
	return variant, nil
}
