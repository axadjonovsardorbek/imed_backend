-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION prevent_order_insert_if_course_active()
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
      AND uc.started_at <= now()
      AND (uc.ended_at IS NULL OR uc.ended_at > now())
  ) THEN
    RAISE EXCEPTION 'Course is still active for this user (cannot create order until ended_at passes)';
  END IF;

  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_prevent_order_insert_if_course_active ON orders;

CREATE TRIGGER trg_prevent_order_insert_if_course_active
BEFORE INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION prevent_order_insert_if_course_active();

CREATE OR REPLACE FUNCTION prevent_order_for_public_course()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM courses c
    WHERE c.id = NEW.course_id
      AND c.is_public = TRUE
      AND c.deleted_at = 0
  ) THEN
    RAISE EXCEPTION 'Cannot create order for public course';
  END IF;

  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_prevent_order_for_public_course ON orders;

CREATE TRIGGER trg_prevent_order_for_public_course
BEFORE INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION prevent_order_for_public_course();


CREATE OR REPLACE FUNCTION prevent_user_course_for_public_course()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM courses c
    WHERE c.id = NEW.course_id
      AND c.is_public = TRUE
      AND c.deleted_at = 0
  ) THEN
    RAISE EXCEPTION 'Cannot assign public course to user';
  END IF;

  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_prevent_user_course_for_public_course ON user_courses;

CREATE TRIGGER trg_prevent_user_course_for_public_course
BEFORE INSERT ON user_courses
FOR EACH ROW
EXECUTE FUNCTION prevent_user_course_for_public_course();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
 
-- +goose StatementEnd
