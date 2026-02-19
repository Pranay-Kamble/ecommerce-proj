package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, authHandler AuthHandler) {
	v1 := router.Group("/api/v1/auth/")
	{
		v1.POST("/register", authHandler.RegisterNormal)
		v1.GET("/ping", authHandler.GetPing)
		v1.POST("/login", authHandler.Login)
	}
}
