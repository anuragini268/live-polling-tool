package config

import (
	"context"
	"crypto/tls"
	"fmt"
        "net/url"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	RedisClient *redis.Client
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
)

func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB connection
	mongoURI := os.Getenv("MONGODB_URI")

	if mongoURI == "" {
		fmt.Println("MONGODB_URI not found")
	} else {
		client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
		if err != nil {
			fmt.Println("MongoDB connection failed:", err)
		} else {
			err = client.Ping(ctx, nil)
			if err != nil {
				fmt.Println("MongoDB ping failed:", err)
			} else {
				MongoClient = client
				MongoDB = client.Database("livepolling")
				fmt.Println("MongoDB connected successfully!")
			}
		}
	}

		// Redis connection
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		fmt.Println("REDIS_URL not found")
		return
	}

	u, err := url.Parse(redisURL)
	if err != nil {
		fmt.Println("Redis URL parse error:", err)
		return
	}

	password, _ := u.User.Password()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Username: u.User.Username(),
		Password: password,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: u.Hostname(),
		},
	})

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer redisCancel()

	_, err = RedisClient.Ping(redisCtx).Result()
	if err != nil {
		fmt.Println("Redis connection failed:", err)
		return
	}

	fmt.Println("Redis connected successfully!")
}
