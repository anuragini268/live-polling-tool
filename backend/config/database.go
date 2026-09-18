package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectDB() {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		fmt.Println("REDIS_URL not found")
		return
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Println("Redis URL error:", err)
		return
	}

	RedisClient = redis.NewClient(opt)

	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Redis connection failed:", err)
		return
	}

	fmt.Println("Redis connected successfully!")
}