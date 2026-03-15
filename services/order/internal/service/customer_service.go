package service

import (
	"context"
	"fmt"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"
)

type CustomerService interface {
	CreateProfile(ctx context.Context, userID, name, phone string) (*domain.CustomerProfile, error)
	GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error)
	AddAddress(ctx context.Context, userID string, address *domain.Address) error
}

type customerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) CustomerService {
	return &customerService{customerRepo: customerRepo}
}

func (s *customerService) CreateProfile(ctx context.Context, userID, name, phone string) (*domain.CustomerProfile, error) {
	profile := &domain.CustomerProfile{
		UserID: userID,
		Name:   name,
		Phone:  phone,
	}

	err := s.customerRepo.CreateProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create profile: %w", err)
	}

	return profile, nil
}

func (s *customerService) GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error) {
	profile, err := s.customerRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get profile: %w", err)
	}
	return profile, nil
}

func (s *customerService) AddAddress(ctx context.Context, userID string, address *domain.Address) error {
	address.UserID = userID

	err := s.customerRepo.AddAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("service: failed to add address: %w", err)
	}

	return nil
}
