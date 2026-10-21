"""
Generate Alfa PTE Practice Questions & Benchmark Samples (Migration 000112).
Complete 100 Questions Edition across all 20 PTE Academic Subtasks (5 questions each).
"""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
UP_MIGRATION = ROOT / "migrations" / "000112_alfa_pte_practice_questions.up.sql"
DOWN_MIGRATION = ROOT / "migrations" / "000112_alfa_pte_practice_questions.down.sql"
GRADEBENCH_SAMPLES = ROOT / "cmd" / "gradebench" / "samples_alfa_pte.json"

VERSION = "pte-2026-01"

def sql_escape(s):
    if s is None:
        return ""
    return str(s).replace("'", "''")

def sql_str_or_null(s):
    if s is None:
        return "NULL"
    return f"'{sql_escape(s)}'"


def normalize_opt(opt):
    clean = {}
    for k, v in opt.items():
        k_norm = str(k).replace(' ', '').replace('\n', '').replace('\r', '')
        clean[k_norm] = v
    return clean

def sql_json_or_null(j):
    if j is None:
        return "NULL"
    s = json.dumps(j, ensure_ascii=False)
    return f"'{sql_escape(s)}'::jsonb"

# Load extracted raw data
DATA_FILE = Path(__file__).parent / "extracted_alfa_100.json"
with open(DATA_FILE, "r", encoding="utf-8") as f:
    RAW_DATA = json.load(f)

BY_ID = {q["id"]: q for q in RAW_DATA}

# ==========================================
# 1.1 Reading & Writing FIB (Dropdown) (5)
# ==========================================
RWFIB_DATA = [
    {
        "id": "12001335",
        "title": "Public Education",
        "clean": "Public education is generally free. However, many schools to choose to charge parents for extra services, such as textbooks, excursions, and music lessons. Low-income families may not be able to afford these costs, which puts their children at a disadvantage, especially when accessing information.",
        "text": "Public education is generally free. However, many schools to [[b1]] to charge parents for extra services, [[b2]] as textbooks, excursions, and music lessons. Low-income families may not be able to afford [[b3]] costs, which puts their children at a disadvantage, especially when [[b4]] information.",
        "blanks": [
            {"id": "b1", "options": ["choose", "decide", "agree", "refuse"], "correctAnswer": "choose"},
            {"id": "b2", "options": ["such", "like", "same", "much"], "correctAnswer": "such"},
            {"id": "b3", "options": ["these", "those", "their", "extra"], "correctAnswer": "these"},
            {"id": "b4", "options": ["accessing", "arranging", "learning", "seeking"], "correctAnswer": "accessing"},
        ],
        "points": 4,
    },
    {
        "id": "12001334",
        "title": "Resultant Force",
        "clean": "The overall result of two or more forces acting on an object is called the resultant force. The resultant of two forces is a single force, which has the same effect as the two forces combined. If two forces pull an object in opposite directions, the size of the resultant can be found by subtracting one force from the other. If the forces are equal, they balance each other.",
        "text": "The overall result of two or [[b1]] forces acting on an object is called the resultant force. The resultant of two forces is a single force, which has the same effect as the two forces combined. If two forces pull an object in [[b2]] directions, the size of the resultant can be found by [[b3]] one force from the other. If the forces are [[b4]], they balance each other.",
        "blanks": [
            {"id": "b1", "options": ["more", "less", "many", "several"], "correctAnswer": "more"},
            {"id": "b2", "options": ["opposite", "different", "both", "various"], "correctAnswer": "opposite"},
            {"id": "b3", "options": ["subtracting", "adding", "sharing", "dividing"], "correctAnswer": "subtracting"},
            {"id": "b4", "options": ["equal", "same", "balanced", "stable"], "correctAnswer": "equal"},
        ],
        "points": 4,
    },
    {
        "id": "12001333",
        "title": "Dogs & Humans",
        "clean": "Dogs, the first domestic animals, evolved from East Asian wolves approximately 20,000-40,000 years ago. Dogs understand human body language, many spoken words, and even abstract symbols. Dogs pay attention to human faces and eyes, and they often turn to humans to help solve problems. Dogs and humans form a mutual emotional bond that reduces stress in both species. Being around dogs improve human physical and mental well-being. Most dogs are very social and play with other dogs if they can. Stray and feral dogs are common on the edges of human society. They form social groups, and sometimes cooperate in territorial defense, much as wolves do.",
        "text": "Dogs, the first [[b1]] animals, evolved from East Asian wolves approximately 20,000-40,000 years ago. Dogs understand human body language, many spoken words, and even abstract symbols. Dogs pay attention to human faces and eyes, and they often [[b2]] to humans to help solve problems. Dogs and humans form a [[b3]] emotional bond that reduces stress in both species. Being [[b4]] dogs improve human physical and mental well-being. Most dogs are very social and play with other dogs if they can. Stray and feral dogs are common on the edges of human society. They form social groups, and sometimes cooperate in territorial defense, much as wolves do.",
        "blanks": [
            {"id": "b1", "options": ["domestic", "wild", "tamed", "native"], "correctAnswer": "domestic"},
            {"id": "b2", "options": ["turn", "look", "refer", "come"], "correctAnswer": "turn"},
            {"id": "b3", "options": ["mutual", "common", "strong", "deep"], "correctAnswer": "mutual"},
            {"id": "b4", "options": ["around", "with", "near", "within"], "correctAnswer": "around"},
        ],
        "points": 4,
    },
    {
        "id": "12001332",
        "title": "Learning",
        "clean": "Learning to ride a bicycle as an adult can be challenging, especially for a beginner. A trained instructor can guide learners through basic techniques and help them feel more comfortable on the road. The purpose is not to become a professional rider immediately, but to build balance and confidence. Even an amateur can improve with regular practice. Later, some learners may join outdoor events or friendly competitions.",
        "text": "Learning to ride a bicycle as an adult can be challenging, especially for a beginner. A trained [[b1]] can guide learners through basic techniques and help them feel more [[b2]] on the road. The purpose is not to become a professional rider immediately, but to build balance and [[b3]]. Even an [[b4]] can improve with regular practice. Later, some learners may join outdoor events or friendly [[b5]].",
        "blanks": [
            {"id": "b1", "options": ["instructor", "coach", "teacher", "trainer"], "correctAnswer": "instructor"},
            {"id": "b2", "options": ["comfortable", "confident", "safe", "easy"], "correctAnswer": "comfortable"},
            {"id": "b3", "options": ["confidence", "strength", "balance", "skill"], "correctAnswer": "confidence"},
            {"id": "b4", "options": ["amateur", "beginner", "adult", "owner"], "correctAnswer": "amateur"},
            {"id": "b5", "options": ["competitions", "races", "parties", "events"], "correctAnswer": "competitions"},
        ],
        "points": 5,
    },
    {
        "id": "12001331",
        "title": "Charity Group",
        "clean": "Landing a position with the acclaimed charity group, often referred to as the angel providers, was a dream come true for Sarah. Although the organization could only offer a meager stipend, she eagerly accepted the terms. She believed that working alongside such selfless individuals was far more rewarding than chasing a high corporate salary. Over the years, this dedicated group has continued to provide essential healthcare to impoverished communities worldwide, relying heavily on passionate workers who value purpose over financial gain.",
        "text": "Landing a position with the acclaimed charity group, often referred to as the [[b1]] providers, was a dream come true for Sarah. Although the organization could only offer a [[b2]] stipend, she eagerly [[b3]] the terms. She believed that working alongside such selfless individuals was far more [[b4]] than chasing a high corporate salary. Over the years, this dedicated group has continued to provide essential healthcare to impoverished communities worldwide, relying heavily on passionate workers who value purpose over financial gain.",
        "blanks": [
            {"id": "b1", "options": ["angel", "care", "health", "people"], "correctAnswer": "angel"},
            {"id": "b2", "options": ["meager", "small", "huge", "modest"], "correctAnswer": "meager"},
            {"id": "b3", "options": ["accepted", "agreed", "signed", "took"], "correctAnswer": "accepted"},
            {"id": "b4", "options": ["rewarding", "valuable", "meaningful", "decreasing"], "correctAnswer": "rewarding"},
        ],
        "points": 4,
    },
]

