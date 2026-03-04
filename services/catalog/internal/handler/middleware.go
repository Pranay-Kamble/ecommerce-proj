package handler

import (
	"net/http"
	"strings"

	"ecommerce/services/catalog/internal/utils"

	"github.com/gin-gonic/gin"
)

func RequireAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}
	tokenString := parts[1]

	claims, err := utils.VerifyJWT(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	
	if userID, ok := claims["id"].(string); ok {
		c.Set("userID", userID)
	}

	c.Set("claims", claims)
	c.Next()
}
