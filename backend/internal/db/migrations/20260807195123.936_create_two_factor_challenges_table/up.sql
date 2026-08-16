CREATE TABLE IF NOT EXISTS two_factor_challenges (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    purpose TEXT NOT NULL,
    token_hash BYTEA NOT NULL UNIQUE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '15 MINUTES',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_two_factor_challenges_expires_at
ON two_factor_challenges(expires_at);