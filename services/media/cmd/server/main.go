package main

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/media/internal/handler"
	"ecommerce/services/media/internal/service"
	"ecommerce/services/media/internal/storage"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
