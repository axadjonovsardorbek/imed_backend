-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX users_email_unique
ON users(email)
WHERE deleted_at = 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
