"""Generate PTE Reading bank 4a (Migration 000102).

This script generates:
  - 10 full-length academic reading passages (6 paragraphs A-F, 700-900 words)
    on natural sciences, medicine, technology, earth and space science,
    linguistics and history
  - 50 Reading & Writing: Fill in the Blanks (dropdown) questions (5 per passage)
  - 50 Reading: Fill in the Blanks (drag and drop) questions (5 per passage)
  - 50 Multiple Choice, Single Answer questions (5 per passage)
  - 50 Multiple Choice, Multiple Answers questions (5 per passage)
Total: 200 questions (50 per task type). Re-order Paragraphs items are not part
of this bank; they are written separately.

Modelled on the passage part of scripts/generate_pte_reading_b3.py (migration
000089). Self-validation enforces:
  - Exactly 10 passages with 6 paragraphs each (A-F), word count 700-900.
  - Exactly 50 questions per task type, 5 per type on every passage.
  - All IDs are unique and use the pr102 prefix.
  - Dropdown blanks: 4 distinct options, only one of which fits; no wrong
    option is used more than twice in the whole bank.
  - Drag-and-drop blanks: a shared pool with at least 2 distractors.
  - MCQ options are rotated with rotate() so keys spread evenly.
  - No passage title repeats a live reading topic (scripts/live_topics.py).
  - Every passage goes through scripts/bank_check.py Batch.passage(), every
    question through Batch.question(), and every option and explanation
    through Batch.style(); Batch.assert_clean() runs before any SQL is written.
  - Writes migrations/000102_pte_reading_bank_4a.up.sql and .down.sql with
    Unix line endings; deterministic and replay-safe
    (INSERT ... ON CONFLICT (id) DO NOTHING).
"""

import hashlib
import json
import re
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from scripts.bank_check import Batch
from scripts.live_topics import topic_clashes
from scripts.pte_b4_reading_a.passages import ALL_PASSAGES
from scripts.pte_b4_reading_a.validator import validate_passage

VERSION = 'pte-2026-01'
PREFIX = 'pr102'
TAGS = ['PTE Reading', 'Original practice', 'Bank 4']
UP_MIGRATION = ROOT / 'migrations' / '000102_pte_reading_bank_4a.up.sql'
DOWN_MIGRATION = ROOT / 'migrations' / '000102_pte_reading_bank_4a.down.sql'

# Live reading titles a passage title may share a word with, after review.
ALLOWED_TOPIC_OVERLAPS = {}

# Bank items bank_check may flag against a passage, after review.
ALLOWED_BANK_OVERLAPS = {}

RWFIB_INSTRUCTIONS = ('Below is a text with blanks. Click on each blank, a list of choices will appear. '
                      'Select the appropriate answer choice for each blank.')
RFIB_INSTRUCTIONS = ('In the text below some words are missing. Drag words from the box below to the '
                     'appropriate place in the text. To undo an answer choice, drag the word back to the '
                     'box below the text.')
SINGLE_INSTRUCTIONS = ('Read the text and answer the multiple-choice question by selecting the correct '
                       'response. Only one response is correct.')
MULTI_INSTRUCTIONS = ('Read the text and answer the question by selecting all the correct responses. '
                      'More than one response is correct.')


def rotate(options, keys, shift, explanation):
    """Moves each option `shift` places up, relabels A, B, C..., and follows the
    keys and any "(X)" letter references in the explanation to their new letters."""
    n = len(options)
    order = [(i + shift) % n for i in range(n)]
    new_options = [{'id': chr(65 + i), 'text': options[order[i]]['text']} for i in range(n)]
    moved = {options[order[i]]['id']: chr(65 + i) for i in range(n)}
    new_keys = sorted(moved[k] for k in keys)
    new_explanation = re.sub(r'\(([A-E])\)', lambda m: '(' + moved.get(m.group(1), m.group(1)) + ')', explanation)
    return new_options, new_keys, new_explanation


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


def blank_text(text):
    """The learner's view of a gapped text, for the bank check."""
    return re.sub(r'\[\[b\d+\]\]', '____', text)


