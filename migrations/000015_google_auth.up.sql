DROP TABLE IF EXISTS otp_codes;

DROP INDEX IF EXISTS idx_users_phone;

ALTER TABLE users
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS phone_verified_at;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS google_sub TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_sub ON users (google_sub) WHERE google_sub IS NOT NULL;

