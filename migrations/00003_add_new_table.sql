-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_course_progress (
  user_id UUID NOT NULL,
  course_id UUID NOT NULL,
  completed_lessons BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, course_id)
);


CREATE OR REPLACE FUNCTION trg_user_lessons_progress_insert()
RETURNS trigger AS $$
DECLARE
  v_course_id uuid;
BEGIN
  v_course_id := get_course_id_by_lesson(NEW.lesson_id); -- oldingi funksiyang bor

  IF v_course_id IS NULL THEN
    RETURN NEW;
  END IF;

  INSERT INTO user_course_progress(user_id, course_id, completed_lessons, updated_at)
  VALUES (NEW.user_id, v_course_id, 1, NOW())
  ON CONFLICT (user_id, course_id)
  DO UPDATE SET
    completed_lessons = user_course_progress.completed_lessons + 1,
    updated_at = NOW();

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS user_lessons_progress_ins ON user_lessons;
CREATE TRIGGER user_lessons_progress_ins
AFTER INSERT ON user_lessons
FOR EACH ROW EXECUTE FUNCTION trg_user_lessons_progress_insert();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
