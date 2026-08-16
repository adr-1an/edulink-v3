CREATE TABLE IF NOT EXISTS portal_storage_objects (
    id BIGINT PRIMARY KEY,
    uploaded_by BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,

    bucket_name TEXT NOT NULL,
    object_key TEXT NOT NULL,
    original_file_name TEXT NOT NULL,
    declared_file_size BIGINT NOT NULL,
    declared_content_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'done', 'failed')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    completion_token BYTEA NOT NULL,

    UNIQUE (bucket_name, object_key)
)