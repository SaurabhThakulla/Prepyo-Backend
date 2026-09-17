"""Generate original IELTS practice tasks and inline SVGs using only stdlib.
Run from any directory. Output is deterministic; existing DB rows are preserved.
Synthetic browser playback is practice-only, not an authentic exam recording.
"""
import json
from pathlib import Path
from urllib.parse import quote
from html import escape

ROOT = Path(__file__).resolve().parents[1]
rows = []

def add(skill, type_id, name, title, prompt, seconds, **extra):
    row = dict(id=f'ielts-audit-{len(rows)+1:03}', exam_version_id='ielts-2026-01',
               exam='IELTS', supported_exams=['IELTS'], skill=skill, type_id=type_id,
               type_name=name, title=title, prompt=prompt, time_limit_seconds=seconds,
               prep_time_seconds=0, points=1, difficulty='medium', tags=['Original practice', 'IELTS'])
    row.update(extra)
    rows.append(row)

def svg_image(title, lines):
    body = ''.join(f'<text x="30" y="{90+i*36}">{escape(line)}</text>' for i, line in enumerate(lines))
    svg = f'<svg xmlns="http://www.w3.org/2000/svg" width="760" height="{130+len(lines)*36}" viewBox="0 0 760 {130+len(lines)*36}"><rect width="100%" height="100%" fill="white"/><g fill="#172554" font-family="Arial" font-size="18"><text x="30" y="40" font-weight="bold">{escape(title)}</text>{body}</g></svg>'
    return 'data:image/svg+xml,' + quote(svg, safe='')

# Task 1 tables: invented data explicitly labelled, each supports comparisons.
figures = [
 ('Library visits by age', 'Annual visits (thousands), fictional city', ['Age group | 2010 | 2020', 'Under 18 | 45 | 62', '18-64 | 110 | 84', '65 and over | 28 | 53'], 'Total visits rose from 183 to 199 thousand. Working-age visits fell while both other groups increased.'),
 ('Household energy use', 'Share of energy use (%), fictional region', ['Purpose | 2005 | 2025', 'Heating | 55 | 42', 'Hot water | 20 | 23', 'Appliances | 15 | 25', 'Lighting | 10 | 10'], 'Heating remained largest but fell 13 percentage points. Appliances rose 10 points and lighting was unchanged. Each column totals 100%.'),
 ('Commuting in Eastford', 'Share of commuters (%), fictional town', ['Mode | 2015 | 2025', 'Car | 60 | 42', 'Bus | 20 | 25', 'Bicycle | 8 | 18', 'Walking | 12 | 15'], 'Cars remained largest but lost share. All other modes gained; cycling more than doubled. Columns total 100%.'),
 ('Museum attendance', 'Annual admissions (thousands), fictional museums', ['Museum | 2018 | 2023', 'History | 120 | 145', 'Science | 95 | 160', 'Art | 140 | 125', 'Transport | 55 | 80'], 'Science overtook art to lead. Art alone declined. Total admissions increased from 410 to 510 thousand.'),
 ('Water consumption by sector', 'Million cubic metres, fictional district', ['Sector | 2000 | 2020', 'Agriculture | 80 | 65', 'Industry | 45 | 30', 'Households | 25 | 40'], 'Overall use fell from 150 to 135 million cubic metres. Agriculture remained largest. Household use rose as the other sectors fell.'),
]
for title, subtitle, lines, guidance in figures:
    add('writing', 'ielts-writing-task1-figure', 'Describe the figure', title,
        'The table shows '+subtitle.lower()+'. Summarise the information by selecting and reporting the main features, and make comparisons where relevant. Write at least 150 words. Suggested time: 20 minutes.', 1200,
        image_url=svg_image(title, [subtitle]+lines), figure_data='\n'.join([subtitle]+lines+[guidance]), explanation=guidance, points=20)

