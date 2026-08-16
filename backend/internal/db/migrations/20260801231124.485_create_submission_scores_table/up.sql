CREATE TABLE IF NOT EXISTS submission_scores (
    submission_id BIGINT PRIMARY KEY REFERENCES assignment_submissions(id) ON DELETE CASCADE,
    graded_by BIGINT NOT NULL REFERENCES users(id) ON DELETE SET DEFAULT DEFAULT 0,
    score_percentage SMALLINT NOT NULL CHECK (score_percentage >= 0 AND score_percentage <= 100),

    notes TEXT,

    graded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)