package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, mediaHandler *MediaHandler) {
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
