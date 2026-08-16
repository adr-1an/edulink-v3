ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS two_factor_status TEXT NOT NULL DEFAULT 'disabled' CHECK (two_factor_status IN ('disabled', 'pending', 'enabled')),
    ADD COLUMN IF NOT EXISTS totp_secret BYTEA;