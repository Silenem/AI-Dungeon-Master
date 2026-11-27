ALTER TABLE users
    DROP COLUMN IF EXISTS email_2fa_enabled,
    DROP COLUMN IF EXISTS email_verified_at;
