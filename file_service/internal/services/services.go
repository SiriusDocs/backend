package services

import (
	"context"
	"log/slog"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/services/files"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/storage"
)

type FileOperations interface {
	GenerateUploadURL(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
	GenerateDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	GetFileMetadata(ctx context.Context, key string) (domain.FileMetadata, error)
	DeleteFile(ctx context.Context, key string) error
	// CopyFile
}

type Service struct {
	FileOperations
}


func NewService(logger *slog.Logger, store *storage.Storage) *Service {
	return &Service{
		FileOperations: files.NewFileOperations(logger, store.FileOperations),
	}
}
