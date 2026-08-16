CREATE TABLE IF NOT EXISTS snowflake_machine_ids (
    machine_id INT PRIMARY KEY,
    owner_id TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
)