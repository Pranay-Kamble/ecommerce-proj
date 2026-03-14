package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ecommerce/services/media/docs"
)

func RegisterRoutes(router *gin.Engine, mediaHandler *MediaHandler) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := router.Group("/api/v1/media")

	public := v1.Group("/")
	{
		public.GET("/ping", mediaHandler.HealthCheck)
	}

	seller := v1.Group("/")
	seller.Use(RequireSeller())
	{
		seller.POST("/upload", mediaHandler.UploadSingleImage)
		seller.POST("/upload-multiple", mediaHandler.UploadMultipleImages)
		seller.DELETE("/image", mediaHandler.DeleteImage)
	}
}
