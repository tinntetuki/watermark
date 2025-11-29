package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"watermark-service/internal/config"
)

// RedisCache implements the ImageCache interface using Redis.
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	log    *logrus.Entry
}

// NewRedisCache creates a new Redis-backed cache.
func NewRedisCache(cfg config.RedisConfig, ttl time.Duration, logger *logrus.Logger) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisCache{
		client: rdb,
		ttl:    ttl,
		log:    logger.WithField("component", "RedisCache"),
	}
}

// Get retrieves an item from the Redis cache.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.log.WithField("key", key).Debug("Getting from redis")
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		c.log.WithField("key", key).Debug("Redis cache miss")
		return nil, nil // Cache miss
	} else if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Redis GET failed")
		return nil, fmt.Errorf("redis GET failed for key %s: %w", key, err)
	}

	c.log.WithField("key", key).Debug("Redis cache hit")
	return val, nil
}

// Set adds an item to the Redis cache with the configured TTL.
func (c *RedisCache) Set(ctx context.Context, key string, data []byte) error {
	c.log.WithField("key", key).Debug("Setting to redis")
	err := c.client.Set(ctx, key, data, c.ttl).Err()
	if err != nil {
		c.log.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("redis SET failed for key %s: %w", key, err)
	}
	return nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
