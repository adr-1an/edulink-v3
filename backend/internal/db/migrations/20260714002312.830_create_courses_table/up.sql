CREATE TABLE IF NOT EXISTS courses (
    id BIGINT PRIMARY KEY,
    grade_id BIGINT NOT NULL REFERENCES grades(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    color TEXT CHECK (length(color) = 6),
    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (grade_id, name)
);