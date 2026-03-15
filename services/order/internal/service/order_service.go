package service

import (
	"context"
	"ecommerce/pkg/logger"
	"fmt"

	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/repository"

	"github.com/sixafter/nanoid"
)

type OrderService interface {
	Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, error)
	GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
	cartRepo  repository.CartRepository
}

func NewOrderService(orderRepo repository.OrderRepository, cartRepo repository.CartRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
	}
}

func (s *orderService) Checkout(ctx context.Context, userID string, name, phone string, address domain.Address) (*domain.Order, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get cart for checkout: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("service: cannot checkout with an empty cart")
	}

	var totalAmount float64
	var orderItems []domain.OrderItem

	for _, item := range cart.Items {
		//In the next phase, we can make a gRPC call to Catalog
		// to verify this price hasn't been maliciously manipulated by the frontend!

		totalAmount += item.Price * float64(item.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	id, err := nanoid.NewWithLength(8)
	if err != nil {
		return nil, fmt.Errorf("service: failed to generate order number: %w", err)
	}
	orderNumber := fmt.Sprintf("ORD-%s", id)

	order := &domain.Order{
		PublicID:        orderNumber,
		UserID:          userID,
		TotalAmount:     totalAmount,
		Status:          "paid",
		ShippingName:    name,
		ShippingPhone:   phone,
		ShippingAddress: address.AddressLine,
		ShippingCity:    address.City,
		ShippingState:   address.State,
		ShippingZip:     address.ZipCode,
		Items:           orderItems,
	}

	if err = s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("service: failed to save order: %w", err)
	}

	if err = s.cartRepo.ClearCart(ctx, userID); err != nil {
		logger.Error("warning: failed to clear cart after successful checkout for user %s" + userID)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, publicID string, userID string) (*domain.Order, error) {
	order, err := s.orderRepo.GetOrderByPublicID(ctx, publicID, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch order: %w", err)
	}
	return order, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	orders, err := s.orderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch user orders: %w", err)
	}
	return orders, nil
}
