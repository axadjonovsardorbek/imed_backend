-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key TEXT NOT NULL,
  value JSONB NOT NULL DEFAULT '{}'::jsonb,
  description TEXT DEFAULT NULL,
  order_num BIGINT DEFAULT 1,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now(),
  deleted_at BIGINT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_app_routes_key_active
ON app_routes(key)
WHERE deleted_at = 0;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
