-- +goose Up
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_prevent_order_if_usercourse_active ON orders;
DROP FUNCTION IF EXISTS prevent_order_if_usercourse_active();

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
      AND (uc.ended_at IS NULL OR uc.ended_at > NOW())
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


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
