-- +goose Up
-- +goose StatementBegin

DO $$ BEGIN
  CREATE TYPE promo_discount_type AS ENUM ('percent', 'fixed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS promocodes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              TEXT NOT NULL,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    discount_type     promo_discount_type NOT NULL,
    discount_value    NUMERIC(12,2) NOT NULL CHECK (discount_value > 0),
    CHECK (
      (discount_type = 'percent' AND discount_value <= 100)
      OR (discount_type = 'fixed')
    ),
    starts_at         TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    ends_at           TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    min_order_amount  NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (min_order_amount >= 0),
    max_discount      NUMERIC(12,2),
    max_uses_total    INT,
    max_uses_per_user INT NOT NULL DEFAULT 1,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at        BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_promo_codes_code
ON promocodes (lower(code))
WHERE deleted_at = 0;

CREATE TABLE IF NOT EXISTS promo_redemptions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  promo_id        UUID NOT NULL REFERENCES promocodes(id),
  user_id         UUID NOT NULL REFERENCES users(id),
  order_id        UUID NOT NULL REFERENCES orders(id),
  order_amount    NUMERIC(12,2) NOT NULL,
  discount_amount NUMERIC(12,2) NOT NULL,
  course_amount   NUMERIC(12,2) NOT NULL,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  deactivated_at  TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 1 orderga 1 ta aktiv promo ishlasin
CREATE UNIQUE INDEX IF NOT EXISTS ux_promo_redemptions_order_active
ON promo_redemptions(order_id)
WHERE is_active = TRUE;

-- total limit tekshiruv tez bo‘lsin
CREATE INDEX IF NOT EXISTS ix_redemptions_promo_active
ON promo_redemptions(promo_id)
WHERE is_active = TRUE;

-- user limit tekshiruv tez bo‘lsin
CREATE INDEX IF NOT EXISTS ix_redemptions_user_promo_active
ON promo_redemptions(user_id, promo_id)
WHERE is_active = TRUE;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_type TEXT NOT NULL DEFAULT 'click',
    ADD COLUMN IF NOT EXISTS promo_id UUID REFERENCES promocodes(id) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS course_amount NUMERIC(12,2) NOT NULL DEFAULT 0;



CREATE OR REPLACE FUNCTION trg_orders_deactivate_redemption_on_cancel_expire()
RETURNS trigger AS $$
BEGIN
  -- status o'zgarmasa hech narsa qilmaymiz
  IF TG_OP = 'UPDATE' AND OLD.status = NEW.status THEN
    RETURN NEW;
  END IF;

  -- faqat CANCELLED/EXPIRED bo'lsa
  IF NEW.status IN ('CANCELLED', 'EXPIRED') THEN
    UPDATE promo_redemptions
    SET is_active = FALSE,
        deactivated_at = now()
    WHERE order_id = NEW.id
      AND is_active = TRUE;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS tr_orders_deactivate_redemption ON orders;

CREATE TRIGGER tr_orders_deactivate_redemption
AFTER UPDATE OF status ON orders
FOR EACH ROW
EXECUTE FUNCTION trg_orders_deactivate_redemption_on_cancel_expire();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
