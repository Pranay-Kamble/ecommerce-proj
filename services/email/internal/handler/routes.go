package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, emailHandler EmailHandler) {
	v1 := router.Group("/api/v1/email/")
	{
		v1.POST("/verification-email", emailHandler.VerificationEmail)
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}
}
