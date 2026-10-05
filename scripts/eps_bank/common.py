"""Shared pieces for the EPS-TOPIK practice bank: Korean number reading, the
pictures (emoji tiles, number tiles, clocks, scenes, signs) and option layout.

Everything here is deterministic, so the generated migration is the same on
every run.
"""

from __future__ import annotations

import html

FONT_KR = "Noto Sans KR,Malgun Gothic,Apple SD Gothic Neo,sans-serif"
FONT_EMOJI = "Apple Color Emoji,Segoe UI Emoji,Noto Color Emoji,sans-serif"

# ---------------------------------------------------------------- Korean numbers

SINO_DIGITS = ["", "일", "이", "삼", "사", "오", "육", "칠", "팔", "구"]


def sino(n: int) -> str:
    """A number read the Sino-Korean way, as used for money, minutes, dates
    and phone numbers: 3500 → 삼천오백, 15000 → 만 오천, 10 → 십."""
    if n == 0:
        return "영"
    if n < 0 or n >= 100_000_000:
        raise ValueError(n)
    parts = []
    man, rest = divmod(n, 10_000)
    if man:
        parts.append(("" if man == 1 else _under_10000(man)) + "만")
    if rest:
        parts.append(_under_10000(rest))
    return " ".join(parts)


def _under_10000(n: int) -> str:
    out = ""
    for value, unit in ((1000, "천"), (100, "백"), (10, "십")):
        digit, n = divmod(n, value)
        if digit:
            out += ("" if digit == 1 else SINO_DIGITS[digit]) + unit
    return out + SINO_DIGITS[n]


NATIVE = {1: "하나", 2: "둘", 3: "셋", 4: "넷", 5: "다섯", 6: "여섯", 7: "일곱", 8: "여덟", 9: "아홉", 10: "열",
          11: "열하나", 12: "열둘"}
# Before a counter (개, 명, 시 …) the first four, and twenty, shorten.
NATIVE_COUNTER = {1: "한", 2: "두", 3: "세", 4: "네", 5: "다섯", 6: "여섯", 7: "일곱", 8: "여덟", 9: "아홉", 10: "열",
                  11: "열한", 12: "열두"}


def native_count(n: int, counter: str) -> str:
    """Native number with a counter: 세 개, 두 명, 열한 시."""
    return f"{NATIVE_COUNTER[n]} {counter}"


MONTHS = {1: "일월", 2: "이월", 3: "삼월", 4: "사월", 5: "오월", 6: "유월", 7: "칠월", 8: "팔월", 9: "구월",
          10: "시월", 11: "십일월", 12: "십이월"}


def date_ko(month: int, day: int) -> str:
    return f"{MONTHS[month]} {sino(day)} 일"


def time_ko(hour: int, minute: int = 0) -> str:
    """Hours in native numbers, minutes in Sino-Korean: 세 시 삼십 분."""
    out = native_count(hour, "시")
    if minute:
        out += f" {sino(minute)} 분"
    return out


def won(n: int) -> str:
    return f"{sino(n)} 원"


def has_batchim(word: str) -> bool:
    """Whether the last syllable ends in a consonant (받침)."""
    last = word.strip()[-1]
    code = ord(last) - 0xAC00
    return 0 <= code < 11172 and code % 28 != 0


def topic(word: str) -> str:
    return word + ("은" if has_batchim(word) else "는")


def subject(word: str) -> str:
    return word + ("이" if has_batchim(word) else "가")


def obj(word: str) -> str:
    return word + ("을" if has_batchim(word) else "를")


def copula(word: str) -> str:
    """word + 이에요/예요."""
    return word + ("이에요" if has_batchim(word) else "예요")


def formal_copula(word: str) -> str:
    return word + "입니다"


# ---------------------------------------------------------------- pictures

def _svg(view: str, body: str) -> str:
    return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{view}">{body}</svg>'


def esc(text: str) -> str:
    return html.escape(text, quote=True)


