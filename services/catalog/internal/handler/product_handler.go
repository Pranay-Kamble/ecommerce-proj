package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/catalog/internal/domain"
	"ecommerce/services/catalog/internal/service"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	productService  service.ProductService
	sellerService   service.SellerService
	categoryService service.CategoryService
}

func NewProductHandler(productService service.ProductService, sellerService service.SellerService, categoryService service.CategoryService) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		sellerService:   sellerService,
		categoryService: categoryService,
	}
}

// ListProducts @Summary      List products
// @Description  Retrieves a paginated list of products. Can filter by category or seller.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        category_id  query     string  false  "Category Public ID"
// @Param        seller_id    query     string  false  "Seller Public ID"
// @Param        page         query     int     false  "Page number" default(1)
// @Param        limit        query     int     false  "Items per page" default(20)
// @Success      200          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {

	categoryId := c.Query("category_id")
	sellerId := c.Query("seller_id")

	page, err := strconv.Atoi(c.DefaultQuery("page", os.Getenv("DEFAULT_PAGE")))
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", os.Getenv("DEFAULT_LIMIT")))
	if err != nil {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	var products []*domain.Product
	if categoryId == "" && sellerId == "" {
		products, err = h.productService.ListAllProducts(c.Request.Context(), page, limit)
	} else if categoryId == "" {
		products, err = h.productService.ListProductsBySeller(c.Request.Context(), sellerId, page, limit)
	} else if sellerId == "" {
		products, err = h.productService.ListProductsByCategory(c.Request.Context(), categoryId, page, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

// GetProductByID @Summary      Get product by ID
// @Description  Retrieves a single product by its Public ID.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product Public ID (itm_...)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productPublicID := c.Param("id")

	if len(productPublicID) < 4 || !strings.HasPrefix(productPublicID, "itm_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	product, err := h.productService.GetProductByPublicID(c.Request.Context(), productPublicID)
	if err != nil {
		if err.Error() == "service: product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// CreateProductRequest @Summary      Create a new product
// @Description  Creates a new product. Requires a registered Seller profile.
// @Tags         Seller Products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      handler.CreateProductRequest  true  "Product creation payload"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /seller/products [post]
type CreateProductRequest struct {
	Title       string                 `json:"title" required:"true" minlength:"3"`
	Description string                 `json:"description" required:"true"`
	Brand       string                 `json:"brand" required:"true"`
	Highlights  []string               `json:"highlights"`
	Dimensions  map[string]interface{} `json:"dimensions"`

	Images   []*domain.Image   `json:"images"`
	Variants []*domain.Variant `json:"variants"`

	CategoryPublicID string `json:"category_public_id" required:"true"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var request CreateProductRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	sellerPublicID, ok := c.Get("seller_public_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	product := &domain.Product{
		Title:       request.Title,
		Description: request.Description,
		Brand:       request.Brand,
		Highlights:  request.Highlights,
		Images:      request.Images,
		Variants:    request.Variants,
	}

	err = h.productService.CreateProduct(c.Request.Context(), sellerPublicID.(string), request.CategoryPublicID, product)
	if err != nil {
		if err.Error() == "service: invalid product details" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product details: " + err.Error()})

		} else if err.Error() == "service: unable to create product:" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			logger.Error("handler: Failed to create product: ", zap.Error(err))

		} else if err.Error() == "not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category/seller public ID"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully"})
}

type UpdateProductRequest struct {
	CategoryPublicID string `json:"categoryId" required:"true"`

	Title       string                 `json:"title" required:"true" minlength:"3"`
	Brand       string                 `json:"brand" required:"true"`
	Description string                 `json:"description" required:"true"`
	Highlights  []string               `json:"highlights" required:"true"`
	Dimensions  map[string]interface{} `json:"dimensions"`

	Variants []*domain.Variant `json:"variants"`
	Images   []*domain.Image   `json:"images"`
}

// UpdateProduct @Summary      Update a product
// @Description  Updates an existing product. Seller must own the product.
// @Tags         Seller Products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Product Public ID (itm_...)"
// @Param        request  body      handler.UpdateProductRequest  true  "Product update payload"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /seller/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {

	sellerPublicID, ok := c.Get("seller_public_id")

	if ok != true {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productPublicID := c.Param("id")
	if len(productPublicID) < 4 || !strings.HasPrefix(productPublicID, "itm_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	var request UpdateProductRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updatedProduct := &domain.Product{
		Title:       request.Title,
		Description: request.Description,
		Brand:       request.Brand,
		Highlights:  request.Highlights,
		Dimensions:  request.Dimensions,
		Images:      request.Images,
		Variants:    request.Variants,
	}

	err = h.productService.UpdateProduct(c.Request.Context(), productPublicID, sellerPublicID.(string), request.CategoryPublicID, updatedProduct)
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "service: failed to find") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		} else if strings.Contains(errorMsg, "does not exist") ||
			strings.Contains(errorMsg, "service: incorrect/invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect or invalid category/seller public ID"})
		} else if strings.Contains(errorMsg, "service: seller does not own this product") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this product"})
		} else if strings.Contains(errorMsg, "service: failed to perform update") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// DeleteProduct @Summary      Delete a product
// @Description  Soft deletes a product. Seller must own the product.
// @Tags         Seller Products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product Public ID (itm_...)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /seller/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productPublicID := c.Param("id")
	if len(productPublicID) < 4 || !strings.HasPrefix(productPublicID, "itm_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}
	sellerPublicID, ok := c.Get("seller_public_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.productService.DeleteProduct(c.Request.Context(), sellerPublicID.(string), productPublicID)
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "service: failed to find") {
			logger.Error("handler: Failed to find by product/seller public id", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		} else if strings.Contains(errorMsg, "service: incorrect/invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect or invalid product/seller public ID"})
		} else if strings.Contains(errorMsg, "service: failed to delete product") {
			logger.Error("handler: Failed to delete product: ", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
