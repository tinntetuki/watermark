package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	ServerPort         int
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
	AWS                AWSConfig
	Redis              RedisConfig
	CacheTTL           time.Duration
	FontPath           string
	FontSize           float64
	WatermarkColor     string
	ImageQuality       int
	LogLevel           string
}

type AWSConfig struct {
	Region string
	Bucket string
	Prefix string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	redisURL := os.Getenv("REDIS_URL")
	var redisConfig RedisConfig

	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
		}
		redisConfig = RedisConfig{
			Addr:     opt.Addr,
			Password: opt.Password,
			DB:       opt.DB,
		}
	} else {
		redisConfig = RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		}
	}

	cfg := &Config{
		ServerPort:         getEnvAsInt("SERVER_PORT", 8080),
		ServerReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		ServerWriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ServerIdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
			Bucket: getEnv("AWS_S3_BUCKET", ""),
			Prefix: getEnv("AWS_S3_PREFIX", "qc-images/"),
		},
		Redis:          redisConfig,
		CacheTTL:       getEnvAsDuration("CACHE_TTL", 7*24*time.Hour),
		FontPath:       getEnv("FONT_PATH", "./fonts/Arial.ttf"),
		FontSize:       getEnvAsFloat("FONT_SIZE", 24.0),
		WatermarkColor: getEnv("WATERMARK_COLOR", "#FFFFFF"),
		ImageQuality:   getEnvAsInt("IMAGE_QUALITY", 90),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if cfg.AWS.Bucket == "" {
		return nil, fmt.Errorf("AWS_S3_BUCKET is required")
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
