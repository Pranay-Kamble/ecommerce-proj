package service

import (
	"context"
	"ecommerce/services/payment/internal/repository"

	"github.com/stripe/stripe-go/v78"
)

type PaymentService interface {
	CreateCheckoutSession(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, string, error)
}

type paymentService struct {
	paymentRepository repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository, stripeSecretKey string) PaymentService {
	stripe.Key = stripeSecretKey
	return &paymentService{paymentRepository: repo}
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, string, error) {
	panic("implement me")
}
