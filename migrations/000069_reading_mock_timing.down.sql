ALTER TABLE reading_mock_sessions
    DROP CONSTRAINT IF EXISTS reading_mock_sessions_draft_answers_array,
    DROP COLUMN IF EXISTS drafts_saved_at,
    DROP COLUMN IF EXISTS draft_answers,
    DROP COLUMN IF EXISTS expires_at;
