-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS app_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_center TEXT NOT NULL,
    support_url TEXT NOT NULL,
    app_version JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at BIGINT DEFAULT 0
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