# ==========================================
# 1.2 Reading MCQ Multiple (5)
# ==========================================
RMCMA_DATA = [
    {
        "id": "9000198",
        "title": "Job Opening",
        "prompt": "Read the text and answer the question by selecting all the correct responses. More than one response is correct. Which of the following are provided according to the job advertisement?",
    },
    {
        "id": "9000197",
        "title": "Disciplined Inquiry",
        "prompt": "Read the text and answer the question by selecting all the correct responses. More than one response is correct. Which of the following limitations of research studies are mentioned in the passage?",
    },
    {
        "id": "9000196",
        "title": "Migration",
        "prompt": "Read the text and answer the question by selecting all the correct responses. More than one response is correct. Which of the following statements about demographic challenges are supported by the text?",
    },
    {
        "id": "9000195",
        "title": "Genome",
        "prompt": "Read the text and answer the question by selecting all the correct responses. More than one response is correct. Which of the following are true of personalized medicine according to the passage?",
    },
    {
        "id": "9000194",
        "title": "Attention",
        "prompt": "Read the text and answer the question by selecting all the correct responses. More than one response is correct. Which of the following findings are supported by the pigeon experiments?",
    },
]

# ==========================================
# 1.3 Re-order Paragraphs (5)
# ==========================================
RO_DATA = [
    {
        "id": "10001082",
        "title": "Farmland and Forest",
        "paragraphs": [
            {"label": "p1", "text": "Seeds sprout, producing seedlings."},
            {"label": "p2", "text": "They grow into saplings, which shade the lawn below."},
            {"label": "p3", "text": "The saplings will grow into big trees, and eventually a forest will form."},
            {"label": "p4", "text": "Farmland can be created by clearing forests."},
            {"label": "p5", "text": "If the farmland is abandoned, the process will start over again."},
        ],
        "correct_order": ["p1", "p2", "p3", "p4", "p5"],
        "points": 4,
    },
    {
        "id": "10001081",
        "title": "Pest Control",
        "paragraphs": [
            {"label": "p1", "text": "Chemical pest control dates back 4,500 years, when the Sumerians used sulfur compounds as insecticides."},
            {"label": "p2", "text": "The Rig Veda, which is about 4,000 years old, also mentions the use of poisonous plants for pest control."},
            {"label": "p3", "text": "And the ancient Chinese and Egyptians are also known to have used chemical pest control."},
            {"label": "p4", "text": "But it was only with the industrialization and mechanization of agriculture in the 18th and 19th centuries, and the introduction of the insecticides pyrethrum and derris that chemical pest control became the method of choice."},
        ],
        "correct_order": ["p1", "p2", "p3", "p4"],
        "points": 3,
    },
    {
        "id": "10001080",
        "title": "Forest Cycle",
        "paragraphs": [
            {"label": "p1", "text": "Seeds sprout from the soil and gradually develop into small seedlings."},
            {"label": "p2", "text": "These seedlings continue to grow and eventually become young saplings."},
            {"label": "p3", "text": "The saplings grow into large trees, and over time, a forest begins to form."},
            {"label": "p4", "text": "The forest may later be cleared to create farmland for agricultural use."},
            {"label": "p5", "text": "If the farmland is abandoned, natural vegetation begins to return and the process starts again."},
        ],
        "correct_order": ["p1", "p2", "p3", "p4", "p5"],
        "points": 4,
    },
    {
        "id": "10001079",
        "title": "Bugs",
        "paragraphs": [
            {"label": "p1", "text": "If the landlord refuses to deal with a bed bug infestation, the tenant can seek assistance from the government."},
            {"label": "p2", "text": "Alternatively, the tenant may hire a professional exterminator directly."},
            {"label": "p3", "text": "Because bed bugs are extremely difficult to eliminate, the work should be carried out by trained pest control professionals."},
            {"label": "p4", "text": "If the tenant hires an exterminator independently, they should choose a reputable company and make sure the treatment is performed properly."},
        ],
        "correct_order": ["p1", "p2", "p3", "p4"],
        "points": 3,
    },
    {
        "id": "10001078",
        "title": "Whale",
        "paragraphs": [
            {"label": "p1", "text": "Whales generally live in the water."},
            {"label": "p2", "text": "However, as mammals, they must come to the surface to breathe."},
            {"label": "p3", "text": "That's why their breathing organ (blowhole) is located on the top of their heads."},
            {"label": "p4", "text": "When they breathe out, they produce a spout of water vapor that looks like a fountain."},
        ],
        "correct_order": ["p1", "p2", "p3", "p4"],
        "points": 3,
    },
]

