package event

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
)

type MinioStorage struct {
	client    *minio.Client
	publicURL string
	bucket    string
	log       *slog.Logger
}

func New(client *minio.Client, bucket string, publicURL string, log *slog.Logger) *MinioStorage {
	return &MinioStorage{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
		log:       log,
	}
}

func (s *MinioStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	const op = "event.cloud.Upload"

	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		s.log.Error("failed to upload file", slog.String("op", op), sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return objectName, nil
}

func (s *MinioStorage) GetURL(fileKey string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucket, fileKey)
}

func (s *MinioStorage) DeleteFile(ctx context.Context, objectName string) error {
	const op = "event.cloud.Delete"

	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		s.log.Error("failed to delete object", slog.String("op", op), sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *MinioStorage) PresignedURL(ctx context.Context, objectName string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, time.Hour*24, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
