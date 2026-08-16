CREATE TABLE IF NOT EXISTS portal_users (
    id BIGINT PRIMARY KEY,
    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,
    notes TEXT,

    account_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    password_hash TEXT,
    account_type TEXT NOT NULL CHECK (account_type IN ('guardian', 'student')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);