# ==========================================
# 1.4 Reading FIB (Drag and Drop) (5)
# ==========================================
RFIB_DATA = [
    {
        "id": "11001117",
        "title": "Mountain Forest",
        "text": "The place had both mountains and a [[b1]]. Every year, over 3.5 million [[b2]] come to enjoy the scenery. There are also several miles of [[b3]] that wind through the area.",
        "clean": "The place had both mountains and a forest. Every year, over 3.5 million visitors come to enjoy the scenery. There are also several miles of trails that wind through the area.",
        "pool": ["visitors", "forest", "desert", "kilometers", "travelers", "trails"],
        "blanks": [
            {"id": "b1", "correctAnswer": "forest"},
            {"id": "b2", "correctAnswer": "visitors"},
            {"id": "b3", "correctAnswer": "trails"},
        ],
        "points": 3,
    },
    {
        "id": "11001116",
        "title": "Recycling Notice",
        "text": "Today, many companies emphasize [[b1]] in their operations. For example, packaging should be made from recyclable materials to ensure proper [[b2]]. In households, using energy-efficient kitchen [[b3]] helps reduce environmental impact.",
        "clean": "Today, many companies emphasize sustainability in their operations. For example, packaging should be made from recyclable materials to ensure proper disposal. In households, using energy-efficient kitchen appliances helps reduce environmental impact.",
        "pool": ["sustainability", "maintenance", "gauge", "appliances", "bargain", "disposal"],
        "blanks": [
            {"id": "b1", "correctAnswer": "sustainability"},
            {"id": "b2", "correctAnswer": "disposal"},
            {"id": "b3", "correctAnswer": "appliances"},
        ],
        "points": 3,
    },
    {
        "id": "11001115",
        "title": "Asking for Leave",
        "text": "My son is [[b1]] to attend school. On the [[b2]] night I found he was [[b3]] and now I think it is better for him not to go to school so that we can avoid infection. I believe after some rest, he will become [[b4]] and will successfully pass his exam. And I am [[b5]] to help him catch up on the missed lessons.",
        "clean": "My son is unable to attend school. On the previous night I found he was sick and now I think it is better for him not to go to school so that we can avoid infection. I believe after some rest, he will become well and will successfully pass his exam. And I am happy to help him catch up on the missed lessons.",
        "pool": ["unable", "difficult", "sick", "well", "painful", "sorry", "previous", "happy"],
        "blanks": [
            {"id": "b1", "correctAnswer": "unable"},
            {"id": "b2", "correctAnswer": "previous"},
            {"id": "b3", "correctAnswer": "sick"},
            {"id": "b4", "correctAnswer": "well"},
            {"id": "b5", "correctAnswer": "happy"},
        ],
        "points": 5,
    },
    {
        "id": "11001114",
        "title": "Lifelong Learning",
        "text": "My grandmother may have an old body, but her [[b1]] hasn't aged a day. She says it's because of all the hobbies she still does, such as reading, baking, and gardening. She knows that learning should never stop. Every time you pick up a hobby, you learn [[b2]], new words, new techniques. Just like your body needs [[b3]] to stay fit, your brain needs learning to stay sharp!",
        "clean": "My grandmother may have an old body, but her mind hasn't aged a day. She says it's because of all the hobbies she still does, such as reading, baking, and gardening. She knows that learning should never stop. Every time you pick up a hobby, you learn new skills, new words, new techniques. Just like your body needs exercise to stay fit, your brain needs learning to stay sharp!",
        "pool": ["new skills", "mind", "Time", "exercise", "thoughts", "items"],
        "blanks": [
            {"id": "b1", "correctAnswer": "mind"},
            {"id": "b2", "correctAnswer": "new skills"},
            {"id": "b3", "correctAnswer": "exercise"},
        ],
        "points": 3,
    },
    {
        "id": "11001113",
        "title": "Work Preparation",
        "text": "It's important to arrive to work early. If your [[b1]] time is 8:30 a.m., you need to be at work by then. Depending on where you work, you might have to [[b2]] your clothes, turn your computer on, or do some other tasks before you're ready to work. You should [[b3]] all of these activities by the time your workday begins.",
        "clean": "It's important to arrive to work early. If your start time is 8:30 a.m., you need to be at work by then. Depending on where you work, you might have to change your clothes, turn your computer on, or do some other tasks before you're ready to work. You should complete all of these activities by the time your workday begins.",
        "pool": ["end", "complete", "explain", "start", "change", "wash"],
        "blanks": [
            {"id": "b1", "correctAnswer": "start"},
            {"id": "b2", "correctAnswer": "change"},
            {"id": "b3", "correctAnswer": "complete"},
        ],
        "points": 3,
    },
]

# ==========================================
# 1.5 Reading MCQ Single (5)
# ==========================================
RMCSA_DATA = [
    {
        "id": "8000287",
        "title": "Sports Readiness",
        "prompt": "Read the text and answer the multiple-choice question by selecting the correct response. What should parents consider when introducing organized sports to their children?",
    },
    {
        "id": "8000286",
        "title": "Knowledge",
        "prompt": "Read the text and answer the multiple-choice question by selecting the correct response. According to the author, what is the primary role of sources in academic writing?",
    },
    {
        "id": "8000285",
        "title": "Persistence",
        "prompt": "Read the text and answer the multiple-choice question by selecting the correct response. Who participated as the subjects in the university study described in the passage?",
    },
    {
        "id": "8000284",
        "title": "Human Development Index",
        "prompt": "Read the text and answer the multiple-choice question by selecting the correct response. Why is the UN Human Development Index considered a more rounded indicator than per capita GDP?",
    },
    {
        "id": "8000283",
        "title": "Glaciers",
        "prompt": "Read the text and answer the multiple-choice question by selecting the correct response. At what intervals do surge glaciers move forward up to 100 times faster than usual?",
    },
]

# ==========================================
# 3.2 Repeat Sentence Sentences
# ==========================================
RS_SENTENCES = {
    "2003378": "Your application for accommodation must be submitted before the course starts.",
    "2003377": "My house is not very far from the hotel.",
    "2003376": "The library is open every day except Sundays and public holidays.",
    "2003375": "The train will not stop at Southeast station due to ongoing construction.",
    "2003374": "The supermarket is located near the theater.",
}

# ==========================================
# 3.5 Answer Short Question Answers
# ==========================================
ASQ_ANSWERS = {
    "5001999": "photographs / photos",
    "5001998": "writer / author",
    "5001997": "cruise",
    "5001996": "diaper / nappy",
    "5001995": "customer",
}

