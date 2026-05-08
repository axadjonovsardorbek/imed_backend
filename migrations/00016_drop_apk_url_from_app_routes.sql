-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_routes DROP COLUMN apk_url;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_routes ADD COLUMN apk_url TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
