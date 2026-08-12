package files

import (
	"context"
	"log/slog"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/storage"
)

type FileService struct {
	log *slog.Logger
	store storage.FileOperations
}

func NewFileOperations(logger *slog.Logger, store storage.FileOperations) *FileService {
	return &FileService{
		log: logger,
		store: store,
	}
}

func (f *FileService) GenerateUploadURL(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	return f.store.PresignPut(ctx, key, contentType, ttl)
}

func (f *FileService) GenerateDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return f.store.PresignGet(ctx, key, ttl)
}

func (f *FileService) GetFileMetadata(ctx context.Context, key string) (domain.FileMetadata, error) {
	return f.store.Head(ctx, key)
}

func (f *FileService) DeleteFile(ctx context.Context, key string) error {
	return f.store.Delete(ctx,key)
}