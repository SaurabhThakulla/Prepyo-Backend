"""Generate IELTS Speaking bank 4 (migration 000092).

  000092  fifty new practice items for each IELTS Speaking part (150 questions)
          and IELTS Speaking mock sets 11-20

All content is original Prepyo practice material written to the public IELTS
Speaking format; none of it is official, recalled or copied test material. The
content lives in scripts/ielts_b4_speaking; this script checks it and writes
SQL. Nothing is written if a check fails:
  - every item and every set passes scripts/bank_check.py (no repeat of
    anything already in the bank, British spelling, no machine-sounding words);
  - counts are exactly fifty per part and ten sets, numbered sm-11 to sm-20;
  - every set has two Part 1 topics of three questions, a cue card with three
    prompts and an "and explain" prompt, and five Part 3 questions.
The practice rows follow scripts/generate_ielts_bank2.py (migration 000083) and
the sets follow its speaking sets (migration 000085). The mock paper builder
(internal/mockpapers) registers every published set that has no number yet in
created_at order, so each new set is stamped a second after the newest set
already in the table and they become Speaking tests 11-20.

Output is deterministic and replay-safe (ON CONFLICT DO NOTHING). Run from any
directory: python scripts/generate_ielts_speaking_b4.py
"""
import json
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(Path(__file__).resolve().parent))

from bank_check import Batch                                            # noqa: E402
from ielts_b4_speaking.speaking import PART1, PART2, PART3              # noqa: E402
from ielts_b4_speaking.speaking_sets import SETS                        # noqa: E402

VERSION = 'ielts-2026-01'
MIGRATION = '000092_ielts_speaking_bank_4'


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


# ------------------------------------------------------------ practice bank

rows = []


def add(prefix, n, type_id, type_name, title, prompt, seconds, **extra):
    row = dict(id=f'ielts-b4-{prefix}-{n:03}', exam_version_id=VERSION, exam='IELTS', supported_exams=['IELTS'],
               skill='speaking', type_id=type_id, type_name=type_name, title=title, prompt=prompt,
               time_limit_seconds=seconds, prep_time_seconds=0, points=15, difficulty='medium',
               tags=['Original practice', 'IELTS', 'Speaking'])
    row.update(extra)
    rows.append(row)


for i, (title, questions) in enumerate(PART1, 1):
    assert len(questions) == 3, title
    add('spk1', i, 'ielts-speaking-part1', 'Introduction', title,
        ' '.join(questions) + ' Give natural answers with relevant detail. This is a short Part 1 practice drill, not a complete 4–5 minute interview.', 90)
for i, (title, cue, points, explain) in enumerate(PART2, 1):
    assert len(points) == 3 and cue.startswith('Describe ') and not explain.startswith('and'), title
    add('spk2', i, 'ielts-speaking-part2', 'Speaking Part 2 (Cue Card)', title,
        f"{cue} You should say: {'; '.join(points)}; and explain {explain}. You have one minute to prepare and may make notes. Speak for one to two minutes.",
        120, prep_time_seconds=60)
for i, (title, questions) in enumerate(PART3, 1):
    assert len(questions) == 3, title
    add('spk3', i, 'ielts-speaking-part3', 'Speaking Part 3 (Discussion)', title,
        ' '.join(questions) + ' Discuss reasons, comparisons and examples. This is a two-minute discussion drill; a full examiner-led Part 3 lasts 4–5 minutes.', 120)

counts = Counter(r['type_id'] for r in rows)
assert counts == {'ielts-speaking-part1': 50, 'ielts-speaking-part2': 50, 'ielts-speaking-part3': 50}, counts
assert len({r['id'] for r in rows}) == len(rows) == 150
for type_id in counts:
    titles = [r['title'].lower() for r in rows if r['type_id'] == type_id]
    assert len(set(titles)) == len(titles), type_id

# ------------------------------------------------------------ speaking sets

assert [s[0] for s in SETS] == [f'sm-{n:02}' for n in range(11, 21)]
assert len({s[1].lower() for s in SETS}) == len(SETS)
assert len({s[2]['part2']['topic'] for s in SETS}) == len(SETS)
assert len({t['topic'] for s in SETS for t in s[2]['part1']}) == 2 * len(SETS)
for set_id, title, content in SETS:
    assert list(content) == ['part1', 'part2', 'part3'], set_id
    assert len(content['part1']) == 2 and all(len(t['questions']) == 3 for t in content['part1']), set_id
    p2 = content['part2']
    assert list(p2) == ['topic', 'cue', 'points', 'explain', 'rounding'], set_id
    assert len(p2['points']) == 3 and p2['explain'].startswith('and explain') and p2['cue'].startswith('Describe '), set_id
    assert list(content['part3']) == ['topic', 'questions'] and len(content['part3']['questions']) == 5, set_id


def set_text(content):
    """Everything the candidate hears in a set, in the order it is asked."""
    p2 = content['part2']
    lines = []
    for topic in content['part1']:
        lines += [topic['topic'] + '.'] + topic['questions']
    lines += [p2['cue'], 'You should say: ' + '; '.join(p2['points']) + '; ' + p2['explain'], p2['rounding']]
    lines += [content['part3']['topic'] + '.'] + content['part3']['questions']
    return '\n'.join(lines)


# ------------------------------------------------------------ bank check

batch = Batch()
for row in rows:
    batch.question('IELTS', row['type_id'], row['title'], row['prompt'], where=row['id'])
for set_id, title, content in SETS:
    batch.passage(title, set_text(content), where=set_id)
batch.assert_clean()

# ------------------------------------------------------------ SQL

columns = list(dict.fromkeys(k for row in rows for k in row))
body = ['-- Generated by scripts/generate_ielts_speaking_b4.py. Original practice content.',
        '-- IELTS Speaking bank 4: fifty new practice items for each Speaking part, and',
        '-- Speaking mock sets 11-20 in the shape of sets 1-10 (migrations 000076, 000085).',
        '-- Not official or recalled IELTS material.',
        '',
        'INSERT INTO questions (' + ', '.join(columns) + ') VALUES']
body.append(',\n'.join('(' + ', '.join(text_array(row[c]) if c in ('supported_exams', 'tags') else lit(row.get(c)) for c in columns) + ')' for row in rows))
body.append('ON CONFLICT (id) DO NOTHING;')
body += ['',
         '-- The mock paper builder numbers unregistered sets in created_at order. Each set',
         '-- is stamped a second after the newest set already present, so sets 11-20 follow',
         '-- sets 1-10 in this order however soon after migration 000085 this one runs.']
for set_id, title, content in SETS:
    body.append(f"INSERT INTO speaking_mock_sets (id, title, content, created_at) VALUES "
                f"({lit(set_id)}, {lit(title)}, {lit(content)}, "
                f"(SELECT greatest(now(), max(created_at)) FROM speaking_mock_sets) + interval '1 second') ON CONFLICT (id) DO NOTHING;")

KEEP = '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n'
(ROOT / f'migrations/{MIGRATION}.up.sql').write_text('\n'.join(body) + '\n', encoding='utf-8', newline='\n')
(ROOT / f'migrations/{MIGRATION}.down.sql').write_text(KEEP, encoding='utf-8', newline='\n')

print('practice items:', dict(counts))
print('speaking sets:', len(SETS), [s[0] for s in SETS])
