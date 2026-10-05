"""The question types built from data rather than written one by one:

- Numbers, Dates and Prices (eps-l-number): a short exchange with a price, a
  time or a date; the options are pictures of numbers that sound alike.
- Picture Question (eps-l-picture-question): a clock, a room with something in
  it, or a group of things to count, and a spoken question about it.

Each builder returns plain dicts the generator turns into rows.
"""

from __future__ import annotations

from .common import MONTHS, copula, calendar_tile, clock, count_scene, has_batchim, native_count, room_scene, sino, text_tile, time_ko, topic, subject, won

EN_HOURS = {1: "one", 2: "two", 3: "three", 4: "four", 5: "five", 6: "six", 7: "seven", 8: "eight", 9: "nine",
            10: "ten", 11: "eleven", 12: "twelve"}
EN_MONTHS = ["", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October",
             "November", "December"]


def en_time(h: int, m: int) -> str:
    if m == 0:
        return f"{EN_HOURS[h]} o'clock"
    if m == 30:
        return f"half past {EN_HOURS[h]}"
    return f"{h}:{m:02d}"


def ordinal(n: int) -> str:
    suffix = "th" if 11 <= n % 100 <= 13 else {1: "st", 2: "nd", 3: "rd"}.get(n % 10, "th")
    return f"{n}{suffix}"


# ---------------------------------------------------------------- prices

# (item, English, price, three prices that sound or look close)
PRICES = [
    ("우산", "this umbrella", 8500, [5800, 8050, 85000]),
    ("라면", "these noodles", 1200, [2100, 1020, 12000]),
    ("사과", "these apples", 6000, [600, 60000, 9000]),
    ("수박", "this watermelon", 15000, [50000, 1500, 25000]),
    ("운동화", "these trainers", 47000, [74000, 4700, 41000]),
    ("셔츠", "this shirt", 23000, [32000, 2300, 13000]),
    ("모자", "this cap", 9900, [9090, 1900, 99000]),
    ("커피", "this coffee", 4500, [5400, 4050, 45000]),
    ("빵", "this bread", 2800, [8200, 2080, 3800]),
    ("충전기", "this charger", 16000, [60000, 1600, 26000]),
    ("안전화", "these safety shoes", 38000, [83000, 3800, 33000]),
    ("가방", "this bag", 72000, [27000, 7200, 70000]),
    ("바지", "these trousers", 19000, [90000, 1900, 29000]),
    ("우유", "this milk", 2600, [6200, 2060, 3600]),
    ("계란", "these eggs", 7400, [4700, 7040, 74000]),
    ("볼펜", "this pen", 1500, [5100, 1050, 15000]),
]

# (what, English, month, day, three confusable dates as (month, day))
DATES = [
    ("시험", "the exam", 6, 10, [(6, 4), (10, 6), (6, 14)]),
    ("축제", "the festival", 10, 25, [(10, 5), (10, 20), (2, 25)]),
    ("회식", "the staff dinner", 3, 17, [(3, 7), (3, 27), (4, 17)]),
    ("휴가", "the holiday", 8, 1, [(8, 11), (7, 1), (8, 21)]),
    ("건강 검진", "the health check", 11, 12, [(11, 2), (12, 11), (11, 22)]),
    ("생일", "your birthday", 9, 30, [(9, 13), (9, 3), (3, 30)]),
    ("이사", "the move", 2, 14, [(2, 4), (4, 14), (2, 24)]),
    ("교육", "the training", 4, 6, [(4, 16), (6, 4), (4, 26)]),
    ("계약 마지막 날", "the last day of the contract", 12, 31, [(12, 13), (12, 3), (12, 21)]),
    ("체육 대회", "the sports day", 5, 18, [(5, 8), (5, 28), (8, 18)]),
    ("비자 연장", "the visa renewal", 7, 9, [(7, 19), (9, 7), (7, 29)]),
    ("한국어 시험", "the Korean test", 1, 26, [(1, 6), (1, 16), (11, 26)]),
    ("출국", "the departure", 6, 23, [(6, 3), (6, 13), (7, 23)]),
    ("결혼식", "the wedding", 10, 3, [(10, 13), (3, 10), (10, 23)]),
    ("면접", "the interview", 3, 2, [(3, 12), (2, 3), (3, 22)]),
    ("검사", "the inspection", 9, 15, [(9, 5), (5, 9), (9, 25)]),
]

