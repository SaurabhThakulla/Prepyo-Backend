"""Writes migrations/000101_eps_practice_bank: 48 new EPS-TOPIK practice
questions for each of the sixteen subtasks, so every subtask has 50 with the
two starters from migration 000097.

Task types, formats and level follow HRD Korea's published EPS-TOPIK standard
(출제기준, epstopik.hrdkorea.or.kr/epstopik/abot/exam/selectStandardDesc.do).
Every question is written for Prepyo; none is copied from HRD Korea's books.

Every question gets an English version (audio_translation): what was said or
written, the question and the four choices. The "Help in English" button
shows it to premium learners only.

The build fails if two questions repeat: the same stimulus (passage, script or
picture) twice in the bank or against the starters, near-identical passages,
or a word tested twice in the same task.

    python scripts/generate_eps_bank.py
"""

from __future__ import annotations

import difflib
import hashlib
import json
import re
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))

from eps_bank.common import emoji_tile, formal_copula, place, sign  # noqa: E402
from eps_bank.dialogue_a import DIALOGUES_A  # noqa: E402
from eps_bank.dialogue_b import DIALOGUES_B  # noqa: E402
from eps_bank.generated import l_number, l_picture_question  # noqa: E402
from eps_bank.lexicon import CATEGORIES, ENTRIES, Entry, fits_with  # noqa: E402
from eps_bank.listening_reply import NEXT, RESPONSE  # noqa: E402
from eps_bank.listening_sound import SETS  # noqa: E402
from eps_bank.reading_blank import BLANKS  # noqa: E402
from eps_bank.reading_detail import DETAILS  # noqa: E402
from eps_bank.reading_grammar import GRAMMAR  # noqa: E402
from eps_bank.reading_practical import PRACTICAL  # noqa: E402
from eps_bank.reading_topic import TOPICS  # noqa: E402
from eps_bank.seed_english import SEED_ENGLISH, SEED_TITLES  # noqa: E402

MIGRATIONS = HERE.parent / "migrations"
NAME = "000101_eps_practice_bank"
SEED_SQL = MIGRATIONS / "000097_eps_topik_seed.up.sql"
VERSION = "eps-topik-2026-01"
PER_TYPE = 48
ASSET_URL = "/api/v1/questions/assets/"

# type_id: (skill, type name, Korean title, instruction, English instruction)
TYPES = {
    "eps-r-picture": ("reading", "Picture to Word or Sentence", "그림 보고 고르기",
                      "다음 그림을 보고 맞는 단어나 문장을 고르십시오.", "Look at the picture and choose the matching word or sentence."),
    "eps-r-text-to-picture": ("reading", "Description to Picture", "설명 보고 그림 고르기",
                              "다음 설명을 읽고 알맞은 그림을 고르십시오.", "Read the description and choose the matching picture."),
    "eps-r-definition": ("reading", "Word from Description", "설명에 맞는 어휘",
                         "다음 설명에 알맞은 어휘를 고르십시오.", "Choose the word that fits the description."),
    "eps-r-word-relation": ("reading", "Related Word", "관계있는 말",
                            "다음 단어와 관계있는 것은 무엇입니까?", "Which is related to this word?"),
    "eps-r-blank": ("reading", "Fill in the Blank", "빈칸 채우기",
                    "빈칸에 들어갈 가장 알맞은 것을 고르십시오.", "Choose the best answer for the blank."),
    "eps-r-grammar": ("reading", "Correct Underlined Grammar", "문법",
                      "다음 중 밑줄 친 부분이 맞는 것은 무엇입니까?", "Which underlined part is correct?"),
    "eps-r-topic": ("reading", "Passage Topic", "글의 주제",
                    "다음 글을 읽고 무엇에 대한 글인지 고르십시오.", "Read the text and choose what it is about."),
    "eps-r-detail": ("reading", "Matching Statement", "내용 일치",
                     "다음 글을 읽고 내용과 같은 것을 고르십시오.", "Read the text and choose the statement that matches it."),
    "eps-r-practical": ("reading", "Signs, Notices and Charts", "안내문", None, None),
    "eps-l-sound": ("listening", "Choose What You Heard", "소리 구별", "들은 것을 고르십시오.", "Choose what you heard."),
    "eps-l-picture": ("listening", "Listen and Choose Picture", "듣고 그림 고르기",
                      "다음을 듣고 들은 내용과 관계있는 그림을 고르십시오.", "Listen and choose the matching picture."),
    "eps-l-response": ("listening", "Choose the Right Reply", "대답 고르기",
                       "다음을 듣고 질문에 알맞은 대답을 고르십시오.", "Listen and choose the right reply."),
    "eps-l-next": ("listening", "What Comes Next", "이어지는 말", "다음을 듣고 이어지는 말을 고르십시오.", "Listen and choose what comes next."),
    "eps-l-number": ("listening", "Numbers, Dates and Prices", "숫자 듣기",
                     "다음을 듣고 들은 내용과 관계있는 그림을 고르십시오.", "Listen and choose the matching picture."),
    "eps-l-picture-question": ("listening", "Picture Question", "그림 보고 대답하기",
                               "다음을 듣고 질문에 알맞은 대답을 고르십시오.", "Look at the picture, listen and choose the right answer."),
    "eps-l-dialogue": ("listening", "Dialogue Comprehension", "대화 듣기", None, None),
}
SHORT = {t: t.removeprefix("eps-") for t in TYPES}

