CREATE TABLE IF NOT EXISTS user_profile_pictures (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_object_id BIGINT NOT NULL REFERENCES storage_objects(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)