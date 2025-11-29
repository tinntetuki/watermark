package storage

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// LocalCache implements the ImageCache interface using the local filesystem.
type LocalCache struct {
	path string
	ttl  time.Duration
	log  *logrus.Entry
}

// NewLocalCache creates a new filesystem-based cache.
// It ensures the cache directory exists.
func NewLocalCache(path string, ttl time.Duration, logger *logrus.Logger) (*LocalCache, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("cannot create cache directory %s: %w", path, err)
	}
	return &LocalCache{
		path: path,
		tl:   ttl,
		log:  logger.WithField("component", "LocalCache"),
	}, nil
}

// getFilePath generates a safe, unique file path for a given cache key.
func (c *LocalCache) getFilePath(key string) string {
	hash := sha1.Sum([]byte(key))
	return filepath.Join(c.path, fmt.Sprintf("%x.jpg", hash))
}

// Get retrieves an item from the cache. It returns nil if the item is not found or expired.
func (c *LocalCache) Get(ctx context.Context, key string) ([]byte, error) {
	filePath := c.getFilePath(key)
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, nil // Cache miss
	}
	if err != nil {
		c.log.WithError(err).WithField("path", filePath).Error("Failed to stat cache file")
		return nil, err // Other error
	}

	if time.Since(info.ModTime()) > c.ttl {
		c.log.WithField("path", filePath).Info("Cache item expired, removing")
		// Attempt to remove the stale file, but don't fail the Get operation if it fails.
		if err := os.Remove(filePath); err != nil {
			c.log.WithError(err).WithField("path", filePath).Warn("Failed to remove stale cache file")
		}
		return nil, nil // Cache miss
	}

	c.log.WithField("path", filePath).Debug("Cache hit")
	return os.ReadFile(filePath)
}

// Set adds an item to the cache.
func (c *LocalCache) Set(ctx context.Context, key string, data []byte) error {
	filePath := c.getFilePath(key)
	c.log.WithField("path", filePath).Debug("Setting cache item")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		c.log.WithError(err).WithField("path", filePath).Error("Failed to write cache file")
		return err
	}
	return nil
}
