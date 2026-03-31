package handler

import (
	"context"

	pb "ecommerce/pkg/protobufs/payment"
	"ecommerce/services/payment/internal/service"
)

type PaymentGrpcHandler struct {
	pb.UnimplementedPaymentServiceServer
	paymentSvc service.PaymentService
}

func NewPaymentGrpcHandler(paymentSvc service.PaymentService) *PaymentGrpcHandler {
	return &PaymentGrpcHandler{paymentSvc: paymentSvc}
}

func (h *PaymentGrpcHandler) CreatePaymentSession(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {

	sessionID, url, err := h.paymentSvc.CreateCheckoutSession(ctx, req.OrderId, req.UserId, req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	return &pb.CreatePaymentResponse{
		PaymentUrl:    url,
		TransactionId: sessionID,
	}, nil
}