# ==========================================
# 4.3 Listening FIB Passages & Blanks
# ==========================================
LFIB_DATA = [
    {
        "id": "15000432",
        "title": "Work in Organization",
        "text": "In organizations where [[b1]] is encouraged, the exchange of knowledge, much of which is gained through environment, including the arrangement of shared spaces in which these experience, often takes place outside formal meetings. Employees, by [[b2]] ideas with colleagues from other departments, can develop a better understanding of problems that, on their own, they might find difficult to solve. The quality of the work [[b3]], including the arrangement of shared spaces in which these exchanges occur, is therefore important. Its effect on [[b4]], a measure of output in relation to the resources used, depends, however, on whether people can concentrate as well as communicate. Although frequent interaction can help groups produce more effective solutions, interruptions, particularly for those engaged in tasks requiring sustained attention, may reduce the quality of their work. The challenge for managers, then, is to optimize the use of space, so that opportunities for [[b5]] do not come at the expense of individual concentration.",
        "transcript": "In organizations where teamwork is encouraged, the exchange of knowledge, much of which is gained through environment, including the arrangement of shared spaces in which these experience, often takes place outside formal meetings. Employees, by swapping ideas with colleagues from other departments, can develop a better understanding of problems that, on their own, they might find difficult to solve. The quality of the work environment, including the arrangement of shared spaces in which these exchanges occur, is therefore important. Its effect on productivity, a measure of output in relation to the resources used, depends, however, on whether people can concentrate as well as communicate. Although frequent interaction can help groups produce more effective solutions, interruptions, particularly for those engaged in tasks requiring sustained attention, may reduce the quality of their work. The challenge for managers, then, is to optimize the use of space, so that opportunities for collaboration do not come at the expense of individual concentration.",
        "blanks": [
            {"id": "b1", "correctAnswer": "teamwork"},
            {"id": "b2", "correctAnswer": "swapping"},
            {"id": "b3", "correctAnswer": "environment"},
            {"id": "b4", "correctAnswer": "productivity"},
            {"id": "b5", "correctAnswer": "collaboration"},
        ],
        "points": 5,
    },
    {
        "id": "15000431",
        "title": "Sugar and Health",
        "text": "In Sri Lanka, as elsewhere, excess [[b1]], salt and fat in food and beverages are the primary culprits for the rise in the number of people that are overweight or [[b2]]. By ensuring clear and accurate nutritional information is provided on food and drink packaging, and by giving each product a green, orange or red label according to its sugar content, Sri Lanka's government is helping [[b3]] understand risks and make healthier choices.",
        "transcript": "In Sri Lanka, as elsewhere, excess sugar, salt and fat in food and beverages are the primary culprits for the rise in the number of people that are overweight or obese. By ensuring clear and accurate nutritional information is provided on food and drink packaging, and by giving each product a green, orange or red label according to its sugar content, Sri Lanka's government is helping consumers understand risks and make healthier choices.",
        "blanks": [
            {"id": "b1", "correctAnswer": "sugar"},
            {"id": "b2", "correctAnswer": "obese"},
            {"id": "b3", "correctAnswer": "consumers"},
        ],
        "points": 3,
    },
    {
        "id": "15000430",
        "title": "Building Responsibility",
        "text": "At school, the students [[b1]] that responsibilities can help build [[b2]] communities. The teacher reminded them that integrity means keeping [[b3]] even in small matters. One day, they made a discovery of ancient words in a library book, which [[b4]] the whole class. They realized that honest actions together with shared [[b5]] can shape a brighter future.",
        "transcript": "At school, the students understood that responsibilities can help build strong communities. The teacher reminded them that integrity means keeping promises even in small matters. One day, they made a discovery of ancient words in a library book, which inspired the whole class. They realized that honest actions together with shared responsibilities can shape a brighter future.",
        "blanks": [
            {"id": "b1", "correctAnswer": "understood"},
            {"id": "b2", "correctAnswer": "strong"},
            {"id": "b3", "correctAnswer": "promises"},
            {"id": "b4", "correctAnswer": "inspired"},
            {"id": "b5", "correctAnswer": "responsibilities"},
        ],
        "points": 5,
    },
    {
        "id": "15000429",
        "title": "Javelin Competition",
        "text": "The [[b1]] at the sports festival was the javelin competition. Athletes from all over the country came to participate. Many [[b2]] were excited to show their skills in [[b3]] the javelin as far as possible. They were organized into different [[b4]] to compete against each other. The biggest team, which had the most experienced members, took the lead. They [[b5]] the competition with an impressive throw.",
        "transcript": "The biggest event at the sports festival was the javelin competition. Athletes from all over the country came to participate. Many people were excited to show their skills in throwing the javelin as far as possible. They were organized into different teams to compete against each other. The biggest team, which had the most experienced members, took the lead. They started the competition with an impressive throw.",
        "blanks": [
            {"id": "b1", "correctAnswer": "biggest event"},
            {"id": "b2", "correctAnswer": "people"},
            {"id": "b3", "correctAnswer": "throwing"},
            {"id": "b4", "correctAnswer": "teams"},
            {"id": "b5", "correctAnswer": "started"},
        ],
        "points": 5,
    },
    {
        "id": "15000428",
        "title": "Local Farmers",
        "text": "Local farmers [[b1]] experiment with new methods to grow their crops more efficiently. They want to improve their harvests each season. Different [[b2]] of farming, such as organic or vertical farming, are becoming popular. Many [[b3]] now offer subscription services to deliver their produce directly to customers. They pack [[b4]] vegetables into a [[b5]] for families to enjoy weekly. This ensures people have access to healthy food.",
        "transcript": "Local farmers frequently experiment with new methods to grow their crops more efficiently. They want to improve their harvests each season. Different styles of farming, such as organic or vertical farming, are becoming popular. Many farms now offer subscription services to deliver their produce directly to customers. They pack fresh vegetables into a box for families to enjoy weekly. This ensures people have access to healthy food.",
        "blanks": [
            {"id": "b1", "correctAnswer": "frequently"},
            {"id": "b2", "correctAnswer": "styles"},
            {"id": "b3", "correctAnswer": "farms"},
            {"id": "b4", "correctAnswer": "fresh"},
            {"id": "b5", "correctAnswer": "box"},
        ],
        "points": 5,
    },
]

