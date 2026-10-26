-- +goose Up
-- +goose StatementBegin
ALTER TABLE sources DROP CONSTRAINT IF EXISTS sources_type_check;
ALTER TABLE sources ADD CONSTRAINT sources_type_check CHECK (type IN ('video', 'document', 'test', 'audio'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sources DROP CONSTRAINT IF EXISTS sources_type_check;
ALTER TABLE sources ADD CONSTRAINT sources_type_check CHECK (type IN ('video', 'document', 'test'));
-- +goose StatementEnd
