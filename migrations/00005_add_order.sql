-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
    CREATE TYPE payment_status AS ENUM ('NEW','RESERVED','PAID','CANCELLED','EXPIRED');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL REFERENCES users(id),
  course_id UUID NOT NULL REFERENCES courses(id),
  tariff_id UUID NOT NULL REFERENCES tariffs(id),

  duration INT NOT NULL CHECK (duration > 0),

  amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),

  status payment_status NOT NULL DEFAULT 'NEW',

  click_trans_id BIGINT,
  click_paydoc_id BIGINT,
  merchant_prepare_id BIGINT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  paid_at TIMESTAMPTZ,

  meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_orders_click_trans
ON orders(click_trans_id) WHERE click_trans_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_user_course ON orders(user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
