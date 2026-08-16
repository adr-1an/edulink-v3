CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,

    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,

    public_profile BOOL NOT NULL DEFAULT FALSE,
    staff_invitations_disabled BOOL NOT NULL DEFAULT FALSE,

    password_hash TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)