-- Learners can delete notifications. Most are deleted outright; a once-only
-- notice is hidden instead, so its dedupe key keeps it from being sent again
-- (clearing "daily limit reached" must not bring it straight back).
ALTER TABLE notifications ADD COLUMN dismissed BOOLEAN NOT NULL DEFAULT FALSE;