HANGUL = re.compile(r"[가-힣]")


# ---------------------------------------------------------------- English versions

def _clean(text: str) -> str:
    """English content with no colon a reader or the voice could take for a
    speaker label, and a closing full stop."""
    text = re.sub(r":\s+", " – ", text.strip())
    return text if text.endswith((".", "?", "!", '"')) else text + "."


def english_version(*, script: str | None = None, text: str | None = None, question: str | None = None,
                    choices: list[str] | None) -> str:
    parts = []
    if script:
        parts.append(script if re.match(r"^[A-Z][a-z]+: ", script) else f"Audio: {_clean(script)}")
    if text:
        parts.append(f"Text: {_clean(text)}")
    if question:
        parts.append(f"Question: {_clean(question)}")
    if choices:
        parts.append("Choices: " + " ".join(f"{letter}. {_clean(choice)}" for letter, choice in zip("ABCD", choices)))
    out = " ".join(parts)
    assert not HANGUL.search(out), out
    return out


# ---------------------------------------------------------------- rows

class Bank:
    def __init__(self) -> None:
        self.rows: list[dict] = []
        self.assets: dict[str, str] = {}
        self.counts: dict[str, int] = {t: 0 for t in TYPES}

    def asset(self, key: str, svg: str) -> str:
        aid = "eps_b2_" + re.sub(r"[^a-z0-9]+", "_", key.lower()).strip("_")
        if aid in self.assets:
            assert self.assets[aid] == svg, f"asset id clash: {aid}"
        self.assets[aid] = svg
        return ASSET_URL + aid

    def add(self, type_id: str, *, options: list[dict], letter: str, explanation: str, english: str, tag: str,
            difficulty: str = "easy", title: str | None = None, prompt: str | None = None, passage: str | None = None,
            script: str | None = None, image_url: str | None = None) -> None:
        skill, type_name, label, instruction, _ = TYPES[type_id]
        self.counts[type_id] += 1
        n = self.counts[type_id]
        self.rows.append({
            "id": f"eps-b2-{SHORT[type_id]}-{n:02d}", "skill": skill, "type_id": type_id, "type_name": type_name,
            "title": title or f"{label} {n + 2}", "prompt": prompt or instruction, "passage": passage, "script": script,
            "image_url": image_url, "options": options, "answer": letter, "explanation": explanation,
            "difficulty": difficulty, "tags": [f"EPS_TOPIK {skill.title()}", type_name, tag], "english": english,
        })


