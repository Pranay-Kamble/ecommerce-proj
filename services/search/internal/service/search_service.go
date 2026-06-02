package service

import (
	"context"
	"fmt"

	"ecommerce/services/search/internal/domain"
	"ecommerce/services/search/internal/repository"
)

// SearchService defines the business logic interface for product search.
type SearchService interface {
	SearchProducts(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error)
	IndexProduct(ctx context.Context, doc domain.ProductDocument) error
	DeleteProduct(ctx context.Context, publicID string) error
}

type searchService struct {
	esRepo repository.ElasticRepository
}

// NewSearchService creates a new SearchService.
func NewSearchService(esRepo repository.ElasticRepository) SearchService {
	return &searchService{esRepo: esRepo}
}

func (s *searchService) SearchProducts(ctx context.Context, filters domain.SearchFilters) (*domain.SearchResult, error) {
	result, err := s.esRepo.SearchProducts(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("service: failed to search products: %w", err)
	}
	return result, nil
}

func (s *searchService) IndexProduct(ctx context.Context, doc domain.ProductDocument) error {
	err := s.esRepo.IndexProduct(ctx, doc)
	if err != nil {
		return fmt.Errorf("service: failed to index product: %w", err)
	}
	return nil
}

func (s *searchService) DeleteProduct(ctx context.Context, publicID string) error {
	err := s.esRepo.DeleteProduct(ctx, publicID)
	if err != nil {
		return fmt.Errorf("service: failed to delete product from index: %w", err)
	}
	return nil
}
