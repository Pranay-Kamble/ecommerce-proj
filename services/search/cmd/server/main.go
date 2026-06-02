package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"ecommerce/pkg/logger"
	"ecommerce/services/search/internal/handler"
	"ecommerce/services/search/internal/repository"
	"ecommerce/services/search/internal/service"
	"ecommerce/services/search/internal/workers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	environment := os.Getenv("ENV_TYPE")
	logger.Init(environment)

	// ── Elasticsearch ────────────────────────────────────────────────
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esURL},
	})
	if err != nil {
		logger.Fatal("Failed to create Elasticsearch client", zap.Error(err))
	}

	// Verify ES connection
	res, err := esClient.Info()
	if err != nil {
		logger.Fatal("Failed to connect to Elasticsearch", zap.Error(err))
	}
	res.Body.Close()
	logger.Info("Connected to Elasticsearch", zap.String("url", esURL))

	// ── Repository / Service / Handler ───────────────────────────────
	esRepo := repository.NewElasticRepository(esClient)

	// Ensure the products index exists with proper mappings
	if err = esRepo.EnsureIndex(ctx); err != nil {
		logger.Fatal("Failed to ensure Elasticsearch index", zap.Error(err))
	}

	searchSvc := service.NewSearchService(esRepo)
	searchHandler := handler.NewSearchHandler(searchSvc)

	// ── RabbitMQ Consumer ────────────────────────────────────────────
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://admin:password@localhost:5672/"
	}

	rabbitConn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		logger.Fatal("Failed to open RabbitMQ channel", zap.Error(err))
	}
	defer rabbitChannel.Close()

	catalogConsumer := workers.NewCatalogConsumer(rabbitChannel, searchSvc)

	go func() {
		logger.Info("Starting Catalog RabbitMQ Consumer...")
		if consumerErr := catalogConsumer.StartListening(ctx); consumerErr != nil {
			logger.Error("Catalog consumer stopped unexpectedly", zap.Error(consumerErr))
		}
	}()

	// ── HTTP Server ──────────────────────────────────────────────────
	router := gin.Default()
	handler.RegisterRoutes(router, searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("Search Service running on port " + port)
		if innerErr := srv.ListenAndServe(); innerErr != nil && !errors.Is(innerErr, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(innerErr))
		}
	}()

	// ── Graceful Shutdown ────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Search Service...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Search Service exited cleanly")
}
