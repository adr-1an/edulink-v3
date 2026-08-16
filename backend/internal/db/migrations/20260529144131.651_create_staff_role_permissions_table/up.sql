CREATE TABLE IF NOT EXISTS staff_role_permissions (
    role_id BIGINT NOT NULL,
    permission TEXT NOT NULL,

    FOREIGN KEY (role_id) REFERENCES staff_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission)
)