def text_options(texts: list[str]) -> list[dict]:
    return [{"id": letter, "text": text} for letter, text in zip("ABCD", texts)]


def hand_choice(bank: Bank, i: int, options: list[tuple[str, str]]):
    """Puts the right answer (first) at the slot for item i; returns Korean
    options, English options and the letter."""
    (ko, en), wrong = options[0], options[1:]
    ordered, letter = place((ko, en), wrong, i)
    return [o[0] for o in ordered], [o[1] for o in ordered], letter


# ---------------------------------------------------------------- word-based types

def round_robin(entries: list[Entry]) -> list[Entry]:
    """Entries in an order that cycles through the categories, so each task
    gets a spread of topics."""
    by_cat: dict[str, list[Entry]] = {}
    for e in entries:
        by_cat.setdefault(e.category, []).append(e)
    out, i = [], 0
    while any(by_cat.values()):
        for cat in list(by_cat):
            if i < len(by_cat[cat]):
                out.append(by_cat[cat][i])
        i += 1
        if all(i >= len(v) for v in by_cat.values()):
            break
    return out


def distractors(target: Entry, salt: int, *, need_emoji: bool, avoid_text: str = "") -> list[Entry]:
    pool = [e for e in ENTRIES if e.category == target.category and fits_with(target, e)
            and (e.emoji or not need_emoji) and e.ko not in avoid_text]
    assert len(pool) >= 3, (target.ko, [e.ko for e in pool])
    start = salt % len(pool)
    chosen: list[Entry] = []
    for k in range(len(pool)):
        cand = pool[(start + k * 2 + (k * 2 >= len(pool))) % len(pool)]
        if cand not in chosen and all(fits_with(cand, c) for c in chosen):
            chosen.append(cand)
        if len(chosen) == 3:
            break
    if len(chosen) < 3:
        chosen = [e for e in pool if all(fits_with(e, c) for c in chosen if c is not e)][:3]
    assert len(chosen) == 3, target.ko
    return chosen


def partition_words():
    seeds = {"eps-r-picture": {"망치", "장갑"}, "eps-r-text-to-picture": {"우산", "소화기"},
             "eps-r-definition": {"톱", "줄자"}, "eps-l-picture": {"우유"}}
    order = round_robin(ENTRIES)
    # A word a starter already describes is not described again, in any task.
    described = seeds["eps-r-text-to-picture"] | seeds["eps-r-definition"]
    ttp = [e for e in order if e.emoji and e.ko not in described][:PER_TYPE]
    rest = [e for e in order if e not in ttp and e.ko not in described]
    definition, relation = [], []
    for e in rest:
        (definition if len(definition) <= len(relation) and len(definition) < PER_TYPE else relation).append(e)
    relation = relation[:PER_TYPE]
    picture_pool = [e for e in order if e.emoji]
    fresh = [e for e in picture_pool if e not in ttp]
    picture = [e for e in fresh if e.ko not in seeds["eps-r-picture"]][:PER_TYPE]
    listen = [e for e in fresh if e not in picture and e.ko not in seeds["eps-l-picture"]]
    listen += [e for e in reversed(ttp) if e.ko not in seeds["eps-l-picture"] and e not in listen][:PER_TYPE - len(listen)]
    for name, group in [("ttp", ttp), ("definition", definition), ("relation", relation), ("picture", picture), ("listen", listen)]:
        assert len(group) == PER_TYPE, (name, len(group))
    return ttp, definition, relation, picture, listen


def a_or_an(word: str) -> str:
    return ("an " if word[0] in "aeiou" else "a ") + word


