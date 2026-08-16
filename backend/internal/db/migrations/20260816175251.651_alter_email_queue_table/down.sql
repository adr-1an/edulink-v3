ALTER TABLE IF EXISTS email_queue
DROP COLUMN IF EXISTS last_retry_at;

ALTER TABLE IF EXISTS email_queue
    ALTER COLUMN subject DROP NOT NULL,
    ALTER COLUMN body DROP NOT NULL;

ALTER TABLE IF EXISTS email_queue
    DROP COLUMN IF EXISTS retries;

ALTER TABLE IF EXISTS email_queue
DROP CONSTRAINT IF EXISTS email_queue_status_check;

ALTER TABLE IF EXISTS email_queue
ADD CONSTRAINT email_queue_status_check
CHECK (status IN ('sent', 'pending', 'failed'));