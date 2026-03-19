-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_routes
ADD COLUMN IF NOT EXISTS feedback_url VARCHAR(255) DEFAULT 'https://t.me/imedteam_feedback';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
