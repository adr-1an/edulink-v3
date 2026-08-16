CREATE TABLE IF NOT EXISTS grades (
    id BIGINT PRIMARY KEY,
    academic_year_id BIGINT NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,

    level INT NOT NULL,
    name TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (academic_year_id, level, name)
)