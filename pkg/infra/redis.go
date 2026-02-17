package infra

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/infra/cache"
	"github.com/redis/go-redis/v9"
)

// InjectRedis sets up the Redis/Dragonfly client in the container.
func InjectRedis(c container.Container) error {
	return c.Singleton(func() (*redis.Client, error) {
		cfg := cache.NewRedisConfigFromEnv()
		return cache.NewRedisClient(cfg)
	})
}
