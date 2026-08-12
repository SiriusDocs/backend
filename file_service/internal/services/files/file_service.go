package files

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/storage"
	logs "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/pkg/logger"
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
	const op = "services.files.GenerateUploadURL"
	start := time.Now()

	url, err := f.store.PresignPut(ctx, key, contentType, ttl)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		f.log.ErrorContext(ctx, "failed to generate upload url",
			slog.String("key", key),
			slog.String("content_type", contentType),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)

		return "", wrappedErr
	}

	f.log.InfoContext(ctx, "upload url generated",
		slog.String("key", key),
		slog.String("content_type", contentType),
		slog.Duration("ttl", ttl),
		logs.ResultAttrs("ok", start),
	)

	return url, nil
}

func (f *FileService) GenerateDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	const op = "services.files.GenerateDownloadURL"
	start := time.Now()

	url, err := f.store.PresignGet(ctx, key, ttl)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		f.log.ErrorContext(ctx, "failed to generate download url",
			slog.String("key", key),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)

		return "", wrappedErr
	}

	f.log.InfoContext(ctx, "download url generated",
		slog.String("key", key),
		slog.Duration("ttl", ttl),
		logs.ResultAttrs("ok", start),
	)

	return url, nil
}

func (f *FileService) GetFileMetadata(ctx context.Context, key string) (domain.FileMetadata, error) {
	const op = "services.files.GetFileMetadata"
	start := time.Now()

	meta, err := f.store.Head(ctx, key)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		f.log.ErrorContext(ctx, "failed to get file metadata",
			slog.String("key", key),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)

		return domain.FileMetadata{}, wrappedErr
	}

	if !meta.Exists {
		f.log.WarnContext(ctx, "file metadata requested but file not found",
			slog.String("key", key),
			logs.ResultAttrs("ok", start),
		)
	} else {
		f.log.InfoContext(ctx, "file metadata retrieved",
			slog.String("key", key),
			slog.Int64("size_bytes", meta.Size),
			slog.String("content_type", meta.ContentType),
			logs.ResultAttrs("ok", start),
		)
	}

	return meta, nil
}

func (f *FileService) DeleteFile(ctx context.Context, key string) error {
	const op = "services.files.DeleteFile"
	start := time.Now()

	err := f.store.Delete(ctx, key)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		f.log.ErrorContext(ctx, "failed to delete file",
			slog.String("key", key),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)

		return wrappedErr
	}

	f.log.InfoContext(ctx, "file deleted",
		slog.String("key", key),
		logs.ResultAttrs("ok", start),
	)

	return nil
}