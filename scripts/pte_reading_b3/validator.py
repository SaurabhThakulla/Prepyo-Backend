"""Validator helper for PTE Reading Batch 3 passages and questions."""

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import topic_clashes, question_clashes, _words

def validate_passage(p):
    pid = p['id']
    title = p['title']
    print(f"\nValidating passage {pid}: '{title}'")

    # Clashes
    clashes = topic_clashes(title)
    if clashes:
        print(f"  WARNING: topic clashes for '{title}': {clashes}")

    # Paragraphs
    paras = p['paragraphs']
    assert len(paras) == 6, f"{pid}: expected 6 paragraphs, got {len(paras)}"
    for idx, expected in enumerate(['A', 'B', 'C', 'D', 'E', 'F']):
        assert paras[idx]['label'] == expected, f"{pid}: paragraph {idx} label != {expected}"

    wc = sum(len(x['text'].split()) for x in paras)
    assert 700 <= wc <= 900, f"{pid}: word count {wc} out of range 700-900"
    print(f"  Word count: {wc} words across 6 paragraphs")

    # RWFIB (5 items)
    rwfib = p['rwfib']
    assert len(rwfib) == 5, f"{pid}: expected 5 RWFIB, got {len(rwfib)}"
    for q_idx, item in enumerate(rwfib, start=1):
        assert item['title'], f"{pid} RWFIB {q_idx}: missing title"
        assert item['text'], f"{pid} RWFIB {q_idx}: missing text"
        assert len(item['blanks']) >= 3, f"{pid} RWFIB {q_idx}: expected >= 3 blanks"
        for b in item['blanks']:
            bid = b['id']
            assert f"[[{bid}]]" in item['text'], f"{pid} RWFIB {q_idx}: [[{bid}]] missing from text"
            assert len(b['options']) == 4, f"{pid} RWFIB {q_idx} blank {bid}: expected 4 options"
            assert b['correctAnswer'] in b['options'], f"{pid} RWFIB {q_idx} blank {bid}: answer not in options"
            assert len(set(b['options'])) == 4, f"{pid} RWFIB {q_idx} blank {bid}: options not distinct"

    # RFIB (5 items)
    rfib = p['rfib']
    assert len(rfib) == 5, f"{pid}: expected 5 RFIB, got {len(rfib)}"
    for q_idx, item in enumerate(rfib, start=1):
        assert item['title'], f"{pid} RFIB {q_idx}: missing title"
        assert item['text'], f"{pid} RFIB {q_idx}: missing text"
        assert len(item['blanks']) >= 3, f"{pid} RFIB {q_idx}: expected >= 3 blanks"
        for b in item['blanks']:
            bid = b['id']
            assert f"[[{bid}]]" in item['text'], f"{pid} RFIB {q_idx}: [[{bid}]] missing from text"
            assert b['correctAnswer'] in item['pool'], f"{pid} RFIB {q_idx} blank {bid}: answer not in pool"
        assert len(item['pool']) >= len(item['blanks']) + 2, f"{pid} RFIB {q_idx}: pool needs >= 2 distractors"
        assert len(set(item['pool'])) == len(item['pool']), f"{pid} RFIB {q_idx}: pool has duplicates"

    # RMCSA (5 items)
    single = p['mcq_single']
    assert len(single) == 5, f"{pid}: expected 5 RMCSA, got {len(single)}"
    for q_idx, item in enumerate(single, start=1):
        assert item['title'], f"{pid} RMCSA {q_idx}: missing title"
        assert item['prompt'], f"{pid} RMCSA {q_idx}: missing prompt"
        assert len(item['options']) == 4, f"{pid} RMCSA {q_idx}: expected 4 options"
        opt_ids = [opt['id'] for opt in item['options']]
        assert opt_ids == ['A', 'B', 'C', 'D'], f"{pid} RMCSA {q_idx}: options must be A, B, C, D"
        assert item['key'] in opt_ids, f"{pid} RMCSA {q_idx}: key not in options"

    # RMCMA (5 items)
    multi = p['mcq_multiple']
    assert len(multi) == 5, f"{pid}: expected 5 RMCMA, got {len(multi)}"
    for q_idx, item in enumerate(multi, start=1):
        assert item['title'], f"{pid} RMCMA {q_idx}: missing title"
        assert item['prompt'], f"{pid} RMCMA {q_idx}: missing prompt"
        assert len(item['options']) == 5, f"{pid} RMCMA {q_idx}: expected 5 options"
        opt_ids = [opt['id'] for opt in item['options']]
        assert opt_ids == ['A', 'B', 'C', 'D', 'E'], f"{pid} RMCMA {q_idx}: options must be A, B, C, D, E"
        assert len(item['keys']) == 2, f"{pid} RMCMA {q_idx}: expected exactly 2 keys"
        for k in item['keys']:
            assert k in opt_ids, f"{pid} RMCMA {q_idx}: key {k} not in options"

    print(f"  All 20 questions for {pid} valid!")
    return True

def validate_reorders(reorders):
    print(f"\nValidating {len(reorders)} Re-order Paragraphs items...")
    assert len(reorders) == 50, f"Expected 50 reorders, got {len(reorders)}"
    titles = [it['title'] for it in reorders]
    assert len(titles) == len(set(titles)), "Duplicate reorder titles found!"

    total_words = 0
    for idx, it in enumerate(reorders, start=1):
        rid = it['id']
        paras = it['paragraphs']
        assert 4 <= len(paras) <= 5, f"{rid}: expected 4-5 paragraphs, got {len(paras)}"
        wc = sum(len(p['text'].split()) for p in paras)
        assert 60 <= wc <= 160, f"{rid}: word count {wc} out of range"
        total_words += wc
        labels = [p['label'] for p in paras]
        expected_labels = [f"p{i}" for i in range(1, len(paras) + 1)]
        assert labels == expected_labels, f"{rid}: labels {labels} != {expected_labels}"

        clashes = question_clashes('PTE', it['title'], type_name='Re-order Paragraphs')
        assert not clashes, f"{rid}: clashes with live: {clashes}"

    print(f"  All 50 Re-order items valid! Total words: {total_words}, average: {total_words/50:.1f} words/item")
    return True

if __name__ == '__main__':
    from scripts.pte_reading_b3.passages import ALL_PASSAGES
    from scripts.pte_reading_b3.reorders import ALL_REORDERS
    for p in ALL_PASSAGES:
        validate_passage(p)
    validate_reorders(ALL_REORDERS)
