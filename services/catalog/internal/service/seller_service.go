package service

import (
	"context"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/repository"
	"fmt"
)

type SellerService interface {
	CreateSeller(c context.Context, seller *domain.Seller) error
	GetByUserID(c context.Context, userID string) (*domain.Seller, error)
	GetByPublicID(c context.Context, publicID string) (*domain.Seller, error)
}

type sellerService struct {
	sellerRepo repository.SellerRepository
}

func NewSellerService(sellerRepo repository.SellerRepository) SellerService {
	return &sellerService{
		sellerRepo: sellerRepo,
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
