package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"roly-poly/config"
	"roly-poly/pkg/logger"
)

func New(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPass,
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}

func HealthCheck(ctx context.Context, client *redis.Client) bool {
	log := logger.New()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Error().Err(err).Msg("Redis health check failed")
		return false
	}
	return true
}
