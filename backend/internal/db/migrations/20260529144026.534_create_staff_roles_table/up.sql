CREATE TABLE IF NOT EXISTS staff_roles (
    id BIGINT PRIMARY KEY,
    position INT NOT NULL CHECK (position >= 0),

    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    color TEXT CHECK (length(color) = 6),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (school_id, position),
    UNIQUE (school_id, name)
)