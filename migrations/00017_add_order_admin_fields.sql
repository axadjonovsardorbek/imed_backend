-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS comment TEXT DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS check_url TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
  DROP COLUMN IF EXISTS comment,
  DROP COLUMN IF EXISTS check_url;
-- +goose StatementEnd
