package service

import (
	"context"
	"fmt"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"
)

type CartService interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	AddItem(ctx context.Context, userID string, item domain.CartItem) (*domain.Cart, error)
	RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error)
	ClearCart(ctx context.Context, userID string) error
}

type cartService struct {
	cartRepo repository.CartRepository
}

func NewCartService(cartRepo repository.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	return s.cartRepo.GetCart(ctx, userID)
}

func (s *cartService) AddItem(ctx context.Context, userID string, newItem domain.CartItem) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get cart: %w", err)
	}

	itemExists := false
	for i, item := range cart.Items {
		if item.ProductVariantID == newItem.ProductVariantID {
			cart.Items[i].Quantity += newItem.Quantity
			cart.Items[i].Price = newItem.Price
			itemExists = true
			break
		}
	}

	if !itemExists {
		cart.Items = append(cart.Items, newItem)
	}

	err = s.cartRepo.SaveCart(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("service: failed to save cart: %w", err)
	}

	return cart, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get cart: %w", err)
	}

	var updatedItems []domain.CartItem
	for _, item := range cart.Items {
		if item.ProductVariantID != productID {
			updatedItems = append(updatedItems, item)
		}
	}

	cart.Items = updatedItems

	err = s.cartRepo.SaveCart(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("service: failed to save cart: %w", err)
	}

	return cart, nil
}

func (s *cartService) ClearCart(ctx context.Context, userID string) error {
	return s.cartRepo.ClearCart(ctx, userID)
}
