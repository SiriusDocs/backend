package domain

// UploadFileResponse - ответ на загрузку файла
type UploadFileResponse struct {
	TaskID string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// TaskStatusResponse - ответ при проверке статуса (когда ещё не готово)
type TaskStatusResponse struct {
	Status string `json:"status" example:"processing"`
}

// TaskResultNames - итоговый результат (обертка для списка)
type TaskResultNames struct {
	Names []string `json:"names"`
}

// CreateParamsRequest - запрос на создание новой таблицы по шаблону
type CreateParamsRequest struct {
	TaskID string            `json:"task_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// swaggertype:"object,string" подсказывает сваггеру, что это мапа строк
	Params map[string]string `json:"params" binding:"required" swaggertype:"object,string" example:"client_name:string,contract_sum:int,is_active:boolean"`
}

// CreateParamsResponse - ответ по поводу создания таблицы по шаблону
type CreateParamsResponse struct {
	TemplateID string `json:"template_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}


