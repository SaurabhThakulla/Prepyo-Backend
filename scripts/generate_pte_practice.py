"""Deterministic original PTE practice bank. Run with Python (stdlib only).
Content is not Pearson material; audio uses labelled browser speech synthesis.
Reading/listening drill timers are local budgets, not official per-item limits.
"""
import json
from collections import Counter
from pathlib import Path
from urllib.parse import quote
from html import escape
from pte_practice_content import TOPICS, REORDERS, ESSAYS, DISCUSSIONS

ROOT = Path(__file__).resolve().parents[1]
rows, passages, groups, reorders = [], [], [], []

def add(skill, tid, name, title, prompt, seconds, **extra):
    row = dict(id=f'pte-expansion-{len(rows)+1:03}', exam_version_id='pte-2026-01',
               exam='PTE', supported_exams=['PTE'], skill=skill, type_id=tid,
               type_name=name, title=title, prompt=prompt, time_limit_seconds=seconds,
               prep_time_seconds=0, group_position=0, points=10, difficulty='medium',
               tags=['PTE', 'Original practice', 'Expansion 2026'])
    row.update(extra)
    rows.append(row)
    return row

def options(texts):
    return [dict(id=chr(65+i), text=text) for i, text in enumerate(texts)]

def image(title, labels, values):
    svg = '<svg xmlns="http://www.w3.org/2000/svg" width="760" height="380" viewBox="0 0 760 380"><rect width="760" height="380" fill="white"/>'
    svg += f'<g fill="#172554" font-family="Arial" font-size="18"><text x="20" y="30">{escape(title)}</text><text x="20" y="58">Fictional survey: percentage of respondents</text>'
    for i, (label, value) in enumerate(zip(labels, values)):
        y = 95+i*60
        svg += f'<text x="20" y="{y+20}">{escape(label)}</text><rect x="220" y="{y}" width="{value*4}" height="30" fill="#2563eb"/><text x="{230+value*4}" y="{y+22}">{value}%</text>'
    return 'data:image/svg+xml,'+quote(svg+'</g></svg>', safe='')

# Reading tasks are linked into the actual passage/reorder bank, not standalone rows.
for n, t in enumerate(TOPICS):
    for tid, name in [('fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks'),
                      ('fill-in-blanks-r', 'Reading: Fill in the Blanks'),
                      ('reading-mcq-single', 'Multiple Choice, Single Answer'),
                      ('reading-mcq-multiple', 'Multiple Choice, Multiple Answers')]:
        pid = f'rp-pte-expansion-{tid}-{n+1}'
        gid = f'g-pte-expansion-{tid}-{n+1}'
        is_gap = tid.startswith('fill-')
        prompt = ('Select the best word for each blank.' if is_gap else
                  'Which statements are supported by the passage? Select all correct answers.' if tid.endswith('multiple') else
                  'Which statement best expresses the main point of the passage?')
        text, blanks = t['text'], []
        for i, choices in enumerate(t['gaps']):
            key = f'b{i+1}'
            text = text.replace(choices[0], '[['+key+']]', 1)
            pool = list(choices)
            pool = pool[(n+i)%4:]+pool[:(n+i)%4]
            blanks.append(dict(id=key, options=pool, correctAnswer=choices[0]))
        resources = []
        if tid == 'fill-in-blanks-r':
            pool = sorted({w for b in blanks for w in b['options']})
            for b in blanks:
                b['options'] = pool
            resources = options(pool)
        passages.append(dict(id=pid, exam_version_id='pte-2026-01', title=t['title'],
            paragraphs=[dict(label='A', text=t['text'])], sources=[], word_count=0,
            difficulty='medium', topic='Community and research', tags=['PTE', 'Original practice'],
            is_published=True, passage_slot='custom'))
        groups.append(dict(id=gid, passage_id=pid, position=1, type_id=tid, type_name=name,
            instructions=prompt, resources=resources, passage_display='hidden' if is_gap else 'full',
            shuffle_questions=False, time_limit_seconds=180))
        extra = dict(passage_id=pid, group_id=gid, group_position=1)
        if is_gap:
            extra.update(context_passage=text, blanks=blanks, points=3)
            explanation = 'In context, the missing words are '+', '.join(b['correctAnswer'] for b in blanks)+'.'
        else:
            values = [t['main']]+t['wrong'] if tid.endswith('single') else [t['facts'][0], t['false'][0], t['facts'][1], t['false'][1]]
            keyed = list(zip(values, [True, False, False, False] if tid.endswith('single') else [True, False, True, False]))
            keyed = keyed[n%4:]+keyed[:n%4]
            extra.update(options=options([x[0] for x in keyed]), correct_answers=[chr(65+i) for i, x in enumerate(keyed) if x[1]], points=1 if tid.endswith('single') else 2)
            explanation = t['main']+' '+' '.join(t['facts'])
        add('reading', tid, name, t['title'], prompt, 180, explanation=explanation, **extra)

