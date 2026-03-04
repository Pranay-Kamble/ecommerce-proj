package service

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"
)

type ProductService interface {
	CreateProduct(ctx context.Context, sellerPublicID, categoryPublicID string, product *domain.Product) error
	UpdateProduct(ctx context.Context, sellerPublicID string, productPublicID string, categoryPublicID string, updatedData *domain.Product) error
	DeleteProduct(ctx context.Context, sellerUserID string, productPublicID string) error

	GetProductByPublicID(ctx context.Context, publicID string) (*domain.Product, error)

	ListProductsByCategory(ctx context.Context, categoryPublicID string, limit, offset int) ([]*domain.Product, error)
	ListProductsBySeller(ctx context.Context, sellerUserID string, limit, offset int) ([]*domain.Product, error)
}

type productService struct {
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
	sellerRepo   repository.SellerRepository
}

func (p *productService) UpdateProduct(ctx context.Context, sellerPublicID string, productPublicID string, categoryPublicID string, updatedData *domain.Product) error {
	existingProduct, err := p.productRepo.GetByPublicID(ctx, productPublicID)

	if err != nil {
		return fmt.Errorf("service: product does not exist")
	}
	if existingProduct == nil {
		return fmt.Errorf("service: product does not exist")
	}

	seller, err := p.sellerRepo.GetByPublicID(ctx, sellerPublicID)
	if err != nil {
		return fmt.Errorf("service: failed to find seller : %w", err)
	} else if seller == nil || seller.ID != existingProduct.SellerID {
		return fmt.Errorf("service: unauthorized. seller not valid or does not own this product %s", sellerPublicID)
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

func NewProductService(categoryRepo repository.CategoryRepository, productRepo repository.ProductRepository, sellerRepo repository.SellerRepository) ProductService {
	return &productService{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		sellerRepo:   sellerRepo,
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