def validate_and_compile():
    print(f"Validating {len(ALL_PASSAGES)} passages...")
    assert len(ALL_PASSAGES) == 10, f"Expected 10 passages, got {len(ALL_PASSAGES)}"

    batch = Batch()
    seen_ids = set()
    seen_titles = set()
    seen_prompts = set()

    passages_sql = []
    groups_sql = []
    questions_sql = []

    type_counts = Counter()
    single_keys = Counter()
    multi_keys = Counter()
    tags = sql_text_array(TAGS)

    global_single_idx = 0
    global_multi_idx = 0

    for p_num, p in enumerate(ALL_PASSAGES, start=1):
        validate_passage(p)
        pid = p['id']
        assert pid == f"rp-{PREFIX}-{p_num:03d}", f"Passage {p_num} has id {pid}"
        assert pid not in seen_ids, f"Duplicate passage ID {pid}"
        seen_ids.add(pid)

        ptitle = p['title']
        assert ptitle not in seen_titles, f"Duplicate passage title '{ptitle}'"
        seen_titles.add(ptitle)

        clashes = topic_clashes(ptitle, allowed=ALLOWED_TOPIC_OVERLAPS.get(ptitle, ()))
        assert not clashes, f"Passage '{ptitle}' clashes with live: {clashes}"

        paras = p['paragraphs']
        assert len(paras) == 6, f"{pid}: expected 6 paragraphs"
        wc = sum(len(x['text'].split()) for x in paras)
        assert 700 <= wc <= 900, f"{pid}: word count {wc} out of range 700-900"

        full_text = '\n\n'.join([p['subtitle']] + [x['text'] for x in paras])
        batch.passage(ptitle, full_text, where=pid, allowed=ALLOWED_BANK_OVERLAPS.get(pid, ()))
        batch.style(p['topic'], where=pid)

        passages_sql.append(
            f"    ({sql_lit(pid)}, {sql_lit(VERSION)}, {sql_lit(ptitle)}, {sql_lit(p['subtitle'])},\n"
            f"     {sql_lit(paras)}, {wc}, 'medium', {sql_lit(p['topic'])},\n"
            f"     {tags}, 'custom', ARRAY['academic'])"
        )

        slug = pid.replace(f'rp-{PREFIX}-', '')

        # Group 1: Reading & Writing: Fill in the Blanks (position 1)
        gid_rwfib = f"g-{PREFIX}-{slug}-rwfib"
        assert gid_rwfib not in seen_ids
        seen_ids.add(gid_rwfib)
        groups_sql.append(
            f"    ({sql_lit(gid_rwfib)}, {sql_lit(pid)}, 1, 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
            f"     {sql_lit(RWFIB_INSTRUCTIONS)},\n"
            f"     '[]'::jsonb, 'hidden', FALSE, 150)"
        )

        for q_idx, item in enumerate(p['rwfib'], start=1):
            qid = f"q-{PREFIX}-rwfib-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qtext = item['text']
            assert qtext not in seen_prompts, f"Duplicate question text in {qid}"
            seen_prompts.add(qtext)

            # Options are dealt in alphabetical order, so the key's place gives nothing away.
            blanks = [{'id': b['id'], 'options': sorted(b['options'], key=str.lower), 'correctAnswer': b['correctAnswer']}
                      for b in item['blanks']]
            assert len(blanks) >= 3, f"{qid}: expected >= 3 blanks"
            for b in blanks:
                assert f"[[{b['id']}]]" in qtext, f"{qid}: [[{b['id']}]] missing from text"
                assert len(b['options']) == 4, f"{qid}: blank {b['id']} must have 4 options"
                assert b['correctAnswer'] in b['options'], f"{qid}: answer not in options"
                assert len(set(b['options'])) == 4, f"{qid}: duplicate options"

            batch.question('PTE', 'fill-in-blanks-rw', qtitle, blank_text(qtext), where=qid)
            batch.style(' / '.join(o for b in blanks for o in b['options']), where=qid)
            batch.style(item['explanation'], where=qid)

            type_counts['fill-in-blanks-rw'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(RWFIB_INSTRUCTIONS)},\n"
                f"     {sql_lit(qtext)}, '[]'::jsonb, '[]'::jsonb, {sql_lit(blanks)},\n"
                f"     {sql_lit(item['explanation'])}, {len(blanks)}, 'medium', {tags},\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_rwfib)}, {q_idx}, TRUE, 150, NULL)"
            )

        # Group 2: Reading: Fill in the Blanks (position 2)
        gid_rfib = f"g-{PREFIX}-{slug}-rfib"
        assert gid_rfib not in seen_ids
        seen_ids.add(gid_rfib)
        groups_sql.append(
            f"    ({sql_lit(gid_rfib)}, {sql_lit(pid)}, 2, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
            f"     {sql_lit(RFIB_INSTRUCTIONS)},\n"
            f"     '[]'::jsonb, 'hidden', FALSE, 120)"
        )

        for q_idx, item in enumerate(p['rfib'], start=1):
            qid = f"q-{PREFIX}-rfib-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qtext = item['text']
            assert qtext not in seen_prompts, f"Duplicate question text in {qid}"
            seen_prompts.add(qtext)

            blanks = item['blanks']
            # The word box is dealt in alphabetical order, not answers first.
            pool = sorted(item['pool'], key=str.lower)
            assert len(blanks) >= 3, f"{qid}: expected >= 3 blanks"
            assert len(pool) >= len(blanks) + 2, f"{qid}: pool needs >= 2 distractors"
            assert len(set(pool)) == len(pool), f"{qid}: duplicate words in pool"
            for b in blanks:
                assert f"[[{b['id']}]]" in qtext, f"{qid}: [[{b['id']}]] missing from text"
                assert b['correctAnswer'] in pool, f"{qid}: answer {b['correctAnswer']} not in pool"

            # Formulate blanks with options = pool for complete client compatibility
            blanks_with_pool = [{'id': b['id'], 'options': pool, 'correctAnswer': b['correctAnswer']} for b in blanks]
            options_pool = [{'id': chr(65 + i), 'text': word} for i, word in enumerate(pool)]

            batch.question('PTE', 'fill-in-blanks-r', qtitle, blank_text(qtext), where=qid)
            batch.style(' / '.join(pool), where=qid)
            batch.style(item['explanation'], where=qid)

            type_counts['fill-in-blanks-r'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(RFIB_INSTRUCTIONS)},\n"
                f"     {sql_lit(qtext)}, {sql_lit(options_pool)}, '[]'::jsonb, {sql_lit(blanks_with_pool)},\n"
                f"     {sql_lit(item['explanation'])}, {len(blanks)}, 'medium', {tags},\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_rfib)}, {q_idx}, TRUE, 120, NULL)"
            )

        # Group 3: Multiple Choice, Single Answer (position 3)
        gid_single = f"g-{PREFIX}-{slug}-single"
        assert gid_single not in seen_ids
        seen_ids.add(gid_single)
        groups_sql.append(
            f"    ({sql_lit(gid_single)}, {sql_lit(pid)}, 3, 'reading-mcq-single', 'Multiple Choice, Single Answer',\n"
            f"     {sql_lit(SINGLE_INSTRUCTIONS)},\n"
            f"     '[]'::jsonb, 'full', FALSE, 90)"
        )

        for q_idx, item in enumerate(p['mcq_single'], start=1):
            qid = f"q-{PREFIX}-single-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qprompt = item['prompt']
            assert qprompt not in seen_prompts, f"Duplicate prompt in {qid}"
            seen_prompts.add(qprompt)

            # Rotate keys
            shift = global_single_idx % 4
            global_single_idx += 1
            new_opts, new_keys, new_exp = rotate(item['options'], [item['key']], shift, item['explanation'])
            single_keys[new_keys[0]] += 1

            batch.question('PTE', 'reading-mcq-single', qtitle,
                           qprompt + ' ' + ' '.join(o['text'] for o in new_opts), where=qid)
            batch.style(new_exp, where=qid)

            type_counts['reading-mcq-single'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single', 'Multiple Choice, Single Answer',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(qprompt)},\n"
                f"     NULL, {sql_lit(new_opts)}, {sql_lit(new_keys)}, '[]'::jsonb,\n"
                f"     {sql_lit(new_exp)}, 1, 'medium', {tags},\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_single)}, {q_idx}, TRUE, 90, NULL)"
            )

        # Group 4: Multiple Choice, Multiple Answers (position 4)
        gid_multi = f"g-{PREFIX}-{slug}-multi"
        assert gid_multi not in seen_ids
        seen_ids.add(gid_multi)
        groups_sql.append(
            f"    ({sql_lit(gid_multi)}, {sql_lit(pid)}, 4, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',\n"
            f"     {sql_lit(MULTI_INSTRUCTIONS)},\n"
            f"     '[]'::jsonb, 'full', FALSE, 150)"
        )

        for q_idx, item in enumerate(p['mcq_multiple'], start=1):
            qid = f"q-{PREFIX}-multi-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qprompt = item['prompt']
            assert qprompt not in seen_prompts, f"Duplicate prompt in {qid}"
            seen_prompts.add(qprompt)

            # Rotate keys
            shift = global_multi_idx % 5
            global_multi_idx += 1
            new_opts, new_keys, new_exp = rotate(item['options'], item['keys'], shift, item['explanation'])
            for k in new_keys:
                multi_keys[k] += 1

            batch.question('PTE', 'reading-mcq-multiple', qtitle,
                           qprompt + ' ' + ' '.join(o['text'] for o in new_opts), where=qid)
            batch.style(new_exp, where=qid)

            type_counts['reading-mcq-multiple'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(qprompt)},\n"
                f"     NULL, {sql_lit(new_opts)}, {sql_lit(new_keys)}, '[]'::jsonb,\n"
                f"     {sql_lit(new_exp)}, 2, 'medium', {tags},\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_multi)}, {q_idx}, TRUE, 150, NULL)"
            )

    # -------------------------------------------------------------------------
    # Assertions on distribution and counts
    # -------------------------------------------------------------------------
    print("\n--- Validation Statistics ---")
    print(f"Passages generated: {len(passages_sql)}")
    print(f"Groups generated:   {len(groups_sql)}")
    print(f"Total questions:    {len(questions_sql)}")
    print("\nQuestions per type:")
    assert set(type_counts) == {'fill-in-blanks-rw', 'fill-in-blanks-r', 'reading-mcq-single', 'reading-mcq-multiple'}
    for k, v in sorted(type_counts.items()):
        print(f"  {k}: {v}")
        assert v == 50, f"Expected exactly 50 questions for {k}, got {v}"

    print(f"\nRMCSA Key Letter Distribution (50 total): {dict(sorted(single_keys.items()))}")
    assert set(single_keys.keys()) == set('ABCD'), "RMCSA keys must use A, B, C, D"
    assert max(single_keys.values()) <= 15 and min(single_keys.values()) >= 10, "RMCSA keys not balanced"

    print(f"RMCMA Letter Distribution (100 total):    {dict(sorted(multi_keys.items()))}")
    assert set(multi_keys.keys()) == set('ABCDE'), "RMCMA keys must use A, B, C, D, E"
    assert max(multi_keys.values()) <= 25 and min(multi_keys.values()) >= 15, "RMCMA letters not balanced"

    # No RWFIB wrong option is used more than twice across the whole bank
    rwfib_wrong_counts = Counter()
    for p in ALL_PASSAGES:
        for item in p['rwfib']:
            for b in item['blanks']:
                wrongs = [opt for opt in b['options'] if opt != b['correctAnswer']]
                assert len(wrongs) == 3, f"Expected 3 wrong options, got {len(wrongs)}"
                for w in wrongs:
                    rwfib_wrong_counts[w] += 1
    top_rwfib_wrongs = rwfib_wrong_counts.most_common(1)
    if top_rwfib_wrongs:
        assert top_rwfib_wrongs[0][1] <= 2, f"RWFIB wrong option '{top_rwfib_wrongs[0][0]}' used {top_rwfib_wrongs[0][1]} times (max 2 allowed)"
    print(f"RWFIB Wrong Option Distribution: {len(rwfib_wrong_counts)} unique words, max frequency = {top_rwfib_wrongs[0][1] if top_rwfib_wrongs else 0}")

    # Nothing is written unless the whole bank passes the bank check.
    batch.assert_clean()
    print("Bank check: clean")

    up_header = (
        "-- Migration: 000102_pte_reading_bank_4a.up.sql\n"
        "-- Generated by scripts/generate_pte_reading_b4a.py; do not edit by hand.\n"
        "-- 10 original academic reading passages (6 paragraphs A-F, 700-900 words).\n"
        "-- 200 PTE Reading questions (50 RWFIB, 50 RFIB, 50 RMCSA, 50 RMCMA).\n"
        "-- Replay-safe: INSERT ... ON CONFLICT (id) DO NOTHING.\n\n"
    )

    passages_insert = (
        "INSERT INTO reading_passages\n"
        "    (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot, modules)\n"
        "VALUES\n" + ",\n".join(passages_sql) + "\nON CONFLICT (id) DO NOTHING;\n\n"
    )

    groups_insert = (
        "INSERT INTO reading_question_groups\n"
        "    (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds)\n"
        "VALUES\n" + ",\n".join(groups_sql) + "\nON CONFLICT (id) DO NOTHING;\n\n"
    )

    questions_insert = (
        "INSERT INTO questions\n"
        "    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,\n"
        "     title, prompt, context_passage, options, correct_answers, blanks,\n"
        "     explanation, points, difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds, reorder_item_id)\n"
        "VALUES\n" + ",\n".join(questions_sql) + "\nON CONFLICT (id) DO NOTHING;\n"
    )

    up_sql = up_header + passages_insert + groups_insert + questions_insert

    down_sql = (
        "-- Migration: 000102_pte_reading_bank_4a.down.sql\n"
        "SELECT 1;\n"
    )

    return up_sql, down_sql


def write_migrations():
    up_sql, down_sql = validate_and_compile()

    UP_MIGRATION.write_text(up_sql, encoding='utf-8', newline='\n')
    DOWN_MIGRATION.write_text(down_sql, encoding='utf-8', newline='\n')

    sha_up = hashlib.sha256(UP_MIGRATION.read_bytes()).hexdigest()
    print(f"\nWrote {UP_MIGRATION} ({UP_MIGRATION.stat().st_size:,} bytes, sha256={sha_up})")
    print(f"Wrote {DOWN_MIGRATION} ({DOWN_MIGRATION.stat().st_size:,} bytes)")


if __name__ == '__main__':
    write_migrations()
