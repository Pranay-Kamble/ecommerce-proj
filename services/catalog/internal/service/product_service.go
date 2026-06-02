package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"

	"ecommerce/pkg/broker"
	"ecommerce/pkg/logger"
	pb "ecommerce/pkg/protobufs/catalog"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"

	"go.uber.org/zap"
)

type ProductService interface {
	CreateProduct(ctx context.Context, sellerPublicID, categoryPublicID string, product *domain.Product) error
	UpdateProduct(ctx context.Context, sellerPublicID string, productPublicID string, categoryPublicID string, updatedData *domain.Product) error
	DeleteProduct(ctx context.Context, sellerUserID string, productPublicID string) error

	GetProductByPublicID(ctx context.Context, publicID string) (*domain.Product, error)
	VerifyVariants(ctx context.Context, variantIDs []string) ([]*domain.Variant, error)
	DecreaseInventory(ctx context.Context, items []*pb.InventoryItem) error

	ListAllProducts(ctx context.Context, page, limit int) ([]*domain.Product, error)
	ListProductsByCategory(ctx context.Context, categoryPublicID string, limit, offset int) ([]*domain.Product, error)
	ListProductsBySeller(ctx context.Context, sellerUserID string, limit, offset int) ([]*domain.Product, error)
}

type productService struct {
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
	sellerRepo   repository.SellerRepository
	variantRepo  repository.VariantRepository
	broker       *broker.RabbitMQClient
}

func (p *productService) ListAllProducts(ctx context.Context, page, limit int) ([]*domain.Product, error) {
	productsPerPage, err := strconv.Atoi(os.Getenv("DEFAULT_LIMIT"))
	if err != nil {
		productsPerPage = 50
	}

	offset := (page - 1) * productsPerPage

	products, err := p.productRepo.GetAll(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get all products: %w", err)
	}
	return products, nil
}

func (p *productService) UpdateProduct(ctx context.Context, sellerPublicID string, productPublicID string, categoryPublicID string, updatedData *domain.Product) error {
	existingProduct, err := p.productRepo.GetByPublicID(ctx, productPublicID)

	if err != nil {
		return fmt.Errorf("service: failed to find product: %w", err)
	}
	if existingProduct == nil {
		return fmt.Errorf("service: product does not exist")
	}

	seller, err := p.sellerRepo.GetByPublicID(ctx, sellerPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to find seller: %w", err)
	} else if seller == nil {
		return fmt.Errorf("service: seller does not exist")
	} else if seller.ID != existingProduct.SellerID {
		return fmt.Errorf("service: unauthorized. seller does not own this product")
	}

	err = p.checkProductValidity(updatedData.Title, updatedData.Description, updatedData.Brand)
	if err != nil {
		return fmt.Errorf("service: incorrect/invalid product data: %w", err)
	}

	category, err := p.categoryRepo.GetByPublicID(ctx, categoryPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to find category id : %w", err)
	} else if category == nil {
		return fmt.Errorf("service: incorrect/invalid category public id %s", productPublicID)
	}

	updatedData.ID = existingProduct.ID
	updatedData.PublicID = existingProduct.PublicID
	updatedData.SellerID = existingProduct.SellerID
	updatedData.CategoryID = category.ID

	err = p.productRepo.Update(ctx, updatedData)
	if err != nil {
		return fmt.Errorf("service: failed to perform update : %w", err)
	}

	// Publish product.updated event for search indexing
	p.publishProductEvent(ctx, "product.updated", updatedData.PublicID)

	return nil
}

func (p *productService) DeleteProduct(ctx context.Context, sellerPublicID string, productPublicID string) error {
	product, err := p.productRepo.GetByPublicID(ctx, productPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to find product public id : %w", err)
	} else if product == nil {
		return fmt.Errorf("service: incorrect/invalid product public id %s", productPublicID)
	}

	seller, err := p.sellerRepo.GetByPublicID(ctx, sellerPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to find seller public id : %w", err)
	} else if seller == nil || seller.ID != product.SellerID {
		return fmt.Errorf("service: incorrect/invalid seller public id %s", productPublicID)
	}

	err = p.productRepo.Delete(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("service: failed to delete product : %w", err)
	}

	// Publish product.deleted event for search indexing
	deleteEvent := map[string]interface{}{
		"event":     "product.deleted",
		"public_id": product.PublicID,
	}
	if pubErr := p.broker.Publish(ctx, "catalog_events", "product.deleted", deleteEvent); pubErr != nil {
		logger.Error("service: failed to publish product.deleted event", zap.Error(pubErr))
	}

	return nil
}

func (p *productService) GetProductByPublicID(ctx context.Context, publicID string) (*domain.Product, error) {
	product, err := p.productRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get product by public id: %w", err)
	} else if product == nil {
		return nil, fmt.Errorf("service: product not found")
	}

	return product, nil
}

func (p *productService) ListProductsByCategory(ctx context.Context, categoryPublicID string, limit, offset int) ([]*domain.Product, error) {
	category, err := p.categoryRepo.GetByPublicID(ctx, categoryPublicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get product by category id: %w", err)
	} else if category == nil {
		return []*domain.Product{}, nil
	}

	products, err := p.productRepo.GetByCategoryID(ctx, category.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get products by category id: %w", err)
	}
	return products, nil
}