opinions = [
 ('Public libraries and digital access', 'Public libraries should spend more money on digital resources than on printed books.'),
 ('Repairing household products', 'Governments should require manufacturers to make household products easy to repair.'),
 ('University community service', 'All university students should complete unpaid community service before graduating.'),
 ('Quiet spaces in cities', 'Providing quiet public spaces is as important as building sports facilities in cities.'),
 ('Flexible working hours', 'Employers should allow employees to choose their working hours whenever the job permits it.'),
]
for title, statement in opinions:
    add('writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree', title,
        statement+' To what extent do you agree or disagree? Give reasons for your answer and include relevant examples from your own knowledge or experience. Write at least 250 words. Suggested time: 40 minutes.', 2400,
        explanation='State a clear position, develop relevant reasons and examples, address qualifications where useful, and sustain the position through the conclusion. There is no preferred opinion.', points=25)


part1 = [
 ('Your neighbourhood', 'What do you like about the area where you live? Has it changed since you first lived there?'),
 ('Cooking at home', 'Do you enjoy cooking? What meal would you like to learn to prepare, and why?'),
 ('Daily travel', 'How do you usually travel to work or study? What do you like or dislike about that journey?'),
 ('Reading for pleasure', 'What do you enjoy reading? Have your reading habits changed since childhood?'),
 ('Spending time outdoors', 'Where do you like to spend time outdoors? How often do you go there?'),
]
for title, questions in part1:
    add('speaking', 'ielts-speaking-part1', 'Introduction', title,
        questions+' Give natural answers with relevant detail. This is a short Part 1 practice drill, not a complete 4–5 minute interview.', 90, points=15)
part2 = [
 ('A useful skill', 'Describe a useful skill you learned outside school.', 'what the skill is; who or what helped you learn it; how you have used it; and explain why it matters to you.'),
 ('A peaceful place', 'Describe a peaceful place you have visited.', 'where it is; when you went there; what you did there; and explain why you found it peaceful.'),
 ('A helpful person', 'Describe a person who helped you solve a problem.', 'who the person is; what the problem was; how the person helped; and explain what you learned from the experience.'),
 ('An enjoyable journey', 'Describe a journey you enjoyed.', 'where you went; how you travelled; who you travelled with; and explain what made the journey enjoyable.'),
 ('Something you repaired', 'Describe something you repaired or had repaired.', 'what the object was; what was wrong with it; how it was repaired; and explain how you felt about the result.'),
]
for title, topic, bullets in part2:
    add('speaking', 'ielts-speaking-part2', 'Speaking Part 2 (Cue Card)', title,
        topic+' You should say: '+bullets+' You have one minute to prepare and may make notes. Speak for one to two minutes.', 120, prep_time_seconds=60, points=15)
part3 = [
 ('Learning throughout life', 'Why do adults continue learning new skills? Should employers or individuals pay for this learning?'),
 ('Protecting quiet places', 'Why can quiet public spaces be valuable in cities? How should cities balance development with preserving such places?'),
 ('Helping others', 'Why do some people volunteer more than others? Should schools encourage volunteering, and how?'),
 ('Travel and society', 'How can tourism benefit a local community? What problems can rapid growth in visitor numbers create?'),
 ('Repair or replace', 'Why do people often replace products rather than repair them? How could businesses and governments encourage repair?'),
]
for title, questions in part3:
    add('speaking', 'ielts-speaking-part3', 'Speaking Part 3 (Discussion)', title,
        questions+' Discuss reasons, comparisons and examples. This is a two-minute discussion drill; a full examiner-led Part 3 lasts 4–5 minutes.', 120, points=15)


