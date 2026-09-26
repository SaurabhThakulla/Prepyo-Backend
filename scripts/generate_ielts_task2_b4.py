"""Generate IELTS Writing Task 2 bank 4 (migration 000094).

Fifty new questions for each of the six IELTS Task 2 essay types (300 total):
  ielts-writing-task2-opinion          Opinion / Agree or Disagree   ielts-b4-opinion-001..050
  ielts-writing-task2-discussion       Discuss Both Views            ielts-b4-discussion-001..050
  ielts-writing-task2-advantages       Advantages / Disadvantages    ielts-b4-advantages-001..050
  ielts-writing-task2-problem-solution Problem / Solution            ielts-b4-problem-001..050
  ielts-writing-task2-cause-effect     Cause / Effect                ielts-b4-cause-001..050
  ielts-writing-task2-two-part         Two-Part Question             ielts-b4-twopart-001..050

Original Prepyo practice content written to the public IELTS Task 2 format;
none of it is official, recalled or copied test material. The content lives in
scripts/ielts_b4_task2.py; this script checks it and writes SQL. Nothing is
written if a check fails:
  - every item passes scripts/bank_check.py (no repeat of anything already in
    the bank or in this batch, British spelling, no machine-sounding words);
  - counts are exactly fifty per type (300 total);
  - each question matches its type's IELTS question format.

Output is deterministic and replay-safe (ON CONFLICT DO NOTHING).
Run from any directory: python scripts/generate_ielts_task2_b4.py
"""
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from scripts.bank_check import Batch                                      # noqa: E402
from ielts_task2_types import BOTH_SIDES, DISCUSS                         # noqa: E402
from ielts_b4_task2 import (                                              # noqa: E402
    ADVANTAGES, AGREE, CAUSE_EFFECT, DISCUSSION, OPINION, PROBLEM_SOLUTION, TWO_PART,
)

VERSION = 'ielts-2026-01'
MIGRATION = '000094_ielts_writing_task2_bank_4'
TAIL = ('Give reasons for your answer and include relevant examples from your own knowledge or experience. '
        'Write at least 250 words. Suggested time: 40 minutes.')

EXPLAIN = {
    'opinion': (
        'State a clear position, develop relevant reasons and examples, address qualifications where useful, '
        'and sustain the position through the conclusion. There is no preferred opinion.'),
    'discussion': (
        'Give each view its own body paragraph with reasons and examples, and state your own opinion clearly '
        'in the introduction and the conclusion. Your opinion may support one view or combine both. Discussing '
        'only one view, or giving no opinion, leaves part of the task unanswered.'),
    'advantages': (
        'Explain the main advantages and the main disadvantages, each with reasons and examples, and say clearly '
        'which side outweighs the other. Keep that judgement consistent from the introduction to the conclusion. '
        'Listing both sides without weighing them leaves the question unanswered.'),
    'advantages-both': (
        'Give balanced coverage of the main advantages and the main disadvantages, each with reasons and '
        'examples, usually in separate body paragraphs. The question does not ask for your opinion, but a '
        'conclusion that sums up the overall picture strengthens the essay.'),
    'problem-solution': (
        'Answer both parts of the question: explain two or three specific problems and their consequences, then '
        'suggest practical solutions that respond directly to those problems. Give both parts similar depth; '
        'solutions that do not match the problems you describe weaken the response.'),
    'cause-effect': (
        'Answer both parts of the question: explain the main causes or reasons and the main effects, each with '
        'examples, usually in separate body paragraphs. Discussing only the causes or only the effects leaves '
        'part of the task unanswered.'),
    'two-part': (
        'Answer each question directly, usually in its own body paragraph, with reasons and examples. Where a '
        'question asks whether something is positive or negative, or asks for your view, give a clear answer '
        'and keep it consistent in the conclusion. Leaving either question unanswered addresses the task '
        'incompletely.'),
}

TYPES = [
    # (id prefix, type id, type name)
    ('opinion', 'ielts-writing-task2-opinion', 'Opinion / Agree or Disagree'),
    ('discussion', 'ielts-writing-task2-discussion', 'Discuss Both Views'),
    ('advantages', 'ielts-writing-task2-advantages', 'Advantages / Disadvantages'),
    ('problem', 'ielts-writing-task2-problem-solution', 'Problem / Solution'),
    ('cause', 'ielts-writing-task2-cause-effect', 'Cause / Effect'),
    ('twopart', 'ielts-writing-task2-two-part', 'Two-Part Question'),
]


