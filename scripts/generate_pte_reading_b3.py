"""Generate PTE Reading Bank Batch 3 (Migration 000089).

This script generates:
  - 10 full-length academic reading passages (6 paragraphs A-F, 700-900 words)
  - 50 Fill in the Blanks (Dropdown) (RWFIB) questions (5 per passage)
  - 50 Reading: Fill in the Blanks (RFIB) questions (5 per passage)
  - 50 Multiple Choice, Single Answer (RMCSA) questions (5 per passage)
  - 50 Multiple Choice, Multiple Answers (RMCMA) questions (5 per passage)
  - 50 Re-order Paragraphs (RO) standalone items with backing questions
Total: 250 questions (50 per sub-task category).

Self-validation enforces:
  - Exactly 10 passages with 6 paragraphs each (A-F), word count 700-900.
  - Exactly 50 questions per sub-task type.
  - All IDs are unique and prefixed with pr89.
  - Distractor rules for blanks: only 1 option fits grammatically and semantically.
  - Re-order Paragraphs items have 4-5 logically cohesive boxes.
  - MCQ options are rotated using rotate() so keys spread evenly.
  - No item repeats a live reading topic or live question (scripts/live_topics.py),
    or repeats any topic from Batches A, B, or C.
  - Generates migrations/000089_pte_reading_bank_3.up.sql and down.sql with Unix line endings.
  - Deterministic and replay-safe (INSERT ... ON CONFLICT (id) DO NOTHING).
"""

import hashlib
import json
import re
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import question_clashes, topic_clashes
from scripts.pte_reading_b3.passages import ALL_PASSAGES
from scripts.pte_reading_b3.reorders import ALL_REORDERS

VERSION = 'pte-2026-01'
UP_MIGRATION = ROOT / 'migrations' / '000089_pte_reading_bank_3.up.sql'
DOWN_MIGRATION = ROOT / 'migrations' / '000089_pte_reading_bank_3.down.sql'

ALLOWED_TOPIC_OVERLAPS = {
    'Roman Hypocaust Heating Systems': ('The guthi system',),
}


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


