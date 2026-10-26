-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Narx admin paneldan qo'yiladi; 0 bo'lsa premium sotib olib bo'lmaydi.
INSERT INTO settings (key, value)
VALUES ('premium', '{"price": 0, "duration_months": 1}')
ON CONFLICT (key) DO NOTHING;

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'course',
  ALTER COLUMN course_id DROP NOT NULL,
  ALTER COLUMN tariff_id DROP NOT NULL;

ALTER TABLE orders
  ADD CONSTRAINT orders_type_check CHECK (
    type = 'premium' OR (type = 'course' AND course_id IS NOT NULL AND tariff_id IS NOT NULL)
  );

CREATE TABLE IF NOT EXISTS user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    order_id UUID UNIQUE REFERENCES orders(id),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user ON user_subscriptions (user_id, ended_at);

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_ustoz BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS ustoz_granted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN IF EXISTS is_ustoz,
  DROP COLUMN IF EXISTS ustoz_granted_at;

DROP TABLE IF EXISTS user_subscriptions;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_type_check;
-- Premium buyurtmalar bo'lsa NOT NULL qaytmaydi; avval ular qo'lda tozalanishi kerak.
ALTER TABLE orders
  ALTER COLUMN course_id SET NOT NULL,
  ALTER COLUMN tariff_id SET NOT NULL,
  DROP COLUMN IF EXISTS type;

DROP TABLE IF EXISTS settings;
-- +goose StatementEnd