def lit(value):
    if isinstance(value, int):
        return str(value)
    return "'" + str(value).replace("'", "''") + "'"


def text_array(values):
    return 'ARRAY[' + ', '.join(lit(v) for v in values) + ']'


def items():
    """Every question as (prefix, title, statement, question, explain key)."""
    for title, statement in OPINION:
        yield 'opinion', title, statement, AGREE, 'opinion'
    for title, statement in DISCUSSION:
        yield 'discussion', title, statement, DISCUSS, 'discussion'
    for title, statement, question in ADVANTAGES:
        yield 'advantages', title, statement, question, 'advantages-both' if question == BOTH_SIDES else 'advantages'
    for title, statement, question in PROBLEM_SOLUTION:
        yield 'problem', title, statement, question, 'problem-solution'
    for title, statement, question in CAUSE_EFFECT:
        yield 'cause', title, statement, question, 'cause-effect'
    for title, statement, question in TWO_PART:
        yield 'twopart', title, statement, question, 'two-part'


FORMS = {
    'opinion': lambda q: q == AGREE,
    'discussion': lambda q: q == DISCUSS,
    'advantages': lambda q: 'advantages' in q or 'benefits' in q,
    'problem': lambda q: 'problem' in q and ('solution' in q or 'measures' in q or 'done' in q),
    'cause': lambda q: ('causes' in q or 'reasons' in q or q.startswith('Why')) and ('effect' in q or 'impact' in q),
    'twopart': lambda q: q.count('?') == 2,
}

batch = Batch()
rows = []
counts = {prefix: 0 for prefix, _, _ in TYPES}
by_prefix = {prefix: (type_id, name) for prefix, type_id, name in TYPES}
seen_titles, seen_statements = set(), set()

for prefix, title, statement, question, explain_key in items():
    counts[prefix] += 1
    n = counts[prefix]
    item_id = f'ielts-b4-{prefix}-{n:03}'
    type_id, type_name = by_prefix[prefix]
    
    assert title.lower() not in seen_titles, ('title repeats in batch', title)
    assert statement.lower() not in seen_statements, ('statement repeats in batch', statement)
    assert statement.endswith('.') and '  ' not in statement, ('statement format', statement)
    assert question.endswith('?') or question == DISCUSS, ('question format', question)
    assert FORMS[prefix](question), (prefix, 'question does not use its type wording', question)
    
    seen_titles.add(title.lower())
    seen_statements.add(statement.lower())
    
    prompt = f'{statement} {question} {TAIL}'
    explanation = EXPLAIN[explain_key]
    
    batch.question('IELTS', type_id, title, statement, where=item_id)
    batch.style(prompt, where=item_id)
    batch.style(explanation, where=item_id)
    
    rows.append(dict(
        id=item_id,
        exam_version_id=VERSION,
        exam='IELTS',
        supported_exams=['IELTS'],
        skill='writing',
        type_id=type_id,
        type_name=type_name,
        title=title,
        prompt=prompt,
        time_limit_seconds=2400,
        prep_time_seconds=0,
        points=25,
        difficulty='medium',
        tags=['Original practice', 'IELTS', 'Writing', 'Task 2'],
        explanation=explanation,
    ))

batch.assert_clean()

assert counts == {prefix: 50 for prefix in counts}, counts
assert len(rows) == 300
assert len({r['id'] for r in rows}) == 300
assert len({r['title'].lower() for r in rows}) == 300

# ------------------------------------------------------------ SQL

def insert(rows):
    columns = list(rows[0])
    out = ['INSERT INTO questions (' + ', '.join(columns) + ') VALUES']
    out.append(',\n'.join('(' + ', '.join(text_array(r[c]) if c in ('supported_exams', 'tags') else lit(r[c]) for c in columns) + ')'
                          for r in rows))
    out.append('ON CONFLICT (id) DO NOTHING;')
    return out


body = [
    '-- Generated by scripts/generate_ielts_task2_b4.py. Original practice content.',
    '-- IELTS Writing Task 2 bank 4: fifty questions for each of the six essay types.',
    '-- Not official or recalled IELTS material.',
    '',
]
body += insert(rows)

KEEP = '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n'
(ROOT / f'migrations/{MIGRATION}.up.sql').write_text('\n'.join(body) + '\n', encoding='utf-8', newline='\n')
(ROOT / f'migrations/{MIGRATION}.down.sql').write_text(KEEP, encoding='utf-8', newline='\n')

print('Generated', len(rows), 'questions:', counts)
