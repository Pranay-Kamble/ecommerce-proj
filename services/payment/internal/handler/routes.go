package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, wh *WebhookHandler) {
	v1 := router.Group("/api/v1/payment/")

	{
		v1.POST("webhook", wh.handleWebhook)
	}
}
