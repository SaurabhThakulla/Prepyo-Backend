-- Institutions (consultancies) and their admin login. Kept apart from the
-- learner tables; all the code that reads them is in internal/institution.

CREATE TABLE institutions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    -- The institution's address on the site: /<code>.
    code        TEXT NOT NULL UNIQUE
                    CHECK (code ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(code) BETWEEN 3 AND 40),
    phone       TEXT NOT NULL,
    -- End of the paid period; NULL until Prepyo activates the institution.
    valid_until DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One admin per institution.
CREATE TABLE institution_admins (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL UNIQUE REFERENCES institutions(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    email          TEXT NOT NULL UNIQUE, -- stored in lower case
    password_hash  TEXT NOT NULL,        -- bcrypt
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE institution_sessions (
    token_hash BYTEA PRIMARY KEY, -- SHA-256 of the cookie token, never the token
    admin_id   UUID NOT NULL REFERENCES institution_admins(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_institution_sessions_admin ON institution_sessions(admin_id);
