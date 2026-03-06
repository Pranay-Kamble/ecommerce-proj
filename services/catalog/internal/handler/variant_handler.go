package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type VariantHandler struct {
	variantService service.VariantService
}

func NewVariantHandler(variantService service.VariantService) *VariantHandler {
	return &VariantHandler{variantService: variantService}
}

func (h *VariantHandler) GetVariantByID(c *gin.Context) {
	variantSKU := c.Param("sku")

	variant, err := h.variantService.GetBySKU(c.Request.Context(), variantSKU)
	if err != nil {
		errorMsg := err.Error()
		if errorMsg == "service: no product variant" {
			c.JSON(http.StatusNotFound, gin.H{"error": "variant not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"variant": variant})
}

type CreateVariantRequest struct {
	SKU            string                 `json:"sku" binding:"required"`
	Title          string                 `json:"title" binding:"required"`
	Price          float64                `json:"price" binding:"required,gt=0"`
	Inventory      int                    `json:"inventory" binding:"gte=0"`
	Specifications map[string]interface{} `json:"specifications"`
	Images         []*domain.Image        `json:"images"`
}

func (h *VariantHandler) CreateVariant(c *gin.Context) {
	productPublicID := c.Param("product_id")

	if len(productPublicID) < 3 || !strings.HasPrefix(productPublicID, "itm_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	sellerPublicID, exists := c.Get("seller_public_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "seller information missing in context"})
		return
	}

	var req CreateVariantRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	newVariant := &domain.Variant{
		Title:     req.Title,
		SKU:       req.SKU,
		Price:     req.Price,
		Inventory: req.Inventory,
		Images:    req.Images,
	}

	err = h.variantService.CreateVariant(c.Request.Context(), productPublicID, sellerPublicID.(string), newVariant)
	if err != nil {
		errorMsg := err.Error()
		if errorMsg == "service: product id does not match" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product and variant are not associated"})
		} else if errorMsg == "service: no product exists" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		} else if errorMsg == "service: variant with the same SKU already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "variant with the same SKU already exists"})
		} else if errorMsg == "service: seller does not own the product" {
			c.JSON(http.StatusConflict, gin.H{"error": "seller does not own the product"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "variant created successfully"})
}

type UpdateVariantRequest struct {
	Title           string                 `json:"title" binding:"required"`
	Price           float64                `json:"price" binding:"required,gte=0"`
	SKU             string                 `json:"sku" binding:"required"`
	Inventory       int                    `json:"inventory" binding:"gte=0"`
	Images          []*domain.Image        `json:"images"`
	Specifications  map[string]interface{} `json:"specifications"`
	ProductPublicID string                 `json:"productId" binding:"required"`
}

func (h *VariantHandler) UpdateVariant(c *gin.Context) {

	variantPublicID := c.Param("id")
	if len(variantPublicID) < 3 || !strings.HasPrefix(variantPublicID, "var_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant id"})
		return
	}

	sellerPublicID, exists := c.Get("seller_public_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "seller information missing in context"})
		return
	}

	var request UpdateVariantRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(request.ProductPublicID) < 3 || !strings.HasPrefix(request.ProductPublicID, "itm_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	newVariant := &domain.Variant{
		Title:          request.Title,
		Price:          request.Price,
		SKU:            request.SKU,
		Inventory:      request.Inventory,
		Images:         request.Images,
		Specifications: request.Specifications,
	}
	err = h.variantService.UpdateVariant(c.Request.Context(), request.ProductPublicID, variantPublicID, sellerPublicID.(string), newVariant)
	if err != nil {
		errorMsg := err.Error()
		if strings.HasPrefix(errorMsg, "service: failed to get") {
			logger.Error("handler: failed to get product or variant or seller during update: %w", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		} else if errorMsg == "service: no product exists" || errorMsg == "service: no variant exists" || errorMsg == "service: no seller exists" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product or variant or seller not found"})
		} else if errorMsg == "service: seller does not own the product" {
			c.JSON(http.StatusConflict, gin.H{"error": "seller does not own the product"})
		} else if errorMsg == "service: price and inventory must be non negative" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price and inventory must be non negative"})
		} else if errorMsg == "service: variant with the same SKU already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "variant with the same SKU already exists"})
		} else {
			logger.Error("handler: failed to update the variant", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "variant updated successfully"})
}

// DeleteVariant handles DELETE /api/v1/catalog/seller/variants/:id
func (h *VariantHandler) DeleteVariant(c *gin.Context) {
	// 1. Extract variant ID from URL
	// 2. Extract seller_public_id from context
	// 3. Call h.variantService.DeleteVariant(ctx, sellerID, variantID)
	// 4. Return 200 OK
}
