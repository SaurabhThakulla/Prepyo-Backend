-- 000082_mock_papers.up.sql
-- Numbered mock tests (Test 1, 2, 3...) for every section mock, PTE and IELTS.

CREATE TABLE mock_papers (
    id             TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    exam           TEXT NOT NULL,
    section        TEXT NOT NULL,
    module         TEXT NOT NULL DEFAULT 'any',
    number         INT  NOT NULL,
    revision       INT  NOT NULL DEFAULT 1,
    supersedes_id  TEXT REFERENCES mock_papers(id) ON DELETE SET NULL,
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'retired')),
    title          TEXT NOT NULL DEFAULT '',
    content_schema INT  NOT NULL DEFAULT 1,
    content        JSONB NOT NULL,
    content_hash   TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ,
    retired_at     TIMESTAMPTZ,
    CONSTRAINT uq_mock_papers_scope_rev UNIQUE (exam, section, module, number, revision),
    CONSTRAINT uq_mock_papers_scope_hash UNIQUE (exam, section, module, content_hash)
);

CREATE UNIQUE INDEX idx_mock_papers_scope_published
    ON mock_papers(exam, section, module, number)
    WHERE status = 'published';

CREATE INDEX idx_mock_papers_list
    ON mock_papers(exam, section, module, number ASC);

CREATE OR REPLACE FUNCTION check_mock_paper_immutability()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status = 'published' THEN
        -- The only allowed change to a published row is setting status = 'retired' and retired_at.
        IF NEW.status = 'retired' THEN
            IF NEW.id <> OLD.id OR
               NEW.exam <> OLD.exam OR
               NEW.section <> OLD.section OR
               NEW.module <> OLD.module OR
               NEW.number <> OLD.number OR
               NEW.revision <> OLD.revision OR
               NEW.supersedes_id IS DISTINCT FROM OLD.supersedes_id OR
               NEW.title <> OLD.title OR
               NEW.content_schema <> OLD.content_schema OR
               NEW.content <> OLD.content OR
               NEW.content_hash <> OLD.content_hash OR
               NEW.created_at <> OLD.created_at OR
               NEW.published_at IS DISTINCT FROM OLD.published_at THEN
                RAISE EXCEPTION 'published mock papers are immutable; only status=retired and retired_at may change';
            END IF;
        ELSE
            RAISE EXCEPTION 'published mock papers cannot be modified; publish a new revision instead';
        END IF;
    ELSIF OLD.status = 'retired' THEN
        RAISE EXCEPTION 'retired mock papers are immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mock_papers_immutable
BEFORE UPDATE ON mock_papers
FOR EACH ROW
EXECUTE FUNCTION check_mock_paper_immutability();

CREATE OR REPLACE FUNCTION check_mock_paper_no_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IN ('published', 'retired') THEN
        RAISE EXCEPTION 'cannot delete published or retired mock papers';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mock_papers_no_delete
BEFORE DELETE ON mock_papers
FOR EACH ROW
EXECUTE FUNCTION check_mock_paper_no_delete();

-- Counter table for gapless permanent numbers per scope
CREATE TABLE mock_paper_counters (
    exam        TEXT NOT NULL,
    section     TEXT NOT NULL,
    module      TEXT NOT NULL DEFAULT 'any',
    next_number INT  NOT NULL DEFAULT 1,
    PRIMARY KEY (exam, section, module)
);

-- paper_id on session tables
ALTER TABLE listening_mock_sessions ADD COLUMN paper_id TEXT REFERENCES mock_papers(id) ON DELETE SET NULL;
CREATE INDEX idx_listening_mock_sessions_user_paper ON listening_mock_sessions(user_id, paper_id);

ALTER TABLE speaking_mock_sessions ADD COLUMN paper_id TEXT REFERENCES mock_papers(id) ON DELETE SET NULL;
CREATE INDEX idx_speaking_mock_sessions_user_paper ON speaking_mock_sessions(user_id, paper_id);

ALTER TABLE writing_mock_sessions ADD COLUMN paper_id TEXT REFERENCES mock_papers(id) ON DELETE SET NULL;
CREATE INDEX idx_writing_mock_sessions_user_paper ON writing_mock_sessions(user_id, paper_id);

ALTER TABLE reading_mock_sessions ADD COLUMN paper_id TEXT REFERENCES mock_papers(id) ON DELETE SET NULL;
CREATE INDEX idx_reading_mock_sessions_user_paper ON reading_mock_sessions(user_id, paper_id);

ALTER TABLE pte_mock_sessions ADD COLUMN paper_id TEXT REFERENCES mock_papers(id) ON DELETE SET NULL;
CREATE INDEX idx_pte_mock_sessions_user_paper ON pte_mock_sessions(user_id, paper_id);

-- Backfill existing fixed papers
-- 1. Listening tests
WITH ranked_listening AS (
    SELECT
        id AS test_id,
        title,
        created_at,
        row_number() OVER (ORDER BY created_at, id) AS rn
    FROM listening_tests
    WHERE is_published
),
inserted_listening AS (
    INSERT INTO mock_papers (
        id, exam, section, module, number, revision, status, title,
        content_schema, content, content_hash, created_at, published_at
    )
    SELECT
        gen_random_uuid()::text,
        'ielts',
        'listening',
        'any',
        rn,
        1,
        'published',
        title,
        1,
        jsonb_build_object('testId', test_id),
        encode(sha256(jsonb_build_object('testId', test_id)::text::bytea), 'hex'),
        created_at,
        created_at
    FROM ranked_listening
    RETURNING id, (content->>'testId') AS test_id
)
UPDATE listening_mock_sessions s
SET paper_id = ins.id
FROM inserted_listening ins
WHERE s.test_id = ins.test_id;

INSERT INTO mock_paper_counters (exam, section, module, next_number)
VALUES ('ielts', 'listening', 'any', COALESCE((SELECT MAX(number) + 1 FROM mock_papers WHERE exam = 'ielts' AND section = 'listening' AND module = 'any'), 1))
ON CONFLICT (exam, section, module) DO UPDATE
SET next_number = EXCLUDED.next_number;

-- 2. Speaking sets
WITH ranked_speaking AS (
    SELECT
        id AS set_id,
        title,
        created_at,
        row_number() OVER (ORDER BY created_at, id) AS rn
    FROM speaking_mock_sets
    WHERE is_published
),
inserted_speaking AS (
    INSERT INTO mock_papers (
        id, exam, section, module, number, revision, status, title,
        content_schema, content, content_hash, created_at, published_at
    )
    SELECT
        gen_random_uuid()::text,
        'ielts',
        'speaking',
        'any',
        rn,
        1,
        'published',
        title,
        1,
        jsonb_build_object('setId', set_id),
        encode(sha256(jsonb_build_object('setId', set_id)::text::bytea), 'hex'),
        created_at,
        created_at
    FROM ranked_speaking
    RETURNING id, (content->>'setId') AS set_id
)
UPDATE speaking_mock_sessions s
SET paper_id = ins.id
FROM inserted_speaking ins
WHERE s.set_id = ins.set_id;

INSERT INTO mock_paper_counters (exam, section, module, next_number)
VALUES ('ielts', 'speaking', 'any', COALESCE((SELECT MAX(number) + 1 FROM mock_papers WHERE exam = 'ielts' AND section = 'speaking' AND module = 'any'), 1))
ON CONFLICT (exam, section, module) DO UPDATE
SET next_number = EXCLUDED.next_number;
