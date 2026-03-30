package main

import (
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
	pb "ecommerce/pkg/protobufs/payment"
	"ecommerce/services/payment/internal/handler"
	"ecommerce/services/payment/internal/repository"
	"ecommerce/services/payment/internal/service"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.Init("dev")
	err := godotenv.Load("../../../.env")
	if err != nil {
		logger.Error("Error loading .env file. Using local environment variables.")
	}

	PostgresDsn := os.Getenv("DATABASE_DSN")
	if PostgresDsn == "" {
		PostgresDsn = "host=localhost user=admin password=password dbname=payment_db port=5432 sslmode=disable"
	}

	db, err := database.NewPostgres(PostgresDsn)
	if err != nil {
		logger.Fatal("main: failed to connect to postgres database: ", zap.Error(err))
	}
	defer db.Close()

	PaymentGatewaySecretKey := os.Getenv("PAYMENT_GATEWAY_SECRET_KEY")

	paymentRepo := repository.NewPaymentRepository(db.DB)
	paymentService := service.NewPaymentService(paymentRepo, PaymentGatewaySecretKey)

	grpcHandler := handler.NewPaymentGrpcHandler(paymentService)
	grpcServer := grpc.NewServer()

	go func() {
		grpcTcpConn, innerErr := net.Listen("tcp", ":50052")
		if innerErr != nil {
			logger.Fatal("main: failed to listen on port 50052", zap.Error(innerErr))
		}

		pb.RegisterPaymentServiceServer(grpcServer, grpcHandler)
		logger.Info("main: gRPC server listening on :50052")

		innerErr = grpcServer.Serve(grpcTcpConn)
		if innerErr != nil {
			logger.Fatal("main: failed to start gRPC server", zap.Error(innerErr))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	logger.Info("main: shutting down gRPC server gracefully...")
	grpcServer.GracefulStop()
	logger.Info("main: server stopped safely")
}
