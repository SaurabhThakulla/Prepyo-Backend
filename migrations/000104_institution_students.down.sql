-- The learner rows go too: without this table they have no way to sign in.
DELETE FROM users WHERE id IN (SELECT user_id FROM institution_students);
DROP TABLE IF EXISTS institution_students;
ALTER TABLE institutions DROP COLUMN IF EXISTS seat_limit;
