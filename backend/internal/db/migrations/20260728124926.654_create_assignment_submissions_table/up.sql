CREATE TABLE IF NOT EXISTS assignment_submissions (
    id BIGINT PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    submitted_by BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,

    status TEXT NOT NULL CHECK (status IN ('pending', 'submitted')) DEFAULT 'pending',
    notes TEXT,

    UNIQUE (assignment_id, submitted_by),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)