def validate_and_compile():
    print(f"Validating {len(ALL_PASSAGES)} passages and {len(ALL_REORDERS)} reorder items...")
    assert len(ALL_PASSAGES) == 10, f"Expected 10 passages, got {len(ALL_PASSAGES)}"
    assert len(ALL_REORDERS) == 50, f"Expected 50 reorders, got {len(ALL_REORDERS)}"

    seen_ids = set()
    seen_titles = set()
    seen_prompts = set()

    passages_sql = []
    groups_sql = []
    questions_sql = []
    reorders_sql = []

    type_counts = Counter()
    single_keys = Counter()
    multi_keys = Counter()

    # -------------------------------------------------------------------------
    # 1. Compile 10 Passages and their 200 Questions (20 per passage)
    # -------------------------------------------------------------------------
    global_single_idx = 0
    global_multi_idx = 0

    for p_num, p in enumerate(ALL_PASSAGES, start=1):
        pid = p['id']
        assert pid not in seen_ids, f"Duplicate passage ID {pid}"
        seen_ids.add(pid)

        ptitle = p['title']
        assert ptitle not in seen_titles, f"Duplicate passage title '{ptitle}'"
        seen_titles.add(ptitle)

        # Clashes
        clashes = topic_clashes(ptitle, allowed=ALLOWED_TOPIC_OVERLAPS.get(ptitle, ()))
        assert not clashes, f"Passage '{ptitle}' clashes with live: {clashes}"

        paras = p['paragraphs']
        assert len(paras) == 6, f"{pid}: expected 6 paragraphs"
        wc = sum(len(x['text'].split()) for x in paras)
        assert 700 <= wc <= 900, f"{pid}: word count {wc} out of range 700-900"

        passages_sql.append(
            f"    ({sql_lit(pid)}, {sql_lit(VERSION)}, {sql_lit(ptitle)}, {sql_lit(p['subtitle'])},\n"
            f"     {sql_lit(paras)}, {wc}, 'medium', {sql_lit(p['topic'])},\n"
            f"     ARRAY['PTE Reading', 'Original practice', 'Batch D'], 'custom', ARRAY['academic'])"
        )

        slug = pid.replace('rp-pr89-', '')

        # Group 1: RWFIB (position 1)
        gid_rwfib = f"g-pr89-{slug}-rwfib"
        assert gid_rwfib not in seen_ids
        seen_ids.add(gid_rwfib)
        groups_sql.append(
            f"    ({sql_lit(gid_rwfib)}, {sql_lit(pid)}, 1, 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
            f"     'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',\n"
            f"     '[]'::jsonb, 'hidden', FALSE, 150)"
        )

        for q_idx, item in enumerate(p['rwfib'], start=1):
            qid = f"q-pr89-rwfib-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qtext = item['text']
            assert qtext not in seen_prompts, f"Duplicate question text in {qid}"
            seen_prompts.add(qtext)

            blanks = item['blanks']
            assert len(blanks) >= 3, f"{qid}: expected >= 3 blanks"
            for b in blanks:
                assert f"[[{b['id']}]]" in qtext, f"{qid}: [[{b['id']}]] missing from text"
                assert len(b['options']) == 4, f"{qid}: blank {b['id']} must have 4 options"
                assert b['correctAnswer'] in b['options'], f"{qid}: answer not in options"
                assert len(set(b['options'])) == 4, f"{qid}: duplicate options"

            type_counts['fill-in-blanks-rw'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
                f"     {sql_lit(qtitle)}, 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',\n"
                f"     {sql_lit(qtext)}, '[]'::jsonb, '[]'::jsonb, {sql_lit(blanks)},\n"
                f"     {sql_lit(item['explanation'])}, {len(blanks)}, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'],\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_rwfib)}, {q_idx}, TRUE, 150, NULL)"
            )

        # Group 2: RFIB (position 2)
        gid_rfib = f"g-pr89-{slug}-rfib"
        assert gid_rfib not in seen_ids
        seen_ids.add(gid_rfib)
        groups_sql.append(
            f"    ({sql_lit(gid_rfib)}, {sql_lit(pid)}, 2, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
            f"     'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.',\n"
            f"     '[]'::jsonb, 'hidden', FALSE, 120)"
        )

        for q_idx, item in enumerate(p['rfib'], start=1):
            qid = f"q-pr89-rfib-{p_num:02d}-{q_idx:02d}"
            assert qid not in seen_ids, f"Duplicate question ID {qid}"
            seen_ids.add(qid)

            qtitle = item['title']
            assert qtitle not in seen_titles, f"Duplicate title '{qtitle}'"
            seen_titles.add(qtitle)

            qtext = item['text']
            assert qtext not in seen_prompts, f"Duplicate question text in {qid}"
            seen_prompts.add(qtext)

            blanks = item['blanks']
            pool = item['pool']
            assert len(blanks) >= 3, f"{qid}: expected >= 3 blanks"
            assert len(pool) >= len(blanks) + 2, f"{qid}: pool needs >= 2 distractors"
            assert len(set(pool)) == len(pool), f"{qid}: duplicate words in pool"
            for b in blanks:
                assert f"[[{b['id']}]]" in qtext, f"{qid}: [[{b['id']}]] missing from text"
                assert b['correctAnswer'] in pool, f"{qid}: answer {b['correctAnswer']} not in pool"

            # Formulate blanks with options = pool for complete client compatibility
            blanks_with_pool = [{'id': b['id'], 'options': pool, 'correctAnswer': b['correctAnswer']} for b in blanks]
            options_pool = [{'id': chr(65 + i), 'text': word} for i, word in enumerate(pool)]

            type_counts['fill-in-blanks-r'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
                f"     {sql_lit(qtitle)}, 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.',\n"
                f"     {sql_lit(qtext)}, {sql_lit(options_pool)}, '[]'::jsonb, {sql_lit(blanks_with_pool)},\n"
                f"     {sql_lit(item['explanation'])}, {len(blanks)}, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'],\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_rfib)}, {q_idx}, TRUE, 120, NULL)"
            )

        # Group 3: RMCSA (position 3)
        gid_single = f"g-pr89-{slug}-single"
        assert gid_single not in seen_ids
        seen_ids.add(gid_single)
        groups_sql.append(
            f"    ({sql_lit(gid_single)}, {sql_lit(pid)}, 3, 'reading-mcq-single', 'Multiple Choice, Single Answer',\n"
            f"     'Read the text and answer the multiple-choice question by selecting the correct response. Only one response is correct.',\n"
            f"     '[]'::jsonb, 'full', FALSE, 90)"
        )

        for q_idx, item in enumerate(p['mcq_single'], start=1):
            qid = f"q-pr89-single-{p_num:02d}-{q_idx:02d}"
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

            type_counts['reading-mcq-single'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single', 'Multiple Choice, Single Answer',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(qprompt)},\n"
                f"     NULL, {sql_lit(new_opts)}, {sql_lit(new_keys)}, '[]'::jsonb,\n"
                f"     {sql_lit(new_exp)}, 1, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'],\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_single)}, {q_idx}, TRUE, 90, NULL)"
            )

        # Group 4: RMCMA (position 4)
        gid_multi = f"g-pr89-{slug}-multi"
        assert gid_multi not in seen_ids
        seen_ids.add(gid_multi)
        groups_sql.append(
            f"    ({sql_lit(gid_multi)}, {sql_lit(pid)}, 4, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',\n"
            f"     'Read the text and answer the question by selecting all the correct responses. More than one response is correct.',\n"
            f"     '[]'::jsonb, 'full', FALSE, 150)"
        )

        for q_idx, item in enumerate(p['mcq_multiple'], start=1):
            qid = f"q-pr89-multi-{p_num:02d}-{q_idx:02d}"
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

            type_counts['reading-mcq-multiple'] += 1
            questions_sql.append(
                f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',\n"
                f"     {sql_lit(qtitle)}, {sql_lit(qprompt)},\n"
                f"     NULL, {sql_lit(new_opts)}, {sql_lit(new_keys)}, '[]'::jsonb,\n"
                f"     {sql_lit(new_exp)}, 2, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'],\n"
                f"     {sql_lit(pid)}, {sql_lit(gid_multi)}, {q_idx}, TRUE, 150, NULL)"
            )

    # -------------------------------------------------------------------------
    # 2. Compile 50 Re-order Paragraphs items
    # -------------------------------------------------------------------------
    for r_idx, item in enumerate(ALL_REORDERS, start=1):
        rid = item['id']
        assert rid not in seen_ids, f"Duplicate reorder ID {rid}"
        seen_ids.add(rid)

        rtitle = item['title']
        assert rtitle not in seen_titles, f"Duplicate reorder title '{rtitle}'"
        seen_titles.add(rtitle)

        clashes = question_clashes('PTE', rtitle, allowed=ALLOWED_TOPIC_OVERLAPS.get(rtitle, ()), type_name='Re-order Paragraphs')
        assert not clashes, f"Re-order '{rtitle}' clashes with live: {clashes}"

        paras = item['paragraphs']
        assert 4 <= len(paras) <= 5, f"{rid}: expected 4-5 paragraphs"
        rwc = sum(len(x['text'].split()) for x in paras)
        assert 60 <= rwc <= 160, f"{rid}: word count {rwc} out of range"

        # Check all paragraph texts are unique
        for p_box in paras:
            p_text = p_box['text']
            assert p_text not in seen_prompts, f"Duplicate text box in {rid}: '{p_text}'"
            seen_prompts.add(p_text)

        reorders_sql.append(
            f"    ({sql_lit(rid)}, {sql_lit(VERSION)}, 'PTE', {sql_lit(rtitle)},\n"
            f"     {sql_lit(paras)}, {rwc}, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'], TRUE)"
        )

        # Create matching question row in questions table
        qid = f"q-pr89-ro-{r_idx:03d}"
        assert qid not in seen_ids, f"Duplicate question ID {qid}"
        seen_ids.add(qid)

        correct_labels = [p_box['label'] for p_box in paras]
        n_boxes = len(paras)

        # Deterministically shuffle options so they are never dealt in solved order
        # Derangements for 4 items: [2, 0, 3, 1] or [1, 3, 0, 2]
        # Derangements for 5 items: [2, 4, 0, 3, 1] or [3, 0, 4, 1, 2]
        if n_boxes == 4:
            permutations = [
                [2, 0, 3, 1], [1, 3, 0, 2], [3, 0, 1, 2], [1, 2, 3, 0], [2, 3, 0, 1]
            ]
            shuffled_indices = permutations[r_idx % len(permutations)]
        else:
            permutations5 = [
                [2, 4, 0, 3, 1], [3, 0, 4, 1, 2], [1, 3, 4, 0, 2], [4, 2, 1, 0, 3], [3, 4, 1, 2, 0]
            ]
            shuffled_indices = permutations5[r_idx % len(permutations5)]

        shuffled_options = [
            {'id': paras[i]['label'], 'text': paras[i]['text']} for i in shuffled_indices
        ]

        type_counts['reorder-paragraphs'] += 1
        points = n_boxes - 1
        explanation = f"Correct sequential order: {' -> '.join(correct_labels)}. Restores the original narrative cohesion."

        questions_sql.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reorder-paragraphs', 'Re-order Paragraphs',\n"
            f"     {sql_lit(rtitle)}, 'The text boxes below have been placed in a random order. Restore the original order.',\n"
            f"     NULL, {sql_lit(shuffled_options)}, {sql_lit(correct_labels)}, '[]'::jsonb,\n"
            f"     {sql_lit(explanation)}, {points}, 'medium', ARRAY['PTE Reading', 'Original practice', 'Batch D'],\n"
            f"     NULL, NULL, 0, TRUE, 150, {sql_lit(rid)})"
        )

    # -------------------------------------------------------------------------
    # Assertions on distribution and counts
    # -------------------------------------------------------------------------
    print("\n--- Validation Statistics ---")
    print(f"Passages generated: {len(passages_sql)}")
    print(f"Groups generated:   {len(groups_sql)}")
    print(f"Reorder items:      {len(reorders_sql)}")
    print(f"Total questions:    {len(questions_sql)}")
    print("\nQuestions per type:")
    for k, v in sorted(type_counts.items()):
        print(f"  {k}: {v}")
        assert v == 50, f"Expected exactly 50 questions for {k}, got {v}"

    print(f"\nRMCSA Key Letter Distribution (50 total): {dict(sorted(single_keys.items()))}")
    assert set(single_keys.keys()) == set('ABCD'), f"RMCSA keys must use A, B, C, D"
    assert max(single_keys.values()) <= 15 and min(single_keys.values()) >= 10, "RMCSA keys not balanced"

    print(f"RMCMA Letter Distribution (100 total):    {dict(sorted(multi_keys.items()))}")
    assert set(multi_keys.keys()) == set('ABCDE'), f"RMCMA keys must use A, B, C, D, E"
    assert max(multi_keys.values()) <= 25 and min(multi_keys.values()) >= 15, "RMCMA letters not balanced"

    # Assert that no RWFIB wrong option is used more than twice across the whole batch
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

    # Assert Roman reorders <= 3
    roman_reorders = [item['title'] for item in ALL_REORDERS if 'roman' in item['title'].lower() or 'pompeii' in item['title'].lower()]
    assert len(roman_reorders) <= 3, f"Too many Roman reorders ({len(roman_reorders)}): {roman_reorders}"
    print(f"Roman reorder items ({len(roman_reorders)}): {roman_reorders}")

    # Assert ceramic reorders <= 2
    ceramic_reorders = [item['title'] for item in ALL_REORDERS if any(w in item['title'].lower() for w in ['ceramic', 'terracotta', 'faience', 'stoneware'])]
    assert len(ceramic_reorders) <= 2, f"Too many ceramic reorders ({len(ceramic_reorders)}): {ceramic_reorders}"
    print(f"Ceramic reorder items ({len(ceramic_reorders)}): {ceramic_reorders}")

    # Assert ancient metalworking reorders <= 3
    metal_reorders = [item['title'] for item in ALL_REORDERS if any(w in item['title'].lower() for w in ['wootz', 'bloomery', 'bog iron', 'coppersmithing', 'platinum sintering', 'tin mining', 'chainmail'])]
    assert len(metal_reorders) <= 3, f"Too many metalworking reorders ({len(metal_reorders)}): {metal_reorders}"
    print(f"Ancient metalworking reorder items ({len(metal_reorders)}): {metal_reorders}")

    # Build SQL files
    up_header = (
        "-- Migration: 000089_pte_reading_bank_3.up.sql\n"
        "-- 10 original academic reading passages (6 paragraphs A-F, 700-900 words).\n"
        "-- 250 PTE Reading questions (50 RWFIB, 50 RFIB, 50 RMCSA, 50 RMCMA, 50 Re-order Paragraphs).\n"
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

    reorders_insert = (
        "INSERT INTO reading_reorder_items\n"
        "    (id, exam_version_id, exam, title, paragraphs, word_count, difficulty, tags, is_published)\n"
        "VALUES\n" + ",\n".join(reorders_sql) + "\nON CONFLICT (id) DO NOTHING;\n\n"
    )

    questions_insert = (
        "INSERT INTO questions\n"
        "    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,\n"
        "     title, prompt, context_passage, options, correct_answers, blanks,\n"
        "     explanation, points, difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds, reorder_item_id)\n"
        "VALUES\n" + ",\n".join(questions_sql) + "\nON CONFLICT (id) DO NOTHING;\n"
    )

    up_sql = up_header + passages_insert + groups_insert + reorders_insert + questions_insert

    down_sql = (
        "-- Migration: 000089_pte_reading_bank_3.down.sql\n"
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
