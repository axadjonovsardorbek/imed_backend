-- +goose Up
-- +goose StatementBegin
ALTER TABLE courses
ADD COLUMN IF NOT EXISTS total_duration  BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_lessons   BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_sources   BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_videos    BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_documents BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_tests     BIGINT NOT NULL DEFAULT 0;

CREATE OR REPLACE FUNCTION get_course_id_by_module(p_module_id uuid)
RETURNS uuid AS $$
DECLARE
  v_course_id uuid;
BEGIN
  SELECT course_id INTO v_course_id
  FROM modules
  WHERE id = p_module_id AND deleted_at = 0;

  RETURN v_course_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_course_id_by_lesson(p_lesson_id uuid)
RETURNS uuid AS $$
DECLARE
  v_course_id uuid;
BEGIN
  SELECT m.course_id INTO v_course_id
  FROM lessons l
  JOIN modules m ON m.id = l.module_id
  WHERE l.id = p_lesson_id
    AND l.deleted_at = 0
    AND m.deleted_at = 0;

  RETURN v_course_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION bump_course_lessons(p_course_id uuid, delta int)
RETURNS void AS $$
BEGIN
  IF p_course_id IS NULL THEN RETURN; END IF;

  UPDATE courses
  SET total_lessons = GREATEST(total_lessons + delta, 0),
      updated_at = NOW()
  WHERE id = p_course_id AND deleted_at = 0;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION bump_course_sources(p_course_id uuid, p_type varchar, delta int)
RETURNS void AS $$
BEGIN
  IF p_course_id IS NULL THEN RETURN; END IF;

  UPDATE courses
  SET
    total_sources = GREATEST(total_sources + delta, 0),
    total_videos = CASE WHEN p_type='video'
      THEN GREATEST(total_videos + delta, 0) ELSE total_videos END,
    total_documents = CASE WHEN p_type='document'
      THEN GREATEST(total_documents + delta, 0) ELSE total_documents END,
    total_tests = CASE WHEN p_type='test'
      THEN GREATEST(total_tests + delta, 0) ELSE total_tests END,
    updated_at = NOW()
  WHERE id = p_course_id AND deleted_at = 0;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION trg_lessons_counter()
RETURNS trigger AS $$
DECLARE
  old_course uuid;
  new_course uuid;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.deleted_at = 0 THEN
      new_course := get_course_id_by_module(NEW.module_id);
      PERFORM bump_course_lessons(new_course, 1);
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF OLD.deleted_at = 0 THEN
      old_course := get_course_id_by_module(OLD.module_id);
      PERFORM bump_course_lessons(old_course, -1);
    END IF;
    RETURN OLD;
  END IF;

  old_course := get_course_id_by_module(OLD.module_id);
  new_course := get_course_id_by_module(NEW.module_id);

  IF OLD.deleted_at = 0 AND NEW.deleted_at <> 0 THEN
    PERFORM bump_course_lessons(old_course, -1);
  ELSIF OLD.deleted_at <> 0 AND NEW.deleted_at = 0 THEN
    PERFORM bump_course_lessons(new_course, 1);
  END IF;

  IF OLD.module_id <> NEW.module_id THEN
    IF OLD.deleted_at = 0 THEN
      PERFORM bump_course_lessons(old_course, -1);
    END IF;
    IF NEW.deleted_at = 0 THEN
      PERFORM bump_course_lessons(new_course, 1);
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS lessons_counter ON lessons;
CREATE TRIGGER lessons_counter
AFTER INSERT OR UPDATE OR DELETE ON lessons
FOR EACH ROW EXECUTE FUNCTION trg_lessons_counter();

