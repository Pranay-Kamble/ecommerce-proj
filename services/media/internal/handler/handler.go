package handler

import (
	"ecommerce/pkg/logger"
	"mime/multipart"
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

type MultipleUploadRequest struct {
	Folder string                  `form:"folder"`
	Images []*multipart.FileHeader `form:"images" binding:"required"`
}

func (h *MediaHandler) UploadMultipleImages(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 80<<20)
	var request MultipleUploadRequest

	err := c.ShouldBind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. Ensure 'images' is an array of files."})
		return
	}

	files := request.Images
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image files provided"})
		return
	}
	if len(files) > 15 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many files. Maximum is 15."})
		return
	}
	for _, file := range files {
		if file.Size > (5 << 20) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more files are too large. Maximum size is 5MB each."})
			return
		}
	}

	folder := c.DefaultPostForm("folder", "general")

	urls, failedFiles := h.mediaService.UploadImages(c.Request.Context(), files, folder)
	if len(failedFiles) == 0 {
		c.JSON(http.StatusCreated, gin.H{"message": "images uploaded successfully", "urls": urls})
		return
	}

	if len(urls) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload images"})
		return
	}

	c.JSON(http.StatusMultiStatus, gin.H{"error": "failed to upload images", "urls": urls, "failedFiles": failedFiles})
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
