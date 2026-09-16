-- Reading tests and modular passage pattern (Passage A, Passage B, Passage C).
--
-- IELTS Academic Reading is 40 questions across three passages:
--   Passage A (Section 1): 13 questions
--     - Sentence Completion (7 questions)
--     - True / False / Not Given (6 questions)
--   Passage B (Section 2): 13 questions
--     - Find the Writer's View (6 questions)
--     - Match the Heading (4 questions)
--     - Yes / No / Not Given (3 questions)
--   Passage C (Section 3): 14 questions
--     - Sentence Completion (6 questions)
--     - Matching Information (4 questions)
--     - Yes / No / Not Given (4 questions)
-- Total: 13 + 13 + 14 = 40 questions.

-- ---------------------------------------------------------------------------
-- 1. Add passage_slot to reading_passages
-- ---------------------------------------------------------------------------

ALTER TABLE reading_passages
    ADD COLUMN passage_slot VARCHAR(10) NOT NULL DEFAULT 'custom'
        CHECK (passage_slot IN ('A', 'B', 'C', 'custom'));

-- ---------------------------------------------------------------------------
-- 2. Reading test bundles (Passage A + Passage B + Passage C = 40 Questions)
-- ---------------------------------------------------------------------------

CREATE TABLE reading_tests (
    id               TEXT PRIMARY KEY,
    exam_version_id  TEXT NOT NULL REFERENCES exam_versions(id),
    exam             TEXT NOT NULL DEFAULT 'IELTS' CHECK (exam IN ('PTE', 'IELTS')),
    title            TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    passage_a_id     TEXT NOT NULL REFERENCES reading_passages(id) ON DELETE RESTRICT,
    passage_b_id     TEXT NOT NULL REFERENCES reading_passages(id) ON DELETE RESTRICT,
    passage_c_id     TEXT NOT NULL REFERENCES reading_passages(id) ON DELETE RESTRICT,
    total_questions  INT  NOT NULL DEFAULT 40,
    duration_minutes INT  NOT NULL DEFAULT 60,
    is_published     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT reading_tests_distinct_passages
        CHECK (passage_a_id <> passage_b_id AND passage_b_id <> passage_c_id AND passage_a_id <> passage_c_id)
);

CREATE INDEX idx_reading_tests_exam ON reading_tests(exam) WHERE is_published;

-- ---------------------------------------------------------------------------
-- 3. Label existing seeded passages with their slot pattern
-- ---------------------------------------------------------------------------

-- Test Set 1:
UPDATE reading_passages SET passage_slot = 'A' WHERE id = 'rp-choc-01';
UPDATE reading_passages SET passage_slot = 'B' WHERE id = 'rp-bees-01';
UPDATE reading_passages SET passage_slot = 'C' WHERE id = 'rp-paper-01';

-- Test Set 2:
UPDATE reading_passages SET passage_slot = 'A' WHERE id = 'rp-container-01';
UPDATE reading_passages SET passage_slot = 'B' WHERE id = 'rp-icecore-01';
UPDATE reading_passages SET passage_slot = 'C' WHERE id = 'rp-seedvault-01';

-- ---------------------------------------------------------------------------
-- 4. Seed curated 40-question IELTS reading tests
-- ---------------------------------------------------------------------------

INSERT INTO reading_tests (
    id, exam_version_id, exam, title, description,
    passage_a_id, passage_b_id, passage_c_id,
    total_questions, duration_minutes, is_published
)
SELECT 'ielts-reading-test-01', 'ielts-2026-01', 'IELTS',
       'IELTS Academic Reading Practice Test 1',
       'Standard 40-question IELTS Reading Test across 3 passages: Passage 1 (A History of Chocolate), Passage 2 (The Return of Urban Beekeeping), and Passage 3 (The Long Road of Papermaking).',
       'rp-choc-01', 'rp-bees-01', 'rp-paper-01',
       40, 60, TRUE
WHERE EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-choc-01')
  AND EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-bees-01')
  AND EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-paper-01')
ON CONFLICT (id) DO NOTHING;

INSERT INTO reading_tests (
    id, exam_version_id, exam, title, description,
    passage_a_id, passage_b_id, passage_c_id,
    total_questions, duration_minutes, is_published
)
SELECT 'ielts-reading-test-02', 'ielts-2026-01', 'IELTS',
       'IELTS Academic Reading Practice Test 2',
       'Standard 40-question IELTS Reading Test across 3 passages: Passage 1 (The Box That Shrank the World), Passage 2 (Reading the Ice), and Passage 3 (Banking the Harvest).',
       'rp-container-01', 'rp-icecore-01', 'rp-seedvault-01',
       40, 60, TRUE
WHERE EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-container-01')
  AND EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-icecore-01')
  AND EXISTS (SELECT 1 FROM reading_passages WHERE id = 'rp-seedvault-01')
ON CONFLICT (id) DO NOTHING;

