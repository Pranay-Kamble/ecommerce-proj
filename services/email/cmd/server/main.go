package main

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/email/internal/handler"
	"ecommerce/services/email/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

func main() {
	r := gin.Default()
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("main: Error loading .env file, relying on system env vars")
	}

	environment := os.Getenv("ENV_TYPE")
	if environment == "" {
		environment = "prod"
	}
	logger.Init(environment)

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		logger.Fatal("main: Error loading RESEND_API_KEY")
		return
	}

	resendClient := resend.NewClient(apiKey)

	emailService := service.NewEmailService(resendClient)
	emailHandler := handler.NewEmailHandler(emailService)

	handler.RegisterRoutes(r, emailHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("main: failed to start server", zap.Error(err))
	}
}
