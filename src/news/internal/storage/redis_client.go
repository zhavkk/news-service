package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zhavkk/news-service/src/news/internal/config"
	"github.com/zhavkk/news-service/src/news/internal/logger"
)

type RedisClient struct {
	redis *redis.Client
}

func NewRedisClient(ctx context.Context, cfg *config.RedisConfig) (*RedisClient, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Log.Error("failed to ping redis", "error", err)
		return nil, err
	}

	return &RedisClient{redis: redisClient}, nil
}

func (r *RedisClient) GetRedis() *redis.Client {
	return r.redis
}
