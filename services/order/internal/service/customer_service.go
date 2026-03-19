package service

import (
	"context"
	"ecommerce/pkg/broker"
	"ecommerce/pkg/logger"
	"encoding/json"
	"fmt"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"

	"go.uber.org/zap"
)

type CustomerService interface {
	CreateProfile(ctx context.Context, userID, name, phone string) (*domain.CustomerProfile, error)
	GetProfile(ctx context.Context, userID string) (*domain.CustomerProfile, error)
	AddAddress(ctx context.Context, userID string, address *domain.Address) error
}

type customerService struct {
	customerRepo repository.CustomerRepository
	broker       *broker.RabbitMQClient
}

func NewCustomerService(customerRepo repository.CustomerRepository, client *broker.RabbitMQClient) CustomerService {
	return &customerService{customerRepo: customerRepo, broker: client}
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

	type CustomerOnboardedEvent struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
	}

	event := CustomerOnboardedEvent{
		UserID: userID,
		Status: "onboarded",
	}

	payload, err := json.Marshal(event)
	if err != nil {
		logger.Error("service: failed to marshal onboard event:", zap.Error(err))
	} else {
		err = s.broker.Publish(ctx, "user_events", "customer.onboarded", payload)
		if err != nil {
			logger.Error("service: failed to publish onboard event", zap.Error(err))
		} else {
			logger.Info("service: published onboard event")
		}
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
