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

CREATE OR REPLACE FUNCTION prevent_user_course_insert_before_end()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM user_courses uc
    WHERE uc.user_id = NEW.user_id
      AND uc.course_id = NEW.course_id
      AND uc.deleted_at = 0
      AND COALESCE(uc.ended_at, 'infinity'::timestamptz) > NOW()
  ) THEN
    RAISE EXCEPTION 'Course is still active for this user (cannot repurchase until ended_at passes)';
  END IF;

  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_prevent_user_course_insert_before_end ON user_courses;

CREATE TRIGGER trg_prevent_user_course_insert_before_end
BEFORE INSERT ON user_courses
FOR EACH ROW
EXECUTE FUNCTION prevent_user_course_insert_before_end();

CREATE OR REPLACE FUNCTION prevent_order_if_usercourse_active()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM user_courses uc
    WHERE uc.user_id   = NEW.user_id
      AND uc.course_id = NEW.course_id
      AND uc.deleted_at = 0
      AND (
        uc.ended_at IS NULL
        OR uc.ended_at > NOW()
      )
  ) THEN
    RAISE EXCEPTION
      'cannot create order: user already has an active course (ended_at not passed)';
  END IF;

  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_prevent_order_if_usercourse_active ON orders;

CREATE TRIGGER trg_prevent_order_if_usercourse_active
BEFORE INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION prevent_order_if_usercourse_active();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