def build_word_types(bank: Bank) -> None:
    ttp, definition, relation, picture, listen = partition_words()

    for i, e in enumerate(picture):
        wrong = distractors(e, i, need_emoji=False)
        ordered, letter = place(e, wrong, i)
        bank.add("eps-r-picture", image_url=bank.asset(f"pic_{e.en}", emoji_tile(e.emoji)), tag=e.tag,
                 options=text_options([formal_copula(o.ko) + "." for o in ordered]), letter=letter,
                 english=english_version(question=TYPES["eps-r-picture"][4], choices=[o.en[0].upper() + o.en[1:] for o in ordered]),
                 explanation=f'{e.ko} means "{e.en}". ' + ", ".join(f"{w.ko} is {w.en}" for w in wrong) + ".")

    for i, e in enumerate(ttp):
        wrong = distractors(e, i + 1, need_emoji=True)
        ordered, letter = place(e, wrong, i)
        bank.add("eps-r-text-to-picture", passage=e.desc_ko, tag=e.tag, letter=letter,
                 options=[{"id": l, "text": o.ko, "imageUrl": bank.asset(f"pic_{o.en}", emoji_tile(o.emoji))} for l, o in zip("ABCD", ordered)],
                 english=english_version(text=e.desc_en, question=TYPES["eps-r-text-to-picture"][4], choices=[o.en for o in ordered]),
                 explanation=f"The description fits {e.ko} ({e.en}): {e.desc_en}")

    for i, e in enumerate(definition):
        wrong = distractors(e, i + 2, need_emoji=False)
        ordered, letter = place(e, wrong, i)
        bank.add("eps-r-definition", passage=e.desc_ko, tag=e.tag, letter=letter, options=text_options([o.ko for o in ordered]),
                 english=english_version(text=e.desc_en, question=TYPES["eps-r-definition"][4], choices=[o.en for o in ordered]),
                 explanation=f"{e.ko} is {a_or_an(e.en) if not e.en.endswith('s') else e.en}: {e.desc_en}")

    for i, e in enumerate(relation):
        wrong = distractors(e, i + 3, need_emoji=False)
        ordered, letter = place(e, wrong, i)
        bank.add("eps-r-word-relation", passage=e.ko, tag=e.tag, letter=letter, options=text_options([o.desc_ko for o in ordered]),
                 english=english_version(text=e.en[0].upper() + e.en[1:], question=TYPES["eps-r-word-relation"][4], choices=[o.desc_en for o in ordered]),
                 explanation=f"{e.ko} ({e.en}): {e.desc_en}")

    for i, e in enumerate(listen):
        wrong = distractors(e, i + 4, need_emoji=True, avoid_text=e.use_ko)
        ordered, letter = place(e, wrong, i)
        bank.add("eps-l-picture", script=e.use_ko, tag=e.tag, letter=letter,
                 options=[{"id": l, "text": o.ko, "imageUrl": bank.asset(f"pic_{o.en}", emoji_tile(o.emoji))} for l, o in zip("ABCD", ordered)],
                 english=english_version(script=e.use_en, choices=[o.en for o in ordered]),
                 explanation=f'You heard {e.use_ko.rstrip(".")} ("{e.use_en}"). {e.ko} is {a_or_an(e.en) if not e.en.endswith("s") else e.en}.')


# ---------------------------------------------------------------- hand-written types

def label_en(script_ko: str, english: str) -> str:
    return english


