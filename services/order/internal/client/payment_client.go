package client

import (
	"context"
	"fmt"
	"time"

	"ecommerce/pkg/logger"
	pb "ecommerce/pkg/protobufs/payment"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentService interface {
	InitiatePayment(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, error)
	Close() error
}

type paymentGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.PaymentServiceClient
}

func NewPaymentClient(paymentServiceAddr string) (PaymentService, error) {
	conn, err := grpc.NewClient(paymentServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial payment service: %w", err)
	}

	client := pb.NewPaymentServiceClient(conn)

	logger.Info("gRPC Payment Client connected", zap.String("address", paymentServiceAddr))

	return &paymentGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (p *paymentGRPCClient) InitiatePayment(ctx context.Context, orderID string, userID string, amount int64, currency string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.CreatePaymentRequest{
		OrderId:  orderID,
		UserId:   userID,
		Amount:   amount,
		Currency: currency,
	}

	res, err := p.client.CreatePaymentSession(ctx, req)
	if err != nil {
		return "", fmt.Errorf("client: gRPC call to payment service failed: %w", err)
	}

	return res.PaymentUrl, nil
}

func (p *paymentGRPCClient) Close() error {
	return p.conn.Close()
}