# (asked by, question, English question, answer frame, English frame, hour, minute, three confusable times)
TIMES = [
    ("남", "회의가 몇 시에 시작해요?", "What time does the meeting start?", "에 시작해요.", "It starts at", 3, 30, [(4, 30), (3, 0), (2, 30)]),
    ("여", "출근 버스가 몇 시에 와요?", "What time does the work bus come?", "에 와요.", "It comes at", 7, 20, [(7, 2), (7, 40), (8, 20)]),
    ("남", "점심시간이 몇 시예요?", "What time is lunch?", "예요.", "It's at", 12, 0, [(11, 0), (12, 30), (2, 0)]),
    ("여", "기차가 몇 시에 출발해요?", "What time does the train leave?", "에 출발해요.", "It leaves at", 9, 15, [(9, 50), (10, 15), (9, 5)]),
    ("남", "영화가 몇 시에 끝나요?", "What time does the film finish?", "에 끝나요.", "It finishes at", 10, 40, [(10, 14), (4, 10), (11, 40)]),
    ("여", "병원 예약이 몇 시예요?", "What time is the hospital appointment?", "예요.", "It's at", 2, 30, [(2, 13), (3, 30), (12, 30)]),
    ("남", "작업이 몇 시에 끝나요?", "What time does the work finish?", "에 끝나요.", "It finishes at", 5, 50, [(5, 15), (6, 50), (5, 5)]),
    ("여", "은행이 몇 시에 문을 닫아요?", "What time does the bank close?", "에 닫아요.", "It closes at", 4, 0, [(10, 0), (3, 0), (4, 30)]),
    ("남", "교육이 몇 시에 있어요?", "What time is the training?", "에 있어요.", "It's at", 1, 0, [(11, 0), (7, 0), (1, 30)]),
    ("여", "비행기가 몇 시에 도착해요?", "What time does the plane arrive?", "에 도착해요.", "It arrives at", 8, 45, [(8, 15), (9, 45), (4, 45)]),
    ("남", "식당이 몇 시에 문을 열어요?", "What time does the restaurant open?", "에 열어요.", "It opens at", 11, 30, [(10, 30), (11, 13), (1, 30)]),
    ("여", "친구를 몇 시에 만나요?", "What time are you meeting your friend?", "에 만나요.", "We're meeting at", 6, 10, [(6, 50), (10, 6), (7, 10)]),
    ("남", "통근 버스가 몇 시에 출발해요?", "What time does the shuttle bus leave?", "에 출발해요.", "It leaves at", 8, 5, [(8, 50), (5, 8), (9, 5)]),
    ("여", "마트가 몇 시에 문을 닫아요?", "What time does the supermarket close?", "에 닫아요.", "It closes at", 10, 0, [(12, 0), (10, 30), (9, 0)]),
    ("남", "야근이 몇 시까지예요?", "Until what time is the overtime?", "까지예요.", "It's until", 9, 0, [(5, 0), (9, 30), (10, 0)]),
    ("여", "수업이 몇 시에 시작해요?", "What time does the class start?", "에 시작해요.", "It starts at", 7, 30, [(8, 30), (7, 3), (1, 30)]),
]


def _price_label(n: int) -> str:
    return f"{n:,}원"


