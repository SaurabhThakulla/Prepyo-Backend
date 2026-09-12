ALTER TABLE users
    ADD COLUMN study_goal     TEXT CHECK (study_goal IN ('university', 'migration', 'work', 'other')),
    ADD COLUMN destination    TEXT CHECK (destination IN ('australia', 'uk', 'canada', 'usa', 'new-zealand', 'other')),
    ADD COLUMN prior_attempt  TEXT CHECK (prior_attempt IN ('first', 'retake')),
    ADD COLUMN previous_score NUMERIC(4, 1),
    ADD COLUMN focus_skill    TEXT CHECK (focus_skill IN ('speaking', 'writing', 'reading', 'listening')),
    ADD COLUMN daily_minutes  INT CHECK (daily_minutes BETWEEN 5 AND 480);
