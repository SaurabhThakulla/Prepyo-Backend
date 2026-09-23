-- IELTS General Training as a module.
--
-- Why this exists. Prepyo modelled one IELTS: Academic. General Training is a
-- different Reading paper (everyday and workplace texts before a longer
-- general-interest text, 2,150-2,375 words), a different Writing Task 1 (a
-- letter), and a different raw-to-band Reading table. Listening, Speaking and
-- Writing Task 2 are the same in both.
--
-- This migration adds the module as data. Content arrives separately.

-- The learner's module. Only meaningful when target_exam is IELTS.
ALTER TABLE users
    ADD COLUMN target_module TEXT NOT NULL DEFAULT 'academic'
        CHECK (target_module IN ('academic', 'general_training'));

-- Which IELTS modules may deal a passage. Every existing passage is Academic.
-- A general-interest text can serve both, as General Training section 3 draws
-- from the same kinds of sources as Academic reading.
ALTER TABLE reading_passages
    ADD COLUMN modules TEXT[] NOT NULL DEFAULT '{academic}';

ALTER TABLE reading_passages
    ADD CONSTRAINT reading_passages_modules_valid
        CHECK (cardinality(modules) > 0
               AND modules <@ ARRAY['academic', 'general_training']);

-- A blueprint belongs to a module, so each module has one live paper shape.
ALTER TABLE reading_mock_blueprints
    ADD COLUMN module TEXT NOT NULL DEFAULT 'academic'
        CHECK (module IN ('academic', 'general_training'));

DROP INDEX IF EXISTS idx_reading_blueprint_active;
CREATE UNIQUE INDEX idx_reading_blueprint_active
    ON reading_mock_blueprints(exam, module) WHERE is_active;

-- A section can ask for a particular kind of text by tag. General Training
-- section 1 needs everyday notices and adverts, section 2 workplace texts;
-- the task counts alone cannot tell those apart from an academic article.
ALTER TABLE reading_mock_blueprint_slots
    ADD COLUMN passage_tag TEXT;

-- ---------------------------------------------------------------------------
-- The General Training Reading paper
-- ---------------------------------------------------------------------------
--
-- 40 one-mark questions in 60 minutes over three sections of increasing
-- difficulty:
--   Section 1 (Q1-14):  everyday texts: matching information 7, TFNG 7
--   Section 2 (Q15-27): workplace texts: sentence completion 5, TFNG 4,
--                       multiple choice 4
--   Section 3 (Q28-40): general-interest text: matching information 4,
--                       YNNG 5, sentence completion 4

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated) VALUES
    ('mock-ielts-gt-reading-gen', 'ielts-2026-01', 'IELTS',
     'IELTS General Training Reading Mock',
     'Three sections of increasing difficulty: everyday texts, workplace texts and a longer general-interest article, drawn fresh each time.',
     60, FALSE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO reading_mock_blueprints
    (id, mock_id, exam, module, passage_count, total_questions, duration_minutes) VALUES
    ('bp-ielts-gt-reading', 'mock-ielts-gt-reading-gen', 'IELTS', 'general_training', 3, 40, 60)
ON CONFLICT (id) DO NOTHING;

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count, passage_tag) VALUES
    ('bp-ielts-gt-reading', 1, 1, 'reading-matching-information', 7, 'GT Section 1'),
    ('bp-ielts-gt-reading', 1, 2, 'reading-true-false',           7, 'GT Section 1'),

    ('bp-ielts-gt-reading', 2, 1, 'reading-sentence-completion',  5, 'GT Section 2'),
    ('bp-ielts-gt-reading', 2, 2, 'reading-true-false',           4, 'GT Section 2'),
    ('bp-ielts-gt-reading', 2, 3, 'reading-mcq-single',           4, 'GT Section 2'),

    ('bp-ielts-gt-reading', 3, 1, 'reading-matching-information', 4, NULL),
    ('bp-ielts-gt-reading', 3, 2, 'reading-yes-no-not-given',     5, NULL),
    ('bp-ielts-gt-reading', 3, 3, 'reading-sentence-completion',  4, NULL)
ON CONFLICT DO NOTHING;

-- The original general-interest articles (000054, 000067) are suitable for
-- General Training section 3 as well as Academic reading.
UPDATE reading_passages
   SET modules = ARRAY['academic', 'general_training']
 WHERE id = 'rp-ir54-library' OR id LIKE 'rp-ir67-%';
