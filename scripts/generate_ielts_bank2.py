"""Generate IELTS content batch 2 (migrations 000083-000085).

  000083  fifty new practice items for each IELTS Speaking, Writing and
          Listening sub-task (400 questions)
  000084  IELTS Listening Practice Tests 2-10 (four parts, forty questions each)
  000085  IELTS Speaking mock sets 5-10

All content is original Prepyo practice material written to the public IELTS
formats; none of it is official, recalled or copied test material, and the
Task 1 data describes invented places. The content lives in scripts/ielts_bank2;
this script checks it and writes SQL. Nothing is written if a check fails:
  - every listening answer is heard in its script, in question order, within
    the word limit, and every multiple-choice/matching answer has quoted
    evidence in the script;
  - every map direction names exactly one room on its plan;
  - pie charts total 100 and chart series match their axes;
  - counts are exactly fifty per practice sub-task, forty per test.
The new Listening tests and Speaking sets become numbered mock tests through
the mock paper builder (internal/mockpapers), which registers every published
test that has no number yet, oldest first.

Output is deterministic and replay-safe (ON CONFLICT DO NOTHING). Run from any
directory: python scripts/generate_ielts_bank2.py
"""
import json
import re
import sys
from collections import Counter
from html import escape
from pathlib import Path
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(Path(__file__).resolve().parent))

from ielts_bank2 import listening_map                                   # noqa: E402
from live_topics import topic_clashes                                   # noqa: E402
from ielts_bank2.common import W1, WN                                   # noqa: E402
from ielts_bank2.listening_fill import FILL                             # noqa: E402
from ielts_bank2.listening_mcq import MCQ                               # noqa: E402
from ielts_bank2.listening_tests_a import TESTS_A                       # noqa: E402
from ielts_bank2.listening_tests_b import TESTS_B                       # noqa: E402
from ielts_bank2.listening_tests_c import TESTS_C                       # noqa: E402
from ielts_bank2.speaking import PART1, PART2, PART3                    # noqa: E402
from ielts_bank2.speaking_sets import SETS                              # noqa: E402
from ielts_bank2.writing_task1 import FIGURES                           # noqa: E402
from ielts_bank2.writing_task2 import TASK2                             # noqa: E402

VERSION = 'ielts-2026-01'
PALETTE = ['#1d4ed8', '#f59e0b', '#059669', '#dc2626', '#7c3aed', '#0891b2']


def lit(value):
    if value is None:
        return 'NULL'
    if isinstance(value, bool):
        return 'TRUE' if value else 'FALSE'
    if isinstance(value, int):
        return str(value)
    if isinstance(value, (list, dict)):
        return "'" + json.dumps(value, ensure_ascii=False).replace("'", "''") + "'::jsonb"
    return "'" + str(value).replace("'", "''") + "'"


def text_array(values):
    return 'ARRAY[' + ', '.join(lit(v) for v in values) + ']'


def data_uri(svg):
    return 'data:image/svg+xml,' + quote(svg, safe='')


def spoken(answer, script):
    return re.search(r'(?<![\w])' + re.escape(answer.lower()) + r'(?![\w])', script.lower()) is not None


def numeric(answer):
    return re.fullmatch(r'[\d.:]+(st|nd|rd|th)?', answer) is not None


def fits(answer, limit):
    if limit == WN:
        return len(answer.split()) <= 1 or any(ch.isdigit() for ch in answer)
    return len(answer.split()) <= 1


# ------------------------------------------------------------------ charts

def nice_max(value):
    for step in (1, 2, 2.5, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000):
        top = step * 5
        if top >= value:
            return top, step
    step = 10 ** len(str(int(value)))
    return step, step / 5


def fmt(v):
    return f'{v:,.0f}' if float(v).is_integer() else f'{v:,.1f}'


