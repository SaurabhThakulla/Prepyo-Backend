-- Writing content: opinion essays for both exams and IELTS Task 1 figures.
--
-- figure_data carries the numbers behind a chart in words. It is never sent to
-- the browser -- the learner gets the image -- but the evaluator needs it to
-- judge whether a description reports the data accurately.

ALTER TABLE questions ADD COLUMN IF NOT EXISTS figure_data TEXT;

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    image_url, figure_data, prep_time_seconds, time_limit_seconds, difficulty,
    tags, points
) VALUES
    ('ielts-wrt-op-001', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'University Tuition Fees',
        'Higher education should be free of charge for every student, regardless of family income.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-001', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'University Tuition Fees',
        'Higher education should be free of charge for every student, regardless of family income.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-002', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Remote Learning',
        'Online classes are just as effective for schoolchildren as lessons in a physical classroom.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-002', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Remote Learning',
        'Online classes are just as effective for schoolchildren as lessons in a physical classroom.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-003', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Homework in Primary School',
        'Children under the age of eleven should not be given any homework.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-003', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Homework in Primary School',
        'Children under the age of eleven should not be given any homework.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-004', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Standardised Testing',
        'Examinations are a poor way of measuring what a student has actually learned.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-004', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Standardised Testing',
        'Examinations are a poor way of measuring what a student has actually learned.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-005', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Learning a Second Language',
        'Every child should be required to learn a foreign language from the age of five.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-005', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Learning a Second Language',
        'Every child should be required to learn a foreign language from the age of five.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-006', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'School Uniforms',
        'Requiring pupils to wear a uniform improves discipline and reduces social division.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-006', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'School Uniforms',
        'Requiring pupils to wear a uniform improves discipline and reduces social division.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-007', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Teaching Practical Skills',
        'Schools should spend less time on academic subjects and more on cooking, budgeting and basic repairs.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-007', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Teaching Practical Skills',
        'Schools should spend less time on academic subjects and more on cooking, budgeting and basic repairs.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-008', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Single-Sex Education',
        'Boys and girls learn better when they are taught in separate schools.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-008', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Single-Sex Education',
        'Boys and girls learn better when they are taught in separate schools.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-009', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Gap Years',
        'Students benefit more from a year of work or travel than from going straight to university.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-009', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Gap Years',
        'Students benefit more from a year of work or travel than from going straight to university.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-010', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Teachers'' Pay',
        'Teachers should be paid as much as doctors and engineers.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-010', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Teachers'' Pay',
        'Teachers should be paid as much as doctors and engineers.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-011', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Screen Time for Children',
        'Children under twelve should not be allowed to own a smartphone.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-011', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Screen Time for Children',
        'Children under twelve should not be allowed to own a smartphone.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-012', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Social Media Regulation',
        'Governments should regulate social media in the same way they regulate newspapers.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-012', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Social Media Regulation',
        'Governments should regulate social media in the same way they regulate newspapers.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-013', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Artificial Intelligence at Work',
        'Machines will eliminate more jobs than they create over the next fifty years.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-013', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Artificial Intelligence at Work',
        'Machines will eliminate more jobs than they create over the next fifty years.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-014', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Online Privacy',
        'People have already given up too much privacy in exchange for free online services.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-014', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Online Privacy',
        'People have already given up too much privacy in exchange for free online services.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-015', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Working From Home',
        'Employees are more productive at home than in an office.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-015', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Working From Home',
        'Employees are more productive at home than in an office.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-016', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'The Four-Day Week',
        'A four-day working week would improve both productivity and wellbeing.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-016', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'The Four-Day Week',
        'A four-day working week would improve both productivity and wellbeing.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-017', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Retirement Age',
        'The retirement age should rise as people live longer.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-017', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Retirement Age',
        'The retirement age should rise as people live longer.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-018', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Unpaid Internships',
        'Unpaid internships should be banned because they favour the wealthy.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-018', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Unpaid Internships',
        'Unpaid internships should be banned because they favour the wealthy.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-019', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Job Automation and Retraining',
        'Companies that automate jobs should be legally required to retrain the workers they replace.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-019', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Job Automation and Retraining',
        'Companies that automate jobs should be legally required to retrain the workers they replace.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-020', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Workplace Dress Codes',
        'Formal dress codes at work are outdated and should be abolished.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-020', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Workplace Dress Codes',
        'Formal dress codes at work are outdated and should be abolished.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-021', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Carbon Taxes',
        'Taxing carbon emissions is the most effective way to fight climate change.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-021', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Carbon Taxes',
        'Taxing carbon emissions is the most effective way to fight climate change.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-022', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Single-Use Plastic',
        'Governments should ban single-use plastic entirely rather than encouraging recycling.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-022', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Single-Use Plastic',
        'Governments should ban single-use plastic entirely rather than encouraging recycling.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-023', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Nuclear Power',
        'Nuclear energy is a necessary part of any realistic plan to cut emissions.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-023', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Nuclear Power',
        'Nuclear energy is a necessary part of any realistic plan to cut emissions.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-024', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Meat Consumption',
        'Reducing meat consumption is the most useful thing an individual can do for the environment.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-024', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Meat Consumption',
        'Reducing meat consumption is the most useful thing an individual can do for the environment.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-025', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Private Cars in Cities',
        'Private cars should be banned from the centre of every large city.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-025', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Private Cars in Cities',
        'Private cars should be banned from the centre of every large city.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-026', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Wildlife Conservation Funding',
        'Money spent protecting endangered species would be better spent on human welfare.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-026', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Wildlife Conservation Funding',
        'Money spent protecting endangered species would be better spent on human welfare.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-027', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Fast Fashion',
        'Clothing companies should be held financially responsible for the waste their products create.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-027', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Fast Fashion',
        'Clothing companies should be held financially responsible for the waste their products create.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-028', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Air Travel',
        'Frequent flyers should pay a much higher rate of tax than occasional travellers.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-028', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Air Travel',
        'Frequent flyers should pay a much higher rate of tax than occasional travellers.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-029', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Renewable Energy Subsidies',
        'Governments should subsidise renewable energy even if it raises household bills.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-029', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Renewable Energy Subsidies',
        'Governments should subsidise renewable energy even if it raises household bills.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-030', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Water Scarcity',
        'Access to clean drinking water should be treated as a human right rather than a commodity.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-030', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Water Scarcity',
        'Access to clean drinking water should be treated as a human right rather than a commodity.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-031', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Public Healthcare Funding',
        'Healthcare should be funded entirely by the state and free at the point of use.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-031', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Public Healthcare Funding',
        'Healthcare should be funded entirely by the state and free at the point of use.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-032', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Sugar Taxes',
        'Taxing sugary drinks is a fair and effective way to improve public health.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-032', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Sugar Taxes',
        'Taxing sugary drinks is a fair and effective way to improve public health.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-033', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Mental Health in Schools',
        'Schools should employ as many counsellors as they do sports coaches.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-033', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Mental Health in Schools',
        'Schools should employ as many counsellors as they do sports coaches.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-034', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Preventive Medicine',
        'Governments should spend more on preventing illness than on treating it.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-034', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Preventive Medicine',
        'Governments should spend more on preventing illness than on treating it.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-035', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Vaccination Policy',
        'Vaccination against serious disease should be compulsory for all children.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-035', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Vaccination Policy',
        'Vaccination against serious disease should be compulsory for all children.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-036', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Alternative Medicine',
        'Public health systems should not fund treatments that lack scientific evidence.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-036', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Alternative Medicine',
        'Public health systems should not fund treatments that lack scientific evidence.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-037', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Sports and Fitness',
        'Daily physical exercise should be a compulsory part of the school timetable.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-037', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Sports and Fitness',
        'Daily physical exercise should be a compulsory part of the school timetable.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-038', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Junk Food Advertising',
        'Advertising unhealthy food to children should be illegal.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-038', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Junk Food Advertising',
        'Advertising unhealthy food to children should be illegal.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-039', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Working Hours and Health',
        'Long working hours do more damage to public health than poor diet.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-039', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Working Hours and Health',
        'Long working hours do more damage to public health than poor diet.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-040', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Ageing Populations',
        'Countries with ageing populations should encourage immigration to sustain their workforce.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-040', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Ageing Populations',
        'Countries with ageing populations should encourage immigration to sustain their workforce.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-041', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Prison Sentences',
        'Long prison sentences are an ineffective way of reducing crime.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-041', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Prison Sentences',
        'Long prison sentences are an ineffective way of reducing crime.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-042', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Rehabilitation Over Punishment',
        'The main purpose of prison should be rehabilitation, not punishment.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-042', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Rehabilitation Over Punishment',
        'The main purpose of prison should be rehabilitation, not punishment.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-043', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Surveillance Cameras',
        'Widespread public surveillance is a fair price to pay for lower crime.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-043', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Surveillance Cameras',
        'Widespread public surveillance is a fair price to pay for lower crime.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-044', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Youth Crime',
        'Young offenders should be treated more leniently than adults who commit the same crime.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-044', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Youth Crime',
        'Young offenders should be treated more leniently than adults who commit the same crime.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-045', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Capital Punishment',
        'No modern justice system should include the death penalty.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-045', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Capital Punishment',
        'No modern justice system should include the death penalty.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-046', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Drug Policy',
        'Possession of small quantities of drugs should be treated as a health issue, not a crime.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-046', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Drug Policy',
        'Possession of small quantities of drugs should be treated as a health issue, not a crime.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-047', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Gun Ownership',
        'Private ownership of firearms should be prohibited in all circumstances.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-047', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Gun Ownership',
        'Private ownership of firearms should be prohibited in all circumstances.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-048', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Police Funding',
        'Money spent on policing would reduce crime more effectively if spent on education.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-048', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Police Funding',
        'Money spent on policing would reduce crime more effectively if spent on education.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-049', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Corporate Crime',
        'Executives should face prison when their companies cause serious environmental damage.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-049', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Corporate Crime',
        'Executives should face prison when their companies cause serious environmental damage.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-050', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Jury Trials',
        'Serious criminal cases should be decided by trained judges rather than juries.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-050', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Jury Trials',
        'Serious criminal cases should be decided by trained judges rather than juries.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-051', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Museum Entry Fees',
        'National museums and galleries should be free to enter.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-051', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Museum Entry Fees',
        'National museums and galleries should be free to enter.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-052', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Funding the Arts',
        'Public money should not be spent on art that most people never see.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-052', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Funding the Arts',
        'Public money should not be spent on art that most people never see.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-053', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Historic Buildings',
        'Preserving old buildings is more important than making room for new development.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-053', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Historic Buildings',
        'Preserving old buildings is more important than making room for new development.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-054', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Traditional Crafts',
        'Governments should actively fund traditional crafts that cannot survive commercially.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-054', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Traditional Crafts',
        'Governments should actively fund traditional crafts that cannot survive commercially.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-055', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Global Culture',
        'Globalisation is eroding local cultures beyond repair.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-055', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Global Culture',
        'Globalisation is eroding local cultures beyond repair.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-056', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Language Extinction',
        'Efforts to revive dying languages are a poor use of public resources.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-056', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Language Extinction',
        'Efforts to revive dying languages are a poor use of public resources.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-057', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Film and National Identity',
        'Countries should require cinemas to show a minimum proportion of domestic films.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-057', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Film and National Identity',
        'Countries should require cinemas to show a minimum proportion of domestic films.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-058', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Sport and National Pride',
        'Hosting international sporting events brings a country more cost than benefit.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-058', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Sport and National Pride',
        'Hosting international sporting events brings a country more cost than benefit.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-059', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Athletes'' Salaries',
        'Professional athletes are paid far more than their contribution to society justifies.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-059', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Athletes'' Salaries',
        'Professional athletes are paid far more than their contribution to society justifies.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-060', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Amateur Sport',
        'Public funding for sport should go to community facilities rather than elite athletes.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-060', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Amateur Sport',
        'Public funding for sport should go to community facilities rather than elite athletes.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-061', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Public Transport Investment',
        'Governments should invest in public transport rather than building new roads.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-061', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Public Transport Investment',
        'Governments should invest in public transport rather than building new roads.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-062', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Cycling Infrastructure',
        'Cities should remove road space from cars and give it to cyclists.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-062', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Cycling Infrastructure',
        'Cities should remove road space from cars and give it to cyclists.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-063', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'High-Speed Rail',
        'High-speed rail is worth its enormous cost.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-063', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'High-Speed Rail',
        'High-speed rail is worth its enormous cost.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-064', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Self-Driving Cars',
        'Driverless vehicles will make roads considerably safer.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-064', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Self-Driving Cars',
        'Driverless vehicles will make roads considerably safer.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-065', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Rural Depopulation',
        'Governments should offer financial incentives for people to move to rural areas.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-065', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Rural Depopulation',
        'Governments should offer financial incentives for people to move to rural areas.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-066', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Affordable Housing',
        'Governments, not private developers, should be responsible for building affordable homes.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-066', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Affordable Housing',
        'Governments, not private developers, should be responsible for building affordable homes.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-067', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Urban Green Space',
        'Every new housing development should be required to include public green space.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-067', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Urban Green Space',
        'Every new housing development should be required to include public green space.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-068', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Skyscrapers',
        'Building upwards is the only sensible answer to urban overcrowding.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-068', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Skyscrapers',
        'Building upwards is the only sensible answer to urban overcrowding.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-069', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Tourism and Local Communities',
        'Popular destinations should limit the number of tourists they admit each year.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-069', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Tourism and Local Communities',
        'Popular destinations should limit the number of tourists they admit each year.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-070', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Remote Work and Cities',
        'The rise of remote work will empty city centres and that is a good thing.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-070', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Remote Work and Cities',
        'The rise of remote work will empty city centres and that is a good thing.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-071', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Parental Leave',
        'Fathers should receive the same amount of paid parental leave as mothers.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-071', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Parental Leave',
        'Fathers should receive the same amount of paid parental leave as mothers.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-072', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Childcare Costs',
        'The state should provide free childcare for all working parents.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-072', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Childcare Costs',
        'The state should provide free childcare for all working parents.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-073', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Extended Families',
        'Older relatives are better cared for at home than in residential care.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-073', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Extended Families',
        'Older relatives are better cared for at home than in residential care.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-074', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Family Size',
        'Governments should not use financial incentives to influence how many children people have.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-074', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Family Size',
        'Governments should not use financial incentives to influence how many children people have.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-075', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Marriage',
        'Marriage as an institution is no longer relevant to modern life.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-075', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Marriage',
        'Marriage as an institution is no longer relevant to modern life.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-076', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Pocket Money',
        'Giving children money for household chores teaches them the wrong lesson.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-076', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Pocket Money',
        'Giving children money for household chores teaches them the wrong lesson.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-077', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Grandparents and Childcare',
        'Relying on grandparents for childcare places an unfair burden on older people.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-077', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Grandparents and Childcare',
        'Relying on grandparents for childcare places an unfair burden on older people.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-078', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Work-Life Balance',
        'Employers should be legally prohibited from contacting staff outside working hours.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-078', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Work-Life Balance',
        'Employers should be legally prohibited from contacting staff outside working hours.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-079', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'News Media Bias',
        'Impartial news reporting is no longer possible in the modern media.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-079', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'News Media Bias',
        'Impartial news reporting is no longer possible in the modern media.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-080', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Paying for Journalism',
        'People should expect to pay for quality journalism rather than read it free.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-080', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Paying for Journalism',
        'People should expect to pay for quality journalism rather than read it free.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-081', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Celebrity Culture',
        'Media attention on celebrities has a damaging effect on young people.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-081', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Celebrity Culture',
        'Media attention on celebrities has a damaging effect on young people.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-082', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Advertising to Children',
        'All advertising aimed at children under twelve should be banned.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-082', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Advertising to Children',
        'All advertising aimed at children under twelve should be banned.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-083', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Streaming Services',
        'Streaming has improved the quality of television and film.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-083', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Streaming Services',
        'Streaming has improved the quality of television and film.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-084', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Video Games',
        'Video games do more to develop useful skills than they do harm.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-084', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Video Games',
        'Video games do more to develop useful skills than they do harm.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-085', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Reading Habits',
        'Digital reading will never replace the value of printed books.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-085', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Reading Habits',
        'Digital reading will never replace the value of printed books.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-086', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Misinformation',
        'Social media companies should be legally liable for false information published on their platforms.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-086', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Misinformation',
        'Social media companies should be legally liable for false information published on their platforms.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-087', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Space Exploration',
        'Money spent exploring space would be better spent solving problems on Earth.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-087', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Space Exploration',
        'Money spent exploring space would be better spent solving problems on Earth.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-088', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Scientific Research Funding',
        'Governments should fund scientific research with no obvious practical application.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-088', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Scientific Research Funding',
        'Governments should fund scientific research with no obvious practical application.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-089', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Animal Testing',
        'Testing medicines on animals is justified when human lives are at stake.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-089', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Animal Testing',
        'Testing medicines on animals is justified when human lives are at stake.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-090', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Genetic Engineering',
        'Genetically modified crops are essential to feeding a growing population.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-090', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Genetic Engineering',
        'Genetically modified crops are essential to feeding a growing population.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-091', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Open Access Research',
        'Publicly funded research should be free for anyone to read.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-091', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Open Access Research',
        'Publicly funded research should be free for anyone to read.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-092', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Foreign Aid',
        'Wealthy countries have a moral duty to give a fixed share of income as foreign aid.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-092', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Foreign Aid',
        'Wealthy countries have a moral duty to give a fixed share of income as foreign aid.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-093', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'International Trade',
        'Free trade agreements benefit large corporations far more than ordinary people.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-093', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'International Trade',
        'Free trade agreements benefit large corporations far more than ordinary people.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-094', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Immigration Policy',
        'Countries should select immigrants primarily on the basis of skills.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-094', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Immigration Policy',
        'Countries should select immigrants primarily on the basis of skills.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-095', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Global Taxation',
        'Multinational companies should pay tax in the countries where they make their sales.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-095', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Global Taxation',
        'Multinational companies should pay tax in the countries where they make their sales.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-096', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'International Organisations',
        'Global problems can only be solved by international institutions, not individual governments.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-096', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'International Organisations',
        'Global problems can only be solved by international institutions, not individual governments.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-097', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Voting Age',
        'The voting age should be lowered to sixteen.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-097', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Voting Age',
        'The voting age should be lowered to sixteen.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-098', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Compulsory Voting',
        'Voting in national elections should be a legal obligation.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'hard',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-098', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Compulsory Voting',
        'Voting in national elections should be a legal obligation.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'hard',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-099', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Political Advertising',
        'Paid political advertising should be banned during election campaigns.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'medium',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-099', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Political Advertising',
        'Paid political advertising should be banned during election campaigns.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'medium',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-op-100', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree',
        'Term Limits',
        'Every elected leader should face a strict limit on time in office.

To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.

Write at least 250 words.',
        NULL, NULL, 0, 2400, 'easy',
        ARRAY['IELTS Writing', 'Task 2', 'Opinion'], 25),

    ('pte-wrt-es-100', 'pte-2026-01', 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
        'Term Limits',
        'Every elected leader should face a strict limit on time in office.

Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.

Write 200-300 words.',
        NULL, NULL, 0, 1200, 'easy',
        ARRAY['PTE Writing', 'Write Essay', 'Opinion'], 15),

    ('ielts-wrt-fg-001', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Internet Users in Three Countries',
        'The line graph shows the percentage of the population using the internet in three countries between 2000 and 2020.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-01-internet-users.svg', 'Line graph. Percentage of population using the internet, 2000-2020, measured every five years. Canada: 2000 = 51%, 2005 = 72%, 2010 = 80%, 2015 = 88%, 2020 = 94%. Mexico: 2000 = 5%, 2005 = 17%, 2010 = 31%, 2015 = 57%, 2020 = 72%. India: 2000 = 1%, 2005 = 2%, 2010 = 7%, 2015 = 27%, 2020 = 43%. Key features: Canada is highest throughout but grows the least in absolute terms; India starts lowest and grows fastest after 2010; the gap between Canada and India narrows from 50 to 51 points but the ratio falls sharply.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-002', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Household Spending by Category',
        'The bar chart shows how an average household divided its monthly spending in 2005 and in 2025.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-02-household-spending.svg', 'Bar chart, percentage of monthly household spending, 2005 versus 2025. Housing: 2005 = 24%, 2025 = 35%. Food: 2005 = 27%, 2025 = 18%. Transport: 2005 = 16%, 2025 = 13%. Healthcare: 2005 = 8%, 2025 = 12%. Leisure: 2005 = 13%, 2025 = 11%. Other: 2005 = 12%, 2025 = 11%. Key features: housing overtakes food as the largest category; food falls by 9 points, the largest decline; healthcare is the only other category to rise.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-003', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Sources of Electricity Generation',
        'The two pie charts show how electricity was generated in one country in 1990 and in 2020.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-03-electricity-pies.svg', 'Two pie charts, share of electricity generation. 1990: coal 52%, gas 14%, nuclear 20%, hydro 11%, wind and solar 3%. 2020: coal 18%, gas 27%, nuclear 15%, hydro 12%, wind and solar 28%. Key features: coal falls from just over half to under a fifth; wind and solar rise almost tenfold and become the second largest source; hydro is almost unchanged.', 0, 1200, 'hard',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-004', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'International Tourist Arrivals',
        'The table shows the number of international tourist arrivals, in millions, in five regions in 2010, 2015 and 2022.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-04-tourist-arrivals.svg', 'Table, international tourist arrivals in millions. Europe: 2010 = 489, 2015 = 605, 2022 = 585. Asia-Pacific: 2010 = 208, 2015 = 284, 2022 = 84. Americas: 2010 = 150, 2015 = 193, 2022 = 156. Africa: 2010 = 50, 2015 = 53, 2022 = 45. Middle East: 2010 = 55, 2015 = 58, 2022 = 78. Key features: Europe dominates in every year; Asia-Pacific grows strongly to 2015 then collapses to under a third of its 2010 figure; the Middle East is the only region above its 2015 level by 2022.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-005', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Average Monthly Temperature',
        'The line graph shows the average monthly temperature in two cities over the course of one year.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-05-temperature.svg', 'Line graph, average monthly temperature in degrees Celsius, January to December. Oslo: Jan -4, Feb -3, Mar 1, Apr 6, May 12, Jun 16, Jul 18, Aug 17, Sep 12, Oct 6, Nov 1, Dec -3. Singapore: Jan 26, Feb 27, Mar 28, Apr 28, May 29, Jun 28, Jul 28, Aug 28, Sep 28, Oct 27, Nov 27, Dec 26. Key features: Singapore is warmer in every month and varies by only 3 degrees across the year; Oslo swings 22 degrees between January and July; the two lines never converge.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-006', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'University Enrolment by Subject',
        'The bar chart shows the number of men and women enrolled in five university subjects in one year.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-06-enrolment.svg', 'Grouped bar chart, enrolment in thousands, men versus women. Engineering: men 82, women 24. Medicine: men 41, women 59. Business: men 66, women 61. Education: men 18, women 72. Computer Science: men 74, women 21. Key features: engineering and computer science are heavily male; education is the most female-dominated; business is close to even; only medicine and education have more women than men.', 0, 1200, 'hard',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-007', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'The Glass Recycling Process',
        'The diagram shows the process by which used glass bottles are recycled into new containers.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-07-glass-recycling.svg', 'Process diagram with seven stages, one cycle. Stage 1: used bottles collected in kerbside bins. Stage 2: transported by lorry to a sorting facility. Stage 3: sorted by colour into clear, green and brown. Stage 4: washed to remove labels, caps and residue. Stage 5: crushed into small fragments called cullet. Stage 6: cullet melted in a furnace at 1,500 degrees Celsius with sand and soda ash. Stage 7: molten glass moulded into new containers, which then re-enter distribution. Key features: the process is cyclical rather than linear; colour sorting happens before washing; melting is the only stage requiring extreme heat.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-008', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Redevelopment of a Town Centre',
        'The two maps show the centre of the town of Halton in 1990 and after redevelopment in 2020.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-08-town-centre.svg', 'Two maps of the same town centre, 1990 versus 2020. 1990: a factory occupies the north-east; a car park sits in the south; a railway station is in the west; housing runs along the eastern edge; the river crosses the south of the site with no bridge. 2020: the factory has been demolished and replaced by a public park; the car park has become a shopping centre; the railway station remains and has been extended; the housing has been retained and expanded northwards; a pedestrian bridge now crosses the river. Key features: industrial land use is entirely replaced by leisure and retail; the station is the only structure unchanged in position; accessibility improves with the new bridge.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-009', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Composition of Household Waste',
        'The stacked bar chart shows the composition of household waste in four countries.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-09-waste-composition.svg', 'Stacked bar chart, percentage composition of household waste by country. Japan: organic 30, paper 34, plastic 16, glass 8, metal 6, other 6. Germany: organic 34, paper 30, plastic 14, glass 12, metal 5, other 5. Brazil: organic 52, paper 18, plastic 17, glass 5, metal 4, other 4. Kenya: organic 65, paper 12, plastic 13, glass 4, metal 3, other 3. Key features: organic waste is the largest fraction everywhere and rises as national income falls; paper is highest in Japan at 34%; plastic is remarkably stable at 13-17% across all four.', 0, 1200, 'hard',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20),

    ('ielts-wrt-fg-010', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task1-figure', 'Describe the figure',
        'Urban and Rural Population',
        'The line graph shows the urban and rural population of one country between 1960 and 2020, with a projection to 2040.

Summarise the information by selecting and reporting the main features, and make comparisons where relevant.

Write at least 150 words.',
        '/images/writing/figure-10-urban-rural.svg', 'Line graph, population in millions, 1960-2040 (2040 is a projection shown as a dashed line). Urban: 1960 = 8, 1980 = 17, 2000 = 34, 2020 = 55, 2040 projected = 71. Rural: 1960 = 32, 1980 = 39, 2000 = 41, 2020 = 34, 2040 projected = 25. Key features: the two lines cross between 2000 and 2020; rural population peaks around 2000 and then declines; urban population grows in every period and is projected to be nearly three times the rural figure by 2040.', 0, 1200, 'medium',
        ARRAY['IELTS Writing', 'Task 1', 'Describe the figure'], 20)
ON CONFLICT (id) DO NOTHING;
