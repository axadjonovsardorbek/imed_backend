-- +goose Up
-- +goose StatementBegin

ALTER TABLE promocodes
ADD COLUMN type VARCHAR(50) NOT NULL 
    DEFAULT 'all'
    CHECK (type IN ('all', 'selected'));

CREATE TABLE IF NOT EXISTS promocode_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promocode_id UUID NOT NULL REFERENCES promocodes(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(promocode_id, course_id)  -- optional but useful
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
