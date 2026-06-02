package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up the HTTP routes for the search service.
func RegisterRoutes(router *gin.Engine, searchHandler *SearchHandler) {
	v1 := router.Group("/api/v1/search")

	v1.GET("/products", searchHandler.SearchProducts)
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
}