def build_hand_types(bank: Bank) -> None:
    for i, (heard, en, close, tag) in enumerate(SETS):
        ko_opts, en_opts, letter = hand_choice(bank, i, [(heard, en)] + close)
        bank.add("eps-l-sound", script=heard, tag=tag, letter=letter, options=text_options(ko_opts),
                 english=english_version(script=en[0].upper() + en[1:], choices=en_opts),
                 explanation=f"You heard {heard} ({en}). " + ", ".join(f"{k} is {e}" for k, e in close) + ".")

    for type_id, items in (("eps-l-response", RESPONSE), ("eps-l-next", NEXT)):
        for i, (speaker, ko, en, options, note, tag) in enumerate(items):
            ko_opts, en_opts, letter = hand_choice(bank, i, options)
            who = "Man" if speaker == "남" else "Woman"
            bank.add(type_id, script=f"{speaker}: {ko}", tag=tag, letter=letter, options=text_options(ko_opts),
                     english=english_version(script=f"{who}: {en}", choices=en_opts), explanation=note,
                     difficulty="medium" if type_id == "eps-l-next" else "easy")

    for i, (script, en, q_ko, q_en, options, note, tag, diff) in enumerate(DIALOGUES_A + DIALOGUES_B):
        ko_opts, en_opts, letter = hand_choice(bank, i, options)
        bank.add("eps-l-dialogue", script=script, prompt=q_ko, tag=tag, difficulty=diff, letter=letter,
                 options=text_options(ko_opts), english=english_version(script=en, question=q_en, choices=en_opts), explanation=note)

    for i, (text, en, options, note, tag, diff) in enumerate(BLANKS):
        ko_opts, en_opts, letter = hand_choice(bank, i, options)
        bank.add("eps-r-blank", passage=text, tag=tag, difficulty=diff, letter=letter, options=text_options(ko_opts),
                 english=english_version(text=en, question=TYPES["eps-r-blank"][4], choices=en_opts), explanation=note)

    for i, (options, note, cat, diff) in enumerate(GRAMMAR):
        meaning = options[0][1]
        options = [(ko, meaning if en == "—" else en) for ko, en in options]
        ko_opts, en_opts, letter = hand_choice(bank, i, options)
        if len(set(en_opts)) == 1:
            english = english_version(text=f"All four sentences try to say this – {meaning}",
                                      question="Which underlined part is correct Korean?", choices=None)
        else:
            english = english_version(question="Which underlined part is correct Korean? Each choice is what the sentence tries to say",
                                      choices=en_opts)
        bank.add("eps-r-grammar", title=f"문법 {i + 3}: {cat}", tag="grammar", difficulty=diff, letter=letter,
                 options=text_options(ko_opts), explanation=note, english=english)

    for type_id, items in (("eps-r-topic", TOPICS), ("eps-r-detail", DETAILS)):
        for i, (text, en, options, note, tag, diff) in enumerate(items):
            ko_opts, en_opts, letter = hand_choice(bank, i, options)
            bank.add(type_id, passage=text, tag=tag, difficulty=diff, letter=letter, options=text_options(ko_opts),
                     english=english_version(text=en, question=TYPES[type_id][4], choices=en_opts), explanation=note)

    for i, (kind, header, lines, emoji, sign_en, q_ko, q_en, options, note, tag, diff) in enumerate(PRACTICAL):
        longest = max(len(line) for line in lines)
        assert longest <= (13 if emoji else 22), (header, longest)
        ko_opts, en_opts, letter = hand_choice(bank, i, options)
        bank.add("eps-r-practical", title=f"안내문: {header}" if kind in ("notice",) else f"안전 표지: {header}",
                 prompt=q_ko, image_url=bank.asset(f"sign_{i + 1}", sign(kind, header, lines, emoji)), tag=tag,
                 difficulty=diff, letter=letter, options=text_options(ko_opts), explanation=note,
                 english=english_version(text=sign_en, question=q_en, choices=en_opts))


def build_generated_types(bank: Bank) -> None:
    for i, item in enumerate(l_number()):
        ordered, letter = place(item["correct"], item["wrong"], i)
        bank.add("eps-l-number", script=item["script"], tag=item["tag"], difficulty=item["difficulty"], letter=letter,
                 options=[{"id": l, "text": o[0], "imageUrl": bank.asset(o[2], o[3])} for l, o in zip("ABCD", ordered)],
                 english=english_version(script=item["english"], choices=[o[1] for o in ordered]), explanation=item["explanation"])

    for i, item in enumerate(l_picture_question()):
        ordered, letter = place(item["correct"], item["wrong"], i)
        key, svg = item["image"]
        bank.add("eps-l-picture-question", script=item["script"], image_url=bank.asset(key, svg), tag=item["tag"],
                 difficulty=item["difficulty"], letter=letter, options=text_options([o[0] for o in ordered]),
                 english=english_version(script=item["english_q"], choices=[o[1] for o in ordered]), explanation=item["explanation"])


