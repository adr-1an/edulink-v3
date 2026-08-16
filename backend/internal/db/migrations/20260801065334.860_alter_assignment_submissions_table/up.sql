-- Drop (assignment_id, submitted_by) constraint
ALTER TABLE IF EXISTS assignment_submissions
    DROP CONSTRAINT IF EXISTS assignment_submissions_assignment_id_submitted_by_key;

-- One non-returned submission per assignment per student
CREATE UNIQUE INDEX IF NOT EXISTS assignment_submissions_one_active_per_user
ON assignment_submissions (assignment_id, submitted_by)
WHERE status <> 'returned';

-- Drop old returned col
ALTER TABLE IF EXISTS assignment_submissions
    DROP COLUMN IF EXISTS returned,
    DROP COLUMN IF EXISTS return_msg;