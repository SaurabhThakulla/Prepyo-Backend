ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone             TEXT,
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL;

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

CREATE TABLE IF NOT EXISTS otp_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone       TEXT NOT NULL,
    code_hash   BYTEA NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    attempts    INT NOT NULL DEFAULT 0,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes (phone, created_at DESC);


-- Only the two accounts kept for testing get a number. Matched on exact email,
-- never a name pattern: two accounts are named Saurabh and a LIKE would set both
-- to one number, which the unique index rejects.

UPDATE users
   SET phone = '+9779876543210', phone_verified_at = now()
 WHERE lower(email) = 'admin@prepyo.com'
   AND phone IS NULL
   AND NOT EXISTS (SELECT 1 FROM users u2 WHERE u2.phone = '+9779876543210');

UPDATE users
   SET phone = '+9779812345678', phone_verified_at = now()
 WHERE lower(email) = 'bistsushant54@gmail.com'
   AND phone IS NULL
   AND NOT EXISTS (SELECT 1 FROM users u2 WHERE u2.phone = '+9779812345678');
