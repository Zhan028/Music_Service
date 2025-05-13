package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var (
	ctx = context.Background()
)

type Redis struct {
	Client *redis.Client
}

func Init() Redis {
	addr := getEnv("REDIS_ADDR", "localhost:6379")

	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")

	result := Redis{Client: redisClient}

	return result
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// todo
func (r *Redis) Get(key string) ([]byte, error) {
	result, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		log.Printf(" Redis GET error: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf(" Redis SET marshal error: %v", err)
		return err
	}

	if err := r.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf(" Redis SET error: %v", err)
		return err
	}

	return nil
}

func (r *Redis) Delete(key string) error {
	if err := r.Client.Del(ctx, key).Err(); err != nil {
		log.Printf(" Redis DEL error: %v", err)
		return err
	}
	return nil
}
