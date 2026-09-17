-- Exact reverse of 000049 up. Each guard keeps the undo from clobbering
-- later editorial edits that no longer contain the corrected phrase.
DELETE FROM questions
WHERE id IN ('pte-wrt-swt-002', 'pte-wrt-swt-003', 'pte-wrt-swt-004', 'pte-wrt-swt-005')
  AND exam = 'PTE' AND skill = 'writing';

UPDATE questions SET type_id = 'ielts-writing-task2', type_name = 'Writing Task 2'
WHERE id = 'ielts-wrt-001' AND exam = 'IELTS' AND skill = 'writing'
  AND type_id = 'ielts-writing-task2-opinion';

UPDATE questions SET model_answer = replace(model_answer,
 'reflective roofing, permeable surfaces and wider tree cover',
 'reflective materials, green roofs and wider tree cover')
WHERE id = 'pte-wrt-001' AND exam = 'PTE' AND skill = 'writing'
  AND model_answer LIKE '%reflective roofing, permeable surfaces and wider tree cover%';

UPDATE questions SET figure_data = replace(figure_data,
 'Canada is highest throughout; Mexico has the largest absolute increase (67 percentage points), followed by Canada (43) and India (42); India starts lowest; from 2010 to 2020 Mexico gains 41 points and India 36; the Canada-India gap widens slightly from 50 to 51 points, while the ratio falls sharply.',
 'Canada is highest throughout but grows the least in absolute terms; India starts lowest and grows fastest after 2010; the gap between Canada and India narrows from 50 to 51 points but the ratio falls sharply.')
WHERE id = 'ielts-wrt-fg-001' AND exam = 'IELTS' AND skill = 'writing';

UPDATE questions SET figure_data = replace(figure_data,
 'under a third of its 2015 figure', 'under a third of its 2010 figure')
WHERE id = 'ielts-wrt-fg-004' AND exam = 'IELTS' AND skill = 'writing'
  AND figure_data LIKE '%under a third of its 2015 figure%';

UPDATE questions SET figure_data = replace(figure_data,
 'organic waste is the largest fraction in Germany, Brazil and Kenya, while paper is largest in Japan; the chart contains no income data, so do not infer an income relationship',
 'organic waste is the largest fraction everywhere and rises as national income falls')
WHERE id = 'ielts-wrt-fg-009' AND exam = 'IELTS' AND skill = 'writing'
  AND figure_data LIKE '%no income data%';
