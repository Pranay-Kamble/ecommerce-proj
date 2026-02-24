package main

import (
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
	"ecommerce/services/auth/internal/client"
	"ecommerce/services/auth/internal/domain"
	"ecommerce/services/auth/internal/handler"
	"ecommerce/services/auth/internal/repository"
	"ecommerce/services/auth/internal/service"
	"ecommerce/services/auth/internal/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	environment := os.Getenv("ENV_TYPE")
	if environment == "" {
		environment = "prod"
	}
	logger.Init(environment)

	err = utils.Init(6)
	if err != nil {
		logger.Fatal("main: failed to load RSA keys or OTP generator", zap.Error(err))
	}

	postgresDSN := os.Getenv("DATABASE_DSN")
	if postgresDSN == "" {
		postgresDSN = "host=localhost user=admin password=password dbname=auth_db port=5432 sslmode=disable"
	}

	pg := &database.Postgres{}
	if err := pg.Connect(postgresDSN); err != nil {
		logger.Fatal("main: failed to connect to postgres database", zap.Error(err))
	}
	defer func() {
		if err := pg.Close(); err != nil {
			logger.Error("main: failed to cleanly close postgres", zap.Error(err))
		}
	}()

	redisDSN := os.Getenv("REDIS_DSN")
	if redisDSN == "" {
		redisDSN = "redis://localhost:6379/0"
	}

	rd := &database.Redis{}
	if err := rd.Connect(redisDSN); err != nil {
		logger.Fatal("main: failed to connect to redis database", zap.Error(err))
	}
	defer func() {
		if err := rd.Close(); err != nil {
			logger.Error("main: failed to cleanly close redis", zap.Error(err))
		}
	}()

	err = pg.DB.AutoMigrate(&domain.User{}, &domain.Token{})
	if err != nil {
		logger.Fatal("main: failed to run database migrations", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(pg.DB)
	tokenRepo := repository.NewTokenRepository(pg.DB)
	otpRepo := repository.NewOTPRepository(rd.Redis)

	authService := service.NewAuthService(userRepo, tokenRepo, otpRepo)

	emailBaseURL := os.Getenv("EMAIL_SERVICE_BASE_URL")
	if emailBaseURL == "" {
		emailBaseURL = "http://localhost:8081/api/v1/email"
	}

	emailClient := client.NewEmailClient(emailBaseURL)
	authHandler := handler.NewAuthHandler(authService, emailClient)

	r := gin.Default()
	handler.RegisterRoutes(r, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("main: failed to start server", zap.Error(err))
	}
}
