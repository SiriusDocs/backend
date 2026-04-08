package domain

import (
	"time"
)

// TaskStatuses — константы для статусов
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusDone       = "done"
	TaskStatusError      = "error"
)

// Task соответствует таблице processing_tasks
type Task struct {
	ID           string    `db:"id"`
	FileStatus   string    `db:"file_status"`
	FileName     string    `db:"file_name"`
	ResultData   []byte    `db:"result_data"`   // JSONB мапится в []byte
	ErrorMessage *string   `db:"error_message"` // Pointer, т.к. может быть NULL
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
