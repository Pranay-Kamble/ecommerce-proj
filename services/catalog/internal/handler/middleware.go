package handler

import (
	"net/http"
	"strings"

	"ecommerce/services/catalog/internal/utils"

	"github.com/gin-gonic/gin"
)

func RequireUser(c *gin.Context) {
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

func RequireSeller(c *gin.Context) {
	claims, err := c.Get("claims")
	if !err {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "claims not found"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims format"})
		return
	}

	role, ok := claimsMap["role"].(string)
	if !ok || role != "seller" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: seller role required"})
		return
	}

	c.Next()
}