def l_number() -> list[dict]:
    items = []
    for item, item_en, price, wrong in PRICES:
        asker, answerer = ("여", "남") if len(items) % 2 == 0 else ("남", "여")
        asker_en, answerer_en = ("Woman", "Man") if asker == "여" else ("Man", "Woman")
        said = won(price)
        verb, pronoun = ("are", "They're") if item_en.startswith("these") else ("is", "It's")
        items.append({
            "title_hint": "가격", "tag": "numbers-time", "difficulty": "medium",
            "script": f"{asker}: 이 {item} 얼마예요? {answerer}: {said}이에요.",
            "english": f"{asker_en}: How much {verb} {item_en}? {answerer_en}: {pronoun} {price:,} won.",
            "correct": (_price_label(price), f"{price:,} won", f"price_{price}", text_tile(_price_label(price))),
            "wrong": [(_price_label(w), f"{w:,} won", f"price_{w}", text_tile(_price_label(w))) for w in wrong],
            "explanation": f"{said} is {price:,} won.",
        })
    for what, what_en, month, day, wrong in DATES:
        asker, answerer = ("남", "여") if len(items) % 2 == 0 else ("여", "남")
        asker_en, answerer_en = ("Man", "Woman") if asker == "남" else ("Woman", "Man")
        said = f"{MONTHS[month]} {sino(day)} 일"
        items.append({
            "title_hint": "날짜", "tag": "numbers-time", "difficulty": "medium",
            "script": f"{asker}: {subject(what)} 언제예요? {answerer}: {said}이에요.",
            "english": f"{asker_en}: When is {what_en}? {answerer_en}: It's {EN_MONTHS[month]} the {ordinal(day)}.",
            "correct": (f"{month}월 {day}일", f"{EN_MONTHS[month]} {day}", f"date_{month}_{day}", calendar_tile(f"{month}월", f"{day}일")),
            "wrong": [(f"{m}월 {d}일", f"{EN_MONTHS[m]} {d}", f"date_{m}_{d}", calendar_tile(f"{m}월", f"{d}일")) for m, d in wrong],
            "explanation": f"{said} is {EN_MONTHS[month]} {day}. Months and days use Sino-Korean numbers; June is 유월 and October is 시월.",
        })
    for asker, q, q_en, frame, frame_en, h, m, wrong in TIMES:
        answerer = "여" if asker == "남" else "남"
        asker_en, answerer_en = ("Man", "Woman") if asker == "남" else ("Woman", "Man")
        said = time_ko(h, m)
        items.append({
            "title_hint": "시간", "tag": "numbers-time", "difficulty": "medium",
            "script": f"{asker}: {q} {answerer}: {copula(said) + '.' if frame == '예요.' else said + frame}",
            "english": f"{asker_en}: {q_en} {answerer_en}: {frame_en} {en_time(h, m)}.",
            "correct": (f"{h}:{m:02d}", f"{h}:{m:02d}", f"clock_{h}_{m}", clock(h, m)),
            "wrong": [(f"{wh}:{wm:02d}", f"{wh}:{wm:02d}", f"clock_{wh}_{wm}", clock(wh, wm)) for wh, wm in wrong],
            "explanation": f"{said} is {h}:{m:02d}. Hours use native numbers (한, 두, 세 …) and minutes Sino-Korean ones (십, 이십 …).",
        })
    return items


# ---------------------------------------------------------------- picture questions

CLOCK_QUESTIONS = [
    ("남", "지금 몇 시입니까?", "What time is it now?", 4, 0),
    ("여", "지금 몇 시예요?", "What time is it?", 9, 30),
    ("남", "회의는 몇 시에 시작합니까?", "What time does the meeting start?", 10, 0),
    ("여", "몇 시에 퇴근합니까?", "What time do you finish work?", 6, 0),
    ("남", "몇 시에 점심을 먹습니까?", "What time do you have lunch?", 12, 30),
    ("여", "버스는 몇 시에 옵니까?", "What time does the bus come?", 7, 10),
    ("남", "몇 시에 일어납니까?", "What time do you get up?", 6, 30),
    ("여", "수업은 몇 시에 끝납니까?", "What time does the class finish?", 8, 50),
    ("남", "가게는 몇 시에 문을 엽니까?", "What time does the shop open?", 11, 0),
    ("여", "몇 시에 잡니까?", "What time do you go to bed?", 10, 30),
    ("남", "기차는 몇 시에 떠납니까?", "What time does the train leave?", 2, 15),
    ("여", "약속이 몇 시입니까?", "What time is the appointment?", 5, 40),
    ("남", "작업은 몇 시에 시작합니까?", "What time does work start?", 8, 0),
    ("여", "영화는 몇 시에 시작합니까?", "What time does the film start?", 1, 20),
    ("남", "몇 시에 출근합니까?", "What time do you go to work?", 7, 45),
    ("여", "저녁은 몇 시에 먹습니까?", "What time do you have dinner?", 6, 20),
]

# (object, English, emoji)
SCENE_OBJECTS = [
    ("고양이", "cat", "🐈"), ("가방", "bag", "👜"), ("사과", "apple", "🍎"), ("책", "book", "📕"),
    ("휴대폰", "mobile phone", "📱"), ("모자", "cap", "🧢"), ("열쇠", "key", "🔑"), ("안경", "glasses", "👓"),
    ("시계", "watch", "⌚"), ("우산", "umbrella", "☂️"), ("컵", "cup", "☕"), ("신발", "shoes", "👟"),
    ("장갑", "gloves", "🧤"), ("공책", "notebook", "📓"), ("공", "ball", "⚽"), ("강아지", "puppy", "🐕"),
]
SPOTS_EN = {"책상 위": "on the desk", "책상 아래": "under the desk", "의자 위": "on the chair", "의자 옆": "next to the chair",
            "상자 안": "in the box", "상자 옆": "next to the box"}
SPOT_ORDER = list(SPOTS_EN)

