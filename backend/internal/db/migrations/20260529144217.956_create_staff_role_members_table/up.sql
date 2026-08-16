CREATE TABLE IF NOT EXISTS staff_role_members (
    staff_id BIGINT NOT NULL REFERENCES school_staff(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES staff_roles(id) ON DELETE CASCADE,

    PRIMARY KEY (staff_id, role_id)
)