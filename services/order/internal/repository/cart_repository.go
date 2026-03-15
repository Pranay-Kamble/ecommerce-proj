package repository

import (
	"context"
	"ecommerce/services/order/internal/domain"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CartRepository interface {
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	SaveCart(ctx context.Context, cart *domain.Cart) error
	ClearCart(ctx context.Context, userID string) error
}

type cartRepository struct {
	redis *redis.Client
}

func NewCartRepository(redisClient *redis.Client) CartRepository {
	return &cartRepository{redis: redisClient}
}

func (r *cartRepository) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	key := fmt.Sprintf("cart:%s", userID)

	data, err := r.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return &domain.Cart{UserID: userID, Items: []domain.CartItem{}}, nil
	} else if err != nil {
		return nil, fmt.Errorf("repository: failed to fetch cart from redis: %w", err)
	}

	var cart domain.Cart
	err = json.Unmarshal([]byte(data), &cart)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to unmarshal cart data: %w", err)
	}

	return &cart, nil
}

func (r *cartRepository) SaveCart(ctx context.Context, cart *domain.Cart) error {
	key := fmt.Sprintf("cart:%s", cart.UserID)

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("repository: failed to marshal cart data: %w", err)
	}

	err = r.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("repository: failed to save cart to redis: %w", err)
	}

	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	key := fmt.Sprintf("cart:%s", userID)

	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("repository: failed to delete cart from redis: %w", err)
	}

	return nil
}
