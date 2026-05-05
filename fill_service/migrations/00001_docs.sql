-- +goose Up
-- +goose StatementBegin

-- TODO: maybe add more states?
CREATE TYPE doc_status AS ENUM(
  'draft',
  'on_review',
  'finished',
  'published'
);

CREATE TABLE documents (
  id UUID PRIMARY KEY, -- Уникальный ID задачи, который мы отдаем фронту
  template UUID REFERENCES templates(id), -- Шаблон этого документа
  status doc_status NOT NULL DEFAULT 'draft', -- pending, processing, done, error
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents;
DROP TYPE IF EXISTS doc_status;
-- +goose StatementEnd
