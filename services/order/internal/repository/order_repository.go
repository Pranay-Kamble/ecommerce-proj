package repository

import (
	"context"
	"fmt"

	"ecommerce/services/order/internal/domain"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByPublicID(ctx context.Context, publicID string) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, order *domain.Order) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return gorm.G[domain.Order](tx).Create(ctx, order)
	})

	if err != nil {
		return fmt.Errorf("repository: failed to create order transaction: %w", err)
	}
	return nil
}

func (r *orderRepository) GetOrderByPublicID(ctx context.Context, publicID string) (*domain.Order, error) {
	order, err := gorm.G[domain.Order](r.db).
		Preload("Items", nil).
		Where("public_id = ?", publicID).
		First(ctx)

	if err != nil {
		return nil, fmt.Errorf("repository: failed to get order: %w", err)
	}
	return &order, nil
}

func (r *orderRepository) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	orders, err := gorm.G[domain.Order](r.db).
		Preload("Items", nil).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("repository: failed to fetch user orders: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) UpdateOrder(ctx context.Context, order *domain.Order) error {
	_, err := gorm.G[*domain.Order](r.db).Where("public_id = ?", order.PublicID).Updates(ctx, order)
	if err != nil {
		return fmt.Errorf("repository: failed to update order: %w", err)
	}
	return nil
}
