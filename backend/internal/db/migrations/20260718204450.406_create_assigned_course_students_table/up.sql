CREATE TABLE assigned_course_students (
    PRIMARY KEY (course_id, student_id),

    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)