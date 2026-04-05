package storage

import (
	"context"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/SiriusDocs/backend/template_service/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

// TaskOperations описывает методы работы с задачами
type TaskOperations interface {
	// CreateTask создает новую задачу со статусом pending
	CreateTask(ctx context.Context, id string, fileName string) error

	// SetStatus обновляет статус задачи (например, переводит в processing)
	SetStatus(ctx context.Context, id string, status string) error

	// SetResult сохраняет результат успешного парсинга и ставит статус done
	SetResult(ctx context.Context, id string, resultJSON []byte) error

	// SetError сохраняет ошибку и ставит статус error
	SetError(ctx context.Context, id string, errorMsg string) error

	// GetTask получает задачу по ID
	GetTask(ctx context.Context, id string) (domain.Task, error)

	// DeleteTask удаляет задачу по ID
	DeleteTask(ctx context.Context, id string) error
}

type TemplateOperations interface {
	// CreateTemplateTable создает динамическую таблицу для шаблона
	CreateTemplateTable(ctx context.Context, templateID string, columns map[string]string) error
}

type Storage struct {
	TaskOperations
	TemplateOperations
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		TaskOperations:     postgres.NewTaskOperationsPostgres(db),
		TemplateOperations: postgres.NewTemplateOperationsPostgres(db),
	}
}