# ==========================================
# 4.7 Highlight Incorrect Words (5)
# ==========================================
HIW_DATA = [
    {
        "id": "16000527",
        "title": "Einstein",
        "displayed_text": "For thousands of years, philosophers and astronomers and thinkers of all sorts have imagined that the university, the space around us, was rather like this floor in front of us. It was fixed and unchangeable, and things happen on it, just as people walk around. So the stars, the comments, and the planets, and the other heavenly bodies moved around and traced down their parts on this completely unchanging stage of space. In the 20th century, as the result of Einstein's work, that view of the universe was completely transformed. We began to understand that there was no absolutely fixed phase of space at all on which all celestial notions were played out. But in some sense, on the larger scale in the universe, the space itself was in this state of a continuous dynasty change. That was a prediction made by Einstein, but it wasn't Einstein who held the honor of making the discovery that the universe was really like that.",
        "spoken_text": "For thousands of years, philosophers and astronomers and thinkers of all sorts have imagined that the universe, the space around us, was rather like this floor in front of us. It was fixed and unchangeable, and things happen on it, just as people walk around. So the stars, the comets, and the planets, and the other heavenly bodies moved around and traced down their parts on this completely unchanging stage of space. In the 20th century, as the result of Einstein's work, that view of the universe was completely transformed. We began to understand that there was no absolutely fixed stage of space at all on which all celestial notions were played out. But in some sense, on the larger scale in the universe, the space itself was in this state of a continuous dynamic change. That was a prediction made by Einstein, but it wasn't Einstein who held the honor of making the discovery that the universe was really like that.",
        "incorrect_words": ["university", "comments", "phase", "dynasty"],
        "spoken_words": ["universe", "comets", "stage", "dynamic"],
    },
    {
        "id": "16000526",
        "title": "Quantum Mechanics",
        "displayed_text": "You know that the incredible thing about quantum mechanics were many incredible things, but one of them, at least to me, is that it really works. It is the science of the universe. It describes what's going on here on Earth, that it describes what happens in the universe that Andre's studies. And so when we think about how those building blocks that Leon was just discussing fit together to make the world around us, we use quantum mechanics to describe atoms, nuclei, molecules, light. It is the theory of the world. And so it's not surprising that we see it in many everyday happenings and many everyday applications. An example might be the lasers that you see in a compact disc machine or the supermarket scanner.",
        "spoken_text": "You know that the incredible thing about quantum mechanics were many incredible things, but one of them, at least to me, is that it really works. It is the theory of the universe. It describes what's going on here on Earth, that it describes what happens in the universe that Andre's studies. And so when we think about how those building blocks that Leon was just describing fit together to make the world around us, we use quantum mechanics to describe atoms, nuclei, molecules, light. It is the theory of the world. And so it's not surprising that we see it in many everyday occurrences and many everyday applications. An example might be the lasers that you see in a compact disc player or the supermarket scanner.",
        "incorrect_words": ["science", "discussing", "happenings", "machine"],
        "spoken_words": ["theory", "describing", "occurrences", "player"],
    },
    {
        "id": "16000525",
        "title": "Sacred Calling",
        "displayed_text": "Now, a professor is also a member of a profession, presumably one that regards its discretion as a vocation in the full sense of the word - from vocare, 'to call'. Despite some modern uses that reduce it to the mere practicalities of motivation practice rather than theory, the product of training rather than education-vocation is more than this. A calling is both active and passive. It implies a sense of intention. It carries with it a guilty to be vocal.",
        "spoken_text": "Now, a professor is also a member of a profession, presumably one that regards its discipline as a vocation in the full sense of the word - from vocare, 'to call'. Despite some modern uses that reduce it to the mere practicalities of employment practice rather than theory, the product of training rather than education-vocation is more than this. A calling is both active and passive. It implies a sense of purpose. It carries with it a duty to be vocal.",
        "incorrect_words": ["discretion", "motivation", "intention", "guilty"],
        "spoken_words": ["discipline", "employment", "purpose", "duty"],
    },
    {
        "id": "16000524",
        "title": "Understanding Color",
        "displayed_text": "Colour theory is the study of colour and its place in art. Humans have been thinking about colours for thousands of years but modern colour theory really arose in the 1800s, when it began to move from science into a pure art. A knowledge of colour theory does require some understanding of basic scientific principles about colour but much of modern colour theory is about the way people perceive, think about and interact with colours - from those used on walls to those selected for a company logo. Colour theory is not only something you see applied to paintings; you can also see it in graphic design, photography, fashion, animation and even video games.",
        "spoken_text": "Colour theory is the study of colour and its position in art. Humans have been thinking about colours for thousands of years but modern colour theory really arose in the 1800s, when it began to move from science into a pure art. A knowledge of colour theory does require some understanding of simple scientific principles about colour but much of modern colour theory is about the way people see, think about and interact with colours - from those used on walls to those chosen for a company logo. Colour theory is not only something you see applied to art; you can also see it in graphic design, photography, fashion, animation and even video games.",
        "incorrect_words": ["place", "basic", "perceive", "selected", "paintings"],
        "spoken_words": ["position", "simple", "see", "chosen", "art"],
    },
    {
        "id": "16000523",
        "title": "Height",
        "displayed_text": "Height is correlated with a lot of things. Up to a certain height, taller people make more money than the vertically challenged. And the taller professional candidate almost always wins. Now a study finds that your height as an adult has a profound effect on your perception of your health. Short people judge their health to be worse than medium or tall people judge theirs. The research was published in the journal Medical Endocrinology. Data for the study came from the 2003 Health Survey for England. More than 14,000 participants filled out answers and had their heights measured. The study only looked at how good the subject thought his or her health was, not their actual health. Questions focused on five areas: mobility, self-care, normal activities, pain or discomfort and anxiety or depression. Men shorter than about 5'4\" and women shorter than 5' reported the worst impressions.",
        "spoken_text": "Height is correlated with a lot of things. Up to a certain height, taller people make more money than the vertically challenged. And the taller presidential candidate almost always wins. Now a study finds that your height as an adult has a profound effect on your perception of your health. Short people judge their health to be worse than average or tall people judge theirs. The research was published in the journal Clinical Endocrinology. Data for the study came from the 2003 Health Survey for England. More than 14,000 participants filled out questionnaires and had their heights measured. The study only looked at how good the subject thought his or her health was, not their actual health. Questions focused on five areas: mobility, self-care, normal activities, pain or discomfort and anxiety or depression. Men shorter than about 5'4\" and women shorter than 5' reported the worst impressions.",
        "incorrect_words": ["professional", "medium", "Medical", "answers"],
        "spoken_words": ["presidential", "average", "Clinical", "questionnaires"],
    },
]

