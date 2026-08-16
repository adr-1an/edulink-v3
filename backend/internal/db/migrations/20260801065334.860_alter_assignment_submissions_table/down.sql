DROP INDEX IF EXISTS assignment_submissions_one_active_per_user;

ALTER TABLE assignment_submissions
    ADD CONSTRAINT assignment_submissions_assignment_id_submitted_by_key
        UNIQUE (assignment_id, submitted_by);

ALTER TABLE assignment_submissions
    ADD COLUMN IF NOT EXISTS returned BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS return_msg TEXT;