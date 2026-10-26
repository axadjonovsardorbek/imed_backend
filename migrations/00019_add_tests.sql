-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id),
    title JSONB NOT NULL,
    time_limit_sec INT DEFAULT NULL,
    order_num BIGINT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS test_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES tests(id),
    text TEXT NOT NULL,
    image_url TEXT DEFAULT NULL,
    order_num BIGINT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS test_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES test_questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    order_num BIGINT DEFAULT 1
);

CREATE TABLE IF NOT EXISTS user_test_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    test_id UUID NOT NULL REFERENCES tests(id),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    correct_count INT DEFAULT 0,
    total_count INT DEFAULT 0,
    percent NUMERIC(5,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_test_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES user_test_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES test_questions(id),
    option_id UUID NOT NULL REFERENCES test_options(id),
    is_correct BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (attempt_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_tests_lesson ON tests (lesson_id) WHERE deleted_at = 0;
CREATE INDEX IF NOT EXISTS idx_test_questions_test ON test_questions (test_id) WHERE deleted_at = 0;
CREATE INDEX IF NOT EXISTS idx_test_options_question ON test_options (question_id);
CREATE INDEX IF NOT EXISTS idx_user_test_attempts_user ON user_test_attempts (user_id, test_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_test_answers;
DROP TABLE IF EXISTS user_test_attempts;
DROP TABLE IF EXISTS test_options;
DROP TABLE IF EXISTS test_questions;
DROP TABLE IF EXISTS tests;
-- +goose StatementEnd
