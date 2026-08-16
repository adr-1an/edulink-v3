CREATE TABLE IF NOT EXISTS school_logs (
    id BIGINT PRIMARY KEY,
    school_id BIGINT REFERENCES schools(id) ON DELETE CASCADE,
    by_user BIGINT REFERENCES users(id) ON DELETE SET DEFAULT DEFAULT 0,

    type TEXT NOT NULL CHECK (type IN ('create', 'edit', 'delete', 'other')),
    action TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    details TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
