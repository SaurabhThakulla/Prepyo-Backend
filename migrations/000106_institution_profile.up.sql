-- What an institution answers in its setup after sign-up, and its logo.

ALTER TABLE institutions
    -- Exams it prepares students for: IELTS, PTE, EPS_TOPIK.
    ADD COLUMN exams            TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN city             TEXT NOT NULL DEFAULT '',
    ADD COLUMN branches         INT NOT NULL DEFAULT 1 CHECK (branches BETWEEN 1 AND 50),
    -- A range such as '25-50': students it prepares in a month.
    ADD COLUMN monthly_students TEXT NOT NULL DEFAULT '',
    ADD COLUMN brand_color      TEXT NOT NULL DEFAULT '' CHECK (brand_color = '' OR brand_color ~ '^#[0-9a-f]{6}$'),
    -- NULL until the setup is finished; the dashboard stays locked until then.
    ADD COLUMN onboarded_at     TIMESTAMPTZ;

CREATE TABLE institution_logos (
    institution_id UUID PRIMARY KEY REFERENCES institutions(id) ON DELETE CASCADE,
    content_type   TEXT NOT NULL,
    data           BYTEA NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Institution students never see the learner onboarding.
UPDATE users SET onboarding_completed_at = now()
WHERE onboarding_completed_at IS NULL
  AND id IN (SELECT user_id FROM institution_students);
