DROP TABLE IF EXISTS email_queue;

DROP INDEX IF EXISTS idx_email_queue_status;
DROP INDEX IF EXISTS idx_email_queue_send_to;
DROP INDEX IF EXISTS idx_email_queue_send_at;
DROP INDEX IF EXISTS idx_email_queue_purpose;