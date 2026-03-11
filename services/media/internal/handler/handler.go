package handler

import (
	"ecommerce/pkg/logger"
	"net/http"

	"ecommerce/services/media/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

func (h *MediaHandler) UploadSingleImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	err := c.Request.ParseMultipartForm(5 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large. Maximum size is 5MB."})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	folder := c.DefaultPostForm("folder", "general")

	url, err := h.mediaService.UploadImage(c.Request.Context(), file, folder)
	if err != nil {
		if err.Error() == "service: invalid file type, only images are allowed" {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Only image files (jpeg, png, webp) are allowed"})
			return
		}

		logger.Error("handler: failed to upload image: ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"url":     url,
	})
}

type DeleteImageRequest struct {
	URL string `binding:"required,url" json:"url"`
}

func (h *MediaHandler) DeleteImage(c *gin.Context) {
	var request DeleteImageRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.mediaService.DeleteImage(c.Request.Context(), request.URL)
	if err != nil {
		logger.Error("handler: failed to delete image: ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

func (h *MediaHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
