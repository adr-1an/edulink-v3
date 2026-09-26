CREATE TABLE IF NOT EXISTS guardian_students (
    PRIMARY KEY (guardian_id, student_id),

    student_id  BIGINT NOT NULL REFERENCES portal_users (id) ON DELETE CASCADE,
    guardian_id BIGINT NOT NULL REFERENCES portal_users (id) ON DELETE CASCADE,

    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE SET DEFAULT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)