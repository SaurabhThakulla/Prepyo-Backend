-- Wind and solar account for 28% in 2020, ahead of gas at 27%.
-- Replace only the inaccurate seed phrase; preserve any other editorial changes.
UPDATE questions
SET figure_data = replace(figure_data,
    'wind and solar rise almost tenfold and become the second largest source',
    'wind and solar rise almost tenfold and become the largest source')
WHERE id = 'ielts-wrt-fg-003'
  AND exam = 'IELTS' AND skill = 'writing'
  AND figure_data LIKE '%wind and solar rise almost tenfold and become the second largest source%';
