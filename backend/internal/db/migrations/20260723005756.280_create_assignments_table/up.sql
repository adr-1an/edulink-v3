CREATE TABLE IF NOT EXISTS assignments (
    id BIGINT PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    referenced_post_id BIGINT REFERENCES course_posts(id) ON DELETE SET NULL,

    title TEXT NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ,
    submissions_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    submissions_close_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)