def svg_frame(title, height=470, width=760):
    return [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">',
            f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
            '<g font-family="Arial" fill="#172554">',
            f'<text x="{width / 2}" y="28" font-size="18" font-weight="bold" text-anchor="middle">{escape(title)}</text>']


def legend(parts, names, x, y):
    for i, name in enumerate(names):
        parts.append(f'<rect x="{x}" y="{y + i * 20}" width="14" height="14" fill="{PALETTE[i % len(PALETTE)]}"/>')
        parts.append(f'<text x="{x + 20}" y="{y + i * 20 + 12}" font-size="13">{escape(name)}</text>')


def axes(parts, top, unit, x0=90, y0=60, w=520, h=310):
    for i in range(6):
        v = top * i / 5
        y = y0 + h - h * i / 5
        parts.append(f'<line x1="{x0}" y1="{y}" x2="{x0 + w}" y2="{y}" stroke="#e2e8f0"/>')
        parts.append(f'<text x="{x0 - 8}" y="{y + 4}" font-size="12" text-anchor="end">{fmt(v)}</text>')
    parts.append(f'<line x1="{x0}" y1="{y0}" x2="{x0}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<line x1="{x0}" y1="{y0 + h}" x2="{x0 + w}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<text x="22" y="{y0 + h / 2}" font-size="12" text-anchor="middle" transform="rotate(-90 22 {y0 + h / 2})">{escape(unit)}</text>')


def bar_svg(f):
    parts = svg_frame(f['title'])
    series = list(f['series'].items())
    top, _ = nice_max(max(max(v) for _, v in series))
    x0, y0, w, h = 90, 60, 520, 310
    axes(parts, top, f['unit'])
    cats = f['categories']
    group_w = w / len(cats)
    bar_w = min(38, (group_w - 16) / len(series))
    for c, cat in enumerate(cats):
        gx = x0 + group_w * c + (group_w - bar_w * len(series)) / 2
        for s, (_, values) in enumerate(series):
            bh = h * values[c] / top
            parts.append(f'<rect x="{gx + s * bar_w:.1f}" y="{y0 + h - bh:.1f}" width="{bar_w - 2:.1f}" height="{bh:.1f}" fill="{PALETTE[s]}"/>')
        parts.append(f'<text x="{x0 + group_w * c + group_w / 2:.1f}" y="{y0 + h + 20}" font-size="12" text-anchor="middle">{escape(cat)}</text>')
    legend(parts, [n for n, _ in series], 625, 70)
    parts.append('</g></svg>')
    return ''.join(parts)


def line_svg(f):
    parts = svg_frame(f['title'])
    series = list(f['series'].items())
    top, _ = nice_max(max(max(v) for _, v in series))
    x0, y0, w, h = 90, 60, 520, 310
    axes(parts, top, f['unit'])
    xs = f['x']
    step = w / (len(xs) - 1)
    for i, x in enumerate(xs):
        parts.append(f'<text x="{x0 + step * i:.1f}" y="{y0 + h + 20}" font-size="12" text-anchor="middle">{escape(str(x))}</text>')
    for s, (_, values) in enumerate(series):
        pts = ' '.join(f'{x0 + step * i:.1f},{y0 + h - h * v / top:.1f}' for i, v in enumerate(values))
        parts.append(f'<polyline points="{pts}" fill="none" stroke="{PALETTE[s]}" stroke-width="3"/>')
        for i, v in enumerate(values):
            parts.append(f'<circle cx="{x0 + step * i:.1f}" cy="{y0 + h - h * v / top:.1f}" r="4" fill="{PALETTE[s]}"/>')
    legend(parts, [n for n, _ in series], 625, 70)
    parts.append('</g></svg>')
    return ''.join(parts)


