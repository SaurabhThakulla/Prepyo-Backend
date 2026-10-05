"""The words behind four reading types and one listening type:
Picture to Word, Description to Picture, Word from Description, Related Word,
and Listen and Choose Picture.

Each entry: (Korean, English, emoji or None, description in Korean,
description in English, a sentence using it in Korean, the same in English).
A description says what the thing is for or what it is like without naming it,
so it can stand as a question. Wrong options come from the same category, so
they are plausible. Everyday and workplace words at EPS-TOPIK level, written
for Prepyo.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Entry:
    ko: str
    en: str
    emoji: str | None
    desc_ko: str | None
    desc_en: str | None
    use_ko: str | None
    use_en: str | None
    category: str
    tag: str


# category key: (Korean label for titles, topic tag)
CATEGORIES = {
    "tools": ("공구", "tools-machines"),
    "safety": ("안전 장비", "industrial-safety"),
    "food": ("음식", "daily-life"),
    "drink": ("음료", "daily-life"),
    "clothes": ("옷과 소지품", "daily-life"),
    "transport": ("교통", "public-services"),
    "home": ("집 안 물건", "basic-life"),
    "kitchen": ("부엌 물건", "basic-life"),
    "office": ("사무용품", "job-korean"),
    "electronics": ("전자 제품", "basic-life"),
    "animals": ("동물", "job-korean"),
    "nature": ("날씨와 자연", "daily-life"),
    "health": ("건강", "public-services"),
    "places": ("장소", "public-services"),
}

_RAW = {
    "tools": [
        ("망치", "hammer", "🔨", "못을 박을 때 쓰는 도구입니다.", "A tool for driving in nails.", "망치로 못을 박아요.", "I hammer in a nail."),
        ("톱", "saw", "🪚", "나무를 자를 때 쓰는 도구입니다. 날카로운 이가 많습니다.", "A tool for cutting wood. It has many sharp teeth.", "톱으로 나무를 잘라요.", "I cut wood with a saw."),
        ("가위", "scissors", "✂️", "종이나 천을 자를 때 씁니다. 손가락을 넣어서 씁니다.", "Used to cut paper or cloth. You put your fingers in it.", "가위로 종이를 잘라요.", "I cut paper with scissors."),
        ("드라이버", "screwdriver", "🪛", "나사를 돌려서 조이거나 풀 때 씁니다.", "Used to turn screws in or out.", "드라이버로 나사를 조여요.", "I tighten a screw with a screwdriver."),
        ("렌치", "wrench", "🔧", "볼트와 너트를 돌릴 때 쓰는 도구입니다.", "A tool for turning bolts and nuts.", "렌치로 볼트를 돌려요.", "I turn a bolt with a wrench."),
        ("줄자", "tape measure", "📏", "길이를 잴 때 씁니다. 당기면 늘어납니다.", "Used to measure length. It stretches when you pull it.", "줄자로 길이를 재요.", "I measure the length with a tape measure."),
        ("사다리", "ladder", "🪜", "높은 곳에 올라갈 때 씁니다.", "Used to climb up to high places.", "사다리를 타고 올라가요.", "I climb up the ladder."),
        ("손전등", "torch", "🔦", "어두운 곳을 밝게 비출 때 씁니다.", "Used to light up dark places.", "손전등을 켜요.", "I turn on the torch."),
        ("도끼", "axe", "🪓", "큰 나무를 쪼갤 때 쓰는 무거운 도구입니다.", "A heavy tool for splitting large pieces of wood.", "도끼로 나무를 쪼개요.", "I split wood with an axe."),
        ("나사", "screw", "🔩", "드라이버로 돌려서 물건을 고정합니다.", "You turn it with a screwdriver to hold things together.", "나사를 하나 더 주세요.", "Please give me one more screw."),
        ("자석", "magnet", "🧲", "쇠를 끌어당기는 물건입니다.", "A thing that pulls iron towards it.", "자석이 쇠를 붙여요.", "The magnet holds the iron."),
        ("삽", "shovel", None, "땅을 파거나 흙을 옮길 때 쓰는 도구입니다.", "A tool for digging or moving soil.", "삽으로 땅을 파요.", "I dig with a shovel."),
        ("펜치", "pliers", None, "철사를 자르거나 구부릴 때 쓰는 도구입니다.", "A tool for cutting or bending wire.", "펜치로 철사를 잘라요.", "I cut the wire with pliers."),
        ("자물쇠", "padlock", "🔒", "문이나 상자를 잠글 때 씁니다. 열쇠로 엽니다.", "Used to lock a door or box. You open it with a key.", "창고 문에 자물쇠를 채워요.", "I put a padlock on the storeroom door."),
    ],
    "safety": [
        ("안전모", "safety helmet", "⛑️", "머리를 보호하기 위해 씁니다. 공사 현장에서 꼭 써야 합니다.", "Worn to protect your head. You must wear it on a building site.", "안전모를 써요.", "I put on a safety helmet."),
        ("장갑", "gloves", "🧤", "손을 보호하기 위해 손에 낍니다.", "Worn on your hands to protect them.", "장갑을 껴요.", "I put on gloves."),
        ("마스크", "face mask", "😷", "먼지를 마시지 않게 코와 입을 가립니다.", "Covers your nose and mouth so you do not breathe in dust.", "마스크를 써요.", "I put on a face mask."),
        ("보안경", "safety goggles", "🥽", "눈을 보호하기 위해 씁니다.", "Worn to protect your eyes.", "보안경을 써요.", "I put on safety goggles."),
        ("안전 조끼", "safety vest", "🦺", "어두운 곳에서도 잘 보이게 입는 밝은 색 조끼입니다.", "A brightly coloured vest worn so people can see you even in the dark.", "현장에서는 안전 조끼를 입어요.", "I wear a safety vest on site."),
        ("소화기", "fire extinguisher", "🧯", "불이 났을 때 불을 끄는 기구입니다.", "Equipment that puts out a fire.", "소화기로 불을 꺼요.", "I put out the fire with an extinguisher."),
        ("안전화", "safety shoes", None, "발을 보호하기 위해 신는 튼튼한 신발입니다.", "Strong shoes worn to protect your feet.", "공장에서는 안전화를 신어요.", "I wear safety shoes in the factory."),
        ("귀마개", "earplugs", None, "시끄러운 곳에서 귀를 보호하기 위해 귀에 넣습니다.", "Put in your ears to protect them in noisy places.", "기계 소리가 커서 귀마개를 해요.", "The machine is loud, so I wear earplugs."),
        ("안전벨트", "safety harness", None, "높은 곳에서 떨어지지 않게 몸에 매는 줄입니다.", "A strap tied to your body so you do not fall from a height.", "높은 곳에서는 안전벨트를 매요.", "I put on a safety harness up high."),
        ("비상구", "emergency exit", None, "불이 났을 때 밖으로 나가는 문입니다.", "The door you leave by when there is a fire.", "비상구로 나가요.", "I go out through the emergency exit."),
    ],
    "food": [
        ("밥", "rice", "🍚", "한국 사람들이 매일 먹는 음식입니다. 쌀로 만듭니다.", "Food Koreans eat every day. It is made from rice grains.", "밥을 먹어요.", "I eat rice."),
        ("빵", "bread", "🍞", "밀가루로 만들어서 굽는 음식입니다.", "Food made from flour and baked.", "아침에 빵을 먹어요.", "I eat bread in the morning."),
        ("라면", "instant noodles", "🍜", "뜨거운 물에 끓여서 먹는 국수입니다. 빨리 만들 수 있습니다.", "Noodles you boil in hot water. They are quick to make.", "라면을 끓여요.", "I cook instant noodles."),
        ("계란", "egg", "🥚", "닭이 낳습니다. 삶거나 부쳐서 먹습니다.", "Hens lay them. You boil or fry them to eat.", "계란을 삶아요.", "I boil an egg."),
        ("고기", "meat", "🥩", "소나 돼지에서 얻는 음식입니다. 구워서 먹습니다.", "Food from cows or pigs. You grill it to eat.", "고기를 구워요.", "I grill meat."),
        ("사과", "apple", "🍎", "빨갛고 둥근 과일입니다.", "A red, round fruit.", "사과를 먹어요.", "I eat an apple."),
        ("바나나", "banana", "🍌", "노랗고 긴 과일입니다. 껍질을 벗겨서 먹습니다.", "A long yellow fruit. You peel it to eat it.", "바나나를 사요.", "I buy bananas."),
        ("수박", "watermelon", "🍉", "여름에 먹는 크고 둥근 과일입니다. 속이 빨갛습니다.", "A big round summer fruit. It is red inside.", "여름에 수박을 먹어요.", "I eat watermelon in summer."),
        ("김밥", "gimbap", None, "밥과 채소를 김으로 싸서 만든 음식입니다.", "Rice and vegetables rolled in seaweed.", "점심에 김밥을 먹어요.", "I eat gimbap for lunch."),
        ("떡", "rice cake", None, "쌀로 만든 쫄깃한 음식입니다. 명절에 많이 먹습니다.", "A chewy food made from rice. People eat a lot of it at holidays.", "설날에 떡을 먹어요.", "I eat rice cakes at New Year."),
        ("케이크", "cake", "🍰", "생일에 먹는 달콤한 음식입니다.", "A sweet food eaten on birthdays.", "생일 케이크를 사요.", "I buy a birthday cake."),
        ("생선", "fish (as food)", "🐟", "바다나 강에서 잡아서 굽거나 끓여 먹습니다.", "Caught in the sea or a river and grilled or stewed.", "저녁에 생선을 구워요.", "I grill fish for dinner."),
    ],
    "drink": [
        ("물", "water", "💧", "목이 마를 때 마십니다. 맛도 색도 없습니다.", "You drink it when you are thirsty. It has no taste or colour.", "물을 마셔요.", "I drink water."),
        ("우유", "milk", "🥛", "소에서 나오는 하얀 음료입니다.", "A white drink that comes from cows.", "우유를 마셔요.", "I drink milk."),
        ("커피", "coffee", "☕", "잠이 깨게 해 주는 검은색 음료입니다.", "A dark drink that wakes you up.", "아침에 커피를 마셔요.", "I drink coffee in the morning."),
        ("녹차", "green tea", "🍵", "찻잎으로 만든 따뜻한 초록색 차입니다.", "A warm green tea made from tea leaves.", "녹차를 마셔요.", "I drink green tea."),
        ("주스", "juice", "🧃", "과일로 만든 달콤한 음료입니다.", "A sweet drink made from fruit.", "오렌지 주스를 마셔요.", "I drink orange juice."),
        ("맥주", "beer", "🍺", "거품이 있는 술입니다. 일이 끝나고 마시기도 합니다.", "An alcoholic drink with a foamy head. People sometimes drink it after work.", "회식에서 맥주를 마셔요.", "I drink beer at the staff dinner."),
        ("콜라", "cola", "🥤", "검은색이고 탄산이 있는 달콤한 음료입니다.", "A sweet, dark, fizzy drink.", "콜라를 마셔요.", "I drink cola."),
        ("와인", "wine", "🍷", "포도로 만든 술입니다.", "An alcoholic drink made from grapes.", "와인을 한 잔 마셔요.", "I drink a glass of wine."),
    ],
    "clothes": [
        ("셔츠", "shirt", "👔", "단추가 있는 윗옷입니다. 회사에 갈 때 많이 입습니다.", "A top with buttons. People often wear it to the office.", "셔츠를 입어요.", "I put on a shirt."),
        ("바지", "trousers", "👖", "두 다리에 입는 옷입니다.", "Clothes you wear on both legs.", "바지를 입어요.", "I put on trousers."),
        ("원피스", "dress", "👗", "위와 아래가 하나로 된 여자 옷입니다.", "A woman's garment with the top and skirt in one piece.", "원피스를 입어요.", "I put on a dress."),
        ("양말", "socks", "🧦", "신발을 신기 전에 발에 신습니다.", "You put them on your feet before your shoes.", "양말을 신어요.", "I put on socks."),
        ("운동화", "trainers", "👟", "운동할 때 신는 편한 신발입니다.", "Comfortable shoes you wear for sport.", "운동화를 신어요.", "I put on trainers."),
        ("모자", "cap", "🧢", "햇빛을 가리려고 머리에 씁니다.", "Worn on your head to keep the sun off.", "모자를 써요.", "I put on a cap."),
        ("장화", "rubber boots", "🥾", "비 오는 날이나 물이 있는 곳에서 신는 긴 신발입니다.", "Tall boots worn in rain or wet places.", "논에서 장화를 신어요.", "I wear rubber boots in the rice paddy."),
        ("외투", "coat", "🧥", "추운 날 옷 위에 입습니다.", "Worn over your clothes on cold days.", "추워서 외투를 입어요.", "It's cold, so I put on a coat."),
        ("목도리", "scarf", "🧣", "추운 날 목에 두릅니다.", "Wrapped round your neck on cold days.", "목도리를 해요.", "I put on a scarf."),
        ("안경", "glasses", "👓", "눈이 나쁠 때 잘 보려고 씁니다.", "Worn to see clearly when your eyesight is poor.", "안경을 써요.", "I put on my glasses."),
        ("가방", "bag", "👜", "물건을 넣어서 들고 다닙니다.", "You put things in it and carry it around.", "가방을 들어요.", "I carry a bag."),
        ("지갑", "wallet", None, "돈이나 카드를 넣어서 가지고 다니는 작은 물건입니다.", "A small thing you carry money and cards in.", "지갑을 잃어버렸어요.", "I've lost my wallet."),
        ("우산", "umbrella", "☂️", "비가 올 때 펴서 머리 위에 듭니다.", "You open it over your head when it rains.", "비가 와서 우산을 써요.", "It's raining, so I use an umbrella."),
        ("손목시계", "wristwatch", "⌚", "손목에 차고 시간을 봅니다.", "Worn on your wrist to tell the time.", "손목시계를 봐요.", "I look at my watch."),
    ],
    "transport": [
        ("버스", "bus", "🚌", "여러 사람이 함께 타는 큰 차입니다. 정류장에서 탑니다.", "A big vehicle many people ride together. You get on at a stop.", "버스를 타요.", "I take the bus."),
        ("택시", "taxi", "🚕", "돈을 내고 원하는 곳까지 타고 가는 차입니다.", "A car you pay to take you where you want to go.", "택시를 타요.", "I take a taxi."),
        ("지하철", "underground train", "🚇", "땅 밑으로 다니는 기차입니다.", "A train that runs under the ground.", "지하철을 타요.", "I take the underground."),
        ("기차", "train", "🚆", "기찻길 위로 먼 도시까지 다닙니다.", "Runs on rails to distant cities.", "기차를 타고 부산에 가요.", "I go to Busan by train."),
        ("비행기", "aeroplane", "✈️", "하늘을 나는 교통수단입니다. 공항에서 탑니다.", "Transport that flies in the sky. You board it at an airport.", "비행기를 타요.", "I take a plane."),
        ("자전거", "bicycle", "🚲", "바퀴가 두 개이고 발로 페달을 밟아서 갑니다.", "It has two wheels and you push the pedals with your feet.", "자전거를 타요.", "I ride a bicycle."),
        ("오토바이", "motorbike", "🏍️", "바퀴가 두 개이고 엔진이 있습니다. 배달할 때 많이 탑니다.", "It has two wheels and an engine. Delivery riders often use it.", "오토바이를 타요.", "I ride a motorbike."),
        ("트럭", "lorry", "🚚", "무거운 짐을 싣고 옮기는 큰 차입니다.", "A big vehicle that carries heavy loads.", "트럭에 짐을 실어요.", "I load the lorry."),
        ("구급차", "ambulance", "🚑", "아픈 사람을 병원으로 데려가는 차입니다.", "A vehicle that takes sick people to hospital.", "구급차를 불러요.", "I call an ambulance."),
        ("소방차", "fire engine", "🚒", "불을 끄러 가는 빨간 차입니다.", "A red vehicle that goes to put out fires.", "소방차가 와요.", "The fire engine is coming."),
        ("자동차", "car", "🚗", "사람이 타는 차입니다. 운전 면허가 있어야 운전할 수 있습니다.", "A vehicle people ride in. You need a licence to drive one.", "자동차를 운전해요.", "I drive a car."),
    ],
    "home": [
        ("침대", "bed", "🛏️", "잠을 잘 때 눕는 가구입니다.", "Furniture you lie on to sleep.", "침대에서 자요.", "I sleep in the bed."),
        ("의자", "chair", "🪑", "앉을 때 쓰는 가구입니다.", "Furniture you sit on.", "의자에 앉아요.", "I sit on the chair."),
        ("거울", "mirror", "🪞", "자기 얼굴을 볼 수 있습니다.", "You can see your own face in it.", "거울을 봐요.", "I look in the mirror."),
        ("열쇠", "key", "🔑", "문을 열거나 잠글 때 씁니다.", "Used to open or lock a door.", "열쇠로 문을 열어요.", "I open the door with a key."),
        ("휴지", "toilet paper", "🧻", "화장실에서 쓰는 얇은 종이입니다.", "Thin paper used in the toilet.", "휴지를 사요.", "I buy toilet paper."),
        ("비누", "soap", "🧼", "손을 씻을 때 거품을 내서 씁니다.", "You make a lather with it to wash your hands.", "비누로 손을 씻어요.", "I wash my hands with soap."),
        ("칫솔", "toothbrush", "🪥", "이를 닦을 때 씁니다.", "Used to brush your teeth.", "칫솔로 이를 닦아요.", "I brush my teeth with a toothbrush."),
        ("텔레비전", "television", "📺", "뉴스나 드라마를 보는 기계입니다.", "A machine for watching news and dramas.", "텔레비전을 봐요.", "I watch television."),
        ("전등", "light", "💡", "방을 밝게 해 줍니다. 스위치로 켭니다.", "It makes a room bright. You turn it on with a switch.", "전등을 켜요.", "I turn on the light."),
        ("냉장고", "fridge", None, "음식을 차갑게 보관하는 큰 전자 제품입니다.", "A large appliance that keeps food cold.", "우유를 냉장고에 넣어요.", "I put the milk in the fridge."),
        ("세탁기", "washing machine", None, "더러운 옷을 넣으면 빨래를 해 줍니다.", "You put dirty clothes in it and it washes them.", "세탁기를 돌려요.", "I run the washing machine."),
        ("빗자루", "broom", None, "바닥의 먼지나 쓰레기를 쓸 때 씁니다.", "Used to sweep dust and rubbish off the floor.", "빗자루로 마당을 쓸어요.", "I sweep the yard with a broom."),
        ("수건", "towel", None, "샤워한 후에 몸을 닦습니다.", "Used to dry your body after a shower.", "수건으로 얼굴을 닦아요.", "I dry my face with a towel."),
    ],
    "kitchen": [
        ("프라이팬", "frying pan", "🍳", "기름을 두르고 음식을 볶거나 부칠 때 씁니다.", "Used with oil to stir-fry or fry food.", "프라이팬에 계란을 부쳐요.", "I fry an egg in the frying pan."),
        ("숟가락", "spoon", "🥄", "밥이나 국을 떠서 먹을 때 씁니다.", "Used to scoop up rice or soup to eat.", "숟가락으로 국을 먹어요.", "I eat soup with a spoon."),
        ("젓가락", "chopsticks", "🥢", "두 개가 한 쌍입니다. 반찬을 집어서 먹습니다.", "They come as a pair. You pick up side dishes with them.", "젓가락으로 반찬을 먹어요.", "I eat side dishes with chopsticks."),
        ("그릇", "bowl", "🥣", "밥이나 국을 담는 오목한 것입니다.", "A deep dish that holds rice or soup.", "그릇에 밥을 담아요.", "I put rice in a bowl."),
        ("칼", "knife", "🔪", "음식 재료를 썰 때 쓰는 날카로운 도구입니다.", "A sharp tool for cutting up ingredients.", "칼로 채소를 썰어요.", "I chop vegetables with a knife."),
        ("접시", "plate", "🍽️", "음식을 올려놓는 납작한 그릇입니다.", "A flat dish you put food on.", "접시를 씻어요.", "I wash the plates."),
        ("주전자", "kettle", "🫖", "물을 끓이거나 차를 따를 때 씁니다.", "Used to boil water or pour tea.", "주전자에 물을 끓여요.", "I boil water in the kettle."),
        ("냄비", "pot", "🍲", "국이나 찌개를 끓일 때 쓰는 깊은 그릇입니다.", "A deep pan for cooking soup or stew.", "냄비에 찌개를 끓여요.", "I cook stew in a pot."),
    ],
    "office": [
        ("연필", "pencil", "✏️", "글씨를 쓰고 지우개로 지울 수 있습니다.", "You write with it and can rub it out with an eraser.", "연필로 이름을 써요.", "I write my name in pencil."),
        ("볼펜", "ballpoint pen", "🖊️", "잉크로 글씨를 씁니다. 지울 수 없습니다.", "You write with ink. It cannot be rubbed out.", "볼펜으로 서류에 써요.", "I fill in the form with a pen."),
        ("공책", "notebook", "📓", "글씨를 쓰는 종이를 묶은 책입니다.", "Sheets of paper bound together for writing in.", "공책에 단어를 써요.", "I write words in my notebook."),
        ("책", "book", "📕", "글이 인쇄되어 있어서 읽습니다.", "It has printed writing that you read.", "책을 읽어요.", "I read a book."),
        ("달력", "calendar", "📅", "날짜와 요일을 볼 수 있습니다.", "You can see dates and days of the week on it.", "달력에 휴일을 표시해요.", "I mark the day off on the calendar."),
        ("편지 봉투", "envelope", "✉️", "편지를 넣어서 보냅니다.", "You put a letter in it and send it.", "편지를 봉투에 넣어요.", "I put the letter in an envelope."),
        ("상자", "box", "📦", "물건을 넣어서 옮기거나 보관합니다.", "You put things in it to move or store them.", "상자를 옮겨요.", "I move the box."),
        ("클립", "paper clip", "📎", "종이 몇 장을 함께 집어 둡니다.", "Holds a few sheets of paper together.", "서류를 클립으로 집어요.", "I clip the papers together."),
        ("서류", "documents", "📄", "회사나 관청에 내는 종이입니다.", "Papers you hand in to a company or government office.", "서류를 내요.", "I hand in the documents."),
    ],
    "electronics": [
        ("휴대폰", "mobile phone", "📱", "들고 다니면서 전화하고 문자를 보냅니다.", "You carry it to make calls and send texts.", "휴대폰으로 전화해요.", "I call on my mobile."),
        ("컴퓨터", "computer", "💻", "인터넷을 하고 문서를 만듭니다.", "You use it for the internet and to make documents.", "컴퓨터로 일해요.", "I work on the computer."),
        ("전화기", "telephone", "☎️", "책상 위에 두고 쓰는 전화입니다.", "A phone that sits on a desk.", "사무실 전화기가 울려요.", "The office phone is ringing."),
        ("카메라", "camera", "📷", "사진을 찍는 기계입니다.", "A machine that takes photos.", "카메라로 사진을 찍어요.", "I take a photo with a camera."),
        ("충전기", "charger", "🔌", "휴대폰 배터리를 채울 때 씁니다.", "Used to fill up a phone's battery.", "충전기를 꽂아요.", "I plug in the charger."),
        ("배터리", "battery", "🔋", "전기를 저장해서 기계를 움직이게 합니다.", "Stores electricity to make machines work.", "배터리가 다 됐어요.", "The battery is flat."),
        ("이어폰", "earphones", "🎧", "다른 사람에게 들리지 않게 귀로 음악을 듣습니다.", "You listen to music in your ears without others hearing.", "이어폰으로 음악을 들어요.", "I listen to music with earphones."),
        ("선풍기", "electric fan", None, "더울 때 바람을 만들어 주는 전자 제품입니다.", "An appliance that makes a breeze when it is hot.", "더워서 선풍기를 켜요.", "It's hot, so I turn on the fan."),
        ("에어컨", "air conditioner", None, "여름에 방을 시원하게 해 줍니다.", "Keeps a room cool in summer.", "에어컨을 켜요.", "I turn on the air conditioner."),
    ],
    "animals": [
        ("소", "cow", "🐄", "농장에서 기르는 큰 동물입니다. 우유를 줍니다.", "A large farm animal that gives milk.", "소에게 사료를 줘요.", "I feed the cow."),
        ("돼지", "pig", "🐖", "농장에서 기르는 동물입니다. 코가 납작합니다.", "A farm animal with a flat nose.", "돼지가 밥을 먹어요.", "The pig is eating."),
        ("닭", "chicken", "🐔", "알을 낳는 새입니다. 아침에 웁니다.", "A bird that lays eggs and crows in the morning.", "닭이 알을 낳아요.", "The hen lays an egg."),
        ("오리", "duck", "🦆", "물에서 헤엄치는 새입니다. 꽥꽥 웁니다.", "A bird that swims on water and quacks.", "오리가 헤엄쳐요.", "The duck is swimming."),
        ("개", "dog", "🐕", "사람과 같이 사는 동물입니다. 집을 지킵니다.", "An animal that lives with people and guards the house.", "개가 짖어요.", "The dog is barking."),
        ("고양이", "cat", "🐈", "야옹 하고 우는 작은 동물입니다.", "A small animal that miaows.", "고양이가 자요.", "The cat is sleeping."),
        ("말", "horse", "🐎", "빨리 달리는 큰 동물입니다. 사람이 탈 수 있습니다.", "A large, fast animal that people can ride.", "말을 타요.", "I ride a horse."),
        ("양", "sheep", "🐑", "털이 많고 하얀 동물입니다. 털로 옷을 만듭니다.", "A white animal with thick wool used for clothes.", "양의 털을 깎아요.", "I shear the sheep."),
        ("물고기", "fish (animal)", "🐠", "물속에 살고 지느러미로 헤엄칩니다.", "It lives in water and swims with fins.", "물고기가 헤엄쳐요.", "The fish is swimming."),
        ("새", "bird", "🐦", "날개가 있어서 하늘을 납니다.", "It has wings and flies in the sky.", "새가 날아요.", "The bird is flying."),
    ],
    "nature": [
        ("비", "rain", "🌧️", "하늘에서 물이 떨어지는 날씨입니다.", "Weather when water falls from the sky.", "비가 와요.", "It's raining."),
        ("눈", "snow", "❄️", "겨울에 하얗게 내립니다.", "Falls white in winter.", "눈이 와요.", "It's snowing."),
        ("구름", "cloud", "☁️", "하늘에 떠 있는 하얀 것입니다.", "The white things floating in the sky.", "구름이 많아요.", "It's very cloudy."),
        ("무지개", "rainbow", "🌈", "비가 그친 후에 하늘에 생기는 일곱 가지 색입니다.", "Seven colours in the sky after the rain stops.", "무지개가 떴어요.", "A rainbow has appeared."),
        ("해", "sun", "☀️", "낮에 하늘에서 빛나고 따뜻하게 해 줍니다.", "Shines in the sky by day and keeps us warm.", "해가 떠요.", "The sun is rising."),
        ("달", "moon", "🌙", "밤하늘에 떠서 빛납니다.", "Shines in the night sky.", "달이 밝아요.", "The moon is bright."),
        ("나무", "tree", "🌳", "땅에 뿌리를 내리고 자랍니다. 잎이 많습니다.", "It grows from roots in the ground and has many leaves.", "나무를 심어요.", "I plant a tree."),
        ("꽃", "flower", "🌸", "봄에 많이 피고 예쁜 색과 향기가 있습니다.", "Blooms in spring with pretty colours and a scent.", "꽃이 피어요.", "The flowers are blooming."),
        ("산", "mountain", "⛰️", "아주 높은 땅입니다. 등산을 합니다.", "Very high land that people climb.", "주말에 산에 가요.", "I go to the mountains at the weekend."),
        ("바다", "sea", "🌊", "아주 넓고 짠 물입니다.", "A very wide area of salt water.", "바다에서 수영해요.", "I swim in the sea."),
    ],
    "health": [
        ("약", "medicine", "💊", "아플 때 먹으면 낫게 도와줍니다.", "You take it when you are ill to help you get better.", "식후에 약을 먹어요.", "I take medicine after meals."),
        ("주사", "injection", "💉", "바늘로 약을 몸에 넣습니다.", "Puts medicine into your body through a needle.", "병원에서 주사를 맞아요.", "I get an injection at the hospital."),
        ("체온계", "thermometer", "🌡️", "열이 있는지 몸의 온도를 잽니다.", "Measures your body temperature to see if you have a fever.", "체온계로 열을 재요.", "I take my temperature."),
        ("반창고", "plaster", "🩹", "작은 상처에 붙입니다.", "You stick it on a small cut.", "손가락에 반창고를 붙여요.", "I put a plaster on my finger."),
        ("휠체어", "wheelchair", "♿", "걸을 수 없는 사람이 앉아서 움직입니다.", "People who cannot walk sit in it to move around.", "휠체어를 밀어요.", "I push the wheelchair."),
        ("치과", "dentist's", None, "이가 아플 때 가는 병원입니다.", "The clinic you go to when your teeth hurt.", "이가 아파서 치과에 가요.", "My tooth hurts, so I go to the dentist."),
        ("감기", "a cold", "🤧", "기침과 콧물이 나는 흔한 병입니다.", "A common illness with a cough and runny nose.", "감기에 걸렸어요.", "I've caught a cold."),
        ("병원", "hospital", "🏥", "아픈 사람이 의사에게 치료를 받는 곳입니다.", "Where sick people are treated by doctors.", "병원에 가요.", "I go to the hospital."),
    ],
    "places": [
        ("은행", "bank", "🏦", "돈을 맡기거나 찾고 다른 나라로 보낼 수 있는 곳입니다.", "Where you deposit or withdraw money and send it abroad.", "은행에서 돈을 찾아요.", "I take out money at the bank."),
        ("학교", "school", "🏫", "학생들이 선생님에게 공부를 배우는 곳입니다.", "Where students learn from teachers.", "학교에 가요.", "I go to school."),
        ("우체국", "post office", "🏤", "편지나 소포를 보내는 곳입니다.", "Where you send letters and parcels.", "우체국에서 소포를 보내요.", "I send a parcel at the post office."),
        ("공장", "factory", "🏭", "기계로 물건을 만드는 곳입니다.", "Where things are made with machines.", "공장에서 일해요.", "I work at the factory."),
        ("편의점", "convenience store", "🏪", "24시간 문을 여는 작은 가게입니다.", "A small shop open 24 hours.", "편의점에서 물을 사요.", "I buy water at the convenience store."),
        ("호텔", "hotel", "🏨", "여행할 때 돈을 내고 잠을 자는 곳입니다.", "Where you pay to sleep when travelling.", "호텔에 묵어요.", "I stay at a hotel."),
        ("교회", "church", "⛪", "기독교 사람들이 기도하러 가는 곳입니다.", "Where Christians go to pray.", "일요일에 교회에 가요.", "I go to church on Sunday."),
        ("집", "home", "🏠", "가족과 함께 살고 쉬는 곳입니다.", "Where you live and rest with your family.", "일이 끝나고 집에 가요.", "I go home after work."),
        ("공항", "airport", "🛫", "비행기를 타고 내리는 곳입니다.", "Where you get on and off planes.", "공항에 도착했어요.", "I've arrived at the airport."),
        ("시장", "market", None, "여러 가게가 모여서 물건을 싸게 파는 곳입니다.", "Many stalls together selling things cheaply.", "시장에서 채소를 사요.", "I buy vegetables at the market."),
        ("식당", "restaurant", "🍴", "돈을 내고 음식을 사 먹는 곳입니다.", "Where you pay to eat a meal.", "식당에서 점심을 먹어요.", "I have lunch at a restaurant."),
        ("주유소", "petrol station", "⛽", "자동차에 기름을 넣는 곳입니다.", "Where you put fuel in a car.", "주유소에서 기름을 넣어요.", "I fill up at the petrol station."),
    ],
}

ENTRIES: list[Entry] = []
for _cat, _items in _RAW.items():
    for ko, en, emoji, dk, de, uk, ue in _items:
        ENTRIES.append(Entry(ko, en, emoji, dk, de, uk, ue, _cat, CATEGORIES[_cat][1]))

# Pairs too close to stand as each other's wrong answer: a description of
# one fits the other as well.
CONFUSABLE = {frozenset(p) for p in [
    ("병원", "치과"), ("자동차", "택시"), ("자동차", "버스"), ("자동차", "트럭"), ("자동차", "구급차"),
    ("자동차", "소방차"), ("전화기", "휴대폰"), ("그릇", "냄비"), ("새", "닭"), ("새", "오리"),
    ("그릇", "접시"), ("기차", "지하철"), ("모자", "안경"),
]}


def fits_with(a: Entry, b: Entry) -> bool:
    return a.ko != b.ko and frozenset((a.ko, b.ko)) not in CONFUSABLE


# Every word and every picture appears once.
assert len({e.ko for e in ENTRIES}) == len(ENTRIES), "duplicate Korean word"
_emojis = [e.emoji for e in ENTRIES if e.emoji]
assert len(set(_emojis)) == len(_emojis), "duplicate emoji"
