CREATE TABLE IF NOT EXISTS staff_invitations (
    id BIGINT PRIMARY KEY,
    school_id BIGINT NOT NULL REFERENCES schools(id),

    token_hash BYTEA NOT NULL UNIQUE,
    user_email TEXT NOT NULL,

    sent_by_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'rejected')) DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days'
)