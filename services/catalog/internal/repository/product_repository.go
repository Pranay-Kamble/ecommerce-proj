package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error

	GetAll(ctx context.Context, offset, limit int) ([]*domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetByPublicID(ctx context.Context, publicID string) (*domain.Product, error)
	GetBySellerID(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	GetByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*domain.Product, error)

	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, product *domain.Product) error
}

type productRepository struct {
	db *gorm.DB
}

func (p *productRepository) GetAll(ctx context.Context, offset, limit int) ([]*domain.Product, error) {
	products, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Offset(offset).Limit(limit).Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("repository: failed to get all products: %w", err)
	}
	return products, nil
}

func (p *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := gorm.G[*domain.Product](p.db).Select("Variants").Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("repository: failed to perform delete: %w", err)
	}
	return nil
}

func (p *productRepository) Update(ctx context.Context, product *domain.Product) error {
	err := p.db.WithContext(ctx).Save(product).Error

	if err != nil {
		return fmt.Errorf("repository: failed to perform update: %w", err)
	}
	return nil
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (p *productRepository) Create(ctx context.Context, product *domain.Product) error {
	err := gorm.G[domain.Product](p.db).Create(ctx, product)
	if err != nil {
		return fmt.Errorf("repository: could not create product: %w", err)
	}
	return nil
}

func (p *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Where("id = ?", id).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product by ID: %w", err)
	}
	return product, nil
}

func (p *productRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Product, error) {
	product, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Where("public_id = ?", publicID).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product by public ID: %w", err)
	}
	return product, nil
}

func (p *productRepository) GetBySellerID(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*domain.Product, error) {
	products, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Where("seller_id = ? ", sellerID).
		Offset(offset).Limit(limit).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get products by seller ID: %w", err)
	}
	return products, nil
}

func (p *productRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Where("slug = ?", slug).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get product by slug: %w", err)
	}
	return product, nil
}

func (p *productRepository) GetByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*domain.Product, error) {
	products, err := gorm.G[*domain.Product](p.db).
		Preload("Category", nil).
		Preload("Seller", nil).
		Preload("Variants", nil).
		Where("category_id = ?", categoryID).
		Offset(offset).Limit(limit).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get products by category ID: %w", err)
	}
	return products, nil
}
