-- +goose Up
-- +goose StatementBegin
CREATE TABLE templates (
  id UUID PRIMARY KEY,
  name VARCHAR(50) NOT NULL DEFAULT 'Шаблон', -- 'Красивое' название для шаблона
  description TEXT, -- Описание шаблона от юзера
  vars TEXT, -- Список переменных шаблона в JSON
  -- TODO: add author_id
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS templates;
-- +goose StatementEnd
