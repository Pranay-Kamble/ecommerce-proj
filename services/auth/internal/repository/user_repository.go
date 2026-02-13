package repository

import (
	"ecommerce/services/auth/internal/domain"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByProviderID(providerID string) (*domain.User, error)
	UpdateVerified(userID string) error
}

type userRepository struct {
	db *gorm.DB
}

func (userRepository *userRepository) CreateUser(user *domain.User) error {
	return nil
}

func (userRepository *userRepository) GetUserByEmail(email string) (*domain.User, error) {
	return nil, nil
}

func (userRepository *userRepository) GetUserByProviderID(providerID string) (*domain.User, error) {
	return nil, nil
}

func (userRepository *userRepository) UpdateVerified(userID string) error {
	return nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
