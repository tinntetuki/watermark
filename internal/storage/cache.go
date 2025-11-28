package storage

import (
	"context"
	"fmt"
	"time"
	"watermark/internal/config"

	"github.com/redis/go-redis/v9"
)

type CacheClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) *CacheClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &CacheClient{client: rdb}
}

func (c *CacheClient) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return data, nil
}

func (c *CacheClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

func (c *CacheClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *CacheClient) DeleteByPattern(ctx context.Context, pattern string) (int, error) {
	// Use SCAN to find all matching keys
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan cache keys: %w", err)
	}

	// Delete all matching keys
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return 0, fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return len(keys), nil
}

func (c *CacheClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *CacheClient) Close() error {
	return c.client.Close()
}
