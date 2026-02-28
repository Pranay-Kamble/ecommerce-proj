package repository

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SellerRepository interface {
	Create(ctx context.Context, seller *domain.Seller) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Seller, error)
	GetByPublicID(ctx context.Context, publicID string) (*domain.Seller, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Seller, error)
}

type sellerRepository struct {
	db *gorm.DB
}

func NewSellerRepository(db *gorm.DB) SellerRepository {
	return &sellerRepository{db: db}
}

func (s *sellerRepository) Create(ctx context.Context, seller *domain.Seller) error {
	err := gorm.G[domain.Seller](s.db).Create(ctx, seller)
	if err != nil {
		return fmt.Errorf("repository: failed to create: %w", err)
	}

	return nil
}

func (s *sellerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Seller, error) {
	seller, err := gorm.G[*domain.Seller](s.db).Where("id = ?", id).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get seller by ID: %w", err)
	}

	return seller, nil
}

func (s *sellerRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.Seller, error) {
	seller, err := gorm.G[*domain.Seller](s.db).Where("public_id = ?", publicID).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get seller by public id: %w", err)
	}

	return seller, nil
}

func (s *sellerRepository) GetByUserID(ctx context.Context, userID string) (*domain.Seller, error) {
	seller, err := gorm.G[*domain.Seller](s.db).Where("user_id = ?", userID).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to get seller by user id: %w", err)
	}

	return seller, nil
}
