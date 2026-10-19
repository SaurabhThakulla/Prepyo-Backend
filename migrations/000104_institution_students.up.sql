-- Students an institution adds. Each one is an ordinary learner (a row in
-- users) who signs in with a username instead of Google.

-- 2 seats is the free plan; Prepyo raises it for a paid one.
ALTER TABLE institutions ADD COLUMN seat_limit INT NOT NULL DEFAULT 2 CHECK (seat_limit >= 0);

CREATE TABLE institution_students (
    user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    institution_id UUID NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
    username       TEXT NOT NULL,
    password_hash  TEXT NOT NULL, -- bcrypt
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A username only has to be unique inside its institution.
CREATE UNIQUE INDEX idx_institution_students_username ON institution_students (institution_id, lower(username));
