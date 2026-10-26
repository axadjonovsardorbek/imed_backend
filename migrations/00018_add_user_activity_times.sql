-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS first_login_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS inactive_push_sent_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

-- Mavjud foydalanuvchilar uchun birinchi kirish vaqti sifatida ro'yxatdan o'tgan vaqt olinadi.
UPDATE users SET first_login_at = created_at WHERE first_login_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_last_seen_at ON users (last_seen_at) WHERE deleted_at = 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_last_seen_at;
ALTER TABLE users
  DROP COLUMN IF EXISTS first_login_at,
  DROP COLUMN IF EXISTS last_seen_at,
  DROP COLUMN IF EXISTS inactive_push_sent_at;
-- +goose StatementEnd
