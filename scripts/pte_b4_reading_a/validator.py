"""Validator helper for PTE Reading bank 4a passages and questions.

Run on its own to check the passages written so far:
    python -m scripts.pte_b4_reading_a.validator
"""

import importlib
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import topic_clashes


def validate_passage(p):
    pid = p['id']
    title = p['title']
    print(f"\nValidating passage {pid}: '{title}'")

    clashes = topic_clashes(title)
    if clashes:
        print(f"  WARNING: topic clashes for '{title}': {clashes}")

    paras = p['paragraphs']
    assert len(paras) == 6, f"{pid}: expected 6 paragraphs, got {len(paras)}"
    for idx, expected in enumerate(['A', 'B', 'C', 'D', 'E', 'F']):
        assert paras[idx]['label'] == expected, f"{pid}: paragraph {idx} label != {expected}"

    wc = sum(len(x['text'].split()) for x in paras)
    assert 700 <= wc <= 900, f"{pid}: word count {wc} out of range 700-900"
    print(f"  Word count: {wc} words across 6 paragraphs")

    # RWFIB (5 items): 3+ blanks, 4 distinct options each, one of them the answer.
    rwfib = p['rwfib']
    assert len(rwfib) == 5, f"{pid}: expected 5 RWFIB, got {len(rwfib)}"
    for q_idx, item in enumerate(rwfib, start=1):
        assert item['title'], f"{pid} RWFIB {q_idx}: missing title"
        assert item['text'], f"{pid} RWFIB {q_idx}: missing text"
        assert len(item['blanks']) >= 3, f"{pid} RWFIB {q_idx}: expected >= 3 blanks"
        for b in item['blanks']:
            bid = b['id']
            assert item['text'].count(f"[[{bid}]]") == 1, f"{pid} RWFIB {q_idx}: [[{bid}]] must appear once"
            assert len(b['options']) == 4, f"{pid} RWFIB {q_idx} blank {bid}: expected 4 options"
            assert b['correctAnswer'] in b['options'], f"{pid} RWFIB {q_idx} blank {bid}: answer not in options"
            assert len(set(b['options'])) == 4, f"{pid} RWFIB {q_idx} blank {bid}: options not distinct"

    # RFIB (5 items): 3+ blanks, a shared pool with at least 2 distractors.
    rfib = p['rfib']
    assert len(rfib) == 5, f"{pid}: expected 5 RFIB, got {len(rfib)}"
    for q_idx, item in enumerate(rfib, start=1):
        assert item['title'], f"{pid} RFIB {q_idx}: missing title"
        assert item['text'], f"{pid} RFIB {q_idx}: missing text"
        assert len(item['blanks']) >= 3, f"{pid} RFIB {q_idx}: expected >= 3 blanks"
        answers = [b['correctAnswer'] for b in item['blanks']]
        assert len(set(answers)) == len(answers), f"{pid} RFIB {q_idx}: two blanks share an answer"
        for b in item['blanks']:
            bid = b['id']
            assert item['text'].count(f"[[{bid}]]") == 1, f"{pid} RFIB {q_idx}: [[{bid}]] must appear once"
            assert b['correctAnswer'] in item['pool'], f"{pid} RFIB {q_idx} blank {bid}: answer not in pool"
        assert len(item['pool']) >= len(item['blanks']) + 2, f"{pid} RFIB {q_idx}: pool needs >= 2 distractors"
        assert len(set(item['pool'])) == len(item['pool']), f"{pid} RFIB {q_idx}: pool has duplicates"

    # RMCSA (5 items): 4 options A-D, one key.
    single = p['mcq_single']
    assert len(single) == 5, f"{pid}: expected 5 RMCSA, got {len(single)}"
    for q_idx, item in enumerate(single, start=1):
        assert item['title'], f"{pid} RMCSA {q_idx}: missing title"
        assert item['prompt'], f"{pid} RMCSA {q_idx}: missing prompt"
        assert len(item['options']) == 4, f"{pid} RMCSA {q_idx}: expected 4 options"
        opt_ids = [opt['id'] for opt in item['options']]
        assert opt_ids == ['A', 'B', 'C', 'D'], f"{pid} RMCSA {q_idx}: options must be A, B, C, D"
        assert item['key'] in opt_ids, f"{pid} RMCSA {q_idx}: key not in options"
        assert len({o['text'] for o in item['options']}) == 4, f"{pid} RMCSA {q_idx}: duplicate option text"

    # RMCMA (5 items): 5 options A-E, exactly two keys.
    multi = p['mcq_multiple']
    assert len(multi) == 5, f"{pid}: expected 5 RMCMA, got {len(multi)}"
    for q_idx, item in enumerate(multi, start=1):
        assert item['title'], f"{pid} RMCMA {q_idx}: missing title"
        assert item['prompt'], f"{pid} RMCMA {q_idx}: missing prompt"
        assert 'TWO' in item['prompt'], f"{pid} RMCMA {q_idx}: prompt must ask for TWO answers"
        assert len(item['options']) == 5, f"{pid} RMCMA {q_idx}: expected 5 options"
        opt_ids = [opt['id'] for opt in item['options']]
        assert opt_ids == ['A', 'B', 'C', 'D', 'E'], f"{pid} RMCMA {q_idx}: options must be A, B, C, D, E"
        assert len(item['keys']) == 2 and len(set(item['keys'])) == 2, f"{pid} RMCMA {q_idx}: expected exactly 2 keys"
        for k in item['keys']:
            assert k in opt_ids, f"{pid} RMCMA {q_idx}: key {k} not in options"
        assert len({o['text'] for o in item['options']}) == 5, f"{pid} RMCMA {q_idx}: duplicate option text"

    print(f"  All 20 questions for {pid} valid!")
    return True


if __name__ == '__main__':
    for n in range(1, 11):
        try:
            mod = importlib.import_module(f'scripts.pte_b4_reading_a.passage_{n}')
        except ModuleNotFoundError:
            continue
        validate_passage(getattr(mod, f'PASSAGE_{n}'))
