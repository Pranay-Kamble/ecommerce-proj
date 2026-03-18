package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"ecommerce/pkg/database"

	"github.com/joho/godotenv"
)

type CartItem struct {
	ProductID string  `json:"product_id"` // This maps to the Variant Public ID
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}

func main() {
	_ = godotenv.Load(".env")
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	rdb := &database.Redis{}
	err := rdb.Connect(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Start Order Redis Seeding")

	testUserID := "usr_test_999"

	mockCart := Cart{
		Items: []CartItem{
			{
				ProductID: "var_test_123",
				Quantity:  2,
				Price:     1.00, //keep price intentionally small, so the gRPC can overwrite this
			},
		},
	}

	cartBytes, err := json.Marshal(mockCart)
	if err != nil {
		log.Fatalf("Failed to marshal cart: %v", err)
	}

	redisKey := "cart:" + testUserID
	ctx := context.Background()
	err = rdb.Redis.Set(ctx, redisKey, cartBytes, 0).Err()
	if err != nil {
		log.Fatalf("Failed to save to Redis: %v", err)
	}

	log.Printf("USER ID: %s has a cart with fake prices ready for checkout.", testUserID)
}
