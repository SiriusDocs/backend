package logger

import (
	"fmt"
	"log/slog"
	"time"
)

// BaseAttrs — набор, который ставится один раз при старте и живёт на каждой записи.
func BaseAttrs(env string) []any {
	return []any{
		slog.String("service", "file_service"), // TODO: Хардкод!! Исправить во всех сервисах
		slog.String("env", env),
	}
}

// FileAttrs — доменный набор: что за файл обрабатываем.
// Имя файла логировать можно, содержимое — нет.
func FileAttrs(id, name string, size int64, contentType string) slog.Attr {
	return slog.Group("file",
		slog.String("id", id),
		slog.String("name", name),
		slog.Int64("size_bytes", size),
		slog.String("content_type", contentType),
	)
}

// S3Attrs — набор внешнего хранилища. Ключи доступа и endpoint провайдера сюда не попадают.
func S3Attrs(bucket, key, operation string, attempt int) slog.Attr {
	return slog.Group("s3",
		slog.String("bucket", bucket),
		slog.String("key", key),
		slog.String("operation", operation),
		slog.Int("attempt", attempt),
	)
}

// ResultAttrs — хвост канонической записи: чем кончилась операция и сколько заняла.
// status — "ok" или "error".
func ResultAttrs(status string, start time.Time) slog.Attr {
	return slog.Group("result",
		slog.String("status", status),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
}

// ErrAttrs — ошибка в лог. Именно err.Error(), а не slog.Any: у большинства типов
// ошибок нет MarshalJSON, и JSONHandler выведет пустой объект.
func ErrAttrs(err error) slog.Attr {
	if err == nil {
		return slog.Attr{} // пустой Attr хендлер отбрасывает
	}

	return slog.Group("error",
		slog.String("message", err.Error()),
		slog.String("type", fmt.Sprintf("%T", err)),
	)
}
