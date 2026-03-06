package handler

import (
	"ecommerce/services/catalog/internal/service"
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

	_, ok := claims["id"].(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims: user ID missing"})
		return
	}
	c.Set("claims", claims)
	c.Next()
}

func RequireSeller(sellerService service.SellerService) gin.HandlerFunc {

	return func(c *gin.Context) {

		claimsObj, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: claims missing"})
			return
		}

		claims, ok := claimsObj.(map[string]interface{})
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims format"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "seller" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: seller access required"})
			return
		}

		userID, ok := claims["id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user ID missing"})
			return
		}

		seller, err := sellerService.GetByUserID(c.Request.Context(), userID)
		if err != nil || seller == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "seller profile not found. please complete onboarding."})
			return
		}

		c.Set("seller_public_id", seller.PublicID)
		c.Next()
	}
}
