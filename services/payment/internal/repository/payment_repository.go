package repository

import (
	"context"
	"ecommerce/services/payment/internal/domain"
	"fmt"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	UpdatePaymentStatusByStripeID(ctx context.Context, stripeID string, status string) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	err := gorm.G[domain.Payment](r.db).Create(ctx, payment)
	if err != nil {
		return fmt.Errorf("repository: could not create payment: %w", err)
	}
	return nil
}

func (r *paymentRepository) UpdatePaymentStatusByStripeID(ctx context.Context, stripeID string, status string) error {
	_, err := gorm.G[*domain.Payment](r.db).Where("stripe_id = ?", stripeID).Update(ctx, "status", status)
	if err != nil {
		return fmt.Errorf("repository: could not update payment status: %w", err)
	}
	return nil
}
