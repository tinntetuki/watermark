package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"watermark-service/internal/processor"
	"watermark-service/internal/storage"
)

var (
	imageProcessDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "image_processing_duration_seconds",
		Help: "Duration of image processing.",
	})
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_hits_total",
		Help: "The total number of cache hits.",
	})
	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "image_cache_misses_total",
		Help: "The total number of cache misses.",
	})
)

// ImageService is the core service for processing images.
// It orchestrates the fetching, processing, and caching of images.	ype ImageService struct {
	storage   storage.ImageStorage
	cache     storage.ImageCache
	processor *processor.WatermarkProcessor
	log       *logrus.Entry
}

// NewImageService creates a new ImageService.
func NewImageService(
	storage storage.ImageStorage,
	cache storage.ImageCache,
	processor *processor.WatermarkProcessor,
	logger *logrus.Logger,
) *ImageService {
	return &ImageService{
		storage:   storage,
		cache:     cache,
		processor: processor,
		log:       logger.WithField("component", "ImageService"),
	}
}

// ProcessImage handles the main logic for fetching, watermarking, and caching an image.
func (s *ImageService) ProcessImage(ctx context.Context, imageKey, watermarkText string) ([]byte, error) {
	cacheKey := fmt.Sprintf("%s-%s", imageKey, watermarkText)

	// 1. Check cache first
	cachedImage, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		// Log the error but continue, as we can still fetch from origin.
		s.log.WithError(err).WithField("cache_key", cacheKey).Error("Cache GET failed")
	}
	if cachedImage != nil {
		cacheHits.Inc()
		s.log.WithField("cache_key", cacheKey).Info("Cache hit")
		return cachedImage, nil
	}

	// 2. Cache miss: fetch original image
	cacheMisses.Inc()
	s.log.WithField("cache_key", cacheKey).Info("Cache miss")

	originalImage, err := s.storage.Get(ctx, imageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get image from storage: %w", err)
	}

	// 3. Process the image
	startTime := time.Now()
	processedImage, err := s.processor.AddWatermark(originalImage, watermarkText)
	if err != nil {
		return nil, fmt.Errorf("failed to add watermark: %w", err)
	}
	imageProcessDuration.Observe(time.Since(startTime).Seconds())

	// 4. Store in cache for future requests (async)
	go func() {
		if err := s.cache.Set(context.Background(), cacheKey, processedImage); err != nil {
			s.log.WithError(err).WithField("cache_key", cacheKey).Error("Failed to set cache")
		}
	}()

	return processedImage, nil
}
