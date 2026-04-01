package main

import (
	"context"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
	pb "ecommerce/pkg/protobufs/payment"
	"ecommerce/services/payment/internal/domain"
	"ecommerce/services/payment/internal/handler"
	"ecommerce/services/payment/internal/repository"
	"ecommerce/services/payment/internal/service"
	"ecommerce/services/payment/internal/workers"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.Init("dev")
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("Error loading .env file. Using local environment variables.")
	}

	postgresDsn := os.Getenv("DATABASE_DSN")
	if postgresDsn == "" {
		postgresDsn = "host=localhost user=admin password=password dbname=payment_db port=5432 sslmode=disable"
	}

	db, err := database.NewPostgres(postgresDsn)
	if err != nil {
		logger.Fatal("main: failed to connect to postgres database: ", zap.Error(err))
	}
	defer db.Close()

	err = db.DB.AutoMigrate(&domain.Payment{}, &domain.OutboxEvent{})
	if err != nil {
		logger.Fatal("main: failed to migrate database: ", zap.Error(err))
	}

	paymentGatewaySecretKey := os.Getenv("PAYMENT_GATEWAY_SECRET_KEY")

	paymentRepo := repository.NewPaymentRepository(db.DB)
	paymentService := service.NewPaymentService(paymentRepo, paymentGatewaySecretKey)

	rabbitmqUrl := os.Getenv("RABBIT_MQ_URL")
	if rabbitmqUrl == "" {
		rabbitmqUrl = "amqp://admin:password@localhost:5672/"
	}
	rabbitConn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}

	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer rabbitChannel.Close()

	err = rabbitChannel.ExchangeDeclare(
		"payment_events", // name of the exchange
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		logger.Fatal("main: failed to declare exchange", zap.Error(err))
	}

	logger.Info("RabbitMQ Exchange initialized successfully!")

	worker := workers.NewOutboxWorker(db.DB, rabbitChannel)
	go worker.StartOutboxWorker(ctx)

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
	cancel()
	grpcServer.GracefulStop()
	logger.Info("main: server stopped safely")
}
