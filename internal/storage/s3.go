package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"

	appConfig "watermark-service/internal/config"
)

// S3Storage implements the ImageStorage interface for AWS S3 and compatible services.
type S3Storage struct {
	client *s3.Client
	bucket string
	prefix string
	log    *logrus.Entry
}

// NewS3Storage creates a new S3 storage backend.
func NewS3Storage(cfg appConfig.S3Config, logger *logrus.Logger) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// For S3-compatible services like R2, a custom endpoint is needed.
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		}
		// fall back to default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolverV2 = customResolver
	})

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
		log:    logger.WithField("component", "S3Storage"),
	}, nil
}

// Get retrieves an image from S3.
func (s *S3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := s.prefix + key
	s.log.WithField("key", fullKey).Info("Getting from S3")

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		s.log.WithError(err).WithField("key", fullKey).Error("Failed to get object from S3")
		return nil, fmt.Errorf("could not get object from s3: %w", err)
	}
	defer result.Body.Close()

	// This is not the most efficient way, but it's simple.
	// For very large files, streaming would be better.
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	s.log.WithField("key", fullKey).Info("Successfully got object from S3")
	return buf.Bytes(), nil
}
