CREATE TABLE IF NOT EXISTS portal_account_activation_tokens (
    token_hash BYTEA PRIMARY KEY,
    portal_user_id BIGINT NOT NULL REFERENCES portal_users ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_portal_account_activation_tokens_portal_user_id
ON portal_account_activation_tokens(portal_user_id);