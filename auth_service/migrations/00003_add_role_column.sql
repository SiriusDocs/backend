-- +goose Up
-- +goose StatementBegin
ALTER TABLE users 
ADD COLUMN user_role VARCHAR(100) NOT NULL DEFAULT 'PENDING';

CREATE INDEX idx_users_role ON users(user_role);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users 
DROP COLUMN user_role;
-- +goose StatementEnd