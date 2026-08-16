CREATE TABLE IF NOT EXISTS students (
    id BIGINT PRIMARY KEY,
    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE,
    email TEXT,
    phone TEXT,
    notes TEXT,

    account_enabled BOOL NOT NULL DEFAULT FALSE,
    password_hash TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,

    UNIQUE (email, school_id)
)