# ---------------------------------------------------------------- starters, for the duplicate check

def sql_values(text: str) -> list[list]:
    """The rows of a VALUES list, as Python values (strings, None, numbers,
    lists for ARRAY[...]). Enough SQL for the seed file, nothing more."""
    rows, i, n = [], 0, len(text)

    def value(i):
        while text[i] in " \n\r\t":
            i += 1
        if text[i] == "'":
            j, out = i + 1, []
            while True:
                k = text.index("'", j)
                out.append(text[j:k])
                if k + 1 < n and text[k + 1] == "'":
                    out.append("'")
                    j = k + 2
                    continue
                i = k + 1
                break
            s = "".join(out)
            if text.startswith("::jsonb", i):
                return json.loads(s), i + 7
            return s, i
        if text.startswith("ARRAY[", i):
            i += 6
            items = []
            while text[i] != "]":
                v, i = value(i)
                items.append(v)
                while text[i] in ", \n":
                    i += 1
            i += 1
            if text.startswith("::text[]", i):
                i += 8
            return items, i
        m = re.match(r"NULL|TRUE|FALSE|-?\d+", text[i:])
        token = m.group(0)
        return {"NULL": None, "TRUE": True, "FALSE": False}.get(token, token if token.isalpha() else int(token)), i + len(token)

    while True:
        start = text.find("(", i)
        if start == -1:
            break
        i, row = start + 1, []
        while True:
            v, i = value(i)
            row.append(v)
            while text[i] in " \n\r\t":
                i += 1
            if text[i] == ",":
                i += 1
                continue
            assert text[i] == ")", text[i - 40:i + 40]
            i += 1
            break
        rows.append(row)
    return rows


def load_seeds() -> tuple[list[dict], dict[str, str]]:
    sql = SEED_SQL.read_text(encoding="utf-8")
    assets_part = sql[sql.index("INSERT INTO question_assets"):sql.index("ON CONFLICT")]
    assets = {}
    for aid, _ctype, _size, svg in sql_values(re.sub(r"convert_to\(('(?:[^']|'')*'), 'UTF8'\)", r"\1", assets_part.split("VALUES", 1)[1])):
        assets[aid] = svg
    q_part = sql[sql.index("INSERT INTO questions"):]
    q_part = q_part.split("VALUES", 1)[1].rsplit("ON CONFLICT", 1)[0]
    seeds = []
    for r in sql_values(q_part):
        seeds.append({"id": r[0], "type_id": r[5], "title": r[7], "prompt": r[8], "passage": r[9], "script": r[10],
                      "image_url": r[11], "options": r[13], "answer": r[14][0], "seed": True})
    return seeds, assets


# ---------------------------------------------------------------- checks

def norm(text: str | None) -> str:
    return re.sub(r"[\s.,!?·:\"'()\[\]~※-]+", "", text or "")


