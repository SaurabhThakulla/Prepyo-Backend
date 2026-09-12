ALTER TABLE users
    DROP COLUMN IF EXISTS study_goal,
    DROP COLUMN IF EXISTS destination,
    DROP COLUMN IF EXISTS prior_attempt,
    DROP COLUMN IF EXISTS previous_score,
    DROP COLUMN IF EXISTS focus_skill,
    DROP COLUMN IF EXISTS daily_minutes;
