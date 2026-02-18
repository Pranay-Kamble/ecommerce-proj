package handler

import (
	"ecommerce/services/auth/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func requireAuth(c *gin.Context) {
	header := c.GetHeader("Authorization")

	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: Missing or invalid token format",
		})
		return
	}

	jwtToken := header[7:]
	claims, err := utils.VerifyJWT(jwtToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: Invalid token",
		})
		return
	}

	expiryTime, err := claims.GetExpirationTime()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: Invalid token",
		})
		return
	}

	if time.Now().UnixMilli() > expiryTime.UnixMilli() {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: Token expired, login again.",
		})
		return
	}

	c.Next()
}
