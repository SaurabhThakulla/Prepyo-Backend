"""Validator helper for PTE Reading bank 4b Re-order Paragraphs items.

Modelled on the reorder part of scripts/pte_reading_b3/validator.py, with extra
checks that the boxes can only be put back in one order: the first box must not
lean on anything before it, and every later box must be tied to the box before.
"""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import question_clashes

# A first box must not open with a word that points back to something earlier.
OPENING_BANNED = {
    'this', 'these', 'that', 'those', 'it', 'its', 'they', 'their', 'them', 'he', 'she', 'his', 'her',
    'such', 'however', 'but', 'and', 'also', 'therefore', 'thus', 'then', 'yet', 'so', 'because',
    'instead', 'afterwards', 'later', 'finally', 'even', 'both', 'further', 'in fact', 'as a result',
}

# Every later box must be tied to the box before it: its first sentence either
# uses a reference word or connective, or picks up a content word from that box.
LINK_WORDS = re.compile(
    r"\b(this|these|that|those|it|its|they|their|them|he|she|his|her|such|however|but|also|therefore|"
    r"thus|then|yet|so|because|instead|afterwards|later|since|further|contrast|fact|indeed|even so|"
    r"reason|reasons|exchange|there|among|both|each|same|above|result)\b",
    re.IGNORECASE,
)


def _content_words(text):
    return {w for w in re.findall(r"[a-z]+", text.lower()) if len(w) >= 5}


def linked_to_previous(box, previous):
    first_sentence = re.split(r'(?<=[.!?])\s', box)[0]
    return bool(LINK_WORDS.search(first_sentence) or _content_words(first_sentence) & _content_words(previous))


def validate_reorders(reorders):
    print(f"\nValidating {len(reorders)} Re-order Paragraphs items...")
    assert len(reorders) == 50, f"Expected 50 reorders, got {len(reorders)}"
    titles = [it['title'] for it in reorders]
    assert len(titles) == len(set(titles)), "Duplicate reorder titles found!"

    total_words = 0
    for idx, it in enumerate(reorders, start=1):
        rid = it['id']
        assert rid == f"ri-pr103-ro-{idx:03d}", f"Item {idx} has id {rid}"
        paras = it['paragraphs']
        assert 4 <= len(paras) <= 5, f"{rid}: expected 4-5 paragraphs, got {len(paras)}"
        wc = sum(len(p['text'].split()) for p in paras)
        assert 60 <= wc <= 160, f"{rid}: word count {wc} out of range"
        total_words += wc
        labels = [p['label'] for p in paras]
        expected_labels = [f"p{i}" for i in range(1, len(paras) + 1)]
        assert labels == expected_labels, f"{rid}: labels {labels} != {expected_labels}"
        texts = [p['text'] for p in paras]
        assert len(set(texts)) == len(texts), f"{rid}: duplicate box text"

        first = paras[0]['text'].lower()
        for word in OPENING_BANNED:
            assert not re.match(re.escape(word) + r'\b', first), f"{rid}: first box opens with '{word}'"
        for prev, p in zip(paras, paras[1:]):
            assert linked_to_previous(p['text'], prev['text']), f"{rid} {p['label']}: no link back to the box before"

        clashes = question_clashes('PTE', it['title'], type_name='Re-order Paragraphs')
        assert not clashes, f"{rid}: clashes with live: {clashes}"

    print(f"  All 50 Re-order items valid! Total words: {total_words}, average: {total_words/50:.1f} words/item")
    return True


if __name__ == '__main__':
    from scripts.pte_b4_reading_b.reorders import ALL_REORDERS
    validate_reorders(ALL_REORDERS)
