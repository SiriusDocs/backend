-- +goose Up
-- +goose StatementBegin
CREATE TABLE processing_tasks (
    id UUID PRIMARY KEY, -- Уникальный ID задачи, который мы отдаем фронту
    file_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, done, error
    file_name VARCHAR(255), -- Имя исходного файла (для истории)
    result_data JSONB, -- Сюда положим результат парсинга, когда будет готово
    error_message TEXT, -- Если упало, запишем сюда текст ошибки
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processing_tasks;
-- +goose StatementEnd
