package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// --- Top Level Config ---

type Config struct {
	ServerPort         int
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
	Storage            StorageConfig
	Cache              CacheConfig
	CacheTTL           time.Duration
	FontPath           string
	FontSize           float64
	WatermarkColor     string
	ImageQuality       int
	LogLevel           string
}

// --- Storage Configuration ---

type StorageConfig struct {
	Provider string
	S3       S3Config
}

// S3Config supports AWS S3 and S3-compatible services like Cloudflare R2.
type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	Prefix          string
	AccessKeyID     string
	SecretAccessKey string
}

// --- Cache Configuration ---

type CacheConfig struct {
	Provider string
	Redis    RedisConfig
	Local    LocalCacheConfig
}

type LocalCacheConfig struct {
	Path string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// --- Load Function ---

func Load() (*Config, error) {
	redisConfig, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}

	storageProvider := getEnv("STORAGE_PROVIDER", "s3")
	cacheProvider := getEnv("CACHE_PROVIDER", "redis")

	cfg := &Config{
		ServerPort:         getEnvAsInt("SERVER_PORT", 8080),
		ServerReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		ServerWriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ServerIdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		Storage: StorageConfig{
			Provider: storageProvider,
			S3: S3Config{
				Endpoint:        getEnv("S3_ENDPOINT", ""), // For R2: https://<accountid>.r2.cloudflarestorage.com
				Region:          getEnv("AWS_REGION", "auto"),
				Bucket:          getEnv("S3_BUCKET", ""),
				Prefix:          getEnv("S3_PREFIX", "qc-images/"),
				AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
				SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			},
		},
		Cache: CacheConfig{
			Provider: cacheProvider,
			Redis:    *redisConfig,
			Local: LocalCacheConfig{
				Path: getEnv("LOCAL_CACHE_PATH", "./cache"),
			},
		},
		CacheTTL:       getEnvAsDuration("CACHE_TTL", 7*24*time.Hour),
		FontPath:       getEnv("FONT_PATH", "./fonts/Arial.ttf"),
		FontSize:       getEnvAsFloat("FONT_SIZE", 24.0),
		WatermarkColor: getEnv("WATERMARK_COLOR", "#FFFFFF"),
		ImageQuality:   getEnvAsInt("IMAGE_QUALITY", 90),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if cfg.Storage.S3.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}

	return cfg, nil
}

func loadRedisConfig() (*RedisConfig, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("could not parse REDIS_URL: %w", err)
		}
		return &RedisConfig{
			Addr:     opts.Addr,
			Password: opts.Password,
			DB:       opts.DB,
		}, nil
	}

	return &RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	}, nil
}

// --- Env Helper Functions ---

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
