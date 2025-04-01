package service

import (
	"context"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var Cache *cache.Cache

// InitRedis initializes the Redis client with OpenTelemetry instrumentation
func InitRedis(redisUrl string) error {
	if redisUrl == "" {
		cfg := config.GetRuntimeConfig()
		redisUrl = cfg.RedisURL
	}

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return err
	}

	redisClient = redis.NewClient(opt)

	// Enable OpenTelemetry instrumentation
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return err
	}

	// Enable OpenTelemetry metrics
	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		return err
	}

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return err
	}

	Cache = cache.New(&cache.Options{
		Redis:      redisClient,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return nil
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