def emoji_tile(emoji: str) -> str:
    """A rounded tile with one emoji, as the seed questions use."""
    return _svg("0 0 240 240",
                '<rect x="4" y="4" width="232" height="232" rx="28" fill="#F8FAFC" stroke="#CBD5E1" stroke-width="4"/>'
                f'<text x="120" y="158" font-size="120" text-anchor="middle" font-family="{FONT_EMOJI}">{emoji}</text>')


def text_tile(text: str, size: int = 40, fill: str = "#FFFBEB", stroke: str = "#F59E0B") -> str:
    """A tile with a short label, for prices, dates and times."""
    return _svg("0 0 240 240",
                f'<rect x="4" y="4" width="232" height="232" rx="28" fill="{fill}" stroke="{stroke}" stroke-width="6"/>'
                f'<text x="120" y="{120 + size // 3}" font-size="{size}" font-weight="700" text-anchor="middle" '
                f'fill="#0F172A" font-family="{FONT_KR}">{esc(text)}</text>')


def calendar_tile(top: str, big: str) -> str:
    return _svg("0 0 240 240",
                '<rect x="20" y="30" width="200" height="190" rx="18" fill="#FFFFFF" stroke="#DC2626" stroke-width="6"/>'
                '<rect x="20" y="30" width="200" height="50" rx="18" fill="#DC2626"/>'
                '<rect x="20" y="62" width="200" height="18" fill="#DC2626"/>'
                f'<text x="120" y="66" font-size="26" font-weight="700" text-anchor="middle" fill="#FFFFFF" font-family="{FONT_KR}">{esc(top)}</text>'
                f'<text x="120" y="176" font-size="{64 if len(big) <= 3 else 50}" font-weight="700" text-anchor="middle" fill="#0F172A" font-family="{FONT_KR}">{esc(big)}</text>')


def clock(hour: int, minute: int, size: int = 240) -> str:
    """An analogue clock face showing hour:minute."""
    import math
    c = size / 2
    r = size / 2 - 14
    ticks = []
    for i in range(12):
        a = math.radians(i * 30)
        x1, y1 = c + math.sin(a) * (r - 4), c - math.cos(a) * (r - 4)
        x2, y2 = c + math.sin(a) * (r - 20), c - math.cos(a) * (r - 20)
        ticks.append(f'<line x1="{x1:.1f}" y1="{y1:.1f}" x2="{x2:.1f}" y2="{y2:.1f}" stroke="#0F172A" stroke-width="{6 if i % 3 == 0 else 3}" stroke-linecap="round"/>')
    ha = math.radians((hour % 12) * 30 + minute * 0.5)
    ma = math.radians(minute * 6)
    hx, hy = c + math.sin(ha) * r * 0.5, c - math.cos(ha) * r * 0.5
    mx, my = c + math.sin(ma) * r * 0.78, c - math.cos(ma) * r * 0.78
    return _svg(f"0 0 {size} {size}",
                f'<circle cx="{c}" cy="{c}" r="{r}" fill="#FFFFFF" stroke="#0F172A" stroke-width="8"/>' + "".join(ticks)
                + f'<line x1="{c}" y1="{c}" x2="{hx:.1f}" y2="{hy:.1f}" stroke="#0F172A" stroke-width="10" stroke-linecap="round"/>'
                + f'<line x1="{c}" y1="{c}" x2="{mx:.1f}" y2="{my:.1f}" stroke="#2563EB" stroke-width="6" stroke-linecap="round"/>'
                + f'<circle cx="{c}" cy="{c}" r="8" fill="#0F172A"/>')


# Where an object can be in the room scene: (x, y) of the emoji's baseline centre.
SCENE_SPOTS = {
    "책상 위": (170, 158),
    "책상 아래": (170, 286),
    "의자 위": (372, 214),
    "의자 옆": (448, 300),
    "상자 안": (80, 244),
    "상자 옆": (24, 300),
}


