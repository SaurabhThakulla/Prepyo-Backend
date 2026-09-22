-- Migration: 000058_ielts_speaking_parts_2_and_3.down.sql
DELETE FROM questions WHERE id IN (
    'ielts-spk-part2-001',
    'ielts-spk-part2-002',
    'ielts-spk-part2-003',
    'ielts-spk-part2-004',
    'ielts-spk-part2-005',
    'ielts-spk-part2-006',
    'ielts-spk-part2-007',
    'ielts-spk-part2-008',
    'ielts-spk-part2-009',
    'ielts-spk-part2-010',
    'ielts-spk-part3-001',
    'ielts-spk-part3-002',
    'ielts-spk-part3-003',
    'ielts-spk-part3-004',
    'ielts-spk-part3-005',
    'ielts-spk-part3-006',
    'ielts-spk-part3-007',
    'ielts-spk-part3-008',
    'ielts-spk-part3-009',
    'ielts-spk-part3-010'
);
