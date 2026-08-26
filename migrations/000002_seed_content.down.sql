-- Reverse of 000002_seed_content.up.sql.
--
-- Only content rows are removed. Learner rows that reference them are left to
-- the foreign keys, so this fails loudly rather than silently deleting
-- someone's attempt history.

DELETE FROM daily_missions;
DELETE FROM mock_sections;
DELETE FROM mocks;
DELETE FROM questions;
DELETE FROM plans WHERE id IN ('free', 'pro', 'elite');
DELETE FROM exam_versions;
