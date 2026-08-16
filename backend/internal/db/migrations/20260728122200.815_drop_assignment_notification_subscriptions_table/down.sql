CREATE TABLE IF NOT EXISTS assignment_notification_subscriptions (
    PRIMARY KEY (assignment_id, portal_user_id),
    assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    portal_user_id BIGINT NOT NULL REFERENCES portal_users(id) ON DELETE CASCADE,
    email BOOL NOT NULL,
    in_app BOOL NOT NULL
)