# (thing, English plural, emoji, counter, count)
COUNT_THINGS = [
    ("사과", "apples", "🍎", "개", 3), ("바나나", "bananas", "🍌", "개", 5), ("계란", "eggs", "🥚", "개", 6),
    ("상자", "boxes", "📦", "개", 4), ("고양이", "cats", "🐈", "마리", 2), ("닭", "chickens", "🐔", "마리", 7),
    ("물고기", "fish", "🐠", "마리", 8), ("돼지", "pigs", "🐖", "마리", 3), ("자동차", "cars", "🚗", "대", 4),
    ("자전거", "bicycles", "🚲", "대", 2), ("트럭", "lorries", "🚚", "대", 5), ("책", "books", "📕", "권", 6),
    ("공책", "notebooks", "📓", "권", 9), ("커피", "cups of coffee", "☕", "잔", 3), ("우유", "glasses of milk", "🥛", "잔", 2),
    ("오리", "ducks", "🦆", "마리", 4),
]
COUNTER_EN = {"개": "(counter for things)", "마리": "(counter for animals)", "대": "(counter for vehicles)",
              "권": "(counter for books)", "잔": "(counter for cups)"}


def _near_times(h: int, m: int) -> list[tuple[int, int]]:
    """Three readings a listener could mistake for h:m."""
    other_m = (m + 30) % 60
    return [((h % 12) + 1, m), (h, other_m), ((h - 2) % 12 or 12, m)]


def _near_counts(n: int) -> list[int]:
    options = [n - 1, n + 1, n + 2] if n > 2 else [n + 1, n + 2, n + 3]
    return options if n > 1 else [2, 3, 4]


def l_picture_question() -> list[dict]:
    items = []
    for asker, q, q_en, h, m in CLOCK_QUESTIONS:
        asker_en = "Man" if asker == "남" else "Woman"
        right = f"{time_ko(h, m)}입니다."
        items.append({
            "title_hint": "시계", "tag": "numbers-time", "difficulty": "easy",
            "script": f"{asker}: {q}", "english_q": f"{asker_en}: {q_en}",
            "image": (f"clock_{h}_{m}", clock(h, m)),
            "correct": (right, f"It's {en_time(h, m)}."),
            "wrong": [(f"{time_ko(wh, wm)}입니다.", f"It's {en_time(wh, wm)}.") for wh, wm in _near_times(h, m)],
            "explanation": f"The clock shows {h}:{m:02d}: {time_ko(h, m)}.",
        })
    for i, (thing, thing_en, emoji) in enumerate(SCENE_OBJECTS):
        spot = SPOT_ORDER[i % len(SPOT_ORDER)]
        others = [s for s in SPOT_ORDER if s != spot]
        wrong = [others[(i + k) % len(others)] for k in (0, 2, 4)]
        asker = "남" if i % 2 == 0 else "여"
        asker_en = "Man" if asker == "남" else "Woman"
        items.append({
            "title_hint": "위치", "tag": "basic-life", "difficulty": "easy",
            "script": f"{asker}: {topic(thing)} 어디에 있습니까?", "english_q": f"{asker_en}: Where is the {thing_en}?",
            "image": (f"room_{i}", room_scene(emoji, spot)),
            "correct": (f"{spot}에 있습니다.", f"It's {SPOTS_EN[spot]}."),
            "wrong": [(f"{s}에 있습니다.", f"It's {SPOTS_EN[s]}.") for s in wrong],
            "explanation": f"In the picture the {thing_en} is {SPOTS_EN[spot]}: {spot}에 있습니다.",
        })
    for i, (thing, thing_en, emoji, counter, n) in enumerate(COUNT_THINGS):
        asker = "여" if i % 2 == 0 else "남"
        asker_en = "Woman" if asker == "여" else "Man"
        question = f"{subject(thing)} 몇 {counter} 있습니까?"
        items.append({
            "title_hint": "세기", "tag": "numbers-time", "difficulty": "easy",
            "script": f"{asker}: {question}", "english_q": f"{asker_en}: How many {thing_en} are there?",
            "image": (f"count_{i}_{n}", count_scene(emoji, n)),
            "correct": (f"{native_count(n, counter)} 있습니다.", f"There are {n}."),
            "wrong": [(f"{native_count(k, counter)} 있습니다.", f"There are {k}.") for k in _near_counts(n)],
            "explanation": f"There are {n} in the picture: {native_count(n, counter)}. {counter} is the {COUNTER_EN[counter][1:-1]}, used with native numbers.",
        })
    return items
