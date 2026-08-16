CREATE TABLE IF NOT EXISTS portal_sessions (
    token_hash BYTEA PRIMARY KEY,
    portal_user_id BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,

    x_forwarded_for INET,
    source_ip INET,
    user_agent TEXT,

    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)