package handler

import (
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type SellerHandler struct {
	sellerService service.SellerService
}

func NewSellerHandler(sellerService service.SellerService) *SellerHandler {
	return &SellerHandler{sellerService: sellerService}
}

type CreateSellerRequest struct {
	Name        string `json:"name" binding:"required,min=1"`
	Description string `json:"description"`
	LogoURL     string `json:"logoUrl"`

	SupportEmail string `json:"supportEmail" binding:"omitempty,email"`
	SupportPhone string `json:"supportPhone"`

	GSTIN             string `json:"gstin" binding:"required"`
	RegisteredAddress string `json:"registeredAddress"`
}

func (h *SellerHandler) CreateSeller(c *gin.Context) {
	var request CreateSellerRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, exists := getUserIdFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing User ID in token"})
		return
	}

	newSeller := &domain.Seller{
		UserID:            userID,
		Name:              request.Name,
		Description:       request.Description,
		LogoURL:           request.LogoURL,
		SupportEmail:      request.SupportEmail,
		SupportPhone:      request.SupportPhone,
		GSTIN:             request.GSTIN,
		RegisteredAddress: request.RegisteredAddress,
	}

	err = h.sellerService.CreateSeller(c.Request.Context(), newSeller)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "service: user already exists" || errMsg == "service: gstin already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "users/gstin already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create seller"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"seller": newSeller})
}

func (h *SellerHandler) GetMyProfile(c *gin.Context) {
	userID, exists := getUserIdFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing User ID in token"})
		return
	}
	seller, err := h.sellerService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seller profile"})
		return
	}

	if seller == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Seller profile not found. Please onboard as a seller."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seller": seller})
}

func getUserIdFromContext(c *gin.Context) (string, bool) {
	claimsObj, exists := c.Get("claims")
	if !exists {
		return "", false
	}

	claimsMap, ok := claimsObj.(map[string]interface{})
	if !ok {
		jwtClaims, jwtOk := claimsObj.(jwt.MapClaims)
		if !jwtOk {
			return "", false
		}
		claimsMap = jwtClaims
	}

	userIDObj, exists := claimsMap["id"]
	if !exists {
		return "", false
	}

	userID, ok := userIDObj.(string)
	return userID, ok
}
