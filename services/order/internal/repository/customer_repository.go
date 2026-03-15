package repository

import (
	"context"
	"errors"
	"fmt"

	"ecommerce/services/order/internal/domain"

	"gorm.io/gorm"
)

type CustomerRepository interface {
	CreateProfile(ctx context.Context, profile *domain.CustomerProfile) error
	GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error)
	AddAddress(ctx context.Context, address *domain.Address) error
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) CreateProfile(ctx context.Context, profile *domain.CustomerProfile) error {
	err := r.db.WithContext(ctx).
		Where("user_id = ?", profile.UserID).
		FirstOrCreate(profile).Error

	if err != nil {
		return fmt.Errorf("repository: failed to create customer profile: %w", err)
	}
	return nil
}

func (r *customerRepository) GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error) {
	profile, err := gorm.G[domain.CustomerProfile](r.db).
		Preload("Addresses", nil).
		Where("user_id = ?", userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: failed to fetch profile: %w", err)
	}

	return &profile, nil
}

func (r *customerRepository) AddAddress(ctx context.Context, address *domain.Address) error {
	err := gorm.G[domain.Address](r.db).Create(ctx, address)
	if err != nil {
		return fmt.Errorf("repository: failed to add address: %w", err)
	}
	return nil
}