CREATE OR REPLACE FUNCTION trg_sources_counter()
RETURNS trigger AS $$
DECLARE
  old_course uuid;
  new_course uuid;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.deleted_at = 0 THEN
      new_course := get_course_id_by_lesson(NEW.lesson_id);
      PERFORM bump_course_sources(new_course, NEW.type, 1);
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF OLD.deleted_at = 0 THEN
      old_course := get_course_id_by_lesson(OLD.lesson_id);
      PERFORM bump_course_sources(old_course, OLD.type, -1);
    END IF;
    RETURN OLD;
  END IF;


  old_course := get_course_id_by_lesson(OLD.lesson_id);
  new_course := get_course_id_by_lesson(NEW.lesson_id);


  IF OLD.deleted_at = 0 AND NEW.deleted_at <> 0 THEN
    PERFORM bump_course_sources(old_course, OLD.type, -1);
  ELSIF OLD.deleted_at <> 0 AND NEW.deleted_at = 0 THEN
    PERFORM bump_course_sources(new_course, NEW.type, 1);
  END IF;


  IF OLD.type <> NEW.type AND NEW.deleted_at = 0 THEN
    PERFORM bump_course_sources(new_course, OLD.type, -1);
    PERFORM bump_course_sources(new_course, NEW.type, 1);
  END IF;


  IF OLD.lesson_id <> NEW.lesson_id THEN
    IF OLD.deleted_at = 0 THEN
      PERFORM bump_course_sources(old_course, OLD.type, -1);
    END IF;
    IF NEW.deleted_at = 0 THEN
      PERFORM bump_course_sources(new_course, NEW.type, 1);
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS sources_counter ON sources;
CREATE TRIGGER sources_counter
AFTER INSERT OR UPDATE OR DELETE ON sources
FOR EACH ROW EXECUTE FUNCTION trg_sources_counter();

CREATE OR REPLACE FUNCTION bump_course_duration(p_course_id uuid, delta bigint)
RETURNS void AS $$
BEGIN
  IF p_course_id IS NULL THEN RETURN; END IF;

  UPDATE courses
  SET total_duration = GREATEST(total_duration + delta, 0),
      updated_at = NOW()
  WHERE id = p_course_id AND deleted_at = 0;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_lessons_counter()
RETURNS trigger AS $$
DECLARE
  old_course uuid;
  new_course uuid;

  old_dur bigint;
  new_dur bigint;
BEGIN
  old_dur := COALESCE(OLD.duration, 0);
  new_dur := COALESCE(NEW.duration, 0);

  IF TG_OP = 'INSERT' THEN
    IF NEW.deleted_at = 0 THEN
      new_course := get_course_id_by_module(NEW.module_id);
      PERFORM bump_course_lessons(new_course, 1);
      PERFORM bump_course_duration(new_course, new_dur);
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF OLD.deleted_at = 0 THEN
      old_course := get_course_id_by_module(OLD.module_id);
      PERFORM bump_course_lessons(old_course, -1);
      PERFORM bump_course_duration(old_course, -old_dur);
    END IF;
    RETURN OLD;
  END IF;

  -- UPDATE
  old_course := get_course_id_by_module(OLD.module_id);
  new_course := get_course_id_by_module(NEW.module_id);

  -- 1) soft delete / restore
  IF OLD.deleted_at = 0 AND NEW.deleted_at <> 0 THEN
    PERFORM bump_course_lessons(old_course, -1);
    PERFORM bump_course_duration(old_course, -old_dur);

  ELSIF OLD.deleted_at <> 0 AND NEW.deleted_at = 0 THEN
    PERFORM bump_course_lessons(new_course, 1);
    PERFORM bump_course_duration(new_course, new_dur);
  END IF;

  -- 2) duration change (faqat active bo‘lsa va module o‘zgarmagan bo‘lsa)
  IF OLD.deleted_at = 0 AND NEW.deleted_at = 0
     AND OLD.module_id = NEW.module_id
     AND old_dur <> new_dur THEN
    PERFORM bump_course_duration(new_course, (new_dur - old_dur));
  END IF;

  -- 3) module_id change (course almashishi mumkin)
  IF OLD.module_id <> NEW.module_id THEN
    -- eski course’dan ayiramiz (agar eski row active bo‘lsa)
    IF OLD.deleted_at = 0 THEN
      PERFORM bump_course_lessons(old_course, -1);
      PERFORM bump_course_duration(old_course, -old_dur);
    END IF;

    -- yangi course’ga qo‘shamiz (agar yangi row active bo‘lsa)
    IF NEW.deleted_at = 0 THEN
      PERFORM bump_course_lessons(new_course, 1);
      PERFORM bump_course_duration(new_course, new_dur);
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS lessons_counter ON lessons;
CREATE TRIGGER lessons_counter
AFTER INSERT OR UPDATE OR DELETE ON lessons
FOR EACH ROW EXECUTE FUNCTION trg_lessons_counter();


CREATE INDEX IF NOT EXISTS idx_modules_course ON modules(course_id) WHERE deleted_at=0;
CREATE INDEX IF NOT EXISTS idx_lessons_module ON lessons(module_id) WHERE deleted_at=0;
CREATE INDEX IF NOT EXISTS idx_sources_lesson ON sources(lesson_id) WHERE deleted_at=0;



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
