package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/media/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtString := c.GetHeader("Authorization")

		if jwtString == "" {
			logger.Error("middleware: failed to get Authorization header")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
		if !strings.HasPrefix(jwtString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		jwtString = jwtString[7:]

		claims, err := utils.VerifyJWT(jwtString)
		if err != nil {
			logger.Error("middleware: failed to verify JWT: ", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "seller" {
			logger.Error("middleware: user does not have seller role")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		isOnboarded, ok := claims["is_onboarded"].(bool)
		if !ok || !isOnboarded {
			logger.Error("middleware: seller is not onboarded")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Set("seller_public_id", claims["id"])

		c.Next()
	}
}
