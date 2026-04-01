package handler

import (
	"ecommerce/services/order/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.VerifyJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			return
		}

		extractedUserID := claims["id"]

		if extractedUserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
			c.Abort()
			return
		}

		c.Set("user_id", extractedUserID)
		c.Next()
	}
}

func MockMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "usr_test_999")
		c.Next()
	}
}
