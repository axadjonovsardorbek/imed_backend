-- +goose Up
-- +goose StatementBegin

ALTER TABLE promocodes 
ADD COLUMN type VARCHAR(50) NOT NULL CHECK (type IN ('all', 'selected')) DEFAULT "all";

CREATE TABLE IF NOT EXISTS promocode_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promocode_id UUID REFERENCES promocodes(id),
    course_id UUID REFERENCES courses(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
