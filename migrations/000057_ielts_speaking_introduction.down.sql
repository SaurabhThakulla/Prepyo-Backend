-- Migration: 000057_ielts_speaking_introduction.down.sql
DELETE FROM questions WHERE id IN (
    'ielts-spk-intro-001',
    'ielts-spk-intro-002',
    'ielts-spk-intro-003',
    'ielts-spk-intro-004',
    'ielts-spk-intro-005',
    'ielts-spk-intro-006',
    'ielts-spk-intro-007',
    'ielts-spk-intro-008',
    'ielts-spk-intro-009',
    'ielts-spk-intro-010'
);
