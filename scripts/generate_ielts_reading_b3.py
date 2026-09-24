"""Generate IELTS Reading Bank Batch 3 (Migration 000086).

This script generates 10 original academic reading passages (6 paragraphs A-F, 700-900 words)
and 25 questions per passage (250 questions total across 6 question types):
  - 3 Multiple Choice Single Answer (options A-D)
  - 2 Multiple Choice Multiple Answers ("Choose TWO", options A-E)
  - 5 True / False / Not Given
  - 5 Yes / No / Not Given
  - 5 Matching Information (options A-F for paragraphs A-F)
  - 5 Sentence Completion (ONE WORD ONLY from the passage)

Self-validation enforces:
  - Exactly 10 passages with 6 paragraphs each (A-F), word count 700-900.
  - Sentence completion answers appear verbatim in the passage text.
  - Matching information evidence quotes appear verbatim in the target paragraph.
  - Every MCQ key is an existing option.
  - TFNG and YNNG keys are valid enum values.
  - Question and group IDs are completely unique and prefixed with ir86.
  - Output is deterministic and replay-safe (INSERT ... ON CONFLICT (id) DO NOTHING).
  - Multiple-choice options are rotated so answer letters spread across A-D
    (A-E for "Choose TWO"); the content was written with most answers at B.
  - Every True/False/Not Given and Yes/No/Not Given group uses all three answers.
  - MCQ, TFNG, YNNG and completion questions follow the order of the passage,
    judged by the paragraph each explanation cites.
  - No passage repeats a topic already live (scripts/live_topics.py).
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from collections import Counter

from scripts.content_b3.ielts_reading import PASSAGES
from scripts.live_topics import topic_clashes

# Live titles sharing a word with a new passage that a person has reviewed and
# judged to be a different topic.
ALLOWED_TOPIC_OVERLAPS = {
    'bioluminescence': {'Keepers of the Light', 'Light pollution'},
    'glaciers': {'Contrails and climate', 'Reducing the Effects of Climate Change'},
}


def rotate(options, keys, shift, explanation):
    """Moves each option `shift` places up, relabels A, B, C..., and follows the
    keys and any "(X)" letter references in the explanation to their new letters."""
    order = list(range(len(options)))
    order = order[shift:] + order[:shift]
    new_options = [{'id': chr(65 + i), 'text': options[j]['text']} for i, j in enumerate(order)]
    moved = {options[j]['id']: chr(65 + i) for i, j in enumerate(order)}
    new_keys = sorted(moved[k] for k in keys)
    new_explanation = re.sub(r'\(([A-E])\)', lambda m: '(' + moved[m.group(1)] + ')', explanation)
    return new_options, new_keys, new_explanation


def cited_paragraph(explanation):
    found = re.search(r'[Pp]aragraph ([A-F])', explanation)
    return found.group(1) if found else None

VERSION = 'ielts-2026-01'
UP_MIGRATION = ROOT / 'migrations' / '000086_ielts_reading_bank_3.up.sql'
DOWN_MIGRATION = ROOT / 'migrations' / '000086_ielts_reading_bank_3.down.sql'


def sql_lit(val):
    if val is None:
        return 'NULL'
    if isinstance(val, bool):
        return 'TRUE' if val else 'FALSE'
    if isinstance(val, (int, float)):
        return str(val)
    if isinstance(val, (dict, list)):
        return "'" + json.dumps(val, ensure_ascii=False).replace("'", "''") + "'::jsonb"
    return "'" + str(val).replace("'", "''") + "'"


def sql_text_array(items):
    return 'ARRAY[' + ', '.join(sql_lit(x) for x in items) + ']'


def validate_and_compile():
    print(f"Validating {len(PASSAGES)} IELTS Reading passages...")
    assert len(PASSAGES) == 10, f"Expected exactly 10 passages, got {len(PASSAGES)}"

    seen_ids = set()
    passages_sql = []
    groups_sql = []
    questions_sql = []

    key_letters = {'single': Counter(), 'multiple': Counter()}
    type_counts = {
        'reading-mcq-single': 0,
        'reading-mcq-multiple': 0,
        'reading-true-false': 0,
        'reading-yes-no-not-given': 0,
        'reading-matching-information': 0,
        'reading-sentence-completion': 0,
    }

    for idx, p in enumerate(PASSAGES, start=1):
        slug = p['id_slug']
        pid = f"rp-ir86-{slug}"
        clashes = topic_clashes(p['title'] + ' ' + p['topic'], ALLOWED_TOPIC_OVERLAPS.get(slug, ()))
        assert not clashes, f"{slug}: repeats a live reading topic {clashes}"
        for group, answers in (('tfng', {'TRUE', 'FALSE', 'NOT_GIVEN'}), ('ynng', {'YES', 'NO', 'NOT_GIVEN'})):
            used = {q['key'] for q in p[group]}
            assert used == answers, f"{slug} {group}: uses {sorted(used)}, needs all of {sorted(answers)}"
        for group in ('mcq_single', 'tfng', 'ynng', 'completion'):
            cited = [c for c in (cited_paragraph(q['explanation']) for q in p[group]) if c]
            assert cited == sorted(cited), f"{slug} {group}: questions out of passage order {cited}"
        assert pid not in seen_ids, f"Duplicate passage ID {pid}"
        seen_ids.add(pid)

        # Validate paragraphs A-F
        paragraphs = p['paragraphs']
        assert len(paragraphs) == 6, f"{pid}: Expected 6 paragraphs A-F, got {len(paragraphs)}"
        labels = [para['label'] for para in paragraphs]
        assert labels == ['A', 'B', 'C', 'D', 'E', 'F'], f"{pid}: Invalid labels {labels}"

        full_text = " ".join(para['text'] for para in paragraphs)
        word_count = len(re.findall(r"\b[\w'-]+\b", full_text))
        assert 700 <= word_count <= 900, f"{pid}: word count {word_count} not in range [700, 900]"

        para_map = {para['label']: para['text'] for para in paragraphs}

        # Passage SQL
        passages_sql.append(
            f"    ({sql_lit(pid)}, {sql_lit(VERSION)}, {sql_lit(p['title'])}, {sql_lit(p['subtitle'])},\n"
            f"     {sql_lit(paragraphs)},\n"
            f"     {word_count}, 'medium', {sql_lit(p['topic'])}, ARRAY['IELTS Reading', 'Original practice'], 'custom', ARRAY['academic'])"
        )

        # Groups configuration
        group_specs = [
            (
                'single',
                1,
                'reading-mcq-single',
                'Multiple Choice, Single Answer',
                'Choose ONE letter, A, B, C or D.',
            ),
            (
                'multiple',
                2,
                'reading-mcq-multiple',
                'Multiple Choice, Multiple Answers',
                'Choose TWO letters for each question.',
            ),
            (
                'tfng',
                3,
                'reading-true-false',
                'True / False / Not Given',
                'Write TRUE if the statement agrees with the information, FALSE if it contradicts it, or NOT GIVEN if there is no information on this.',
            ),
            (
                'ynng',
                4,
                'reading-yes-no-not-given',
                'Yes / No / Not Given',
                'Write YES if the statement agrees with the views of the writer, NO if it contradicts them, or NOT GIVEN if the view is not stated.',
            ),
            (
                'matching',
                5,
                'reading-matching-information',
                'Matching Information',
                'Which paragraph, A-F, contains the following information? You may use any letter more than once.',
            ),
            (
                'completion',
                6,
                'reading-sentence-completion',
                'Sentence Completion',
                'Complete each sentence. Write ONE WORD ONLY from the passage in each gap.',
            ),
        ]

        for g_code, g_pos, g_tid, g_tname, g_instr in group_specs:
            gid = f"g-ir86-{slug}-{g_code}"
            assert gid not in seen_ids, f"Duplicate group ID {gid}"
            seen_ids.add(gid)

            groups_sql.append(
                f"    ({sql_lit(gid)}, {sql_lit(pid)}, {g_pos}, {sql_lit(g_tid)}, {sql_lit(g_tname)},\n"
                f"     {sql_lit(g_instr)}, '[]'::jsonb, 'full', FALSE, 420)"
            )

        # 1. MCQ Single (3 questions)
        mcq_single = p['mcq_single']
        assert len(mcq_single) == 3, f"{pid}: expected 3 mcq_single, got {len(mcq_single)}"
        for q_idx, q in enumerate(mcq_single, start=1):
            qid = f"q-ir86-{slug}-single-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            opt_ids = [opt['id'] for opt in q['options']]
            assert opt_ids == ['A', 'B', 'C', 'D'], f"{qid}: options must be A, B, C, D"
            assert q['key'] in opt_ids, f"{qid}: key {q['key']} not in options"
            assert q['explanation'], f"{qid}: explanation missing"
            options, (key,), explanation = rotate(q['options'], [q['key']], (idx * 3 + q_idx) % 4, q['explanation'])
            q = dict(q, options=options, key=key, explanation=explanation)
            key_letters['single'][key] += 1

            type_counts['reading-mcq-single'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-mcq-single', 'Multiple Choice, Single Answer',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['prompt'])}, {sql_lit(q['options'])}, {sql_lit([q['key']])},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-single')}, {q_idx}, TRUE, 0)"
            )

        # 2. MCQ Multiple (2 questions)
        mcq_multiple = p['mcq_multiple']
        assert len(mcq_multiple) == 2, f"{pid}: expected 2 mcq_multiple, got {len(mcq_multiple)}"
        for q_idx, q in enumerate(mcq_multiple, start=1):
            qid = f"q-ir86-{slug}-multiple-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            opt_ids = [opt['id'] for opt in q['options']]
            assert opt_ids == ['A', 'B', 'C', 'D', 'E'], f"{qid}: options must be A-E"
            assert len(q['keys']) == 2, f"{qid}: keys must have exactly 2 elements"
            for k in q['keys']:
                assert k in opt_ids, f"{qid}: key {k} not in options"
            assert q['explanation'], f"{qid}: explanation missing"
            options, keys, explanation = rotate(q['options'], q['keys'], (idx + q_idx * 2) % 5, q['explanation'])
            q = dict(q, options=options, keys=keys, explanation=explanation)
            key_letters['multiple'].update(keys)

            type_counts['reading-mcq-multiple'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['prompt'])}, {sql_lit(q['options'])}, {sql_lit(sorted(q['keys']))},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-multiple')}, {q_idx}, TRUE, 0)"
            )

        # 3. True / False / Not Given (5 questions)
        tfng = p['tfng']
        assert len(tfng) == 5, f"{pid}: expected 5 tfng, got {len(tfng)}"
        tfng_options = [
            {"id": "TRUE", "text": "True"},
            {"id": "FALSE", "text": "False"},
            {"id": "NOT_GIVEN", "text": "Not Given"},
        ]
        for q_idx, q in enumerate(tfng, start=1):
            qid = f"q-ir86-{slug}-tfng-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            assert q['key'] in ['TRUE', 'FALSE', 'NOT_GIVEN'], f"{qid}: invalid key {q['key']}"
            assert q['explanation'], f"{qid}: explanation missing"

            type_counts['reading-true-false'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false', 'True / False / Not Given',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['statement'])}, {sql_lit(tfng_options)}, {sql_lit([q['key']])},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-tfng')}, {q_idx}, TRUE, 0)"
            )

        # 4. Yes / No / Not Given (5 questions)
        ynng = p['ynng']
        assert len(ynng) == 5, f"{pid}: expected 5 ynng, got {len(ynng)}"
        ynng_options = [
            {"id": "YES", "text": "Yes"},
            {"id": "NO", "text": "No"},
            {"id": "NOT_GIVEN", "text": "Not Given"},
        ]
        for q_idx, q in enumerate(ynng, start=1):
            qid = f"q-ir86-{slug}-ynng-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            assert q['key'] in ['YES', 'NO', 'NOT_GIVEN'], f"{qid}: invalid key {q['key']}"
            assert q['explanation'], f"{qid}: explanation missing"

            type_counts['reading-yes-no-not-given'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-yes-no-not-given', 'Yes / No / Not Given',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['statement'])}, {sql_lit(ynng_options)}, {sql_lit([q['key']])},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-ynng')}, {q_idx}, TRUE, 0)"
            )

        # 5. Matching Information (5 questions)
        matching = p['matching']
        assert len(matching) == 5, f"{pid}: expected 5 matching, got {len(matching)}"
        matching_options = [{"id": ch, "text": f"Paragraph {ch}"} for ch in ['A', 'B', 'C', 'D', 'E', 'F']]
        for q_idx, q in enumerate(matching, start=1):
            qid = f"q-ir86-{slug}-matching-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            assert q['key'] in ['A', 'B', 'C', 'D', 'E', 'F'], f"{qid}: invalid matching key {q['key']}"
            assert q['explanation'], f"{qid}: explanation missing"

            # Check that evidence quote is contained in the target paragraph!
            evidence = q['evidence']
            target_para_text = para_map[q['key']]
            assert evidence.lower() in target_para_text.lower(), (
                f"{qid}: Evidence quote '{evidence}' not found verbatim in Paragraph {q['key']}!"
            )

            type_counts['reading-matching-information'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-matching-information', 'Matching Information',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['prompt'])}, {sql_lit(matching_options)}, {sql_lit([q['key']])},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-matching')}, {q_idx}, TRUE, 0)"
            )

        # 6. Sentence Completion (5 questions)
        completion = p['completion']
        assert len(completion) == 5, f"{pid}: expected 5 completion, got {len(completion)}"
        for q_idx, q in enumerate(completion, start=1):
            qid = f"q-ir86-{slug}-completion-{q_idx}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            ans = q['answer']
            assert len(ans.split()) == 1, f"{qid}: completion answer '{ans}' must be exactly ONE WORD ONLY"
            assert '________' in q['prompt'], f"{qid}: completion prompt must contain '________'"
            assert q['explanation'], f"{qid}: explanation missing"

            # Check that answer appears verbatim in the passage!
            pattern = r"(?<!\w)" + re.escape(ans) + r"(?!\w)"
            assert re.search(pattern, full_text, re.IGNORECASE), (
                f"{qid}: Completion answer '{ans}' not found verbatim in passage {pid}!"
            )

            type_counts['reading-sentence-completion'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'IELTS', ARRAY['IELTS'], 'reading', 'reading-sentence-completion', 'Sentence Completion',\n"
                f"     {sql_lit(p['title'])}, {sql_lit(q['prompt'])}, '[]'::jsonb, {sql_lit([ans])},\n"
                f"     {sql_lit(q['explanation'])}, 1, 'medium', ARRAY['IELTS Reading', 'Original practice'], {sql_lit(pid)}, {sql_lit(f'g-ir86-{slug}-completion')}, {q_idx}, TRUE, 0)"
            )

    # No letter may carry much more than its share of the answers.
    assert set(key_letters['single']) == set('ABCD') and max(key_letters['single'].values()) <= 10, key_letters['single']
    assert set(key_letters['multiple']) == set('ABCDE') and max(key_letters['multiple'].values()) <= 12, key_letters['multiple']

    print("\n--- Validation Succeeded ---")
    print(f"Answer letters: single {dict(sorted(key_letters['single'].items()))}, "
          f"multiple {dict(sorted(key_letters['multiple'].items()))}")
    print(f"Passages: {len(passages_sql)}")
    print(f"Groups:   {len(groups_sql)}")
    print(f"Questions:{len(questions_sql)}")
    print("\nQuestions per type:")
    for tid, count in type_counts.items():
        print(f"  {tid:<30} : {count}")

    # Build SQL file content
    sql_header = (
        "-- Migration: 000086_ielts_reading_bank_3.up.sql\n"
        "-- 10 original IELTS Academic reading passages (6 paragraphs, 700-900 words each).\n"
        "-- Each passage carries 25 questions: 5 Multiple Choice (3 single, 2 multiple),\n"
        "-- 5 True/False/Not Given, 5 Yes/No/Not Given, 5 Matching Information and\n"
        "-- 5 Sentence Completion. Total 250 questions (50 per sub-task category).\n"
        "-- All questions use the ir86 namespace and strictly follow passage order.\n"
        "-- Replay-safe: INSERT ... ON CONFLICT (id) DO NOTHING.\n\n"
    )

    passages_block = (
        "INSERT INTO reading_passages\n"
        "    (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot, modules)\n"
        "VALUES\n" + ",\n".join(passages_sql) + "\nON CONFLICT (id) DO NOTHING;\n\n"
    )

    groups_block = (
        "INSERT INTO reading_question_groups\n"
        "    (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds)\n"
        "VALUES\n" + ",\n".join(groups_sql) + "\nON CONFLICT (id) DO NOTHING;\n\n"
    )

    questions_block = (
        "INSERT INTO questions\n"
        "    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,\n"
        "     title, prompt, options, correct_answers, explanation, points,\n"
        "     difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds)\n"
        "VALUES\n" + ",\n".join(questions_sql) + "\nON CONFLICT (id) DO NOTHING;\n"
    )

    full_sql = sql_header + passages_block + groups_block + questions_block

    print(f"\nWriting up migration to {UP_MIGRATION}...")
    UP_MIGRATION.write_text(full_sql, encoding='utf-8', newline='\n')

    print(f"Writing down migration to {DOWN_MIGRATION}...")
    DOWN_MIGRATION.write_text("-- Migration: 000086_ielts_reading_bank_3.down.sql\n-- Additive content migration: preserve user sessions and practice records.\nSELECT 1;\n", encoding='utf-8', newline='\n')

    print("Generation complete!")


if __name__ == '__main__':
    validate_and_compile()
