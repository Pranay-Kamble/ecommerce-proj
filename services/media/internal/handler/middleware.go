package handler

import (
	"ecommerce/services/media/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtString := c.GetHeader("Authorization")

		if jwtString == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad request: no Authorization header provided"})
			return
		}
		if !strings.HasPrefix(jwtString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		jwtString = jwtString[7:]

		claims, err := utils.VerifyJWT(jwtString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "seller" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		isOnboarded, ok := claims["is_onboarded"].(bool)
		if !ok || !isOnboarded {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Set("seller_public_id", claims["id"])

		c.Next()
	}
}
