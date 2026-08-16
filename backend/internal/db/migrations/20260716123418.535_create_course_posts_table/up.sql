CREATE TABLE IF NOT EXISTS course_posts (
    id BIGINT PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE SET DEFAULT DEFAULT 0,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,

    title TEXT NOT NULL,
    body TEXT NOT NULL,
    accent_color TEXT CHECK (length(accent_color) = 6),

    show_until TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMPTZ
)