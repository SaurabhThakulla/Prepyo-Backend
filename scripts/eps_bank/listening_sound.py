"""Choose What You Heard: one word is read, and three words that sound close
to it stand beside it. Each set contrasts a sound EPS-TOPIK learners mix up:
plain, tense and aspirated consonants (ㅂ/ㅃ/ㅍ, ㄷ/ㄸ/ㅌ, ㅈ/ㅉ/ㅊ, ㄱ/ㄲ/ㅋ,
ㅅ/ㅆ), close vowels (ㅓ/ㅗ, ㅐ/ㅔ, ㅡ/ㅜ) and final consonants (ㄴ/ㅇ/ㅁ).

Each entry: (word heard, its English, [three words that sound close, each
with its English], topic tag).
"""

SETS = [
    ("불", "fire", [("풀", "grass"), ("뿔", "horn"), ("볼", "cheek")], "daily-life"),
    ("달", "moon", [("탈", "mask"), ("딸", "daughter"), ("돌", "stone")], "daily-life"),
    ("방", "room", [("빵", "bread"), ("밤", "night"), ("반", "half")], "basic-life"),
    ("자다", "to sleep", [("차다", "to kick"), ("짜다", "to be salty"), ("사다", "to buy")], "basic-life"),
    ("굴", "oyster", [("꿀", "honey"), ("귤", "tangerine"), ("글", "writing")], "daily-life"),
    ("개", "dog", [("깨", "sesame"), ("게", "crab"), ("귀", "ear")], "daily-life"),
    ("공", "ball", [("콩", "bean"), ("곰", "bear"), ("강", "river")], "daily-life"),
    ("살", "flesh", [("쌀", "rice grain"), ("술", "alcohol"), ("설", "New Year")], "daily-life"),
    ("비", "rain", [("피", "blood"), ("배", "pear"), ("뼈", "bone")], "daily-life"),
    ("문", "door", [("물", "water"), ("무", "radish"), ("눈", "eye")], "basic-life"),
    ("손", "hand", [("솜", "cotton"), ("선", "line"), ("산", "mountain")], "basic-life"),
    ("사람", "person", [("사랑", "love"), ("사탕", "sweet"), ("사장", "company owner")], "basic-life"),
    ("가게", "shop", [("가방", "bag"), ("가발", "wig"), ("가장", "head of the family")], "daily-life"),
    ("회사", "company", [("회의", "meeting"), ("화가", "painter"), ("회비", "membership fee")], "job-korean"),
    ("운동", "exercise", [("운전", "driving"), ("온도", "temperature"), ("운명", "fate")], "daily-life"),
    ("공원", "park", [("병원", "hospital"), ("정원", "garden"), ("직원", "employee")], "public-services"),
    ("학교", "school", [("학과", "department"), ("학생", "student"), ("하교", "going home from school")], "daily-life"),
    ("이유", "reason", [("우유", "milk"), ("여유", "spare time"), ("오이", "cucumber")], "daily-life"),
    ("의자", "chair", [("이자", "interest on money"), ("의사", "doctor"), ("여자", "woman")], "basic-life"),
    ("지하철", "underground train", [("지하실", "basement"), ("지하도", "underpass"), ("지하수", "groundwater")], "public-services"),
    ("안전", "safety", [("안정", "stability"), ("안경", "glasses"), ("안내", "guidance")], "industrial-safety"),
    ("기계", "machine", [("시계", "clock"), ("기회", "chance"), ("기차", "train")], "tools-machines"),
    ("월급", "monthly pay", [("월세", "monthly rent"), ("월말", "end of the month"), ("일급", "daily wage")], "job-korean"),
    ("출근", "going to work", [("퇴근", "leaving work"), ("출구", "exit"), ("출국", "leaving the country")], "workplace-culture"),
    ("입구", "entrance", [("입국", "entering the country"), ("인구", "population"), ("입금", "deposit")], "public-services"),
    ("계산", "paying the bill", [("계단", "stairs"), ("계란", "egg"), ("개선", "improvement")], "daily-life"),
    ("물건", "goods", [("물결", "wave"), ("문건", "document"), ("물감", "paint")], "daily-life"),
    ("사진", "photo", [("사전", "dictionary"), ("사정", "circumstances"), ("사장", "company owner")], "daily-life"),
    ("편지", "letter", [("펜치", "pliers"), ("판지", "cardboard"), ("벤치", "bench")], "daily-life"),
    ("휴지", "tissue", [("휴가", "holiday"), ("휴일", "day off"), ("유지", "upkeep")], "basic-life"),
    ("청소", "cleaning", [("장소", "place"), ("청년", "young person"), ("정상", "normal")], "basic-life"),
    ("주소", "address", [("주스", "juice"), ("주사", "injection"), ("조사", "survey")], "public-services"),
    ("바지", "trousers", [("바다", "sea"), ("바위", "rock"), ("바닥", "floor")], "daily-life"),
    ("신발", "shoes", [("신문", "newspaper"), ("신분", "identity"), ("신부", "bride")], "daily-life"),
    ("감자", "potato", [("감사", "thanks"), ("감기", "a cold"), ("과자", "snack")], "daily-life"),
    ("요리", "cooking", [("유리", "glass"), ("오리", "duck"), ("우리", "we")], "daily-life"),
    ("전기", "electricity", [("정기", "regular"), ("전구", "light bulb"), ("정지", "stop")], "industrial-safety"),
    ("가위", "scissors", [("거위", "goose"), ("가구", "furniture"), ("가요", "pop songs")], "tools-machines"),
    ("빨래", "laundry", [("발레", "ballet"), ("벌레", "insect"), ("빨대", "straw")], "basic-life"),
    ("짐", "luggage", [("집", "house"), ("김", "seaweed"), ("침", "saliva")], "basic-life"),
    ("서울", "Seoul", [("저울", "scales"), ("겨울", "winter"), ("거울", "mirror")], "korean-culture"),
    ("날씨", "weather", [("날짜", "date"), ("말씨", "way of speaking"), ("낱말", "word")], "daily-life"),
    ("숙제", "homework", [("축제", "festival"), ("국제", "international"), ("수제", "handmade")], "daily-life"),
    ("연습", "practice", [("예습", "preparing for a lesson"), ("연필", "pencil"), ("염소", "goat")], "daily-life"),
    ("회식", "staff dinner", [("휴식", "rest"), ("회색", "grey"), ("음식", "food")], "workplace-culture"),
    ("작업", "work task", [("직업", "occupation"), ("졸업", "graduation"), ("작별", "farewell")], "job-korean"),
    ("위험", "danger", [("시험", "exam"), ("체험", "experience"), ("위협", "threat")], "industrial-safety"),
    ("고장", "breakdown", [("고향", "hometown"), ("공항", "airport"), ("고생", "hardship")], "tools-machines"),
]

assert len(SETS) == 48, len(SETS)
