package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/Shiva936/newsletter-service/internal/config"
)

var ctx = context.Background()

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: "", // no password by default; update if needed
		DB:       0,  // default DB
	})

	// Test connection
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v\n", err)
		return nil, err
	}

	log.Printf("Connected to Redis: %s\n", pong)
	return client, nil
}
