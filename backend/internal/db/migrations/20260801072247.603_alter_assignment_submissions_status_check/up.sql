ALTER TABLE assignment_submissions
    DROP CONSTRAINT assignment_submissions_status_check;

ALTER TABLE assignment_submissions
    ADD CONSTRAINT assignment_submissions_status_check
        CHECK (status IN ('pending', 'submitted', 'returned'));
