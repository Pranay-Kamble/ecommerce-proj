package database_test

import (
	"context"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logger"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	logger.Init("prod")
	t.Run("Postgres Test", func(t *testing.T) {
		dsn := "host=localhost user=admin password=password dbname=auth_db port=5432 sslmode=disable"
		db, err := database.Connect(dsn)

		assert.NoError(t, err, "Could not connect to database")
		assert.NotNil(t, db, "Database should not be nil")

	})

	t.Run("Redis Check", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		status, err := rdb.Ping(ctx).Result()
		assert.NoError(t, err, "Redis should be reachable")
		assert.Equal(t, "PONG", status, "Redis should reply with PONG")
	})
}
