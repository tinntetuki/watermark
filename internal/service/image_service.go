package service

import (
	"context"
	"fmt"
	"time"
	"watermark/internal/config"
	"watermark/internal/processor"
	"watermark/internal/storage"
	"watermark/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_hits_total",
		Help: "Total number of cache hits",
	})

	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_misses_total",
		Help: "Total number of cache misses",
	})

	processingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "image_processing_duration_seconds",
		Help:    "Duration of image processing",
		Buckets: prometheus.DefBuckets,
	})
)

type ImageService struct {
	s3Client  *storage.S3Client
	cache     *storage.CacheClient
	processor *processor.WatermarkProcessor
	config    *config.Config
	logger    *logger.Logger
}

type ProcessRequest struct {
	ImageID    string
	Weight     float64
	Dimensions string
}

func NewImageService(
	s3Client *storage.S3Client,
	cache *storage.CacheClient,
	processor *processor.WatermarkProcessor,
	cfg *config.Config,
	logger *logger.Logger,
) *ImageService {
	return &ImageService{
		s3Client:  s3Client,
		cache:     cache,
		processor: processor,
		config:    cfg,
		logger:    logger,
	}
}

func (s *ImageService) ProcessImage(ctx context.Context, req ProcessRequest) ([]byte, error) {
	startTime := time.Now()
	defer func() {
		processingDuration.Observe(time.Since(startTime).Seconds())
	}()

	cacheKey := s.generateCacheKey(req)

	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("Cache get error", "error", err)
	}
	if cached != nil {
		cacheHits.Inc()
		s.logger.Info("Cache hit", "imageID", req.ImageID)
		return cached, nil
	}
	cacheMisses.Inc()

	s.logger.Info("Downloading from S3", "imageID", req.ImageID)
	imageData, err := s.s3Client.GetObject(ctx, req.ImageID)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	s.logger.Info("Adding watermark", "imageID", req.ImageID)
	processed, err := s.processor.AddWatermark(imageData, processor.WatermarkOptions{
		Weight:     req.Weight,
		Dimensions: req.Dimensions,
		Quality:    s.config.ImageQuality,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, processed, s.config.CacheTTL); err != nil {
		s.logger.Warn("Cache set error", "error", err)
	}

	s.logger.Info("Image processed successfully",
		"imageID", req.ImageID,
		"duration", time.Since(startTime),
	)

	return processed, nil
}

func (s *ImageService) generateCacheKey(req ProcessRequest) string {
	return fmt.Sprintf("image:%s:%.2f:%s", req.ImageID, req.Weight, req.Dimensions)
}

func (s *ImageService) WarmupCache(ctx context.Context, requests []ProcessRequest) error {
	s.logger.Info("Starting cache warmup", "count", len(requests))

	for _, req := range requests {
		go func(r ProcessRequest) {
			if _, err := s.ProcessImage(ctx, r); err != nil {
				s.logger.Error("Cache warmup failed",
					"imageID", r.ImageID,
					"error", err,
				)
			}
		}(req)
	}

	return nil
}

func (s *ImageService) InvalidateCache(ctx context.Context, imageID string) error {
	s.logger.Info("Invalidating cache", "imageID", imageID)

	// Pattern to match all cache keys for this imageID
	pattern := fmt.Sprintf("image:%s:*", imageID)

	// Delete all matching keys
	count, err := s.cache.DeleteByPattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	s.logger.Info("Cache invalidated", "imageID", imageID, "keysDeleted", count)
	return nil
}
