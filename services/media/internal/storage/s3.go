package storage

import (
	"context"
	"ecommerce/pkg/logger"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Client     *s3.Client
	BucketName string
	Endpoint   string
}

func NewS3Storage() (*S3Storage, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("AWS_REGION")
	bucketName := os.Getenv("S3_BUCKET_NAME")

	if bucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME is required in .env")
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

	logger.Info("Successfully initialized S3 Storage Client (Endpoint: " + endpoint + " )")

	return &S3Storage{
		Client:     client,
		BucketName: bucketName,
		Endpoint:   endpoint,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, file io.Reader, objectKey string, contentType string) (string, error) {

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		fmt.Println(s.BucketName, s.Endpoint, s.Client)
		return "", fmt.Errorf("s3: failed to upload to s3: %w", err)
	}

	//        <-----Endpoint------>/<---BucketName-->/<----ObjectKey---->
	// e.g., http://localhost:9000/ecommerce-images/products/nano123.jpg
	url := fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.Endpoint, "/"), s.BucketName, objectKey)
	return url, nil
}