for n, (title, sentences) in enumerate(REORDERS):
    rid = f'ri-pte-expansion-{n+1}'
    labels = ['p3', 'p1', 'p4', 'p2']
    paragraphs = [dict(label=label, text=text) for label, text in zip(labels, sentences)]
    reorders.append(dict(id=rid, exam_version_id='pte-2026-01', exam='PTE', title=title,
        paragraphs=paragraphs, topic='Practical investigations', word_count=len(' '.join(sentences).split()),
        difficulty='medium', tags=['PTE', 'Original practice'], is_published=True))
    opts = [dict(id=p['label'], text=p['text']) for p in paragraphs]
    add('reading', 'reorder-paragraphs', 'Re-order Paragraphs', title,
        'Restore the logical order of the text boxes.', 180, reorder_item_id=rid,
        options=[opts[i] for i in [2,0,3,1]], correct_answers=labels, points=3,
        explanation='Introduce the project, describe its method, report the observation, then show the response to that observation.')

for n, t in enumerate(TOPICS):
    for tid, name, multi in [('pte-listening-mcma','MCQ Multiple Answer',True),
                              ('pte-listening-mcsa','MCQ Single Answer',False),
                              ('pte-highlight-correct-summary','Highlight Correct Summary',False)]:
        values = [t['facts'][0], t['false'][0], t['facts'][1], t['false'][1]] if multi else [t['summary'] if 'summary' in tid else t['main']]+t['wrong']
        keys = [True, False, True, False] if multi else [True, False, False, False]
        pairs = list(zip(values, keys))
        pairs = pairs[n%4:]+pairs[:n%4]
        add('listening', tid, name, t['title'],
            'Listen and select all supported statements.' if multi else 'Listen and select the best summary.' if 'summary' in tid else 'What is the main point of the recording?',
            150, audio_transcript=t['text'], options=options([p[0] for p in pairs]),
            correct_answers=[chr(65+i) for i, p in enumerate(pairs) if p[1]],
            explanation=t['summary'], points=2 if multi else 1)
    text, blanks = t['text'], []
    for i, choices in enumerate(t['gaps']):
        bid = f'b{i+1}'
        text = text.replace(choices[0], '[['+bid+']]', 1)
        blanks.append(dict(id=bid, options=[], correctAnswer=choices[0]))
    add('listening', 'pte-listening-fib', 'Listening Fill in the Blanks', t['title'],
        'Listen and type the missing word in each blank.', 150, context_passage=text,
        audio_transcript=t['text'], blanks=blanks, correct_answers=[b['correctAnswer'] for b in blanks],
        explanation='The recording says: '+', '.join(b['correctAnswer'] for b in blanks)+'.', points=3)
    words = t['text'].split()
    incorrect = []
    for choices in t['gaps']:
        i = next(i for i, w in enumerate(words) if w.strip('.,;:!?') == choices[0])
        words[i] = words[i].replace(choices[0], choices[1])
        incorrect.append(f'w{i+1}')
    add('listening', 'pte-highlight-incorrect-word', 'Highlight Incorrect Word', t['title'],
        'Listen and select words in the displayed transcript that differ from the recording.', 150,
        audio_transcript=t['text'], context_passage=' '.join(words),
        options=[dict(id=f'w{i+1}', text=w) for i, w in enumerate(words)], correct_answers=incorrect,
        explanation='The recording uses '+', '.join(c[0] for c in t['gaps'])+' instead of the substituted words.', points=3)
    choices = [t['missing'], 'permission', 'competition', 'decoration']
    choices = choices[n%4:]+choices[:n%4]
    add('listening', 'pte-select-missing-word', 'Select Missing Word', t['title'],
        'The final word is omitted from this practice recording. Select the word that completes the speaker\'s meaning.', 120,
        audio_transcript=t['text']+' '+t['ending']+' ...', options=options(choices),
        correct_answers=[chr(65+choices.index(t['missing']))],
        explanation='The conclusion ends with '+t['missing']+'. This practice uses a pause, not a recorded beep.', points=1)
    add('listening', 'write-from-dictation', 'Write from Dictation', t['title'],
        'Listen to the sentence and type exactly what you hear.', 60,
        audio_transcript=t['dictation'], correct_answers=[t['dictation']],
        explanation='Reference sentence: '+t['dictation'], points=len(t['dictation'].split()))
    # The existing summary grader is a keyword/form practice heuristic, not Pearson scoring.
    model = t['summary']+' The speaker illustrates why a simple measure can be misleading and argues that decisions should reflect the purpose of the activity as well as the limitations of the available evidence.'
    assert 50 <= len(model.split()) <= 70
    add('listening', 'summarize-spoken-text', 'Summarize Spoken Text', t['title'],
        'Listen to the talk and summarise its main points in 50-70 words. You have 10 minutes. Practice feedback is an approximate content/form check, not an official PTE score.', 600,
        audio_transcript=t['text']+' '+t['extra'], model_answer=model,
        correct_answers=t['keywords'], explanation=t['summary'], points=10)