def check(bank: Bank, seeds: list[dict], seed_assets: dict[str, str]) -> None:
    assets = {**seed_assets, **bank.assets}

    def image_key(url):
        return hashlib.sha256(assets[url.removeprefix(ASSET_URL)].encode()).hexdigest()[:16] if url else ""

    problems = []
    every = seeds + bank.rows
    ids = [q["id"] for q in every]
    if len(ids) != len(set(ids)):
        problems.append("repeated question id")
    for t in TYPES:
        count = sum(1 for q in every if q["type_id"] == t)
        if count != 50:
            problems.append(f"{t} has {count} questions, not 50")

    seen: dict[tuple, str] = {}
    for q in every:
        options = q["options"]
        texts = [o["text"] for o in options]
        if len(options) != 4 or len(set(texts)) != 4:
            problems.append(f"{q['id']}: options must be four different choices")
        if [o["id"] for o in options] != list("ABCD") or q["answer"] not in "ABCD":
            problems.append(f"{q['id']}: options must be A-D with one answer")
        if any(o.get("imageUrl") for o in options):
            pics = [image_key(o["imageUrl"]) for o in options]
            if len(set(pics)) != 4:
                problems.append(f"{q['id']}: two picture options are the same picture")
        stimulus = norm(q["passage"]) + norm(q["script"]) + image_key(q["image_url"])
        # A task with nothing to read, hear or see but its choices (grammar) is its choices.
        choices = "" if stimulus else "|".join(sorted(norm(t) for t in texts))
        key = (q["type_id"], norm(q["prompt"]), norm(q["passage"]), norm(q["script"]), image_key(q["image_url"]), choices)
        if key in seen:
            problems.append(f"{q['id']} repeats {seen[key]}")
        seen.setdefault(key, q["id"])
        # The same stimulus and answer must not come back as another task either.
        answer = next(o["text"] for o in options if o["id"] == q["answer"])
        cross = ("any", norm(q["passage"]) + "|" + norm(q["script"]) + "|" + image_key(q["image_url"]), norm(answer))
        if cross in seen and cross[1] != "||":
            problems.append(f"{q['id']} asks the same thing as {seen[cross]}")
        seen.setdefault(cross, q["id"])

    # Near-identical passages or scripts, in any task. Pictures aside, and the
    # price, date and time exchanges, which share a frame by design and differ
    # in the number (the exact check above still covers them).
    texts = [(q["id"], norm(q["passage"]) + norm(q["script"])) for q in every
             if not q["image_url"] and q["type_id"] != "eps-l-number"]
    texts = [(i, t) for i, t in texts if len(t) >= 12]
    for a in range(len(texts)):
        for b in range(a + 1, len(texts)):
            ratio = difflib.SequenceMatcher(None, texts[a][1], texts[b][1]).ratio()
            if ratio >= 0.8:
                problems.append(f"{texts[a][0]} and {texts[b][0]} are near-identical ({ratio:.2f})")

    # The same right answer behind a similar passage or script, in any task.
    keyed = [(q["id"], norm(next(o["text"] for o in q["options"] if o["id"] == q["answer"])).removesuffix("입니다"),
              norm(q["passage"]) + norm(q["script"])) for q in every if not q["image_url"]]
    for a in range(len(keyed)):
        for b in range(a + 1, len(keyed)):
            (ida, ansa, ta), (idb, ansb, tb) = keyed[a], keyed[b]
            if ansa == ansb and ta and tb:
                ratio = difflib.SequenceMatcher(None, ta, tb).ratio()
                if ratio >= 0.5:
                    problems.append(f"{ida} and {idb} share an answer and a similar text ({ratio:.2f})")

    for q in bank.rows:
        if HANGUL.search(q["english"]) or not q["english"].startswith(("Audio:", "Text:", "Question:", "Man:", "Woman:")):
            problems.append(f"{q['id']}: English version is malformed")
    if problems:
        print("\n".join(problems))
        raise SystemExit(f"{len(problems)} problem(s); nothing written.")


# ---------------------------------------------------------------- SQL

def lit(value) -> str:
    if value is None:
        return "NULL"
    return "'" + str(value).replace("'", "''") + "'"


