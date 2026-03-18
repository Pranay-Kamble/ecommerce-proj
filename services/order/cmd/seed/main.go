package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type CartItem struct {
	ProductVariantID string  `json:"product_variant_id"`
	Quantity         int     `json:"quantity"`
	Price            float64 `json:"price"`
}

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

func main() {
	_ = godotenv.Load(".env")

	redisDSN := os.Getenv("REDIS_DSN")
	if redisDSN == "" {
		redisDSN = "redis://localhost:6379/1"
	}

	opt, err := redis.ParseURL(redisDSN)
	if err != nil {
		log.Fatalf("Failed to parse Redis DSN: %v", err)
	}

	rdb := redis.NewClient(opt)
	ctx := context.Background()

	log.Println("Start Order Redis Seeding...")

	testUserID := "usr_test_999"

	mockCart := Cart{
		UserID: testUserID,
		Items: []CartItem{
			{
				ProductVariantID: "var_test_123",
				Quantity:         2,
				Price:            1.00,
			},
		},
	}

	cartBytes, err := json.Marshal(mockCart)
	if err != nil {
		log.Fatalf("Failed to marshal cart: %v", err)
	}

	redisKey := "cart:" + testUserID
	err = rdb.Set(ctx, redisKey, cartBytes, 0).Err()
	if err != nil {
		log.Fatalf("Failed to save to Redis: %v", err)
	}

	log.Printf("USER ID: %s has a cart ready for checkout.", testUserID)
}
