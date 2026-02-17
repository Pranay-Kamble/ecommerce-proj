package main

import (
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
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
	environment := os.Getenv("ENV_TYPE")
	if environment == "" {
		environment = "prod"
	}
	logger.Init(environment)

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=admin password=password dbname=auth_db port=5432 sslmode=disable"
	}

	db, err := database.Connect(dsn)
	if err != nil {
		logger.Fatal("main: failed to connect to database: %v", zap.Error(err))
	}

	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		return
	}

	repo := repository.NewUserRepository(db)
	authService := service.NewAuthService(repo)
	authHandler := handler.NewAuthHandler(authService)

	err = utils.InitKeys()
	if err != nil {
		logger.Fatal("main: failed to load RSA keys: %v", zap.Error(err))
	}

	r := gin.Default()
	handler.RegisterRoutes(r, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err = r.Run(":" + port)
	if err != nil {
		logger.Fatal("main: failed to start server: %v", zap.Error(err))
	}
}
