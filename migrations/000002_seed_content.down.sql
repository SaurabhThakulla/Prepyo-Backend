-- Reverse of 000002_seed_content.up.sql. Learner rows that reference these are
-- left to the foreign keys, so this fails loudly rather than deleting history.

DELETE FROM daily_missions WHERE id IN ('mis-speaking-2', 'mis-writing-1', 'mis-reading-3', 'mis-listening-2');
DELETE FROM plans WHERE id IN ('free', 'pro', 'elite');
DELETE FROM exam_versions WHERE id IN ('pte-2026-01', 'ielts-2026-01');
