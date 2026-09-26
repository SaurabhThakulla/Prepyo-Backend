"""Generate migration 000091: the IELTS Writing Task 2 essay types beyond Opinion.

Until now every Task 2 question was an Opinion (agree or disagree) essay, but
the real test sets any of six kinds of question. This adds fifty original
questions for each of the other five:

  ielts-writing-task2-discussion       Discuss Both Views
  ielts-writing-task2-advantages       Advantages / Disadvantages
  ielts-writing-task2-problem-solution Problem / Solution
  ielts-writing-task2-cause-effect     Cause / Effect
  ielts-writing-task2-two-part         Two-Part Question

It also gives the existing numbered Writing tests (mock papers) the six types
in turn, so the sectional Writing test and the full mock deal every kind of
Task 2 question rather than only Opinion essays. See ROTATE_PAPERS below.

The content lives in scripts/ielts_task2_types.py; this script checks it and
writes SQL. Nothing is written if a check fails:
  - exactly fifty questions per type;
  - no title or statement repeats another, in this set or in the existing
    Task 2 bank (read from the earlier migrations);
  - each question uses its type's IELTS wording.

Output is deterministic and replay-safe (ON CONFLICT DO NOTHING). Run from any
directory: python scripts/generate_ielts_task2_types.py
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(Path(__file__).resolve().parent))

from ielts_task2_types import (                                        # noqa: E402
    ADVANTAGES, BOTH_SIDES, CAUSE_EFFECT, DISCUSS, DISCUSSION, PROBLEM_SOLUTION, TWO_PART,
)

MIGRATION = '000091_ielts_writing_task2_types'
VERSION = 'ielts-2026-01'
TAIL = ('Give reasons for your answer and include relevant examples from your own knowledge or experience. '
        'Write at least 250 words. Suggested time: 40 minutes.')

EXPLAIN = {
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


def existing_task2():
    """Titles and statements of every Task 2 question in earlier migrations."""
    titles, statements = set(), set()
    row = re.compile(r"'ielts-writing-task2[a-z-]*',\s*'(?:[^']|'')*',\s*'((?:[^']|'')*)',\s*'((?:[^']|'')*)'")
    for path in sorted((ROOT / 'migrations').glob('*.up.sql')):
        if path.name.startswith(MIGRATION):
            continue
        for title, prompt in row.findall(path.read_text(encoding='utf-8')):
            titles.add(title.replace("''", "'").lower())
            statements.add(prompt.replace("''", "'").split('?')[0].lower())
    return titles, statements


def items():
    """Every question as (prefix, type id, type name, title, statement, question, explanation key)."""
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


# The wording each type's question must use.
FORMS = {
    'discussion': lambda q: q == DISCUSS,
    'advantages': lambda q: 'advantages' in q or 'benefits' in q,
    'problem': lambda q: 'problem' in q and ('solution' in q or 'measures' in q or 'done' in q),
    'cause': lambda q: ('causes' in q or 'reasons' in q or q.startswith('Why')) and ('effect' in q or 'impact' in q),
    'twopart': lambda q: q.count('?') == 2,
}

old_titles, old_statements = existing_task2()
rows, seen_titles, seen_statements = [], set(), set()
counts = {prefix: 0 for prefix, _, _ in TYPES}
by_prefix = {prefix: (type_id, name) for prefix, type_id, name in TYPES}

for prefix, title, statement, question, explain in items():
    counts[prefix] += 1
    n = counts[prefix]
    key = title.lower()
    assert key not in seen_titles and key not in old_titles, ('title repeats', title)
    assert statement.lower() not in seen_statements, ('statement repeats', statement)
    assert not any(statement.lower().rstrip('.') in old for old in old_statements), ('statement repeats bank', statement)
    assert statement.endswith('.') and '  ' not in statement, ('statement format', statement)
    assert question.endswith('?') or question == DISCUSS, ('question format', question)
    assert FORMS[prefix](question), (prefix, 'question does not use its type wording', question)
    seen_titles.add(key)
    seen_statements.add(statement.lower())
    type_id, type_name = by_prefix[prefix]
    rows.append(dict(
        id=f'ielts-wrt-{prefix}-{n:03}', type_id=type_id, type_name=type_name, title=title,
        prompt=f'{statement} {question} {TAIL}', explanation=EXPLAIN[explain]))

assert counts == {prefix: 50 for prefix in counts}, counts

# The numbered Writing tests were composed when every Task 2 question was an
# Opinion essay. Give each module's tests the six types in turn -- Test 1
# Opinion, Test 2 Discuss Both Views, and so on, Test 7 Opinion again -- as a
# new revision of each test whose type changes. Published papers are
# immutable, so this does what mockpapers.Repository.Revise does: retire the
# old revision and publish the next one under the same number. The content
# JSON and hash match what the Go builder writes. Each new revision takes a
# question no published paper uses yet, a different one for every test.
KINDS = ['ielts-writing-task2-opinion'] + [type_id for _, type_id, _ in TYPES]
ROTATE_PAPERS = f"""
CREATE TEMP TABLE ielts_task2_rotation ON COMMIT DROP AS
WITH kinds (pos, type_id) AS (
    VALUES {', '.join(f"({i}, '{t}')" for i, t in enumerate(KINDS))}
),
papers AS (
    SELECT p.id, p.module, p.number, p.content->>'task1Id' AS task1_id, q.type_id AS current_type,
           (row_number() OVER (PARTITION BY p.module ORDER BY p.number) - 1) % {len(KINDS)} AS pos
      FROM mock_papers p
      JOIN questions q ON q.id = p.content->>'task2Id'
     WHERE p.exam = 'ielts' AND p.section = 'writing' AND p.status = 'published'
),
wanted AS (
    SELECT p.id, p.task1_id, k.type_id,
           row_number() OVER (PARTITION BY k.type_id ORDER BY p.module, p.number) AS nth
      FROM papers p
      JOIN kinds k ON k.pos = p.pos
     WHERE p.current_type <> k.type_id
),
candidates AS (
    SELECT q.id, q.type_id, row_number() OVER (PARTITION BY q.type_id ORDER BY q.id) AS nth
      FROM questions q
     WHERE q.is_published AND q.skill = 'writing' AND 'IELTS' = ANY(q.supported_exams)
       AND q.type_id IN (SELECT type_id FROM kinds)
       AND NOT EXISTS (SELECT 1 FROM mock_papers p
                        WHERE p.exam = 'ielts' AND p.section = 'writing' AND p.status = 'published'
                          AND p.content->>'task2Id' = q.id)
)
SELECT w.id AS paper_id,
       '{{"task1Id":"' || w.task1_id || '","task2Id":"' || c.id || '"}}' AS content
  FROM wanted w
  JOIN candidates c ON c.type_id = w.type_id AND c.nth = w.nth;

