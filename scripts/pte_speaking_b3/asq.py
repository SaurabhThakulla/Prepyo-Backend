"""50 Answer Short Question (ASQ) items for PTE Speaking Batch 3.

Rewritten after review: the first version repeated about fifteen live
questions, used academic vocabulary far above the task (the real task tests
everyday words), and its titles named the answer. These are everyday questions
answered in one to three words, none of them asked by a live question
(checked with short_question_repeats in scripts/live_topics.py). Titles are
category labels numbered on from the live bank's, so the list never shows the
answer before the question is heard.
"""


def _item(title, question, answer, variants=''):
    also = f' {variants}' if variants else ''
    return {
        "title": title,
        "audio_transcript": question,
        "model_answer": answer,
        "explanation": f"Expected answer: {answer}.{also} Any natural equivalent is accepted.",
    }


ASQ_ITEMS = [
    _item("Everyday life 19", "What do you use to cut paper?", "Scissors"),
    _item("Everyday life 20", "What do you put on an envelope before you post a letter?", "A stamp"),
    _item("Everyday life 21", "What machine in a kitchen washes plates and cups?", "A dishwasher"),
    _item("Everyday life 22", "What do you wear on your feet inside your shoes?", "Socks"),
    _item("Everyday life 23", "What do you use to dry yourself after a shower?", "A towel"),
    _item("Everyday life 24", "What clock makes a noise to wake you up in the morning?", "An alarm clock", "Also: an alarm."),
    _item("Everyday life 25", "At a railway station, where do passengers stand to wait for a train?", "The platform"),
    _item("Everyday life 26", "What do teachers usually write with on a whiteboard?", "A marker", "Also: a pen or marker pen."),
    _item("Everyday life 27", "What do you look into to see your own face when you brush your hair?", "A mirror"),
    _item("Geography 15", "What do we call a very large area of land covered with trees?", "A forest"),
    _item("Geography 16", "What do we call the place where a river begins?", "The source"),
    _item("Geography 17", "What do we call the land along the edge of the sea?", "The coast", "Also: the coastline or the shore."),
    _item("Geography 18", "How many continents are there on Earth?", "Seven"),
    _item("Health 13", "Where do you go to buy medicine that a doctor has prescribed?", "A pharmacy", "Also: a chemist."),
    _item("Health 14", "Which part of your mouth do you use to chew food?", "Your teeth"),
    _item("Health 15", "What is the piece of paper a doctor writes so that you can get medicine?", "A prescription"),
    _item("Health 16", "What vehicle takes injured people to hospital in an emergency?", "An ambulance"),
    _item("Health 17", "What do we call a doctor who treats animals?", "A vet", "Also: a veterinarian."),
    _item("Language 08", "What is the opposite of 'arrive'?", "Leave", "Also: depart."),
    _item("Language 09", "What is the plural of 'mouse'?", "Mice"),
    _item("Language 10", "What is the past tense of 'go'?", "Went"),
    _item("Language 11", "What is the opposite of 'narrow'?", "Wide", "Also: broad."),
    _item("Language 12", "What do we call a short form of a word, such as 'Dr' for 'doctor'?", "An abbreviation"),
    _item("Nature 18", "What is a young dog called?", "A puppy"),
    _item("Nature 19", "What do we call small balls of ice that fall from the sky during a storm?", "Hail"),
    _item("Nature 20", "Which animal has a very long neck and eats leaves from tall trees?", "A giraffe"),
    _item("Nature 21", "What sweet liquid do bees collect from flowers?", "Nectar"),
    _item("Nature 22", "What do we call a baby frog before it grows legs?", "A tadpole"),
    _item("People and jobs 13", "Who looks after the books in a library?", "A librarian"),
    _item("People and jobs 14", "Who delivers letters to people's homes?", "A postman", "Also: a postal worker or mail carrier."),
    _item("People and jobs 15", "What do we call a person who makes bread and cakes to sell?", "A baker"),
    _item("People and jobs 16", "Who puts out fires and rescues people from burning buildings?", "Firefighters"),
    _item("People and jobs 17", "Who makes and repairs water pipes in houses?", "A plumber"),
    _item("Science 21", "Which planet is known as the Red Planet?", "Mars"),
    _item("Science 22", "What do we call water when it is heated until it becomes a gas?", "Steam", "Also: water vapour."),
    _item("Science 23", "Which part of a plant is usually green and flat?", "The leaf", "Also: leaves."),
    _item("Science 24", "What do we call a scientist who studies animals?", "A zoologist"),
    _item("Science 25", "What do we use a magnet to pick up: wood or iron?", "Iron"),
    _item("Time and numbers 17", "How many days are there in a week?", "Seven"),
    _item("Time and numbers 18", "What is ten multiplied by ten?", "One hundred", "Also: 100."),
    _item("Time and numbers 19", "How many wheels does a car usually have?", "Four"),
    _item("Time and numbers 20", "What is the last month of the year?", "December"),
    _item("Time and numbers 21", "How many centimetres are there in a metre?", "One hundred", "Also: 100."),
    _item("University life 15", "What do we call the last date by which an assignment must be handed in?", "The deadline"),
    _item("University life 16", "What do we call the list of topics covered in a course?", "The syllabus", "Also: the course outline."),
    _item("University life 17", "What do we call a meeting where one student discusses their work with a tutor?", "A tutorial"),
    _item("University life 18", "What card do students show to get into the university library?", "A student card", "Also: a student ID."),
    _item("Nepal 11", "Which river flows through the city of Kathmandu?", "The Bagmati"),
    _item("Nepal 12", "Which lake in Pokhara is famous for its view of the mountains?", "Phewa Lake"),
    _item("Nepal 13", "In which city is the Pashupatinath temple?", "Kathmandu"),
]

assert len(ASQ_ITEMS) == 50
