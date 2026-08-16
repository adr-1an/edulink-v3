ALTER TABLE IF EXISTS email_queue
ADD COLUMN IF NOT EXISTS last_retry_at TIMESTAMPTZ;

ALTER TABLE IF EXISTS email_queue
    ALTER COLUMN subject SET NOT NULL,
    ALTER COLUMN body SET NOT NULL;

ALTER TABLE IF EXISTS email_queue
    ADD COLUMN IF NOT EXISTS retries INT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS email_queue
DROP CONSTRAINT IF EXISTS email_queue_status_check;

ALTER TABLE IF EXISTS email_queue
ADD CONSTRAINT email_queue_status_check
CHECK (status IN ('sent', 'pending', 'processing', 'failed'));