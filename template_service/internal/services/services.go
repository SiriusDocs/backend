package templates

import (
	"context"
	"log/slog"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/SiriusDocs/backend/template_service/internal/config"
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

type ParamsOperations interface {
	// CreateTemplate создает таблицу в БД на основе параметров
	CreateTemplate(ctx context.Context, taskID string, params map[string]string) (string, error)

	// ListTemplates возвращает список зарегистрированных шаблонов
	ListTemplates(ctx context.Context, perPage int, page int) ([]domain.Template, error)
}

type Service struct {
	TaskOperations
	ParamsOperations
}

func NewService(log *slog.Logger, store *storage.Storage, cfg config.TasksConfig) *Service {
	return &Service{
		TaskOperations:   temp.NewTasksService(log, store.TaskOperations, cfg),
		ParamsOperations: temp.NewParamsService(log, store.TemplateOperations, store.TaskOperations),
	}
}