def room_scene(emoji: str, spot: str) -> str:
    """A desk, a chair and an open box, with one object placed somewhere."""
    x, y = SCENE_SPOTS[spot]
    box_front = ('<rect x="40" y="250" width="80" height="56" rx="6" fill="#D97706" stroke="#92400E" stroke-width="4"/>'
                 '<polygon points="40,250 54,232 106,232 120,250" fill="#B45309"/>')
    obj_svg = f'<text x="{x}" y="{y}" font-size="58" text-anchor="middle" font-family="{FONT_EMOJI}">{emoji}</text>'
    # An object in the box sits behind the box's front.
    layers = (obj_svg + box_front) if spot == "상자 안" else (box_front + obj_svg)
    return _svg("0 0 480 320",
                '<rect x="6" y="6" width="468" height="308" rx="22" fill="#F8FAFC" stroke="#CBD5E1" stroke-width="4"/>'
                '<rect x="110" y="170" width="160" height="16" rx="4" fill="#92400E"/>'
                '<rect x="122" y="186" width="12" height="118" fill="#92400E"/><rect x="246" y="186" width="12" height="118" fill="#92400E"/>'
                f'<text x="372" y="300" font-size="110" text-anchor="middle" font-family="{FONT_EMOJI}">🪑</text>'
                + layers)


def count_scene(emoji: str, n: int) -> str:
    """n copies of an emoji in tidy rows."""
    per_row = 5 if n > 4 else n
    rows = (n + per_row - 1) // per_row
    items = []
    for i in range(n):
        r, col = divmod(i, per_row)
        in_row = min(per_row, n - r * per_row)
        x = 240 + (col - (in_row - 1) / 2) * 84
        y = 160 + (r - (rows - 1) / 2) * 92 + 26
        items.append(f'<text x="{x:.0f}" y="{y:.0f}" font-size="70" text-anchor="middle" font-family="{FONT_EMOJI}">{emoji}</text>')
    return _svg("0 0 480 320", '<rect x="6" y="6" width="468" height="308" rx="22" fill="#F8FAFC" stroke="#CBD5E1" stroke-width="4"/>' + "".join(items))


SIGN_STYLE = {
    # kind: (frame colour, header colour, header text colour)
    "notice": ("#2563EB", "#2563EB", "#FFFFFF"),
    "warning": ("#CA8A04", "#FACC15", "#0F172A"),
    "prohibition": ("#DC2626", "#DC2626", "#FFFFFF"),
    "mandatory": ("#1D4ED8", "#1D4ED8", "#FFFFFF"),
    "info": ("#16A34A", "#16A34A", "#FFFFFF"),
}


def sign(kind: str, header: str, lines: list[str], emoji: str | None = None) -> str:
    """A notice or safety sign: a coloured header and up to five lines of text,
    with an optional symbol."""
    frame, head, head_text = SIGN_STYLE[kind]
    body = [f'<rect x="6" y="6" width="468" height="308" rx="22" fill="#FFFFFF" stroke="{frame}" stroke-width="8"/>',
            f'<rect x="6" y="6" width="468" height="74" rx="22" fill="{head}"/>',
            f'<rect x="6" y="58" width="468" height="22" fill="{head}"/>',
            f'<text x="240" y="58" font-size="{38 if len(header) <= 12 else 30}" font-weight="700" text-anchor="middle" '
            f'fill="{head_text}" font-family="{FONT_KR}">{esc(header)}</text>']
    left = 40
    if emoji:
        body.append(f'<text x="400" y="230" font-size="96" text-anchor="middle" font-family="{FONT_EMOJI}">{emoji}</text>')
    size = 26 if max((len(l) for l in lines), default=0) <= 18 else 22
    for i, line in enumerate(lines[:5]):
        colour = "#DC2626" if line.startswith("※") else "#0F172A"
        body.append(f'<text x="{left}" y="{124 + i * 42}" font-size="{size}" fill="{colour}" font-family="{FONT_KR}">{esc(line)}</text>')
    return _svg("0 0 480 320", "".join(body))


# ---------------------------------------------------------------- options

LETTERS = "ABCD"


def place(correct, wrong: list, index: int):
    """Four options with the right one at a position that cycles A, B, C, D
    through the bank. Returns (options in order, letter of the right one)."""
    assert len(wrong) == 3, wrong
    slot = index % 4
    ordered = list(wrong)
    ordered.insert(slot, correct)
    return ordered, LETTERS[slot]
