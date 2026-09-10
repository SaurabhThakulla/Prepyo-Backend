UPDATE questions
SET supported_exams = ARRAY['PTE', 'IELTS']
WHERE skill = 'reading'
  AND group_id IS NOT NULL
  AND type_id IN ('reading-mcq-single', 'reading-mcq-multiple');

DO $$
DECLARE
    shared INT;
BEGIN
    SELECT count(*) INTO shared
    FROM questions
    WHERE skill = 'reading'
      AND group_id IS NOT NULL
      AND type_id IN ('reading-mcq-single', 'reading-mcq-multiple')
      AND NOT (ARRAY['PTE', 'IELTS'] <@ supported_exams);

    IF shared > 0 THEN
        RAISE EXCEPTION 'shared multiple choice: % questions are still set by one exam', shared;
    END IF;
END $$;
