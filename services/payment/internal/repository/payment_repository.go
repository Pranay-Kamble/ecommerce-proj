package repository

import (
	"context"
	"ecommerce/services/payment/internal/domain"
	"fmt"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	UpdatePaymentStatusBySessionID(ctx context.Context, sessionID string, status string) error
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

func (r *paymentRepository) UpdatePaymentStatusBySessionID(ctx context.Context, sessionID string, status string) error {
	rowsAffected, err := gorm.G[*domain.Payment](r.db).Where("stripe_id = ?", sessionID).Update(ctx, "status", status)
	if err != nil {
		return fmt.Errorf("repository: could not update payment status: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository: sessionID not found :%w", err)
	}
	return nil
}
