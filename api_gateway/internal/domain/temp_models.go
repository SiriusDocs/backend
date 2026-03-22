package domain

// UploadFileResponse - ответ на загрузку файла
type UploadFileResponse struct {
	TaskID string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// TaskStatusResponse - ответ при проверке статуса (когда ещё не готово)
type TaskStatusResponse struct {
	Status string `json:"status" example:"processing"`
}

// TaskResultResponse - ответ при успешном парсинге
// Swagger не умеет описывать произвольный JSON, поэтому используем map
type TaskResultResponse map[string]interface{}