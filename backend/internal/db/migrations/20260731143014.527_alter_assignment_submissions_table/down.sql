ALTER TABLE IF EXISTS assignment_submissions
    DROP COLUMN IF EXISTS returned,
    DROP COLUMN IF EXISTS return_msg;