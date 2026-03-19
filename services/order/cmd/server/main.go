package main

import (
	"ecommerce/services/order/internal/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"ecommerce/pkg/broker"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
	"ecommerce/services/order/internal/domain"
	"ecommerce/services/order/internal/handler"
	"ecommerce/services/order/internal/repository"
	"ecommerce/services/order/internal/service"

	pb "ecommerce/pkg/protobufs/catalog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	environment := os.Getenv("ENV_TYPE")
	logger.Init(environment)

	pgDSN := os.Getenv("DATABASE_DSN")
	if pgDSN == "" {
		pgDSN = "host=localhost user=admin password=password dbname=order_db port=5432 sslmode=disable"
	}
	pg := &database.Postgres{}
	if err = pg.Connect(pgDSN); err != nil {
		logger.Fatal("Failed to connect to Postgres", zap.Error(err))
	}
	defer pg.Close()

	redisDSN := os.Getenv("REDIS_DSN")
	if redisDSN == "" {
		redisDSN = "redis://localhost:6379/1"
	}
	rd := &database.Redis{}
	if err = rd.Connect(redisDSN); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer rd.Close()

	err = pg.DB.AutoMigrate(
		&domain.Order{},
		&domain.OrderItem{},
		&domain.CustomerProfile{},
		&domain.Address{},
	)
	if err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	err = utils.GetPublicKey()
	if err != nil {
		logger.Fatal("Failed to get public key", zap.Error(err))
	}

	cartRepo := repository.NewCartRepository(rd.Redis)
	customerRepo := repository.NewCustomerRepository(pg.DB)
	orderRepo := repository.NewOrderRepository(pg.DB)

	catalogGrpcURL := os.Getenv("CATALOG_GRPC_URL")
	if catalogGrpcURL == "" {
		catalogGrpcURL = "localhost:50051"
	}

	catalogConn, err := grpc.NewClient(catalogGrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to Catalog gRPC server", zap.Error(err))
	}
	defer catalogConn.Close()

	catalogClient := pb.NewCatalogServiceClient(catalogConn)

	cartSvc := service.NewCartService(cartRepo)

	rabbitMQURL := os.Getenv("RABBIT_MQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://admin:password@localhost:5672/"
	}

	rabbitMQ, err := broker.NewRabbitMQClient(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	defer rabbitMQ.Close()

	err = rabbitMQ.DeclareExchange("user_events", "topic")
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	customerSvc := service.NewCustomerService(customerRepo, rabbitMQ)

	orderSvc, err := service.NewOrderService(orderRepo, cartRepo, catalogClient)
	if err != nil {
		logger.Fatal("Failed to initialize order service", zap.Error(err))
	}

	cartHandler := handler.NewCartHandler(cartSvc)
	customerHandler := handler.NewCustomerHandler(customerSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.Default()
	handler.RegisterRoutes(router, cartHandler, customerHandler, orderHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	logger.Info("Order Service running on port " + port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