UPDATE mock_papers p
   SET status = 'retired', retired_at = now()
  FROM ielts_task2_rotation r
 WHERE p.id = r.paper_id AND p.status = 'published';

INSERT INTO mock_papers (exam, section, module, number, revision, supersedes_id, status, title,
                         content_schema, content, content_hash, created_at, published_at)
SELECT old.exam, old.section, old.module, old.number, old.revision + 1, old.id, 'published', old.title,
       old.content_schema, r.content::jsonb, encode(sha256(convert_to(r.content, 'UTF8')), 'hex'), now(), now()
  FROM ielts_task2_rotation r
  JOIN mock_papers old ON old.id = r.paper_id;
"""

COLUMNS = ('id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, '
           'time_limit_seconds, prep_time_seconds, points, difficulty, tags, explanation')
values = []
for r in rows:
    values.append('(' + ', '.join([
        lit(r['id']), lit(VERSION), lit('IELTS'), text_array(['IELTS']), lit('writing'),
        lit(r['type_id']), lit(r['type_name']), lit(r['title']), lit(r['prompt']),
        '2400', '0', '25', lit('medium'),
        text_array(['Original practice', 'IELTS', 'Writing', 'Task 2']), lit(r['explanation']),
    ]) + ')')

up = [
    '-- Generated by scripts/generate_ielts_task2_types.py. Original practice content.',
    '-- Fifty IELTS Writing Task 2 questions for each essay type beyond Opinion: Discuss',
    '-- Both Views, Advantages / Disadvantages, Problem / Solution, Cause / Effect and',
    '-- Two-Part Question. Not official or recalled IELTS material.',
    f'INSERT INTO questions ({COLUMNS}) VALUES',
    ',\n'.join(values),
    'ON CONFLICT (id) DO NOTHING;',
    '',
    '-- Rotate the six Task 2 types through the existing numbered Writing tests.',
    ROTATE_PAPERS.strip(),
    '',
]
down = [
    f'-- Migration: {MIGRATION}.down.sql',
    '-- Preserve authored content, learner references and published test revisions on',
    '-- rollback: published mock papers are immutable and learners may have answered',
    '-- these questions.',
    'SELECT 1;',
    '',
]

(ROOT / 'migrations' / f'{MIGRATION}.up.sql').write_text('\n'.join(up), encoding='utf-8', newline='\n')
(ROOT / 'migrations' / f'{MIGRATION}.down.sql').write_text('\n'.join(down), encoding='utf-8', newline='\n')
print(f'wrote {len(rows)} questions to {MIGRATION}.up.sql')
