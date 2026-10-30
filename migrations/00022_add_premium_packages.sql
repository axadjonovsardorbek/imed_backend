-- +goose Up
-- +goose StatementBegin
-- Premium 3 paketga bo'lindi (T-11, T-12): fan qaysi paketga tegishli ekanini admin belgilaydi.
ALTER TABLE subjects
  ADD COLUMN IF NOT EXISTS package VARCHAR(20) DEFAULT NULL
  CHECK (package IN ('fundamental', 'preclinical', 'clinical'));

-- Eski umumiy premium obunalar 'all' bo'lib qoladi va muddati tugaguncha hamma kursni ochadi.
ALTER TABLE user_subscriptions
  ADD COLUMN IF NOT EXISTS package VARCHAR(20) NOT NULL DEFAULT 'all'
  CHECK (package IN ('all', 'fundamental', 'preclinical', 'clinical'));

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS package VARCHAR(20) DEFAULT NULL;

DROP INDEX IF EXISTS idx_user_subscriptions_user;
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user ON user_subscriptions (user_id, package, ended_at);

INSERT INTO settings (key, value) VALUES
  ('premium:fundamental', '{"price": 0, "duration_months": 1}'),
  ('premium:preclinical', '{"price": 0, "duration_months": 1}'),
  ('premium:clinical',    '{"price": 0, "duration_months": 1}')
ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM settings WHERE key IN ('premium:fundamental', 'premium:preclinical', 'premium:clinical');

DROP INDEX IF EXISTS idx_user_subscriptions_user;
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user ON user_subscriptions (user_id, ended_at);

ALTER TABLE orders DROP COLUMN IF EXISTS package;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS package;
ALTER TABLE subjects DROP COLUMN IF EXISTS package;
-- +goose StatementEnd
