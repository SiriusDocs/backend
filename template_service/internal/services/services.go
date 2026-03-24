package templates

import (
	"context"
	"log/slog"

	temp "github.com/SiriusDocs/backend/template_service/internal/services/templates"
	"github.com/SiriusDocs/backend/template_service/internal/storage"
)

// TaskOperations описывает бизнес-логику работы с задачами
type TaskOperations interface {
	// UploadAndStartTask принимает имя файла и данные, создает задачу и запускает парсинг в фоне
	UploadAndStartTask(ctx context.Context, filename string, fileData []byte) (string, error)
	
	// CheckTaskStatus возвращает статус задачи и результат (список имен), если готов
	CheckTaskStatus(ctx context.Context, taskID string) (string, []string, error)
}

type Service struct {
	TaskOperations
}

func NewService(log *slog.Logger, store *storage.Storage) *Service {
	return &Service{
		TaskOperations: temp.NewTemplateService(log, store.TaskOperations),
	}
}