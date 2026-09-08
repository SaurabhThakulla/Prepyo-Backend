ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

UPDATE users SET role = CASE WHEN role = 'admin' THEN 'admin' ELSE 'learner' END;

ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('learner', 'admin'));
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'learner';

DROP INDEX IF EXISTS idx_users_plan_valid_until;
ALTER TABLE users DROP COLUMN IF EXISTS plan_started_at;
