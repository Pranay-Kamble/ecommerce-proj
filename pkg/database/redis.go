package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Redis *redis.Client // Exported so you can pass it to repositories
}

func (r *Redis) Connect(dsn string) error {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return fmt.Errorf("database (redis): failed to parse redis url: %w", err)
	}

	r.Redis = redis.NewClient(options)

	if err := r.Redis.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("database (redis): failed to ping server: %w", err)
	}

	return nil
}

func (r *Redis) Close() error {
	if r.Redis == nil {
		return nil
	}
	return r.Redis.Close()
}
