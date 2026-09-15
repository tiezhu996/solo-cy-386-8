package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/config"
	"github.com/redis/go-redis/v9"
)

// ConnectRedis 建立 Redis 连接（消息实时推送 pub/sub 通道）。
func ConnectRedis(cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis %s: %w", cfg.RedisAddr, err)
	}
	log.Info("redis connected", "addr", cfg.RedisAddr)
	return client, nil
}