func (p *productService) ListProductsBySeller(ctx context.Context, sellerUserID string, limit, offset int) ([]*domain.Product, error) {
	seller, err := p.sellerRepo.GetByPublicID(ctx, sellerUserID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get products by seller id: %w", err)
	} else if seller == nil {
		return []*domain.Product{}, nil
	}
	products, err := p.productRepo.GetBySellerID(ctx, seller.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get products by seller id: %w", err)
	}
	return products, nil
}

func NewProductService(categoryRepo repository.CategoryRepository, productRepo repository.ProductRepository, sellerRepo repository.SellerRepository, variantRepo repository.VariantRepository, broker *broker.RabbitMQClient) ProductService {
	return &productService{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		sellerRepo:   sellerRepo,
		variantRepo:  variantRepo,
		broker:       broker,
	}
}

func (p *productService) CreateProduct(c context.Context, sellerPublicID, categoryPublicID string, product *domain.Product) error {
	err := p.checkProductValidity(product.Title, product.Description, product.Brand)
	if err != nil {
		return fmt.Errorf("service: invalid product details")
	}

	seller, err := p.sellerRepo.GetByPublicID(c, sellerPublicID)
	if err != nil {
		return fmt.Errorf("service: unable to find seller by public id: %w", err)
	} else if seller == nil {
		return fmt.Errorf("service: seller not found")
	}

	category, err := p.categoryRepo.GetByPublicID(c, categoryPublicID)

	if err != nil {
		return fmt.Errorf("service: unable to find category by public id: %w", err)
	} else if category == nil {
		return fmt.Errorf("service: category not found")
	}

	product.SellerID = seller.ID
	product.CategoryID = category.ID

	err = p.productRepo.Create(c, product)
	if err != nil {
		return fmt.Errorf("service: unable to create product: %w", err)
	}

	// Publish product.created event for search indexing
	p.publishProductEvent(c, "product.created", product.PublicID)

	return nil
}

func (p *productService) checkProductValidity(title, description, brand string) error {
	if len(title) < 3 {
		return fmt.Errorf("product name is too short (minimum 3 characters)")
	}
	if len(description) > 500 {
		return fmt.Errorf("product description is too long (maximum 500 characters)")
	}
	if len(brand) == 0 {
		return fmt.Errorf("product brand cannot be empty")
	}

	return nil
}

func (p *productService) VerifyVariants(ctx context.Context, variantIDs []string) ([]*domain.Variant, error) {
	if len(variantIDs) == 0 {
		return nil, nil
	}

	variants, err := p.productRepo.GetVariantsByPublicIDs(ctx, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("service: failed to verify variants: %w", err)
	}

	return variants, nil
}

func (p *productService) DecreaseInventory(ctx context.Context, items []*pb.InventoryItem) error {
	for _, item := range items {
		err := p.variantRepo.UpdateInventoryByPublicID(ctx, item.VariantId, -int(item.Quantity))
		if err != nil {
			return fmt.Errorf("service: failed to decrease inventory for variant %s: %w", item.VariantId, err)
		}
	}
	return nil
}

// publishProductEvent re-fetches the product with all preloads, computes
// min/max price from variants, and publishes to the catalog_events exchange.
func (p *productService) publishProductEvent(ctx context.Context, eventType, publicID string) {
	product, err := p.productRepo.GetByPublicID(ctx, publicID)
	if err != nil || product == nil {
		logger.Error("service: failed to fetch product for event publishing",
			zap.String("public_id", publicID), zap.Error(err))
		return
	}

	minPrice := math.MaxFloat64
	maxPrice := 0.0
	inStock := false

	for _, v := range product.Variants {
		if v.Price < minPrice {
			minPrice = v.Price
		}
		if v.Price > maxPrice {
			maxPrice = v.Price
		}
		if v.Inventory > 0 {
			inStock = true
		}
	}

	// If no variants exist, set prices to 0
	if len(product.Variants) == 0 {
		minPrice = 0
	}

	// Extract primary image URL
	imageURL := ""
	for _, img := range product.Images {
		if img.IsPrimary {
			imageURL = img.URL
			break
		}
	}
	if imageURL == "" && len(product.Images) > 0 {
		imageURL = product.Images[0].URL
	}

	event := map[string]interface{}{
		"event": eventType,
		"product": map[string]interface{}{
			"public_id":     product.PublicID,
			"title":         product.Title,
			"brand":         product.Brand,
			"description":   product.Description,
			"category_name": product.Category.Name,
			"seller_name":   product.Seller.Name,
			"slug":          product.Slug,
			"min_price":     minPrice,
			"max_price":     maxPrice,
			"in_stock":      inStock,
			"image_url":     imageURL,
		},
	}

	if pubErr := p.broker.Publish(ctx, "catalog_events", eventType, event); pubErr != nil {
		logger.Error("service: failed to publish catalog event",
			zap.String("event", eventType),
			zap.String("public_id", publicID),
			zap.Error(pubErr))
	} else {
		logger.Info("service: published catalog event",
			zap.String("event", eventType),
			zap.String("public_id", publicID))
	}
}
