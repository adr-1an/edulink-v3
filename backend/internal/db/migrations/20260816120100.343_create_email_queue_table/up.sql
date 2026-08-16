CREATE TABLE IF NOT EXISTS email_queue (
    id BIGINT PRIMARY KEY,
    send_to TEXT NOT NULL,

    subject TEXT,
    content_type TEXT NOT NULL DEFAULT 'text/plain',
    body TEXT,
    priority INT NOT NULL DEFAULT 1,

    status TEXT NOT NULL CHECK (status IN ('sent', 'pending', 'failed')) DEFAULT 'pending',
    max_retries INT NOT NULL DEFAULT 1,
    purpose TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    send_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_queue_status
ON email_queue(status);

CREATE INDEX IF NOT EXISTS idx_email_queue_send_to
ON email_queue(send_to);

CREATE INDEX IF NOT EXISTS idx_email_queue_send_at
ON email_queue(send_at);

CREATE INDEX IF NOT EXISTS idx_email_queue_purpose
ON email_queue(purpose);