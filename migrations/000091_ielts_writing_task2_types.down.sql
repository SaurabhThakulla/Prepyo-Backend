-- Migration: 000091_ielts_writing_task2_types.down.sql
-- Preserve authored content, learner references and published test revisions on
-- rollback: published mock papers are immutable and learners may have answered
-- these questions.
SELECT 1;
