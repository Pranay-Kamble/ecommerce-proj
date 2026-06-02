package handler

import (
	"net/http"
	"strconv"

	"ecommerce/services/search/internal/domain"
	"ecommerce/services/search/internal/service"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles HTTP requests for product search.
type SearchHandler struct {
	searchService service.SearchService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// SearchProducts handles GET /api/v1/search/products
// Query params: q, category, min_price, max_price, in_stock, page, limit
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	filters := domain.SearchFilters{
		Query:    c.Query("q"),
		Category: c.Query("category"),
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if v, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filters.MinPrice = v
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if v, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filters.MaxPrice = v
		}
	}

	if inStock := c.Query("in_stock"); inStock != "" {
		if v, err := strconv.ParseBool(inStock); err == nil {
			filters.InStock = &v
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	filters.Page = page

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filters.Limit = limit

	result, err := h.searchService.SearchProducts(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