completion = [
 ('Evening pottery class', 'Receptionist: The daytime pottery course is full, but the evening class has places. Caller: Which day is it? Receptionist: It used to be on Tuesday; from next month it will be Thursday. Caller: What should I bring? Receptionist: An apron. Clay and tools are included.', 'Class day: ____ (1)\nBring an: ____ (2)', ['Thursday', 'apron']),
 ('Coastal field trip', 'Tutor: Our field trip will investigate erosion, not the fishing industry as originally planned. Meet beside the library rather than at the laboratory. Wear boots because the path may be wet. We will provide recording sheets.', 'Research topic: ____ (1)\nMeeting point: the ____ (2)', ['erosion', 'library']),
 ('Community garden', 'Coordinator: The garden needs volunteers to water young plants. Watering takes place in the morning, before the heat increases. Use the entrance beside the bakery; the gate near the school is locked during repairs.', 'Watering time: ____ (1)\nEntrance beside the: ____ (2)', ['morning', 'bakery']),
 ('Student project planning', 'Student: We first thought of interviewing shop owners. Tutor: That could work, but your question concerns customers. Student: Then we will use questionnaires to reach more customers. Tutor: Good. Test the questions at the market before collecting data elsewhere.', 'Chosen research method: ____ (1)\nPilot location: the ____ (2)', ['questionnaires', 'market']),
 ('Urban tree research', 'Lecturer: Tree shade can lower surface temperatures, but our study measured noise. We compared streets with similar traffic. The greatest reduction occurred where trees were combined with hedges, rather than where trees stood alone.', 'Variable measured: ____ (1)\nTrees were most effective with: ____ (2)', ['noise', 'hedges']),
]
for title, script, context, keys in completion:
    add('listening', 'ielts-listening-fill-blanks', 'Fill in the blanks', title,
        'Listen and complete the notes. Write ONE WORD ONLY for each answer. This is a short practice drill with synthesised playback.', 180,
        audio_transcript=script, context_passage=context,
        blanks=[dict(id=str(i+1), correctAnswer=k) for i,k in enumerate(keys)], correct_answers=keys,
        explanation='The recording gives these answers in order: '+', '.join(keys)+'. Do not include extra words.', points=len(keys))

mcqs = [
 ('Museum visit', 'Guide: The main gallery is closed while its lighting is replaced. Some visitors thought the roof was damaged, but that work was finished last year. The new exhibition opens next month, after the lights have been tested.', 'Why is the main gallery closed?', ['Its roof is being repaired.', 'Its lighting is being replaced.', 'Its exhibition is being moved.'], 'B'),
 ('Choosing a research topic', 'Student: I wanted to study national transport policy, but the tutor said the scope was too wide for six weeks. I have narrowed it to bus reliability in one district. Access to passengers is straightforward, and the library has all the background material.', 'Why did the student change the topic?', ['The original topic was too broad.', 'There were too few sources.', 'Passengers refused interviews.'], 'A'),
 ('A walking tour', 'Organiser: Tomorrow we will start outside the station as usual. Rain is unlikely, so the route stays the same. However, the guide has a later appointment, so we need everyone there at nine instead of half past nine.', 'What has changed about tomorrow’s tour?', ['The meeting place.', 'The route.', 'The starting time.'], 'C'),
 ('Booking a study room', 'Librarian: You can reserve a room online. You do not need to pay a deposit or show a student card at reception. But you must confirm the reservation using the email we send; otherwise it expires after an hour.', 'What must a student do to keep a room reservation?', ['Pay a deposit.', 'Confirm by email.', 'Show a card at reception.'], 'B'),
 ('Testing packaging', 'Researcher: Both materials kept the food fresh equally well. The paper package was cheaper, but it tore during delivery. We selected the stronger plant-based package even though it cost more, because fewer items would be damaged.', 'Why was the plant-based package selected?', ['It was cheaper.', 'It kept food fresh longer.', 'It resisted damage better.'], 'C'),
]
for title, script, question, options, key in mcqs:
    add('listening', 'ielts-listening-mcq', 'Choose the correct answer', title,
        question+' Choose ONE letter, A, B or C. Synthesised practice audio.', 120,
        audio_transcript=script, options=[dict(id=chr(65+i), text=text) for i,text in enumerate(options)],
        correct_answers=[key], explanation='The speaker explicitly supports '+key+': '+options[ord(key)-65])

