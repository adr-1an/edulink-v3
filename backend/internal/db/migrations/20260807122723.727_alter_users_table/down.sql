ALTER TABLE IF EXISTS users
    DROP COLUMN IF EXISTS two_factor_enabled,
    DROP COLUMN IF EXISTS totp_secret;