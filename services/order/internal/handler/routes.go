package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, cartHandler *CartHandler, customerHandler *CustomerHandler, orderHandler *OrderHandler) {

	v1 := router.Group("/api/v1")

	v1.Use(RequireUser())
	{
		v1.GET("/cart", cartHandler.GetCart)
		v1.POST("/cart/add", cartHandler.AddItem)
		v1.DELETE("/cart/remove/:product_id", cartHandler.RemoveItem)
		v1.DELETE("/cart", cartHandler.ClearCart)

		v1.GET("/profile", customerHandler.GetProfile)
		v1.POST("/profile", customerHandler.CreateProfile)
		v1.POST("/profile/addresses", customerHandler.AddAddress)

		v1.POST("/checkout", orderHandler.Checkout)
		v1.GET("/orders/:public_id", orderHandler.GetOrder)
		v1.GET("/orders", orderHandler.GetUserOrders)
	}
}
