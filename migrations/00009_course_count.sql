-- +goose Up
-- +goose StatementBegin

ALTER TABLE subjects
ADD COLUMN IF NOT EXISTS total_courses BIGINT NOT NULL DEFAULT 0;

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email VARCHAR(255) DEFAULT NULL;

CREATE OR REPLACE FUNCTION trg_courses_update_subject_total()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.subject_id IS NOT NULL AND NEW.deleted_at = 0 THEN
      UPDATE subjects
      SET total_courses = total_courses + 1,
          updated_at = CURRENT_TIMESTAMP
      WHERE id = NEW.subject_id;
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF OLD.subject_id IS NOT NULL AND OLD.deleted_at = 0 THEN
      UPDATE subjects
      SET total_courses = GREATEST(total_courses - 1, 0),
          updated_at = CURRENT_TIMESTAMP
      WHERE id = OLD.subject_id;
    END IF;
    RETURN OLD;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    IF OLD.subject_id IS NOT NULL AND OLD.deleted_at = 0 THEN
      IF (NEW.subject_id IS DISTINCT FROM OLD.subject_id) OR (NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0) THEN
        UPDATE subjects
        SET total_courses = GREATEST(total_courses - 1, 0),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = OLD.subject_id;
      END IF;
    END IF;

    IF NEW.subject_id IS NOT NULL AND NEW.deleted_at = 0 THEN
      IF (NEW.subject_id IS DISTINCT FROM OLD.subject_id) OR (NEW.deleted_at <> OLD.deleted_at AND OLD.deleted_at <> 0) THEN
        UPDATE subjects
        SET total_courses = total_courses + 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.subject_id;
      END IF;
    END IF;

    RETURN NEW;
  END IF;

  RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS courses_subject_total_trg ON courses;

CREATE TRIGGER courses_subject_total_trg
AFTER INSERT OR UPDATE OR DELETE ON courses
FOR EACH ROW
EXECUTE FUNCTION trg_courses_update_subject_total();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