def write(bank: Bank) -> None:
    out = [
        "-- EPS-TOPIK practice bank: 48 new questions for each of the sixteen subtasks,",
        "-- so each has 50 with the two starters from 000097 (800 in all).",
        "--",
        "-- Task types, format and level follow HRD Korea's published EPS-TOPIK",
        "-- standard (출제기준, epstopik.hrdkorea.or.kr/epstopik/abot/exam/selectStandardDesc.do).",
        "-- The questions are written for Prepyo; none is copied from HRD Korea's books.",
        "-- Generated by scripts/generate_eps_bank.py from scripts/eps_bank/: edit the",
        "-- data there and regenerate rather than editing this file.",
        "--",
        "-- Every question carries an English version in audio_translation (what was",
        "-- said or written, the question and the choices), shown by the premium-only",
        "-- Help in English. Spoken recordings of both are added later by",
        "-- prepyo-voicegen, which fills audio_url and translation_audio_url.",
        "",
        "INSERT INTO question_assets (id, content_type, byte_size, data) VALUES",
    ]
    asset_lines = [f"    ({lit(aid)}, 'image/svg+xml', {len(svg.encode())}, convert_to({lit(svg)}, 'UTF8'))"
                   for aid, svg in sorted(bank.assets.items())]
    out.append(",\n".join(asset_lines))
    out.append("ON CONFLICT (id) DO NOTHING;")
    out.append("")
    out.append("INSERT INTO questions")
    out.append("    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,")
    out.append("     context_passage, audio_transcript, image_url, time_limit_seconds, options, correct_answers,")
    out.append("     explanation, difficulty, tags, points, is_published, audio_translation)")
    out.append("VALUES")
    rows = []
    for q in bank.rows:
        options = json.dumps(q["options"], ensure_ascii=False)
        tags = "ARRAY[" + ", ".join(lit(t) for t in q["tags"]) + "]::text[]"
        rows.append(
            f"    ({lit(q['id'])}, '{VERSION}', 'EPS_TOPIK', ARRAY['EPS_TOPIK'], {lit(q['skill'])}, {lit(q['type_id'])}, "
            f"{lit(q['type_name'])}, {lit(q['title'])}, {lit(q['prompt'])}, {lit(q['passage'])}, {lit(q['script'])}, "
            f"{lit(q['image_url'])}, 75, {lit(options)}::jsonb, '[\"{q['answer']}\"]'::jsonb, {lit(q['explanation'])}, "
            f"{lit(q['difficulty'])}, {tags}, 1, TRUE, {lit(q['english'])})")
    out.append(",\n".join(rows))
    out.append("ON CONFLICT (id) DO NOTHING;")
    out.append("")
    out.append("-- English versions for the starter reading questions, which had none.")
    out.append("UPDATE questions AS q SET audio_translation = t.english")
    out.append("FROM (VALUES")
    seed_rows = []
    for qid, (text, question, choices) in SEED_ENGLISH.items():
        seed_rows.append(f"    ({lit(qid)}, {lit(english_version(text=text, question=question, choices=choices))})")
    out.append(",\n".join(seed_rows))
    out.append(") AS t(id, english)")
    out.append("WHERE q.id = t.id AND COALESCE(q.audio_translation, '') = '';")
    out.append("")
    out.append("-- Starter titles that named the answer. Only a title nobody has changed is replaced.")
    out.append("UPDATE questions AS q SET title = t.new_title")
    out.append("FROM (VALUES")
    out.append(",\n".join(f"    ({lit(qid)}, {lit(old)}, {lit(new)})" for qid, (old, new) in SEED_TITLES.items()))
    out.append(") AS t(id, old_title, new_title)")
    out.append("WHERE q.id = t.id AND q.title = t.old_title;")
    out.append("")
    up = MIGRATIONS / f"{NAME}.up.sql"
    down = MIGRATIONS / f"{NAME}.down.sql"
    up.write_bytes("\r\n".join("\n".join(out).split("\n")).encode("utf-8"))
    down.write_bytes("-- Content migration: questions may already have attempts, so nothing is removed.\r\nSELECT 1;\r\n".encode("utf-8"))
    print(f"Wrote {up.name}: {len(bank.rows)} questions, {len(bank.assets)} pictures.")


def main() -> None:
    bank = Bank()
    build_word_types(bank)
    build_hand_types(bank)
    build_generated_types(bank)
    seeds, seed_assets = load_seeds()
    for s in seeds:
        old_new = SEED_TITLES.get(s["id"])
        if old_new:
            s["title"] = old_new[1]
    check(bank, seeds, seed_assets)
    if "--check" not in sys.argv:
        write(bank)
    by_type = {t: c for t, c in bank.counts.items()}
    print("per task:", ", ".join(f"{SHORT[t]} {c}+2" for t, c in by_type.items()))


if __name__ == "__main__":
    main()
