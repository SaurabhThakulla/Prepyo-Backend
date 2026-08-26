-- Reverse of 000001_init_schema.up.sql. Dropped in dependency order.

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_plan_id_fkey;

DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS user_mission_progress;
DROP TABLE IF EXISTS daily_missions;
DROP TABLE IF EXISTS xp_transactions;
DROP TABLE IF EXISTS ai_evaluations;
DROP TABLE IF EXISTS mistakes;
DROP TABLE IF EXISTS mock_attempts;
DROP TABLE IF EXISTS practice_attempts;
DROP TABLE IF EXISTS mock_sections;
DROP TABLE IF EXISTS mocks;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS exam_versions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
