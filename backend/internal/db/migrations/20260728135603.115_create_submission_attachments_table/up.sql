CREATE TABLE IF NOT EXISTS submission_attachments (
    id BIGINT PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES assignment_submissions(id) ON DELETE CASCADE,
    storage_object_id BIGINT NOT NULL REFERENCES portal_storage_objects(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)