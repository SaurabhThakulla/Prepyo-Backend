-- Restore the seed wording only when the corrected phrase is still present.
UPDATE questions
SET figure_data = replace(figure_data,
    'wind and solar rise almost tenfold and become the largest source',
    'wind and solar rise almost tenfold and become the second largest source')
WHERE id = 'ielts-wrt-fg-003'
  AND exam = 'IELTS' AND skill = 'writing'
  AND figure_data LIKE '%wind and solar rise almost tenfold and become the largest source%';
