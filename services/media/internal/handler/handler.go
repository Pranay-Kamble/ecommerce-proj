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

func (h *MediaHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
