package service

import (
	"context"
	"ecommerce/pkg/broker"
	"ecommerce/pkg/logger"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"

	"go.uber.org/zap"
)

type SellerService interface {
	CreateSeller(c context.Context, seller *domain.Seller) error
	GetByUserID(c context.Context, userID string) (*domain.Seller, error)
	GetByPublicID(c context.Context, publicID string) (*domain.Seller, error)
}

type sellerService struct {
	sellerRepo repository.SellerRepository
	broker     *broker.RabbitMQClient
}

func NewSellerService(sellerRepo repository.SellerRepository, broker *broker.RabbitMQClient) SellerService {
	return &sellerService{
		sellerRepo: sellerRepo,
		broker:     broker,
	}
}

func (s *sellerService) CreateSeller(c context.Context, seller *domain.Seller) error {

	existingSeller, err := s.sellerRepo.GetByUserID(c, seller.UserID)
	if err != nil {
		return fmt.Errorf("service: failed to get seller by user ID: %w", err)
	} else if existingSeller != nil {
		return fmt.Errorf("service: user already exists")
	}

	existingSeller, err = s.sellerRepo.GetByGSTIN(c, seller.GSTIN)
	if err != nil {
		return fmt.Errorf("service: failed to get user by gstin: %w", err)
	} else if existingSeller != nil {
		return fmt.Errorf("service: gstin already exists")
	}
	err = s.sellerRepo.Create(c, seller)
	if err != nil {
		return fmt.Errorf("service: failed to create seller: %w", err)
	}

	event := map[string]interface{}{
		"user_id": seller.UserID,
		"status":  "onboarded",
	}
	err = s.broker.Publish(c, "user_events", "user.onboarded", event)
	if err != nil {
		logger.Error("service: failed to publish user onboarded event: %v", zap.Error(err))
	}

	return nil
}

func (s *sellerService) GetByUserID(c context.Context, userID string) (*domain.Seller, error) {
	seller, err := s.sellerRepo.GetByUserID(c, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get seller by user ID: %w", err)
	} else if seller == nil {
		return nil, nil
	}
	return seller, nil
}

func (s *sellerService) GetByPublicID(c context.Context, publicID string) (*domain.Seller, error) {
	seller, err := s.sellerRepo.GetByPublicID(c, publicID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get seller by public ID: %w", err)
	}
	if seller == nil {
		return nil, nil
	}
	return seller, nil
}
