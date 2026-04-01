package repository

import (
	"context"
	"ecommerce/services/payment/internal/domain"
	"errors"
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

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		//Save to Payment Service Database
		var payment domain.Payment
		innerErr := tx.Where("gateway_session_id = ?", sessionID).First(&payment).Error
		if innerErr != nil {
			if errors.Is(innerErr, gorm.ErrRecordNotFound) {
				return errors.New("payment record not found for the given session ID")
			}
			return fmt.Errorf("could not find payment: %w", innerErr)
		}

		payment.Status = status
		innerErr = tx.Save(&payment).Error
		if innerErr != nil {
			return fmt.Errorf("could not update payment status: %w", innerErr)
		}

		//Save to Outbox Database for Message Broker
		if status == "success" {
			payload := fmt.Sprintf(`{"order_id": "%s", "status": "paid"}`, payment.OrderID)

			outboxEvent := &domain.OutboxEvent{
				EventType: "OrderPaid",
				Payload:   payload,
				Processed: false,
			}

			innerErr = tx.Create(outboxEvent).Error
			if innerErr != nil {
				return fmt.Errorf("failed to save event to outbox: %w", innerErr)
			}
		}

		return nil
	})

	return err
}
