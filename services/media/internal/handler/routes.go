package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, mediaHandler *MediaHandler) {
	v1 := router.Group("/api/v1/media")

	public := v1.Group("/")
	{
		public.GET("/ping", mediaHandler.HealthCheck)
		public.POST("/upload", mediaHandler.UploadSingleImage)
	}
}
