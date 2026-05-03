-- +goose Up
-- +goose StatementBegin
ALTER TABLE sources ADD COLUMN IF NOT EXISTS video_url JSONB DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
