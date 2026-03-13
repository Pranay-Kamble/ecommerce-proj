package main

import (
	"ecommerce/pkg/broker"
	"ecommerce/pkg/logger"
	"ecommerce/services/catalog/internal/domain"
	"log"
	"os"

	"ecommerce/services/catalog/internal/handler"
	"ecommerce/services/catalog/internal/repository"
	"ecommerce/services/catalog/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           Catalog Microservice API
// @version         1.0
// @description     E-commerce Catalog management including Products, Variants, Categories, and Sellers.

// @contact.name    Pranay Kamble
// @contact.email   iampranaykamble1@gmail.com

// @host      localhost:8082
// @BasePath  /api/v1/catalog

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by a space and your JWT token.

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("No .env file found, relying on environment variables")
	}
	logger.Init(os.Getenv("ENV_TYPE"))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=catalog port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	defer func() {
		if err = sqlDb.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS ltree;").Error
	if err != nil {
		log.Fatalf("Failed to create ltree extension: %v", err)
	}

	err = db.AutoMigrate(
		&domain.Seller{},
		&domain.Category{},
		&domain.Product{},
		&domain.Variant{},
	)

	if err != nil {
		logger.Fatal("Failed to auto-migrate database schema: %v", zap.Error(err))
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
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

	categoryRepo := repository.NewCategoryRepository(db)
	sellerRepo := repository.NewSellerRepository(db)
	productRepo := repository.NewProductRepository(db)
	variantRepo := repository.NewProductVariantRepository(db)

	categoryService := service.NewCategoryService(categoryRepo)
	sellerService := service.NewSellerService(sellerRepo, rabbitMQ)
	productService := service.NewProductService(categoryRepo, productRepo, sellerRepo)
	variantService := service.NewVariantService(variantRepo, productRepo, sellerRepo)

	categoryHandler := handler.NewCategoryHandler(categoryService)
	sellerHandler := handler.NewSellerHandler(sellerService)
	productHandler := handler.NewProductHandler(productService, sellerService, categoryService)
	variantHandler := handler.NewVariantHandler(variantService)

	router := gin.Default()

	handler.RegisterRoutes(
		router,
		categoryHandler,
		productHandler,
		sellerHandler,
		variantHandler,
		sellerService,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Catalog Microservice running on port %s", port)
	err = router.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
