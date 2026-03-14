package main

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/media/internal/handler"
	"ecommerce/services/media/internal/service"
	"ecommerce/services/media/internal/storage"
	"ecommerce/services/media/internal/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title           E-Commerce Media Microservice API
// @version         1.0
// @description     Media upload and management service for sellers.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Pranay Kamble
// @contact.email  iampranaykamble1@gmail.com

// @host      localhost:8083
// @BasePath  /api/v1/media

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	environment := os.Getenv("ENV_TYPE")
	logger.Init(environment)

	s3, err := storage.NewS3Storage()

	if err != nil {
		logger.Fatal("Failed to initialize S3 Storage: " + err.Error())
	}

	mediaService := service.NewMediaService(s3)
	mediaHandler := handler.NewMediaHandler(mediaService)

	err = utils.GetPublicKey()

	if err != nil {
		logger.Fatal("Failed to get public key from auth service: " + err.Error())
	}

	router := gin.Default()
	handler.RegisterRoutes(router, mediaHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Info("Server is running on port " + port)
	err = router.Run(":" + port)
	if err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
