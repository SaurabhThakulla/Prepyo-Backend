-- Institutions buy Basic (10 students), Extended (25) or Ultimate (50); there is no free plan.
-- Seats come from the plan, so seat_limit goes. An institution already paid for keeps the plan its seats fit.
ALTER TABLE institutions ADD COLUMN plan TEXT CHECK (plan IN ('basic', 'extended', 'ultimate'));
UPDATE institutions
SET plan = CASE WHEN seat_limit <= 10 THEN 'basic' WHEN seat_limit <= 25 THEN 'extended' ELSE 'ultimate' END
WHERE valid_until IS NOT NULL;
ALTER TABLE institutions DROP COLUMN seat_limit;

-- A disabled student cannot sign in; their history and seat are kept.
ALTER TABLE institution_students ADD COLUMN disabled_at TIMESTAMPTZ;
