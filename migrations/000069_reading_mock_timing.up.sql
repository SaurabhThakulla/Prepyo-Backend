-- Reading mocks keep time on the server and save answers as they go.
--
-- Why this exists. The 60-minute limit lived only in the browser: the timer
-- started again at 60:00 every time a paper was reopened, and the server
-- accepted a submission at any time with a duration the client reported. The
-- learner's answers lived only in the tab, so reopening a paper lost them.
--
-- expires_at is when the paper's time runs out. draft_answers is the latest
-- set of answers the browser saved before then; a submission that arrives
-- after the deadline is graded from these, not from whatever it carries.

ALTER TABLE reading_mock_sessions
    ADD COLUMN expires_at    TIMESTAMPTZ,
    ADD COLUMN draft_answers JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN drafts_saved_at TIMESTAMPTZ;

UPDATE reading_mock_sessions
   SET expires_at = created_at + make_interval(mins => duration_minutes)
 WHERE expires_at IS NULL;

ALTER TABLE reading_mock_sessions
    ALTER COLUMN expires_at SET NOT NULL,
    ADD CONSTRAINT reading_mock_sessions_draft_answers_array
        CHECK (jsonb_typeof(draft_answers) = 'array');
