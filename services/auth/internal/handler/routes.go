package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRoutes(router *gin.Engine, authHandler *AuthHandler) {
	v1 := router.Group("/api/v1/auth/")
	{
		v1.POST("/register", authHandler.Register)
	}
}
