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

// UploadSingleImage godoc
// @Summary      Upload a single image
// @Description  Uploads a single image (jpeg, png, webp) to a specified folder. Maximum file size is 5MB.
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        image   formData  file    true  "The image file to upload (Max 5MB)"
// @Param        folder  formData  string  false "Target folder name (default: 'general')"
// @Success      200     {object}  map[string]interface{} "Image uploaded successfully with URL"
// @Failure      400     {object}  map[string]interface{} "Bad request (file too large, no image provided)"
// @Failure      401     {object}  map[string]interface{} "Unauthorized (missing or invalid JWT)"
// @Failure      403     {object}  map[string]interface{} "Forbidden (not a seller or incomplete onboarding)"
// @Failure      415     {object}  map[string]interface{} "Unsupported media type"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /upload [post]
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

// UploadMultipleImages godoc
// @Summary      Upload multiple images
// @Description  Uploads a batch of up to 15 images. Maximum size per image is 5MB.
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        images  formData  file    true  "Array of image files to upload (Max 15 files)"
// @Param        folder  formData  string  false "Target folder name (default: 'general')"
// @Success      201     {object}  map[string]interface{} "All images uploaded successfully"
// @Success      207     {object}  map[string]interface{} "Multi-Status (Some uploads succeeded, some failed)"
// @Failure      400     {object}  map[string]interface{} "Bad request (files too large, too many files)"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      403     {object}  map[string]interface{} "Forbidden"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /upload-multiple [post]
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
	URL string `binding:"required,url" json:"url" example:"https://my-bucket.s3.amazonaws.com/image.jpg"`
}

// DeleteImage godoc
// @Summary      Delete an image
// @Description  Deletes an image from the storage provider using its absolute URL.
// @Tags         Media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      DeleteImageRequest  true  "The exact URL of the image to delete"
// @Success      200     {object}  map[string]interface{} "Image deleted successfully"
// @Failure      400     {object}  map[string]interface{} "Invalid request body or malformed URL"
// @Failure      401     {object}  map[string]interface{} "Unauthorized"
// @Failure      403     {object}  map[string]interface{} "Forbidden"
// @Failure      500     {object}  map[string]interface{} "Internal server error"
// @Router       /image [delete]
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

// HealthCheck godoc
// @Summary      Health check the server
// @Description  Use this to check if the Media service is active and running.
// @Tags         System
// @Produce      json
// @Success      200     {object}  map[string]interface{} "pong"
// @Router       /ping [get]
func (h *MediaHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
