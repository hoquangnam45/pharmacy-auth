package service

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Redis struct {
	client *redis.Client
}

func NewRedisClient(configService *Config) *Redis {
	addr := configService.GetOrDefault("REDIS_HOST", "localhost")
	port := configService.GetOrDefault("REDIS_PORT", "6379")
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", addr, port),
			Username: "", // No username
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
}

func (s *Redis) Get(ctx context.Context, namespace string, key string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if namespace == "" {
		return s.client.Get(ctx, key).Result()
	}
	return s.client.Get(ctx, fmt.Sprintf("%s:%s", namespace, key)).Result()
}
