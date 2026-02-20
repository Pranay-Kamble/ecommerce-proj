package repository

import (
	"context"
	"ecommerce/services/auth/internal/domain"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByProviderID(ctx context.Context, providerID string) (*domain.User, error)
	UpdateVerified(ctx context.Context, userID string) error
}

type userRepository struct {
	db *gorm.DB
}

func (u *userRepository) CreateUser(ctx context.Context, user *domain.User) error {

	err := gorm.G[domain.User](u.db).Create(ctx, user)

	if err != nil {
		return fmt.Errorf("repository: could not create user: %w", err)
	}

	return nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	res, err := gorm.G[domain.User](u.db).Where("email = ?", email).First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("repository: could not get user by email: %w", err)
	}

	return &res, nil
}

func (u *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	res, err := gorm.G[domain.User](u.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: could not get user by id: %w", err)
	}

	return &res, nil
}

func (u *userRepository) GetUserByProviderID(ctx context.Context, providerID string) (*domain.User, error) {
	res, err := gorm.G[domain.User](u.db).Where("provider_id = ?", providerID).First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("repository: could not get user by provider id: %w", err)
	}

	return &res, nil
}

func (u *userRepository) UpdateVerified(ctx context.Context, userID string) error {
	_, err := gorm.G[domain.User](u.db).Where("id = ?", userID).Update(ctx, "is_verified", true)

	if err != nil {
		return fmt.Errorf("repository: could not update verified user: %w", err)
	}

	return nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
