-- PTE Academic mock tests: the full test, and one sectional test per skill.
--
-- Why this exists. PTE could be practised item by item, and its only mock was a
-- reading paper run like an IELTS paper: a page of passages the learner moves
-- around freely. The real test is nothing like that. It is one item per screen,
-- in a fixed order, with no way back; speaking items run their own preparation
-- and recording windows; Summarize Written Text, the Essay and Summarize Spoken
-- Text each have their own clock; the rest of Reading and Listening share one
-- clock per part. Scores are reported on the 10-90 scale for each communicative
-- skill, and most items count towards two skills at once.
--
-- One engine runs every kind of PTE mock (package ptemock). What a kind is made
-- of - which tasks, how many of each, how they are timed - is a blueprint in
-- that package; this migration only stores the papers dealt from it.

-- The attempt rows results are recorded against. None is offered in the
-- generic mock list, which starts papers through a different flow, and all are
-- marked generated so that billing does not count their attempts: a PTE full
-- mock spends the full-mock allowance when it starts (pte_mock_sessions), and a
-- sectional test pays in sub-tests.
INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated, is_available) VALUES
    ('mock-pte-full', 'pte-2026-01', 'PTE', 'PTE Academic Full Mock Test',
     'Speaking & Writing, Reading and Listening in the order of the test, scored 10-90.',
     125, FALSE, TRUE, FALSE),
    ('mock-pte-speaking', 'pte-2026-01', 'PTE', 'PTE Speaking Sectional Test',
     'Every speaking task in the order of the test, with a Speaking score.',
     35, FALSE, TRUE, FALSE),
    ('mock-pte-writing', 'pte-2026-01', 'PTE', 'PTE Writing Sectional Test',
     'Summarize Written Text and Write Essay, each on its own clock.',
     40, FALSE, TRUE, FALSE),
    ('mock-pte-reading', 'pte-2026-01', 'PTE', 'PTE Reading Sectional Test',
     'The reading part, one item per screen on a shared clock.',
     30, FALSE, TRUE, FALSE),
    ('mock-pte-listening', 'pte-2026-01', 'PTE', 'PTE Listening Sectional Test',
     'The listening part: each recording plays once.',
     30, FALSE, TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE pte_mock_sessions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind               TEXT NOT NULL
                           CHECK (kind IN ('full', 'speaking', 'writing', 'reading', 'listening')),
    mock_id            TEXT NOT NULL REFERENCES mocks(id),
    exam_version_id    TEXT NOT NULL REFERENCES exam_versions(id),
    status             TEXT NOT NULL DEFAULT 'in_progress'
                           CHECK (status IN ('in_progress', 'scoring', 'completed', 'abandoned')),
    -- The item on screen. Items are taken in order and never revisited, so the
    -- learner's place in the test is one number.
    current_position   INT  NOT NULL DEFAULT 1 CHECK (current_position > 0),
    total_items        INT  NOT NULL CHECK (total_items > 0),
    -- Each clock's deadline, set when the first item it times is shown:
    -- {"reading": "...", "item:14": "..."}. Kept by the server, so a browser
    -- with a wrong clock or a reload does not buy time.
    deadlines          JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- Tasks of the blueprint the bank could not supply. The paper is dealt
    -- without them and says so, rather than failing outright.
    missing_tasks      TEXT[] NOT NULL DEFAULT '{}',
    -- The score report, written once scoring finishes.
    result             JSONB,
    mock_attempt_id    UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    scoring_started_at TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ
);

-- One open paper per learner and kind: a retried start resumes it instead of
-- dealing and charging for another.
CREATE UNIQUE INDEX idx_pte_mock_live
    ON pte_mock_sessions(user_id, kind) WHERE status = 'in_progress';

CREATE INDEX idx_pte_mock_user ON pte_mock_sessions(user_id, created_at DESC);

-- Full mocks started, for the plan allowance.
CREATE INDEX idx_pte_mock_full ON pte_mock_sessions(user_id, created_at) WHERE kind = 'full';

CREATE TABLE pte_mock_items (
    session_id       UUID NOT NULL REFERENCES pte_mock_sessions(id) ON DELETE CASCADE,
    position         INT  NOT NULL CHECK (position > 0),
    question_id      TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    -- The blueprint task the item was dealt for ('RA', 'WFD', ...), which fixes
    -- its timing and which skills it counts towards.
    task             TEXT NOT NULL,
    part             TEXT NOT NULL CHECK (part IN ('speaking_writing', 'reading', 'listening')),
    -- Re-order Paragraphs boxes in the order they were dealt. The bank stores
    -- them correct, so the shuffle is kept here to survive a reload.
    option_order     TEXT[],
    status           TEXT NOT NULL DEFAULT 'pending'
                         CHECK (status IN ('pending', 'answered', 'skipped', 'timed_out')),
    -- The answer as models.AnswerSubmission; for a pending item, the draft
    -- saved so far, which is what is marked if time runs out.
    response         JSONB,
    -- A spoken answer: what the device heard, and how it was delivered.
    transcript       TEXT,
    duration_seconds INT,
    delivery         JSONB,
    shown_at         TIMESTAMPTZ,
    answered_at      TIMESTAMPTZ,
    -- The mark, once scored: score out of max_score. unscored_reason says why an
    -- item has none (its rating was unavailable), and it is left out of the
    -- skill scores rather than counted as zero.
    scored           BOOLEAN NOT NULL DEFAULT FALSE,
    score            DOUBLE PRECISION,
    max_score        DOUBLE PRECISION,
    unscored_reason  TEXT,
    feedback         TEXT,
    correct_display  TEXT,
    user_display     TEXT,
    evaluation_id    UUID REFERENCES ai_evaluations(id) ON DELETE SET NULL,
    PRIMARY KEY (session_id, position)
);

-- "Have I been dealt this before", for preferring unseen items.
CREATE INDEX idx_pte_mock_items_question ON pte_mock_items(question_id);
