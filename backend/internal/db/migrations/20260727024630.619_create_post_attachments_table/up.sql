CREATE TABLE IF NOT EXISTS post_attachments (
    id BIGINT PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES course_posts(id) ON DELETE CASCADE,
    storage_object_id BIGINT NOT NULL REFERENCES storage_objects(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)