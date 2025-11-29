package storage

import (
	"context"
)

// ImageStorage defines the interface for an object storage backend.
// It is responsible for fetching the original images.
type ImageStorage interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// ImageCache defines the interface for a cache backend.
// It is responsible for storing and retrieving processed images to improve performance.
type ImageCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}
