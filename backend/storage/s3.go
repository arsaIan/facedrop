package storage

import (
	"archive/zip"
	"bytes"
	"context"
	"facedrop/config"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client     *s3.Client
	buckets    []string
}

func NewS3Client(cfg *config.Config) (*S3Client, error) {
	// Create custom endpoint resolver for MinIO
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               cfg.StorageConfig.Endpoint,
			HostnameImmutable: true,
		}, nil
	})

	// Load AWS configuration
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.StorageConfig.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.StorageConfig.AccessKeyID,
			cfg.StorageConfig.SecretAccessKey,
			"",
		)),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	return &S3Client{
		client:     client,
		buckets: []string{cfg.StorageConfig.EventBucket, cfg.StorageConfig.UserBucket},
	}, nil
}

func (s *S3Client) UploadFile(ctx context.Context, fileKey string, data []byte, bucket string) (string, error) {
	// Generate a unique file key with timestamp
	timestamp := time.Now().Unix()
	uniqueKey := fmt.Sprintf("%d_%s", timestamp, fileKey)

	// Upload the file
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uniqueKey),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate presigned URL for the uploaded file
	presignClient := s3.NewPresignClient(s.client)
	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uniqueKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 7 * 24 * time.Hour // URL expires in 7 days
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.URL, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, fileKey string, bucket string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *S3Client) GetFileURL(fileKey string, bucket string) string {
	return fmt.Sprintf("%s/%s/%s", *s.client.Options().BaseEndpoint, bucket, fileKey)
}

func (s *S3Client) GetZippedFiles(ctx context.Context, fileURLs []string, bucket string) ([]byte, error) {
	// Create a buffer to write our zip file to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Download each file and add it to the zip
	for _, fileURL := range fileURLs {
		// Extract file key from URL
		// URL format: .../mofoto/{key}?...
		parts := strings.Split(fileURL, "/mofoto/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid file URL format: %s", fileURL)
		}
		
		// Get everything before the ? character
		keyParts := strings.Split(parts[1], "?")
		if len(keyParts) == 0 {
			return nil, fmt.Errorf("invalid file URL format: %s", fileURL)
		}
		
		fileKey := keyParts[0]

		// Get the file from S3
		result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileKey),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get file %s: %w", fileURL, err)
		}
		defer result.Body.Close()

		// Create a new file in the zip archive with just the filename
		filename := filepath.Base(fileKey)
		zipFile, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", fileURL, err)
		}

		// Copy the file contents to the zip
		_, err = io.Copy(zipFile, result.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %s to zip: %w", fileURL, err)
		}
	}

	// Close the zip writer
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *S3Client) UploadZipFile(ctx context.Context, zipData []byte, eventID string, bucket string) (string, error) {
	// Generate a unique file key with timestamp and event ID
	timestamp := time.Now().Unix()
	zipKey := fmt.Sprintf("events/%s/%d_photos.zip", eventID, timestamp)

	// Upload the zip file
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(zipKey),
		Body:   bytes.NewReader(zipData),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload zip file: %w", err)
	}

	// Generate presigned URL for the uploaded zip file
	presignClient := s3.NewPresignClient(s.client)
	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(zipKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 7 * 24 * time.Hour // URL expires in 7 days
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.URL, nil
} 