for n, t in enumerate(TOPICS):
    add('writing', 'summarize-written-text', 'Summarize Written Text', t['title'],
        'Summarise the passage in ONE sentence of 5-75 words. You have 10 minutes.', 600,
        context_passage=t['text']+' '+t['extra'], model_answer=t['summary'], explanation=t['summary'])
    aloud = '. '.join(t['text'].split('. ')[:3])+'.'
    assert len(aloud.split()) <= 60
    add('speaking', 'pte-read-aloud', 'Read Aloud', t['title'],
        'Prepare for 35 seconds, then read the passage aloud clearly and naturally.', 40,
        prep_time_seconds=35, context_passage=aloud, explanation='Read all words accurately; group words into meaningful phrases rather than pausing after each word.')
    add('speaking', 'pte-repeat-sentence', 'Repeat Sentence', t['title'],
        'Listen, then repeat the sentence exactly as spoken.', 15, audio_transcript=t['dictation'],
        model_answer=t['dictation'], explanation='Preserve the original word sequence and grammatical endings.')
    add('speaking', 'pte-answer-short-question', 'Answer Short Questions', t['title'],
        'Listen to the question and answer with one or a few words.', 10,
        audio_transcript=t['short'], model_answer=t['answer'], explanation='Expected answer: '+t['answer']+'. Equivalent natural phrasing is acceptable.')
    add('speaking', 'pte-retell-lecture', 'Re-tell Lecture', t['title'],
        'Listen to the lecture. After 10 seconds of preparation, retell the main points in your own words in 40 seconds.', 40,
        prep_time_seconds=10, audio_transcript=t['text'], model_answer=t['summary'], explanation=t['summary'])
    labels = ['Very useful','Somewhat useful','Not very useful','Not useful']
    values = [[45,30,15,10],[38,34,20,8],[52,28,12,8],[41,35,18,6],[36,40,16,8]][n]
    caption = 'Fictional survey on '+t['title'].lower()
    add('speaking', 'pte-describe-image', 'Describe Image', caption,
        'Study the bar chart for 25 seconds. Describe its main features and comparisons in 40 seconds. Data are fictional.', 40,
        prep_time_seconds=25, image_url=image(caption, labels, values),
        figure_data='; '.join(f'{label}: {value}%' for label,value in zip(labels,values)),
        explanation='Identify the survey topic, compare the largest and smallest categories, and note that positive responses total '+str(sum(values[:2]))+'%.',
        model_answer=f'The chart presents responses to a fictional survey about {t["title"].lower()}. {values[0]} percent considered it very useful and {values[1]} percent somewhat useful. The two less positive categories accounted for {values[2]} and {values[3]} percent. Overall, positive responses made up {sum(values[:2])} percent of the total.')

for n, (title, statement, guidance) in enumerate(ESSAYS):
    tid, name = ('pte-write-essay','Write Essay') if n < 5 else ('pte-repeated-essay-questions','Repeated Essay Questions')
    add('writing', tid, name, title,
        statement+' To what extent do you agree or disagree? Support your view with reasons and examples. Write 200-300 words in 20 minutes.'+(' This is an original practice prompt, not a recalled exam question.' if n >= 5 else ''),
        1200, explanation=guidance+' State and sustain a clear position. There is no preferred opinion.', points=15)

