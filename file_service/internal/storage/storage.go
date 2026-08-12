package storage

import (
	"context"
	"fmt"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/storage/s3"
)

type FileOperations interface {
	// загрузка файла
	PresignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
	// скачать файл
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	// получить метаданный
	Head(ctx context.Context, key string) (domain.FileMetadata, error)
	// удалить файл
	Delete(ctx context.Context, key string) error
	// copy
}

type Storage struct {
	FileOperations
}

func NewS3Storage(ctx context.Context, cfg config.S3) (*Storage, error) {
	const op = "storage.NewS3Storage"

	s3Ops, err := s3.New(ctx, cfg)
	if err != nil { return nil, fmt.Errorf("%s: %w", op, err) }

	return &Storage{
		FileOperations: s3Ops,
	}, nil
}