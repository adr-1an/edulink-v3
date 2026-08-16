-- This is what the table looks like after the migration.
CREATE TABLE IF NOT EXISTS assignment_submissions (
    id BIGINT PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    submitted_by BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,

    status TEXT NOT NULL CHECK (status IN ('pending', 'submitted')) DEFAULT 'pending',
    notes TEXT,

    returned BOOL NOT NULL DEFAULT FALSE, -- ADDED
    return_msg TEXT, -- ADDED

    UNIQUE (assignment_id, submitted_by),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
-- This file is not used by the migration tool.