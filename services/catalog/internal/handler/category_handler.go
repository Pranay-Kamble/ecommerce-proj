package handler

import (
	"ecommerce/services/catalog/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// GetAllCategories @Summary      Get all categories
// @Description  Retrieves a list of categories. Optionally filter by parent_id.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        parent_id  query     string  false  "Parent Category Public ID"
// @Success      200        {object}  map[string]interface{}
// @Failure      404        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]interface{}
// @Router       /categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	parentIDParams := c.Query("parent_id")
	categories, err := h.categoryService.GetAllCategories(c.Request.Context(), parentIDParams)
	if err != nil {
		if err.Error() == "service: category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetBreadcrumbs @Summary      Get category breadcrumbs
// @Description  Retrieves the hierarchical path for a specific category.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        category_id  path      string  true  "Category Public ID"
// @Success      200          {object}  map[string]interface{}
// @Failure      400          {object}  map[string]interface{}
// @Failure      404          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /categories/{category_id}/breadcrumbs [get]

func (h *CategoryHandler) GetBreadcrumbs(c *gin.Context) {
	categoryPublicID := c.Param("category_id")
	if categoryPublicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}
	data, err := h.categoryService.GetCategoryBreadCrumbs(c.Request.Context(), categoryPublicID)
	if err != nil {
		if err.Error() == "service: category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
	return
}
