package cache

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds connection settings for Redis/Dragonfly.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisConfigFromEnv reads Redis connection config from environment variables.
func NewRedisConfigFromEnv() *RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	return &RedisConfig{
		Addr:     addr,
		Password: password,
		DB:       0,
	}
}

// NewRedisClient creates a new Redis client from the given config.
func NewRedisClient(cfg *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	slog.Info("Redis client created", "addr", cfg.Addr)

	return client, nil
}

// ActiveQueueKeyPrefix is the Redis key prefix for active queue entries.
const ActiveQueueKeyPrefix = "matchmaking:active_queue:"

// ActiveQueueSetKey is the Redis key for the set of all active player IDs.
const ActiveQueueSetKey = "matchmaking:active_queue:players"

// PlayerKey returns the Redis key for a specific player's queue entry.
func PlayerKey(playerID string) string {
	return fmt.Sprintf("%s%s", ActiveQueueKeyPrefix, playerID)
}
