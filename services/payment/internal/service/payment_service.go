package service

import (
	"context"
	"ecommerce/services/payment/internal/domain"
	"ecommerce/services/payment/internal/repository"
	"fmt"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

type PaymentService interface {
	CreateCheckoutSession(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, string, error)
	MarkPaymentAsSuccess(ctx context.Context, sessionID string) error
}

type paymentService struct {
	paymentRepository repository.PaymentRepository
}

func (s *paymentService) MarkPaymentAsSuccess(ctx context.Context, sessionID string) error {
	err := s.paymentRepository.UpdatePaymentStatusBySessionID(ctx, sessionID, "success")
	if err != nil {
		return fmt.Errorf("service: failed to mark payment as success: %w", err)
	}

	return nil
}

func NewPaymentService(repo repository.PaymentRepository, paymentGatewaySecretKey string) PaymentService {
	stripe.Key = paymentGatewaySecretKey
	return &paymentService{paymentRepository: repo}
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, string, error) {

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String("http://localhost:8085/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:          stripe.String("http://localhost:8085/cancel"),
		ClientReferenceID:  stripe.String(orderID),

		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Order #" + orderID),
					},
					UnitAmount: stripe.Int64(amount), // Remember: This is PAISE (300000 = ₹3000)
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", err
	}

	paymentRecord := &domain.Payment{
		OrderID:          orderID,
		UserID:           userID,
		GatewaySessionID: sess.ID,
		Amount:           amount,
		Currency:         currency,
		Status:           "pending",
	}

	err = s.paymentRepository.CreatePayment(ctx, paymentRecord)
	if err != nil {
		return "", "", err
	}

	return sess.ID, sess.URL, nil
}
