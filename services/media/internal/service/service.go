package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"ecommerce/services/media/internal/storage"

	"github.com/sixafter/nanoid"
)

type MediaService interface {
	UploadImage(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error)
	DeleteImage(ctx context.Context, fileUrl string) error
}

type mediaService struct {
	s3Storage *storage.S3Storage
}

func NewMediaService(s3Storage *storage.S3Storage) MediaService {
	return &mediaService{s3Storage: s3Storage}
}

func (s *mediaService) UploadImage(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("service: failed to open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("service: failed to read file header: %w", err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("service: failed to reset file pointer: %w", err)
	}

	contentType := http.DetectContentType(buffer)
	if !strings.HasPrefix(contentType, "image/") {
		return "", errors.New("service: invalid file type, only images are allowed")
	}

	id, _ := nanoid.New()
	extension := filepath.Ext(fileHeader.Filename)

	// e.g., "products/nano123456789.jpg" (S3 Key)
	objectKey := fmt.Sprintf("%s/%s%s", folder, id, extension)

	url, err := s.s3Storage.Upload(ctx, file, objectKey, contentType)
	if err != nil {

		return "", fmt.Errorf("service: failed to upload file: %w", err)
	}

	return url, nil
}

func (s *mediaService) DeleteImage(ctx context.Context, fileUrl string) error {
	objectKey, err := url.Parse(fileUrl)
	if err != nil {
		return fmt.Errorf("service: failed to parse file URL: %w", err)
	}
	prefixToTrim := fmt.Sprintf("/%s/", s.s3Storage.BucketName)
	objectKeyString := strings.TrimPrefix(objectKey.Path, prefixToTrim)
	
	err = s.s3Storage.Delete(ctx, objectKeyString)

	if err != nil {
		return fmt.Errorf("service: failed to delete file: %w", err)
	}

	return nil
}
