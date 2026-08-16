CREATE TABLE IF NOT EXISTS academic_years (
    id BIGINT PRIMARY KEY,
    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,

    start_year INT NOT NULL,
    end_year INT NOT NULL,

    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '1 year',

    is_active BOOLEAN NOT NULL DEFAULT false,

    UNIQUE (school_id, start_year, end_year),
    CHECK (end_year >= start_year)
);

CREATE UNIQUE INDEX one_active_year_per_school
ON academic_years (school_id)
WHERE is_active = true;