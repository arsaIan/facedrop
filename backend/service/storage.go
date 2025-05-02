package service

import (
	"context"
	"mofoto/storage"
)

type StorageClient struct {
	s3Client *storage.S3Client
}

func NewStorageClient(s3Client *storage.S3Client) *StorageClient {
	return &StorageClient{s3Client: s3Client}
}

func (s *StorageClient) UploadFile(ctx context.Context, fileKey string, data []byte, bucket string) (string, error) {
	return s.s3Client.UploadFile(ctx, fileKey, data, bucket)
}

func (s *StorageClient) DeleteFile(ctx context.Context, fileKey string, bucket string) error {
	return s.s3Client.DeleteFile(ctx, fileKey, bucket)
}

func (s *StorageClient) GetFileURL(fileKey string, bucket string) string {
	return s.s3Client.GetFileURL(fileKey, bucket)
} 