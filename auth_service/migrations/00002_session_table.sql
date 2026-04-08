-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_sessions (
    "id" SERIAL PRIMARY KEY,
    "user_id" INT REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    "refresh_token" VARCHAR(128) NOT NULL UNIQUE,
    "expires_in" timestamp with time zone NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE refresh_sessions;
-- +goose StatementEnd
