CREATE TABLE assigned_course_students (
    PRIMARY KEY (course_id, portal_user_id),

    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    portal_user_id BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)