ALTER TABLE institution_students DROP COLUMN IF EXISTS disabled_at;

ALTER TABLE institutions ADD COLUMN seat_limit INT NOT NULL DEFAULT 2 CHECK (seat_limit >= 0);
UPDATE institutions
SET seat_limit = CASE plan WHEN 'basic' THEN 10 WHEN 'extended' THEN 25 WHEN 'ultimate' THEN 50 ELSE 2 END;
ALTER TABLE institutions DROP COLUMN IF EXISTS plan;
