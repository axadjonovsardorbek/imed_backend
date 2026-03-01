-- +goose Up
-- +goose StatementBegin
ALTER TABLE courses
ADD COLUMN IF NOT EXISTS can_buy BOOLEAN DEFAULT FALSE;

ALTER TABLE app_routes
ADD COLUMN IF NOT EXISTS buy_course BOOLEAN DEFAULT FALSE;

-- 1) Function
CREATE OR REPLACE FUNCTION sync_courses_can_buy_from_app_routes()
RETURNS trigger AS $$
BEGIN
  -- faqat buy_course o'zgargan va app_routes active bo'lsa
  IF NEW.deleted_at = 0 
     AND NEW.buy_course IS DISTINCT FROM OLD.buy_course THEN

    -- true -> false
    IF OLD.buy_course = TRUE AND NEW.buy_course = FALSE THEN
      UPDATE courses
      SET can_buy = FALSE,
          updated_at = now()
      WHERE deleted_at = 0
        AND is_public = FALSE;

    -- false -> true
    ELSIF OLD.buy_course = FALSE AND NEW.buy_course = TRUE THEN
      UPDATE courses
      SET can_buy = TRUE,
          updated_at = now()
      WHERE deleted_at = 0
        AND is_public = FALSE;

    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2) Trigger
DROP TRIGGER IF EXISTS trg_sync_courses_can_buy ON app_routes;

CREATE TRIGGER trg_sync_courses_can_buy
AFTER UPDATE OF buy_course ON app_routes
FOR EACH ROW
WHEN (
  NEW.deleted_at = 0
  AND OLD.buy_course IS DISTINCT FROM NEW.buy_course
)
EXECUTE FUNCTION sync_courses_can_buy_from_app_routes();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
