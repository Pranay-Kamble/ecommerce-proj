package handler

import (
	"github.com/gin-gonic/gin"

	_ "ecommerce/services/auth/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, authHandler AuthHandler) {

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1/auth/")
	{
		v1.POST("/register", authHandler.RegisterNormal)
		v1.GET("/ping", authHandler.GetPing)
		v1.POST("/login", authHandler.Login)
		v1.POST("/refresh", authHandler.Refresh)
		v1.POST("/logout", authHandler.Logout)
		v1.POST("/verify", authHandler.Verify)
		v1.POST("/resend-otp", authHandler.ResendOTP)
		v1.GET("/google/login", authHandler.GoogleLogin)
		v1.GET("/google/callback", authHandler.GoogleCallback)
		v1.GET("/public-key", authHandler.GetPublicKey)
	}
}