def generate_all():
    passages_sql = []
    groups_sql = []
    reorder_sql = []
    questions_sql = []

    # 1.1 RWFIB
    for item in RWFIB_DATA:
        qid = f"pte-alfa-rwfib-{item['id']}"
        pid = f"rp-{qid}"
        gid = f"g-{qid}"
        paras = [{"label": "A", "text": item["clean"]}]
        passages_sql.append(
            f"('{pid}', '{VERSION}', '{sql_escape(item['title'])}', 'PTE Reading & Writing Fill in Blanks', "
            f"'{sql_escape(json.dumps(paras, ensure_ascii=False))}'::jsonb, {len(item['clean'].split())}, "
            f"'medium', 'Academic Reading', ARRAY['PTE', 'Alfa PTE', 'Reading'], 'custom', ARRAY['academic'])"
        )
        groups_sql.append(
            f"('{gid}', '{pid}', 1, 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)', "
            f"'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.', "
            f"'[]'::jsonb, 'hidden', FALSE, 150)"
        )
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'reading', 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)', "
            f"'{sql_escape(item['title'])}', 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.', "
            f"'{sql_escape(item['text'])}', NULL, NULL, NULL, NULL, NULL, '[]'::jsonb, "
            f"'{sql_escape(json.dumps(item['blanks'], ensure_ascii=False))}'::jsonb, NULL, "
            f"'Dropdown fill in the blanks.', 'medium', {item['points']}, ARRAY['PTE', 'Alfa PTE', 'Reading']::text[], TRUE, 0, 150, "
            f"'{pid}', '{gid}', 1, NULL)"
        )

    # 1.2 RMCMA
    for item in RMCMA_DATA:
        raw = BY_ID[item["id"]]
        qid = f"pte-alfa-rmcma-{item['id']}"
        pid = f"rp-{qid}"
        gid = f"g-{qid}"
        clean_text = raw["prompt_passage"]
        paras = [{"label": "A", "text": clean_text}]
        passages_sql.append(
            f"('{pid}', '{VERSION}', '{sql_escape(item['title'])}', 'PTE Multiple Choice (Multiple)', "
            f"'{sql_escape(json.dumps(paras, ensure_ascii=False))}'::jsonb, {len(clean_text.split())}, "
            f"'medium', 'Academic Reading', ARRAY['PTE', 'Alfa PTE', 'Reading'], 'custom', ARRAY['academic'])"
        )
        groups_sql.append(
            f"('{gid}', '{pid}', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers', "
            f"'{sql_escape(item['prompt'])}', '[]'::jsonb, 'full', FALSE, 150)"
        )
        # Parse options
        opts_raw = raw["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E", "F", "G"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'reading', 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers', "
            f"'{sql_escape(item['title'])}', '{sql_escape(item['prompt'])}', "
            f"NULL, NULL, NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Select all correct options.', 'medium', {len(correct)}, ARRAY['PTE', 'Alfa PTE', 'Reading']::text[], TRUE, 0, 150, "
            f"'{pid}', '{gid}', 1, NULL)"
        )

    # 1.3 RO
    for item in RO_DATA:
        qid = f"pte-alfa-ro-{item['id']}"
        ri_id = f"ri-{qid}"
        reorder_sql.append(
            f"('{ri_id}', '{VERSION}', 'PTE', '{sql_escape(item['title'])}', "
            f"'{sql_escape(json.dumps(item['paragraphs'], ensure_ascii=False))}'::jsonb, "
            f"{sum(len(p['text'].split()) for p in item['paragraphs'])}, 'medium', ARRAY['PTE', 'Alfa PTE', 'Reorder']::text[], TRUE)"
        )
        # options are the paragraphs with id and text
        opts = [{"id": p["label"], "text": p["text"]} for p in item["paragraphs"]]
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'reading', 'reorder-paragraphs', 'Re-order Paragraphs', "
            f"'{sql_escape(item['title'])}', 'The text boxes below have been placed in a random order. Restore the original order.', "
            f"NULL, NULL, NULL, NULL, NULL, '{sql_escape(json.dumps(opts, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(item['correct_order'], ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Restore the original logical order of paragraphs.', 'medium', {item['points']}, ARRAY['PTE', 'Alfa PTE', 'Reading']::text[], TRUE, 0, 150, "
            f"NULL, NULL, NULL, '{ri_id}')"
        )

    # 1.4 RFIB
    for item in RFIB_DATA:
        qid = f"pte-alfa-rfib-{item['id']}"
        pid = f"rp-{qid}"
        gid = f"g-{qid}"
        paras = [{"label": "A", "text": item["clean"]}]
        passages_sql.append(
            f"('{pid}', '{VERSION}', '{sql_escape(item['title'])}', 'PTE Reading Fill in Blanks (Drag & Drop)', "
            f"'{sql_escape(json.dumps(paras, ensure_ascii=False))}'::jsonb, {len(item['clean'].split())}, "
            f"'medium', 'Academic Reading', ARRAY['PTE', 'Alfa PTE', 'Reading'], 'custom', ARRAY['academic'])"
        )
        groups_sql.append(
            f"('{gid}', '{pid}', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks', "
            f"'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.', "
            f"'[]'::jsonb, 'hidden', FALSE, 120)"
        )
        letters = ["A", "B", "C", "D", "E", "F", "G", "H"]
        opts = [{"id": letters[idx], "text": word} for idx, word in enumerate(item["pool"])]
        blanks = []
        for b in item["blanks"]:
            blanks.append({"id": b["id"], "options": item["pool"], "correctAnswer": b["correctAnswer"]})
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'reading', 'fill-in-blanks-r', 'Reading: Fill in the Blanks', "
            f"'{sql_escape(item['title'])}', 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.', "
            f"'{sql_escape(item['text'])}', NULL, NULL, NULL, NULL, '{sql_escape(json.dumps(opts, ensure_ascii=False))}'::jsonb, '[]'::jsonb, "
            f"'{sql_escape(json.dumps(blanks, ensure_ascii=False))}'::jsonb, NULL, "
            f"'Drag words into appropriate blanks.', 'medium', {item['points']}, ARRAY['PTE', 'Alfa PTE', 'Reading']::text[], TRUE, 0, 120, "
            f"'{pid}', '{gid}', 1, NULL)"
        )

    # 1.5 RMCSA
    for item in RMCSA_DATA:
        raw = BY_ID[item["id"]]
        qid = f"pte-alfa-rmcsa-{item['id']}"
        pid = f"rp-{qid}"
        gid = f"g-{qid}"
        clean_text = raw["prompt_passage"]
        paras = [{"label": "A", "text": clean_text}]
        passages_sql.append(
            f"('{pid}', '{VERSION}', '{sql_escape(item['title'])}', 'PTE Multiple Choice (Single)', "
            f"'{sql_escape(json.dumps(paras, ensure_ascii=False))}'::jsonb, {len(clean_text.split())}, "
            f"'medium', 'Academic Reading', ARRAY['PTE', 'Alfa PTE', 'Reading'], 'custom', ARRAY['academic'])"
        )
        groups_sql.append(
            f"('{gid}', '{pid}', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer', "
            f"'{sql_escape(item['prompt'])}', '[]'::jsonb, 'full', FALSE, 90)"
        )
        opts_raw = raw["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'reading', 'reading-mcq-single', 'Multiple Choice, Single Answer', "
            f"'{sql_escape(item['title'])}', '{sql_escape(item['prompt'])}', "
            f"NULL, NULL, NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Select the single best answer.', 'medium', 1, ARRAY['PTE', 'Alfa PTE', 'Reading']::text[], TRUE, 0, 90, "
            f"'{pid}', '{gid}', 1, NULL)"
        )

    # 2.1 SWT
    for q in [BY_ID[qid] for qid in ["6000485", "6000484", "6000483", "6000482", "6000481"]]:
        qid = f"pte-alfa-swt-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'writing', 'summarize-written-text', 'Summarize Written Text', "
            f"'{sql_escape(q['title'])}', 'Read the passage below and summarize it using one sentence. You have 10 minutes to finish this task. Your response will be judged on the quality of your writing and on how well your response presents the key points in the passage.', "
            f"'{sql_escape(q['prompt_passage'])}', NULL, NULL, NULL, NULL, NULL, NULL, "
            f"NULL, NULL, 'Write a single-sentence summary of 5-75 words covering key ideas.', 'medium', 10, ARRAY['PTE', 'Alfa PTE', 'Writing']::text[], TRUE, 0, 600, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 2.2 Write Essay
    for q in [BY_ID[qid] for qid in ["7000503", "7000502", "7000501", "7000500", "7000499"]]:
        qid = f"pte-alfa-we-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'writing', 'write-essay', 'Write Essay', "
            f"'{sql_escape(q['title'])}', 'You will have 20 minutes to plan, write and revise an essay about the topic below. Your response will be judged on how well you develop a position, organize your ideas, present supporting details, and control the elements of standard written English. You should write 200-300 words.\\n\\n{sql_escape(q['prompt_passage'])}', "
            f"'{sql_escape(q['prompt_passage'])}', NULL, NULL, NULL, NULL, NULL, NULL, "
            f"NULL, NULL, 'Write a well-structured essay of 200-300 words with clear introduction, body paragraphs, and conclusion.', 'medium', 15, ARRAY['PTE', 'Alfa PTE', 'Writing']::text[], TRUE, 0, 1200, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 3.1 Read Aloud
    for q in [BY_ID[qid] for qid in ["1001648", "1001647", "1001646", "1001645", "1001644"]]:
        qid = f"pte-alfa-ra-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'speaking', 'read-aloud', 'Read Aloud', "
            f"'{sql_escape(q['title'])}', 'Look at the text below. In 35 seconds, you must read this text aloud as naturally and clearly as possible. You have 40 seconds to read aloud.', "
            f"'{sql_escape(q['prompt_passage'])}', '{sql_escape(q['audio_url'])}', '{sql_escape(q['prompt_passage'])}', NULL, NULL, NULL, NULL, "
            f"NULL, NULL, 'Read clearly and smoothly at a natural pace with correct pronunciation and intonation.', 'medium', 15, ARRAY['PTE', 'Alfa PTE', 'Speaking']::text[], TRUE, 35, 40, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 3.2 Repeat Sentence
    for qid_num in ["2003378", "2003377", "2003376", "2003375", "2003374"]:
        q = BY_ID[qid_num]
        sentence = RS_SENTENCES[qid_num]
        qid = f"pte-alfa-rs-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'speaking', 'pte-repeat-sentence', 'Repeat Sentence', "
            f"'{sql_escape(q['title'])}', 'You will hear a sentence. Please repeat the sentence exactly as you hear it. You will hear the sentence only once.', "
            f"'{sql_escape(sentence)}', '{sql_escape(q['audio_url'])}', '{sql_escape(sentence)}', NULL, NULL, NULL, NULL, "
            f"NULL, '{sql_escape(sentence)}', 'Repeat the sentence word for word with proper fluency and pronunciation.', 'medium', 13, ARRAY['PTE', 'Alfa PTE', 'Speaking']::text[], TRUE, 3, 15, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 3.3 Describe Image
    for q in [BY_ID[qid] for qid in ["3001446", "3001445", "3001444", "3001443", "3001442"]]:
        qid = f"pte-alfa-di-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'speaking', 'pte-describe-image', 'Describe Image', "
            f"'{sql_escape(q['title'])}', 'Look at the image below. In 25 seconds, please speak into the microphone and describe in detail what the image is showing. You will have 40 seconds to give your response.', "
            f"NULL, NULL, NULL, '{sql_escape(q['image_url'])}', NULL, NULL, NULL, "
            f"NULL, NULL, 'Describe the key trends, data points, axes, and conclude with an overall summary.', 'medium', 15, ARRAY['PTE', 'Alfa PTE', 'Speaking']::text[], TRUE, 25, 40, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 3.4 Re-tell Lecture
    for q in [BY_ID[qid] for qid in ["4000604", "4000603", "4000602", "4000601", "4000599"]]:
        qid = f"pte-alfa-rl-{q['id']}"
        audio_val = sql_str_or_null(q["audio_url"])
        image_val = sql_str_or_null(q["image_url"])
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'speaking', 'pte-retell-lecture', 'Re-tell Lecture', "
            f"'{sql_escape(q['title'])}', 'You will hear a lecture. After listening to the lecture, in 10 seconds, please speak into the microphone and retell what you have just heard from the lecture in your own words. You will have 40 seconds to give your response.', "
            f"NULL, {audio_val}, NULL, {image_val}, NULL, NULL, NULL, "
            f"NULL, NULL, 'Retell the main points and supporting ideas from the spoken lecture.', 'medium', 15, ARRAY['PTE', 'Alfa PTE', 'Speaking']::text[], TRUE, 10, 40, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 3.5 Answer Short Question
    for qid_num in ["5001999", "5001998", "5001997", "5001996", "5001995"]:
        q = BY_ID[qid_num]
        ans = ASQ_ANSWERS[qid_num]
        qid = f"pte-alfa-asq-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'speaking', 'pte-answer-short-question', 'Answer Short Questions', "
            f"'{sql_escape(q['title'])}', 'You will hear a question. Please give a simple and short answer. Often just one or a few words is enough.', "
            f"NULL, '{sql_escape(q['audio_url'])}', NULL, NULL, NULL, NULL, NULL, "
            f"NULL, '{sql_escape(ans)}', 'Answer with one or a few words.', 'medium', 1, ARRAY['PTE', 'Alfa PTE', 'Speaking']::text[], TRUE, 3, 10, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.1 Summarize Spoken Text
    for q in [BY_ID[qid] for qid in ["13000701", "13000700", "13000699", "13000698", "13000697"]]:
        qid = f"pte-alfa-sst-{q['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text', "
            f"'{sql_escape(q['title'])}', 'You will hear a short lecture. Write a summary for a fellow student who was not present at the lecture. You should write 50-70 words. You have 10 minutes to finish this task. Your response will be judged on the quality of your writing and on how well your response presents the key points in the lecture.', "
            f"NULL, '{sql_escape(q['audio_url'])}', '{sql_escape(q['model_answer'])}', NULL, NULL, NULL, NULL, "
            f"NULL, '{sql_escape(q['model_answer'])}', 'Write a 50-70 word summary capturing key ideas of the spoken recording.', 'medium', 10, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 0, 600, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.2 Listening MCQ Multiple
    for q in [BY_ID[qid] for qid in ["14000228", "14000225", "14000223", "14000222", "14000221"]]:
        qid = f"pte-alfa-lmcma-{q['id']}"
        opts_raw = q["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-listening-mcma', 'MCQ Multiple Answer', "
            f"'{sql_escape(q['title'])}', '{sql_escape(q['prompt_passage'])}', "
            f"NULL, '{sql_escape(q['audio_url'])}', NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Listen to the audio and select all correct options.', 'medium', {len(correct)}, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 7, 90, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.3 Listening FIB
    for item in LFIB_DATA:
        q = BY_ID[item["id"]]
        qid = f"pte-alfa-lfib-{item['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-listening-fib', 'Listening Fill in the Blanks', "
            f"'{sql_escape(item['title'])}', 'You will hear a recording. Type the missing words in each blank.', "
            f"'{sql_escape(item['text'])}', '{sql_escape(q['audio_url'])}', '{sql_escape(item['transcript'])}', NULL, NULL, NULL, NULL, "
            f"'{sql_escape(json.dumps(item['blanks'], ensure_ascii=False))}'::jsonb, NULL, 'Type the exact words heard in the audio.', 'medium', {item['points']}, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 7, 120, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.4 Highlight Correct Summary
    for q in [BY_ID[qid] for qid in ["19000162", "19000160", "19000159", "19000158", "19000157"]]:
        qid = f"pte-alfa-hcs-{q['id']}"
        opts_raw = q["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-highlight-correct-summary', 'Highlight Correct Summary', "
            f"'{sql_escape(q['title'])}', 'You will hear a recording. Click on the paragraph that best relates to the recording.', "
            f"NULL, '{sql_escape(q['audio_url'])}', NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Click on the paragraph that best summarizes the audio.', 'medium', 1, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 10, 90, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.5 Listening MCQ Single
    for q in [BY_ID[qid] for qid in ["17000219", "17000218", "17000217", "17000216", "17000215"]]:
        qid = f"pte-alfa-lmcsa-{q['id']}"
        opts_raw = q["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-listening-mcsa', 'MCQ Single Answer', "
            f"'{sql_escape(q['title'])}', '{sql_escape(q['prompt_passage'])}', "
            f"NULL, '{sql_escape(q['audio_url'])}', NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Select the single correct answer based on the recording.', 'medium', 1, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 5, 90, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.6 Select Missing Word
    for q in [BY_ID[qid] for qid in ["18000214", "18000213", "18000212", "18000211", "18000210"]]:
        qid = f"pte-alfa-smw-{q['id']}"
        opts_raw = q["options_json"]
        options = []
        correct = []
        letters = ["A", "B", "C", "D", "E"]
        for idx, opt_raw in enumerate(opts_raw):
            opt = normalize_opt(opt_raw)
            lbl = letters[idx]
            options.append({"id": lbl, "text": opt.get("anstext", "").strip()})
            if int(opt.get("anscheck", 0)) == 1:
                correct.append(lbl)
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-select-missing-word', 'Select Missing Word', "
            f"'{sql_escape(q['title'])}', 'You will hear a recording. At the end of the recording the last word or group of words has been replaced by a beep. Select the correct option to complete the recording.', "
            f"NULL, '{sql_escape(q['audio_url'])}', NULL, NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Select the phrase that best completes the audio.', 'medium', 1, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 7, 90, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.7 Highlight Incorrect Word
    for item in HIW_DATA:
        q = BY_ID[item["id"]]
        qid = f"pte-alfa-hiw-{item['id']}"
        # Tokenize displayed text
        tokens = item["displayed_text"].split()
        options = []
        correct = []
        for idx, tok in enumerate(tokens):
            tok_id = f"w{idx+1}"
            options.append({"id": tok_id, "text": tok})
            clean_tok = re.sub(r"[^\w]", "", tok).lower()
            for bad_word in item["incorrect_words"]:
                if clean_tok == bad_word.lower():
                    correct.append(tok_id)
                    break
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'pte-highlight-incorrect-word', 'Highlight Incorrect Word', "
            f"'{sql_escape(item['title'])}', 'You will hear a recording. Below is a transcription of the recording. Some words in the transcription differ from what the speaker said. Please click on the words that are different.', "
            f"'{sql_escape(item['displayed_text'])}', '{sql_escape(q['audio_url'])}', '{sql_escape(item['spoken_text'])}', NULL, NULL, '{sql_escape(json.dumps(options, ensure_ascii=False))}'::jsonb, '{sql_escape(json.dumps(correct, ensure_ascii=False))}'::jsonb, "
            f"'[]'::jsonb, NULL, 'Click on each word in the transcription that differs from the audio.', 'medium', {len(correct)}, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 10, 120, "
            f"NULL, NULL, NULL, NULL)"
        )

    # 4.8 Write From Dictation
    WFD_DATA = [
        {"id": "20003760", "text": "Nursing is defined as a combination of medical practice and clinical work."},
        {"id": "20003759", "text": "The manager will have a meeting in this room today."},
        {"id": "20003757", "text": "The conference will be held on Thursday morning."},
        {"id": "20003756", "text": "Canada has a long history of immigration from many different parts of the world."},
        {"id": "20003754", "text": "We are encouraged to write on each page."},
    ]
    for item in WFD_DATA:
        q = BY_ID[item["id"]]
        qid = f"pte-alfa-wfd-{item['id']}"
        questions_sql.append(
            f"('{qid}', '{VERSION}', 'PTE', ARRAY['PTE']::text[], 'listening', 'write-from-dictation', 'Write from Dictation', "
            f"'{sql_escape(q['title'])}', 'You will hear a sentence. Type the sentence in the box below exactly as you hear it. Write as much of the sentence as you can. You will hear the sentence only once.', "
            f"NULL, '{sql_escape(q['audio_url'])}', '{sql_escape(item['text'])}', NULL, NULL, NULL, NULL, "
            f"NULL, '{sql_escape(item['text'])}', 'Type the sentence in the box below exactly as you hear it.', 'medium', {len(item['text'].split())}, ARRAY['PTE', 'Alfa PTE', 'Listening']::text[], TRUE, 7, 60, "
            f"NULL, NULL, NULL, NULL)"
        )

    print(f"Generated: {len(passages_sql)} passages, {len(groups_sql)} groups, {len(reorder_sql)} reorder items, {len(questions_sql)} questions.")

    # Write UP Migration
    up_content = [
        "-- Migration: 000112_alfa_pte_practice_questions.up.sql",
        "-- Alfa PTE Practice Questions & AI Score Compendium (100 Questions across all 20 PTE Subtasks).",
        "-- 5 Questions per Subtask. Replay-safe via ON CONFLICT (id) DO NOTHING.",
        "",
    ]
    if passages_sql:
        up_content.append("INSERT INTO reading_passages")
        up_content.append("    (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot, modules)")
        up_content.append("VALUES\n    " + ",\n    ".join(passages_sql))
        up_content.append("ON CONFLICT (id) DO NOTHING;\n")

    if groups_sql:
        up_content.append("INSERT INTO reading_question_groups")
        up_content.append("    (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds)")
        up_content.append("VALUES\n    " + ",\n    ".join(groups_sql))
        up_content.append("ON CONFLICT (id) DO NOTHING;\n")

    if reorder_sql:
        up_content.append("INSERT INTO reading_reorder_items")
        up_content.append("    (id, exam_version_id, exam, title, paragraphs, word_count, difficulty, tags, is_published)")
        up_content.append("VALUES\n    " + ",\n    ".join(reorder_sql))
        up_content.append("ON CONFLICT (id) DO NOTHING;\n")

    if questions_sql:
        up_content.append("INSERT INTO questions (")
        up_content.append("    id, exam_version_id, exam, supported_exams, skill, type_id, type_name,")
        up_content.append("    title, prompt, context_passage, audio_url, audio_transcript, image_url, figure_data,")
        up_content.append("    options, correct_answers, blanks, model_answer, explanation, difficulty, points,")
        up_content.append("    tags, is_published, prep_time_seconds, time_limit_seconds, passage_id, group_id, group_position, reorder_item_id")
        up_content.append(") VALUES\n    " + ",\n    ".join(questions_sql))
        up_content.append("ON CONFLICT (id) DO NOTHING;\n")

    with open(UP_MIGRATION, "w", encoding="utf-8") as f:
        f.write("\n".join(up_content))
    print(f"Wrote UP migration: {UP_MIGRATION}")

    # Write DOWN Migration
    down_content = """-- Migration: 000112_alfa_pte_practice_questions.down.sql
DELETE FROM questions WHERE id LIKE 'pte-alfa-%';
DELETE FROM reading_question_groups WHERE id LIKE 'g-pte-alfa-%';
DELETE FROM reading_passages WHERE id LIKE 'rp-pte-alfa-%';
DELETE FROM reading_reorder_items WHERE id LIKE 'ri-pte-alfa-%';
"""
    with open(DOWN_MIGRATION, "w", encoding="utf-8") as f:
        f.write(down_content)
    print(f"Wrote DOWN migration: {DOWN_MIGRATION}")

if __name__ == "__main__":
    generate_all()