for title, *turns in DISCUSSIONS:
    transcript = ' '.join(f'{speaker}: {turn}' for speaker, turn in zip(['Alex','Blair','Casey','Blair','Alex'],turns))
    add('speaking', 'pte-summarise-group-discussion', 'Summarise Group Discussion', title,
        'Listen to three students discussing a proposal. After 10 seconds of preparation, summarise their views and how the discussion develops. Speak for up to 2 minutes. This is a shortened synthetic-voice practice discussion, not an exam recording.',
        120, prep_time_seconds=10, audio_transcript=transcript,
        explanation='Include the initial proposal, Blair\'s concern, Casey\'s compromise and the agreed next step. Attribute views accurately rather than listing disconnected keywords.',
        model_answer='Alex proposes: '+turns[0]+' Blair raises a concern: '+turns[1]+' Casey suggests a compromise: '+turns[2]+' The agreed next step is: '+turns[4])

situations = [
 ('Overlapping deadlines', 'You have two assignments due on Friday and your tutor has offered a meeting on Thursday. You need advice before you can finish one assignment. Call the tutor, explain the conflict and politely ask whether an earlier meeting is possible.', 'Explain the deadline, request an earlier appointment and offer flexibility without demanding an extension.'),
 ('Incorrect delivery', 'You ordered ten folders for a student event tomorrow, but only six arrived. Speak to the supplier, explain what is missing and ask for a practical solution before the event.', 'State the quantity ordered and received, mention tomorrow\'s event and request delivery of the remaining four or an alternative.'),
 ('Shared kitchen', 'Your flatmate often leaves dishes in the shared sink overnight, making it difficult for you to prepare breakfast. Speak to your flatmate politely, describe the problem and suggest an arrangement.', 'Use a respectful tone, explain the practical effect and propose an agreed cleaning routine.'),
 ('Room change', 'You organise a study group of eight people. The library has moved your booking to a room with only four chairs. Speak to the librarian and ask for a suitable arrangement without disturbing other users.', 'Explain the booking size, identify the seating problem and ask about another room or additional seating.'),
 ('Missed shift', 'Your train has been cancelled and you will arrive thirty minutes late for your volunteer shift. Phone the coordinator, explain the delay and ask how you can minimise the disruption.', 'Give timely notice, apologise, estimate arrival and offer a practical way to help rather than making excuses.'),
]
for title, situation, guidance in situations:
    add('speaking', 'pte-respond-to-situation', 'Respond to a Situation', title,
        situation+' Prepare for 10 seconds, then respond directly to the person for up to 40 seconds.', 40,
        prep_time_seconds=10, context_passage=situation, audio_transcript=situation, explanation=guidance)


# EMIT

def sql(value, column):
    if value is None:
        return 'NULL'
    if isinstance(value, bool):
        return 'TRUE' if value else 'FALSE'
    if column in ('tags', 'supported_exams'):
        return 'ARRAY['+','.join("'"+v.replace("'", "''")+"'" for v in value)+']'
    if isinstance(value, (dict, list)):
        return "'"+json.dumps(value, ensure_ascii=False).replace("'", "''")+"'::jsonb"
    if isinstance(value, int):
        return str(value)
    return "'"+value.replace("'", "''")+"'"

def insert(table, records):
    columns = list(dict.fromkeys(k for row in records for k in row))
    return 'INSERT INTO '+table+' ('+', '.join(columns)+') VALUES\n'+',\n'.join(
        '('+', '.join(sql(row.get(k), k) for k in columns)+')' for row in records
    )+'\nON CONFLICT (id) DO NOTHING;\n'

if __name__ == '__main__':
    counts = Counter(r['type_id'] for r in rows)
    assert len(rows) == 115 and len(counts) == 23 and set(counts.values()) == {5}, counts
    assert len({r['id'] for r in rows}) == 115
    for r in rows:
        assert r.get('explanation'), r['id']
        if r['skill'] == 'listening':
            assert r.get('audio_transcript') and r.get('correct_answers')
        if r.get('options'):
            assert set(r['correct_answers']) <= {o['id'] for o in r['options']}
        for b in r.get('blanks', []):
            assert '[['+b['id']+']]' in r['context_passage']
            assert not b['options'] or b['correctAnswer'] in b['options']
    body = '-- Generated by scripts/generate_pte_practice.py; original, unofficial practice.\n-- 23 categories x 5 tasks. Synthetic playback; practice scores are not official PTE scores.\n'
    for table, records in [('reading_passages', passages), ('reading_question_groups', groups),
                           ('reading_reorder_items', reorders), ('questions', rows)]:
        body += insert(table, records)
    (ROOT/'migrations/000053_pte_practice_expansion.up.sql').write_text(body, encoding='utf-8')
    (ROOT/'migrations/000053_pte_practice_expansion.down.sql').write_text(
        '-- Preserve learner references and editorial changes on rollback.\nSELECT 1;\n', encoding='utf-8')
    print(dict(counts))