# Shared spatial layout, five distinct targets and directions. No target label
# is printed on the map: learners identify the letter from the recording.
def map_image(title):
    svg = '<svg xmlns="http://www.w3.org/2000/svg" width="640" height="470" viewBox="0 0 640 470"><rect width="640" height="470" fill="white"/><g font-family="Arial" font-size="20" fill="#172554"><text x="20" y="30">'+escape(title)+'</text><text x="555" y="70">N ↑</text><path d="M320 425 V95 M100 270 H540" stroke="#94a3b8" stroke-width="30" fill="none"/>'
    for label,x,y in [('A',140,110),('B',400,110),('C',140,195),('D',400,195),('E',140,320),('F',400,320)]:
        svg += f'<rect x="{x}" y="{y}" width="100" height="45" fill="#dbeafe" stroke="#1e3a8a"/><text x="{x+40}" y="{y+30}">{label}</text>'
    svg += '<text x="255" y="450">Entrance ↑</text></g></svg>'
    return 'data:image/svg+xml,'+quote(svg, safe='')

maps = [
 ('Community centre plan', 'the art room', 'Go north from the entrance and cross the east-west corridor. The art room is the first room on your left after the crossing, not the one at the far end.', 'C'),
 ('College welcome day', 'the advice desk', 'From the south entrance, walk north. Before you reach the east-west corridor, turn into the room on your right. That is where the advice desk is today.', 'F'),
 ('Sports centre plan', 'the equipment store', 'Walk north from the entrance. Pass the crossing and continue past the first pair of rooms. The equipment store is the last room on your right.', 'B'),
 ('Arts venue plan', 'the rehearsal room', 'Starting at the entrance on the south side, take the northbound corridor. The rehearsal room is on your left before the east-west crossing.', 'E'),
 ('Training centre plan', 'the computer room', 'Go north from the entrance, cross the east-west corridor, and pass the first rooms on both sides. The computer room is the last room on your left.', 'A'),
]
for title, target, script, key in maps:
    add('listening', 'ielts-listening-map', 'Navigating according to audio', title,
        'Locate '+target+' on the plan. Choose ONE letter, A–F. North is at the top; start at the south entrance. Synthesised practice audio.', 180,
        image_url=map_image(title), audio_transcript=script,
        options=[dict(id=k,text='Location '+k) for k in 'ABCDEF'], correct_answers=[key],
        explanation='Following the directions from the marked entrance leads to location '+key+'.')

# Verify catalogue coverage before emitting SQL; no placeholders or answerless
# deterministic tasks are allowed in the generated migration.
from collections import Counter
assert len(rows) == 40
assert set(Counter(row['type_id'] for row in rows).values()) == {5}
for row in rows:
    if row['skill'] == 'listening':
        assert row.get('audio_transcript') and row.get('correct_answers')
    if row['type_id'] == 'ielts-listening-map':
        assert row.get('image_url')

def sql(value, column):
    if value is None: return 'NULL'
    if column in ('supported_exams','tags'):
        return 'ARRAY['+','.join("'"+v.replace("'", "''")+"'" for v in value)+']'
    if isinstance(value, (list,dict)):
        return "'"+json.dumps(value, ensure_ascii=False).replace("'", "''")+"'::jsonb"
    if isinstance(value,int): return str(value)
    return "'"+value.replace("'", "''")+"'"

columns = list(dict.fromkeys(k for row in rows for k in row))
body = '-- Generated by scripts/generate_ielts_practice.py. Original practice content.\n-- Five new tasks for each IELTS non-reading sidebar subtask; no reading inserts.\n'
body += 'INSERT INTO questions ('+', '.join(columns)+') VALUES\n'
body += ',\n'.join('('+', '.join(sql(row.get(k),k) for k in columns)+')' for row in rows)
body += '\nON CONFLICT (id) DO NOTHING;\n'
(ROOT/'migrations/000052_ielts_practice_bank.up.sql').write_text(body, encoding='utf-8')
(ROOT/'migrations/000052_ielts_practice_bank.down.sql').write_text('-- Preserve authored content and attempt references on rollback.\nSELECT 1;\n', encoding='utf-8')
print('Generated 40 tasks: '+str(dict(Counter(row['type_id'] for row in rows))))