def pie_svg(f):
    import math
    pies = list(f['pies'].items())
    categories = list(dict.fromkeys(c for _, p in pies for c in p))
    colour = {c: PALETTE[i % len(PALETTE)] if i < len(PALETTE) else '#94a3b8' for i, c in enumerate(categories)}
    width = 760
    parts = svg_frame(f['title'], height=430, width=width)
    r = 105
    for n, (label, pie) in enumerate(pies):
        cx = 150 + n * 260 if len(pies) > 1 else 260
        cy = 200
        parts.append(f'<text x="{cx}" y="80" font-size="15" font-weight="bold" text-anchor="middle">{escape(label)}</text>')
        angle = -math.pi / 2
        for cat, pct in pie.items():
            sweep = 2 * math.pi * pct / 100
            x1, y1 = cx + r * math.cos(angle), cy + r * math.sin(angle)
            x2, y2 = cx + r * math.cos(angle + sweep), cy + r * math.sin(angle + sweep)
            large = 1 if sweep > math.pi else 0
            parts.append(f'<path d="M{cx},{cy} L{x1:.1f},{y1:.1f} A{r},{r} 0 {large} 1 {x2:.1f},{y2:.1f} Z" fill="{colour[cat]}" stroke="#ffffff" stroke-width="2"/>')
            mid = angle + sweep / 2
            parts.append(f'<text x="{cx + r * 0.65 * math.cos(mid):.1f}" y="{cy + r * 0.65 * math.sin(mid) + 4:.1f}" font-size="12" fill="#ffffff" font-weight="bold" text-anchor="middle">{pct}%</text>')
            angle += sweep
    ly = 330
    for i, cat in enumerate(categories):
        x = 40 + (i % 3) * 240
        y = ly + (i // 3) * 22
        parts.append(f'<rect x="{x}" y="{y}" width="14" height="14" fill="{colour[cat]}"/>')
        parts.append(f'<text x="{x + 20}" y="{y + 12}" font-size="13">{escape(cat)}</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


def table_svg(f):
    cols = len(f['header'])
    rows = [f['header']] + f['rows']
    col_w = 700 / cols
    height = 70 + 40 * len(rows) + 20
    parts = svg_frame(f['title'], height=height)
    for r, row in enumerate(rows):
        y = 50 + 40 * r
        fill = '#dbeafe' if r == 0 else ('#f8fafc' if r % 2 else '#ffffff')
        for c, cell in enumerate(row):
            x = 30 + col_w * c
            parts.append(f'<rect x="{x:.1f}" y="{y}" width="{col_w:.1f}" height="40" fill="{fill}" stroke="#94a3b8"/>')
            weight = 'bold' if r == 0 or c == 0 else 'normal'
            size = 12 if r == 0 else 14
            # Headers wrap onto two lines so a long one stays inside its cell.
            lines = wrap(cell, max(8, int(col_w / 7))) if r == 0 else [cell]
            for k, line in enumerate(lines[:2]):
                dy = 25 + (k - (len(lines[:2]) - 1) / 2) * 13
                parts.append(f'<text x="{x + col_w / 2:.1f}" y="{y + dy:.1f}" font-size="{size}" font-weight="{weight}" text-anchor="middle">{escape(line)}</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


def wrap(text, width):
    words, lines, line = text.split(), [], ''
    for word in words:
        if len(line) + len(word) + 1 > width and line:
            lines.append(line)
            line = word
        else:
            line = (line + ' ' + word).strip()
    lines.append(line)
    return lines


def process_svg(f):
    steps = f['steps']
    per_row = 4
    rows = (len(steps) + per_row - 1) // per_row
    height = 80 + rows * 150
    parts = svg_frame(f['title'], height=height)
    parts.append('<defs><marker id="a" markerWidth="10" markerHeight="10" refX="8" refY="5" orient="auto"><path d="M0,0 L10,5 L0,10 z" fill="#1e3a8a"/></marker></defs>')
    box_w, box_h, gap = 150, 90, 38
    for i, step in enumerate(steps):
        row, col = divmod(i, per_row)
        if row % 2:
            col = per_row - 1 - col          # snake so each arrow is short
        x = 30 + col * (box_w + gap)
        y = 60 + row * 150
        parts.append(f'<rect x="{x}" y="{y}" width="{box_w}" height="{box_h}" rx="10" fill="#eff6ff" stroke="#1e3a8a"/>')
        parts.append(f'<circle cx="{x + 16}" cy="{y + 16}" r="11" fill="#1e3a8a"/><text x="{x + 16}" y="{y + 20}" font-size="12" fill="#ffffff" text-anchor="middle">{i + 1}</text>')
        for k, line in enumerate(wrap(step, 20)[:4]):
            parts.append(f'<text x="{x + box_w / 2}" y="{y + 36 + k * 15}" font-size="12" text-anchor="middle">{escape(line)}</text>')
        if i + 1 < len(steps):
            nrow, ncol = divmod(i + 1, per_row)
            if nrow % 2:
                ncol = per_row - 1 - ncol
            if nrow == row:
                if ncol > col:
                    parts.append(f'<line x1="{x + box_w}" y1="{y + box_h / 2}" x2="{x + box_w + gap - 4}" y2="{y + box_h / 2}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
                else:
                    parts.append(f'<line x1="{x}" y1="{y + box_h / 2}" x2="{x - gap + 4}" y2="{y + box_h / 2}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
            else:
                parts.append(f'<line x1="{x + box_w / 2}" y1="{y + box_h}" x2="{x + box_w / 2}" y2="{y + 146}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
    parts.append('</g></svg>')
    return ''.join(parts)


MAP_FILL = {'water': '#bfdbfe', 'sand': '#fde68a', 'building': '#cbd5e1', 'road': '#e5e7eb', 'green': '#bbf7d0'}


def map_svg(f):
    panels = list(f['panels'].items())
    cell = 32
    parts = svg_frame(f['title'], height=380)
    for p, (label, areas) in enumerate(panels):
        ox, oy = 30 + p * 370, 70
        parts.append(f'<text x="{ox + 160}" y="{oy - 12}" font-size="15" font-weight="bold" text-anchor="middle">{escape(label)}</text>')
        parts.append(f'<rect x="{ox}" y="{oy}" width="{10 * cell}" height="{8 * cell}" fill="#f8fafc" stroke="#172554"/>')
        for name, x, y, w, h, kind in areas:
            parts.append(f'<rect x="{ox + x * cell}" y="{oy + y * cell}" width="{w * cell}" height="{h * cell}" fill="{MAP_FILL[kind]}" stroke="#475569"/>')
            lines = wrap(name, max(6, int(w * 3.2)))
            for k, line in enumerate(lines[:2]):
                parts.append(f'<text x="{ox + (x + w / 2) * cell}" y="{oy + (y + h / 2) * cell + 4 + (k - (len(lines[:2]) - 1) / 2) * 13}" font-size="11" text-anchor="middle">{escape(line)}</text>')
        parts.append(f'<text x="{ox + 10 * cell - 14}" y="{oy + 16}" font-size="13">N&#8593;</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


RENDER = {'bar': bar_svg, 'line': line_svg, 'pie': pie_svg, 'table': table_svg, 'process': process_svg, 'map': map_svg}
LEAD = {'bar': 'The bar chart below shows', 'line': 'The graph below shows', 'table': 'The table below shows',
        'process': 'The diagram below shows', 'map': 'The maps below show'}


def figure_data(f):
    lines = [f['title'], f['desc'][0].upper() + f['desc'][1:] + '.']
    kind = f['kind']
    if kind in ('bar', 'line'):
        axis = f['categories'] if kind == 'bar' else f['x']
        lines.append(f"Unit: {f['unit']}")
        lines.append(' | '.join(['Series'] + [str(a) for a in axis]))
        for name, values in f['series'].items():
            lines.append(' | '.join([name] + [fmt(v) for v in values]))
    elif kind == 'pie':
        for label, pie in f['pies'].items():
            lines.append(f'{label}: ' + ', '.join(f'{c} {p}%' for c, p in pie.items()))
    elif kind == 'table':
        for row in [f['header']] + f['rows']:
            lines.append(' | '.join(row))
    elif kind == 'process':
        lines += [f'Stage {i + 1}: {s}' for i, s in enumerate(f['steps'])]
    elif kind == 'map':
        for label, areas in f['panels'].items():
            lines.append(f'{label}: ' + '; '.join(a[0] for a in areas))
    lines.append('Main features: ' + f['features'])
    return '\n'.join(lines)


# ------------------------------------------------------------ practice bank

rows = []


def add(prefix, n, skill, type_id, type_name, title, prompt, seconds, **extra):
    row = dict(id=f'ielts-b2-{prefix}-{n:03}', exam_version_id=VERSION, exam='IELTS', supported_exams=['IELTS'],
               skill=skill, type_id=type_id, type_name=type_name, title=title, prompt=prompt,
               time_limit_seconds=seconds, prep_time_seconds=0, points=15, difficulty='medium',
               tags=['Original practice', 'IELTS', skill.capitalize()])
    row.update(extra)
    rows.append(row)


for i, (title, questions) in enumerate(PART1, 1):
    add('spk1', i, 'speaking', 'ielts-speaking-part1', 'Introduction', title,
        ' '.join(questions) + ' Give natural answers with relevant detail. This is a short Part 1 practice drill, not a complete 4–5 minute interview.', 90)
for i, (title, cue, points, explain) in enumerate(PART2, 1):
    add('spk2', i, 'speaking', 'ielts-speaking-part2', 'Speaking Part 2 (Cue Card)', title,
        f"{cue} You should say: {'; '.join(points)}; and explain {explain}. You have one minute to prepare and may make notes. Speak for one to two minutes.",
        120, prep_time_seconds=60)
for i, (title, questions) in enumerate(PART3, 1):
    add('spk3', i, 'speaking', 'ielts-speaking-part3', 'Speaking Part 3 (Discussion)', title,
        ' '.join(questions) + ' Discuss reasons, comparisons and examples. This is a two-minute discussion drill; a full examiner-led Part 3 lasts 4–5 minutes.', 120)

for i, (title, statement) in enumerate(TASK2, 1):
    add('wt2', i, 'writing', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree', title,
        statement + ' To what extent do you agree or disagree? Give reasons for your answer and include relevant examples from your own knowledge or experience. Write at least 250 words. Suggested time: 40 minutes.',
        2400, points=25,
        explanation='State a clear position, develop relevant reasons and examples, address qualifications where useful, and sustain the position through the conclusion. There is no preferred opinion.')

for i, f in enumerate(FIGURES, 1):
    lead = LEAD.get(f['kind'], '')
    if f['kind'] == 'pie':
        lead = 'The pie charts below show' if len(f['pies']) > 1 else 'The pie chart below shows'
    add('wt1', i, 'writing', 'ielts-writing-task1-figure', 'Describe the figure', f['title'],
        f"{lead} {f['desc']}.\n\nSummarise the information by selecting and reporting the main features, and make comparisons where relevant.\n\nWrite at least 150 words. You should spend about 20 minutes on this task.",
        1200, points=20, difficulty='hard' if f['kind'] in ('process', 'map') else 'medium',
        image_url=data_uri(RENDER[f['kind']](f)), figure_data=figure_data(f), explanation=f['features'])

for i, (title, script, notes, limit) in enumerate(FILL, 1):
    context, blanks, keys, last = [], [], [], -1
    for n, (text, answers) in enumerate(notes, 1):
        assert '____' in text, (title, text)
        context.append(text.replace('____', f'____ ({n})', 1))
        heard = [a for a in answers if spoken(a, script)]
        assert heard or any(numeric(a) for a in answers), (title, 'answer never heard', answers)
        for a in answers:
            assert fits(a, limit) or numeric(a), (title, a, limit)
        if heard:
            # The first time the answer is said after the previous answer: a
            # word can come up earlier as a distractor or in passing.
            later = [m.start() for a in heard
                     for m in re.finditer(r'(?<![\w])' + re.escape(a.lower()) + r'(?![\w])', script.lower()) if m.start() > last]
            assert later, (title, 'answers out of order', answers)
            last = min(later)
        blank = dict(id=str(n), correctAnswer=answers[0])
        if len(answers) > 1:
            blank['acceptedAnswers'] = answers[1:]
        blanks.append(blank)
        keys.append(answers[0])
    add('lfb', i, 'listening', 'ielts-listening-fill-blanks', 'Fill in the blanks', title,
        f'Listen and complete the notes. Write {limit} for each answer. This is a short practice drill with synthesised playback.',
        180, points=len(keys), audio_transcript=script, context_passage='\n'.join(context), blanks=blanks,
        correct_answers=keys, explanation='The recording gives these answers in order: ' + ', '.join(keys) + '. Do not include extra words.')

for i, (title, script, question, options, key, evidence) in enumerate(MCQ, 1):
    assert evidence in script, (title, evidence)
    assert len(options) == 3 and key in 'ABC', title
    # The answers were written mostly as C; rotate the options so the key
    # falls evenly on A, B and C.
    right = options['ABC'.index(key)]
    order = options[i % 3:] + options[:i % 3]
    new_key = 'ABC'[order.index(right)]
    add('lmcq', i, 'listening', 'ielts-listening-mcq', 'Choose the correct answer', title,
        question + ' Choose ONE letter, A, B or C. Synthesised practice audio.', 120, points=1,
        audio_transcript=script, options=[dict(id=chr(65 + k), text=t) for k, t in enumerate(order)],
        correct_answers=[new_key], explanation=f'The speaker says: "{evidence}". The answer is {new_key}: {right}')

listening_map.check()
for i, item in enumerate(listening_map.items(), 1):
    add('lmap', i, 'listening', 'ielts-listening-map', 'Navigating according to audio', item['title'],
        f"Locate {item['facility']} on the plan. Choose ONE letter, A–F. North is at the top; start at the main entrance at the bottom. Synthesised practice audio.",
        180, points=1, image_url=listening_map.svg(item['venue'], item['landmarks'], item['letter_slots']),
        audio_transcript=item['script'], options=[dict(id=k, text='Location ' + k) for k in 'ABCDEF'],
        correct_answers=[item['key']], explanation=f"Following the directions from the main entrance leads to location {item['key']}.")

counts = Counter(r['type_id'] for r in rows)
assert len(counts) == 8 and set(counts.values()) == {50}, counts
assert len({r['id'] for r in rows}) == len(rows)
assert Counter(r['correct_answers'][0] for r in rows if r['type_id'] == 'ielts-listening-mcq').most_common(1)[0][1] <= 20
for f in FIGURES:
    if f['kind'] == 'pie':
        for label, pie in f['pies'].items():
            assert sum(pie.values()) == 100, (f['title'], label)
    if f['kind'] in ('bar', 'line'):
        n = len(f['categories'] if f['kind'] == 'bar' else f['x'])
        assert all(len(v) == n for v in f['series'].values()), f['title']

columns = list(dict.fromkeys(k for row in rows for k in row))
body = ['-- Generated by scripts/generate_ielts_bank2.py. Original practice content.',
        '-- Fifty new items for each IELTS Speaking, Writing and Listening sub-task.',
        '-- Not official or recalled IELTS material; Task 1 data describes invented places.',
        'INSERT INTO questions (' + ', '.join(columns) + ') VALUES']
body.append(',\n'.join('(' + ', '.join(text_array(row[c]) if c in ('supported_exams', 'tags') else lit(row.get(c)) for c in columns) + ')' for row in rows))
body.append('ON CONFLICT (id) DO NOTHING;')

# ------------------------------------------------------------ listening tests

TYPE_NAMES = {
    'ielts-listening-completion': 'Listening completion',
    'ielts-listening-mcq': 'Listening multiple choice',
    'ielts-listening-map': 'Listening map labelling',
    'ielts-listening-matching': 'Listening matching',
}


def check_test(test):
    total = 0
    for part_no, part in enumerate(test['parts'], 1):
        script, last, count = part['script'], -1, 0
        for group in part['groups']:
            if group['type_id'] == 'ielts-listening-map':
                name, landmarks, letter_slots = group['image']
                letters = dict(zip('ABCDEF', letter_slots))
                for q in group['questions']:
                    _, _, key, evidence = q
                    slot = letters[key]
                    assert listening_map.position(slot, 'corner' in evidence or 'side just' in evidence) in evidence, \
                        (test['key'], part_no, 'map evidence does not describe the answer', q)
            for q in group['questions']:
                count += 1
                evidence = q[-1]
                where = script.find(evidence)
                assert where >= 0, (test['key'], part_no, 'evidence not in script', evidence)
                assert where > last, (test['key'], part_no, 'question out of recording order', evidence)
                last = where
                if group['type_id'] == 'ielts-listening-completion':
                    prompt, answers, _ = q
                    assert '________' in prompt, prompt
                    for a in answers:
                        assert len(a.split()) <= group['limit'] or numeric(a), (a, group['limit'])
                    assert any(spoken(a, script) for a in answers) or any(numeric(a) for a in answers), \
                        (test['key'], part_no, 'answer never heard', answers)
                else:
                    _, options, key, _ = q
                    assert key in {o['id'] for o in options}, (test['key'], q)
        assert count == 10, (test['key'], part_no, count)
        total += count
    assert total == 40, test['key']


TESTS = TESTS_A + TESTS_B + TESTS_C

# Lectures teach facts, so a lecture on a topic already live as a reading
# passage would give its answers away. Pairs a person has reviewed and judged
# to be different topics are listed here; any other shared word fails.
ALLOWED_TOPIC_OVERLAPS = {
    'Lecture: Desert plants': {'Plant-based meat'},
    'Lecture: Noise in offices': {'Noise on hospital wards'},
    'Autumn colours in trees': {'Westfield Adult Education: Autumn Short Courses'},
}
lecture_topics = [t['parts'][3]['groups'][0]['heading'] for t in TESTS] + [f[0] for f in FILL if f[0].startswith('Lecture:')]
for topic in lecture_topics:
    clashes = topic_clashes(topic.replace('Lecture:', ''), ALLOWED_TOPIC_OVERLAPS.get(topic, ()))
    assert not clashes, (topic, 'repeats a live reading topic', clashes)
assert [t['key'] for t in TESTS] == [f't{n}' for n in range(2, 11)]
for test in TESTS:
    check_test(test)

out = ['-- Generated by scripts/generate_ielts_bank2.py. Original practice content.',
       '-- IELTS Listening Practice Tests 2-10: four parts, forty one-mark questions each.',
       '-- Not official or recalled IELTS material. The mock paper builder numbers them in',
       '-- created_at order, so each test is stamped a second after the one before.', '']
for n, test in enumerate(TESTS, 1):
    tid = f"lt-{test['key']}"
    out.append(f"INSERT INTO listening_tests (id, exam_version_id, title, created_at) VALUES "
               f"({lit(tid)}, {lit(VERSION)}, {lit(test['title'])}, now() + interval '{n} seconds') ON CONFLICT (id) DO NOTHING;")
    number = 0
    for part_no, part in enumerate(test['parts'], 1):
        pid = f'{tid}-p{part_no}'
        out.append(f"INSERT INTO listening_parts (id, test_id, part_no, setting, script, reading_seconds) VALUES "
                   f"({lit(pid)}, {lit(tid)}, {part_no}, {lit(part['setting'])}, {lit(part['script'])}, 30) ON CONFLICT (id) DO NOTHING;")
        for position, group in enumerate(part['groups'], 1):
            gid = f'{pid}-g{position}'
            image = listening_map.svg(*group['image']) if group['type_id'] == 'ielts-listening-map' else None
            out.append(f"INSERT INTO listening_question_groups (id, part_id, position, type_id, instructions, heading, image_url) VALUES "
                       f"({lit(gid)}, {lit(pid)}, {position}, {lit(group['type_id'])}, {lit(group['instructions'])}, {lit(group['heading'])}, {lit(image)}) ON CONFLICT (id) DO NOTHING;")
            values = []
            for index, q in enumerate(group['questions'], 1):
                number += 1
                qid = f'{gid}-q{index}'
                if group['type_id'] == 'ielts-listening-completion':
                    prompt, answers, evidence = q
                    options, correct = [], answers
                else:
                    prompt, options, key, evidence = q
                    correct = [key]
                values.append('(' + ', '.join([
                    lit(qid), lit(VERSION), lit('IELTS'), "ARRAY['IELTS']", lit('listening'), lit(group['type_id']),
                    lit(TYPE_NAMES[group['type_id']]), lit(f"{test['title']}, Question {number}"), lit(prompt),
                    lit(options) if options else "'[]'::jsonb", lit(correct), lit(f'The recording says: "{evidence}".'),
                    '1', lit('medium'), "ARRAY['IELTS Listening', 'Mock test', 'Original practice']",
                    lit(gid), lit(index), 'TRUE', '0']) + ')')
            out.append('INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, options, '
                       'correct_answers, explanation, points, difficulty, tags, listening_group_id, group_position, is_published, time_limit_seconds) VALUES')
            out.append(',\n'.join(values))
            out.append('ON CONFLICT (id) DO NOTHING;')
    assert number == 40

# ------------------------------------------------------------ speaking sets

for set_id, title, content in SETS:
    assert len(content['part1']) == 2 and all(len(t['questions']) == 3 for t in content['part1']), set_id
    assert len(content['part2']['points']) == 3 and content['part2']['explain'].startswith('and explain'), set_id
    assert len(content['part3']['questions']) == 5, set_id
assert [s[0] for s in SETS] == [f'sm-{n:02}' for n in range(5, 11)]

sets = ['-- Generated by scripts/generate_ielts_bank2.py. Original practice content.',
        '-- IELTS Speaking mock sets 5-10, in the shape of sets 1-4 (migration 000076).',
        '-- Not official or recalled IELTS material. Stamped a second apart so the mock',
        '-- paper builder numbers them in this order.', '']
for n, (set_id, title, content) in enumerate(SETS, 1):
    sets.append(f"INSERT INTO speaking_mock_sets (id, title, content, created_at) VALUES "
                f"({lit(set_id)}, {lit(title)}, {lit(content)}, now() + interval '{n} seconds') ON CONFLICT (id) DO NOTHING;")

KEEP = '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n'
for name, text in [('000083_ielts_practice_bank_2', body), ('000084_ielts_listening_tests_2_to_10', out),
                   ('000085_ielts_speaking_sets_5_to_10', sets)]:
    (ROOT / f'migrations/{name}.up.sql').write_text('\n'.join(text) + '\n', encoding='utf-8', newline='\n')
    (ROOT / f'migrations/{name}.down.sql').write_text(KEEP, encoding='utf-8', newline='\n')

print('practice items:', dict(counts))
print('listening tests:', len(TESTS), 'x 40 questions; script words per test:',
      [sum(len(p['script'].split()) for p in t['parts']) for t in TESTS])
print('speaking sets:', len(SETS))
