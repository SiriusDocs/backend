package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type TaskOperationsPostgres struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewTaskOperationsPostgres(db *sqlx.DB, logger *slog.Logger) *TaskOperationsPostgres {
	return &TaskOperationsPostgres{
		db:     db,
		logger: logger,
	}
}

// CreateTask создает запись. ID мы передаем извне (UUID генерируется в сервисе)
func (r *TaskOperationsPostgres) CreateTask(ctx context.Context, id string, fileName string) error {
	const op = "storage.postgres.task.CreateTask"

	query := fmt.Sprintf(`
		INSERT INTO %s (id, file_name, file_status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)`,
		tasksTable)

	now := time.Now()
	// Статус по умолчанию pending
	_, err := r.db.ExecContext(ctx, query, id, fileName, domain.TaskStatusPending, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// SetStatus просто меняет статус (например, pending -> processing)
func (r *TaskOperationsPostgres) SetStatus(ctx context.Context, id string, status string) error {
	const op = "storage.postgres.task.SetStatus"

	query := fmt.Sprintf(`
		UPDATE %s 
		SET file_status = $1, updated_at = $2 
		WHERE id = $3`,
		tasksTable)

	res, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка, что такая задача вообще была
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrTaskNotFound)
	}

	return nil
}

// SetResult сохраняет JSON и помечает как done
func (r *TaskOperationsPostgres) SetResult(ctx context.Context, id string, resultJSON []byte) error {
	const op = "storage.postgres.task.SetResult"

	query := fmt.Sprintf(`
		UPDATE %s 
		SET file_status = $1, result_data = $2, updated_at = $3 
		WHERE id = $4`,
		tasksTable)

	// resultJSON (type []byte) отлично ложится в поле типа JSONB в Postgres
	res, err := r.db.ExecContext(ctx, query, domain.TaskStatusDone, resultJSON, time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrTaskNotFound)
	}

	return nil
}

// SetError записывает сообщение об ошибке
func (r *TaskOperationsPostgres) SetError(ctx context.Context, id string, errorMsg string) error {
	const op = "storage.postgres.task.SetError"

	query := fmt.Sprintf(`
		UPDATE %s 
		SET file_status = $1, error_message = $2, updated_at = $3 
		WHERE id = $4`,
		tasksTable)

	res, err := r.db.ExecContext(ctx, query, domain.TaskStatusError, errorMsg, time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrTaskNotFound)
	}

	return nil
}

// GetTask возвращает всю информацию о задаче
func (r *TaskOperationsPostgres) GetTask(ctx context.Context, id string) (domain.Task, error) {
	const op = "storage.postgres.task.GetTask"

	var task domain.Task
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tasksTable)

	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, fmt.Errorf("%s: %w", op, domain.ErrTaskNotFound)
		}
		return domain.Task{}, fmt.Errorf("%s: %w", op, err)
	}

	return task, nil
}