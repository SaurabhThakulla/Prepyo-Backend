-- 000044_seed_pte_reading_bank.down.sql

DELETE FROM questions WHERE id IN (
    'q-pte-fbrw-01', 'q-pte-fbrw-02', 'q-pte-fbrw-03', 'q-pte-fbrw-04', 'q-pte-fbrw-05',
    'q-pte-mcqm-01', 'q-pte-mcqm-02', 'q-pte-mcqm-03', 'q-pte-mcqm-04', 'q-pte-mcqm-05',
    'q-ri-pte-ro-01', 'q-ri-pte-ro-02', 'q-ri-pte-ro-03', 'q-ri-pte-ro-04', 'q-ri-pte-ro-05',
    'q-pte-fbr-01', 'q-pte-fbr-02', 'q-pte-fbr-03', 'q-pte-fbr-04', 'q-pte-fbr-05',
    'q-pte-mcqs-01', 'q-pte-mcqs-02', 'q-pte-mcqs-03', 'q-pte-mcqs-04', 'q-pte-mcqs-05'
);

DELETE FROM reading_reorder_items WHERE id IN (
    'ri-pte-ro-01', 'ri-pte-ro-02', 'ri-pte-ro-03', 'ri-pte-ro-04', 'ri-pte-ro-05'
);

DELETE FROM reading_question_groups WHERE id IN (
    'g-pte-fbrw-01', 'g-pte-fbrw-02', 'g-pte-fbrw-03', 'g-pte-fbrw-04', 'g-pte-fbrw-05',
    'g-pte-mcqm-01', 'g-pte-mcqm-02', 'g-pte-mcqm-03', 'g-pte-mcqm-04', 'g-pte-mcqm-05',
    'g-pte-fbr-01', 'g-pte-fbr-02', 'g-pte-fbr-03', 'g-pte-fbr-04', 'g-pte-fbr-05',
    'g-pte-mcqs-01', 'g-pte-mcqs-02', 'g-pte-mcqs-03', 'g-pte-mcqs-04', 'g-pte-mcqs-05'
);

DELETE FROM reading_passages WHERE id IN (
    'rp-pte-fbrw-01', 'rp-pte-fbrw-02', 'rp-pte-fbrw-03', 'rp-pte-fbrw-04', 'rp-pte-fbrw-05',
    'rp-pte-mcqm-01', 'rp-pte-mcqm-02', 'rp-pte-mcqm-03', 'rp-pte-mcqm-04', 'rp-pte-mcqm-05',
    'rp-pte-fbr-01', 'rp-pte-fbr-02', 'rp-pte-fbr-03', 'rp-pte-fbr-04', 'rp-pte-fbr-05',
    'rp-pte-mcqs-01', 'rp-pte-mcqs-02', 'rp-pte-mcqs-03', 'rp-pte-mcqs-04', 'rp-pte-mcqs-05'
);

DELETE FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-pte-reading';

UPDATE reading_mock_blueprints
   SET passage_count = 4,
       total_questions = 13,
       duration_minutes = 30
 WHERE id = 'bp-pte-reading';

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count, source) VALUES
    ('bp-pte-reading', 1, 1, 'fill-in-blanks-rw',    2, 'passage'),
    ('bp-pte-reading', 2, 1, 'fill-in-blanks-r',     2, 'passage'),
    ('bp-pte-reading', 3, 1, 'reading-mcq-single',   4, 'passage'),
    ('bp-pte-reading', 4, 1, 'reading-mcq-multiple', 2, 'passage'),
    ('bp-pte-reading', 5, 1, 'reorder-paragraphs',   3, 'reorder');
