CREATE TABLE IF NOT EXISTS two_factor_recovery_codes (
    recovery_code_hash BYTEA PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)