package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPRepository interface {
	Create(ctx context.Context, otp, email string, ttl time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

type otpRepository struct {
	redis *redis.Client
}

func NewOTPRepository(redis *redis.Client) OTPRepository {
	return &otpRepository{redis: redis}
}

func (o *otpRepository) Create(ctx context.Context, otp, email string, ttl time.Duration) error {
	_, err := o.redis.Set(ctx, fmt.Sprintf("otp:%s", email), otp, ttl).Result()

	if err != nil {
		return fmt.Errorf("repository: otp create failure: %w", err)
	}
	return nil
}

func (o *otpRepository) Get(ctx context.Context, email string) (string, error) {
	res, err := o.redis.Get(ctx, fmt.Sprintf("otp:%s", email)).Result()
	if errors.Is(err, redis.Nil) {
		return "", errors.New("repository: otp not found")
	} else if err != nil {
		return "", err
	}
	return res, nil
}

func (o *otpRepository) Delete(ctx context.Context, email string) error {
	_, err := o.redis.Del(ctx, fmt.Sprintf("otp:%s", email)).Result()

	if err != nil {
		return err
	}
	return nil
}
