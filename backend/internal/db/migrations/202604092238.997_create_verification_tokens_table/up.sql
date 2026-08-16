CREATE TABLE IF NOT EXISTS verification_tokens (
    token_hash BYTEA PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose TEXT CHECK (purpose IN ('email_change', 'password_reset')),
    email_change_new_email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)