package client

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ecommerce/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	UploadPDF(ctx context.Context, buffer *bytes.Buffer, objectKey string) (string, error)
	GeneratePresignedURL(ctx context.Context, objectKey string, lifetime time.Duration) (string, error)
}

type s3Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
	endpoint      string
}

func NewS3Client() (S3Client, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("AWS_REGION")
	bucketName := os.Getenv("S3_INVOICE_BUCKET_NAME")

	if endpoint == "" {
		endpoint = "http://localhost:9000"
	}
	if accessKey == "" {
		accessKey = "admin"
	}
	if secretKey == "" {
		secretKey = "password"
	}
	if region == "" {
		region = "us-east-1"
	}
	if bucketName == "" {
		bucketName = "invoices"
	}

	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)

	logger.Info("Successfully initialized S3 Client for Order Service (Endpoint: " + endpoint + " )")

	return &s3Client{
		client:        client,
		presignClient: presignClient,
		bucketName:    bucketName,
		endpoint:      endpoint,
	}, nil
}

func (s *s3Client) UploadPDF(ctx context.Context, buffer *bytes.Buffer, objectKey string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", fmt.Errorf("s3: failed to upload PDF to s3: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.endpoint, "/"), s.bucketName, objectKey)
	return url, nil
}

func (s *s3Client) GeneratePresignedURL(ctx context.Context, objectKey string, lifetime time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = lifetime
	})
	if err != nil {
		return "", fmt.Errorf("s3: failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}
