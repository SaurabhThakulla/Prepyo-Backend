UPDATE questions
SET supported_exams = ARRAY[exam]
WHERE skill = 'reading'
  AND group_id IS NOT NULL
  AND type_id IN ('reading-mcq-single', 'reading-mcq-multiple');
