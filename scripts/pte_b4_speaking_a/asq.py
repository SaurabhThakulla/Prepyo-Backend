"""50 original Answer Short Question items for PTE Speaking bank 4a.

Everyday questions answered in one to three words, as in the live task. None
asks what a live question or a batch-3 question asks (checked with
short_question_repeats in scripts/live_topics.py, and the same test against
the batch-3 questions in the bank snapshot). Titles are category labels, so the
practice list never shows the answer. Categories carried on from batch 3 keep
their numbering; each title has at most one distinctive word, so that
bank_check does not read two numbered titles of the same category as a repeat.
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
    # Household (everyday objects at home)
    _item("Household 01", "What do you use in the kitchen to boil water for a cup of tea?", "A kettle"),
    _item("Household 02", "What do you rest your head on when you sleep in bed?", "A pillow"),
    _item("Household 03", "What large piece of furniture do you hang your clothes in?", "A wardrobe"),
    _item("Household 04", "What do you use to sweep dust and dirt from a floor?", "A broom", "Also: a brush."),
    _item("Household 05", "What do you plug into the wall to put power back into a mobile phone?", "A charger"),
    _item("Household 06", "What tool do you use to open a tin of beans?", "A tin opener", "Also: a can opener."),
    # Food
    _item("Food 01", "What hot drink is made from roasted and ground beans?", "Coffee"),
    _item("Food 02", "Which fruit is crushed to make wine?", "Grapes"),
    _item("Food 03", "What white powder, made by grinding wheat, is the main ingredient of bread?", "Flour"),
    _item("Food 04", "What do we call the meal eaten in the middle of the day?", "Lunch"),
    # Geography
    _item("Geography 19", "What is the capital city of Australia?", "Canberra"),
    _item("Geography 20", "Which ocean lies between Africa and Australia?", "The Indian Ocean"),
    _item("Geography 21", "Which country has the largest area in the world?", "Russia"),
    _item("Geography 22", "On a map, what are the lines that join places of the same height called?", "Contour lines", "Also: contours."),
    _item("Geography 23", "What do we call a book of maps?", "An atlas"),
    # Health
    _item("Health 18", "In Britain, which doctor do you usually see first when you feel unwell?", "A GP", "Also: a general practitioner, a family doctor."),
    _item("Health 19", "Which joint connects your hand to your arm?", "The wrist"),
    _item("Health 20", "Which vitamin does your skin make when it is in sunlight?", "Vitamin D"),
    _item("Health 21", "What do we call a crack or break in a bone?", "A fracture"),
    _item("Health 22", "What is given by injection to protect people from catching a disease?", "A vaccine", "Also: a vaccination."),
    # Language
    _item("Language 13", "What is the opposite of 'maximum'?", "Minimum"),
    _item("Language 14", "What do we call words that sound the same but have different meanings, like 'sea' and 'see'?", "Homophones"),
    _item("Language 15", "What punctuation marks go around the exact words that someone said?", "Quotation marks", "Also: inverted commas, speech marks."),
    _item("Language 16", "What is the first letter of the English alphabet?", "A"),
    # Nature
    _item("Nature 23", "What do we call a baby sheep?", "A lamb"),
    _item("Nature 24", "What is the home that bees live in called?", "A hive", "Also: a beehive."),
    _item("Nature 25", "Which part of a tree is covered in bark and holds up the branches?", "The trunk"),
    _item("Nature 26", "Which colourful bird can learn to copy human speech?", "A parrot"),
    # Science
    _item("Science 26", "What is the chemical symbol for water?", "H2O"),
    _item("Science 27", "Which is the largest planet in our solar system?", "Jupiter"),
    _item("Science 28", "What instrument do weather forecasters use to measure air pressure?", "A barometer"),
    _item("Science 29", "What is the smallest unit of a chemical element called?", "An atom"),
    # Transport
    _item("Transport 01", "What must you buy before you travel on a train?", "A ticket"),
    _item("Transport 02", "What do you fasten across your body in a car to keep you safe in a crash?", "A seat belt"),
    _item("Transport 03", "What is the long strip of ground where aeroplanes take off and land?", "A runway"),
    _item("Transport 04", "What do cars need to stop at when the traffic light is red?", "The stop line", "Also: the white line, the traffic lights."),
    # People and jobs
    _item("People and jobs 18", "Who tests your eyesight and sells glasses?", "An optician"),
    _item("People and jobs 19", "Who is in charge of a school?", "The head teacher", "Also: the principal, the headmaster, the headmistress."),
    _item("People and jobs 20", "What do we call a person whose job is writing software for computers?", "A programmer", "Also: a software developer."),
    _item("People and jobs 21", "What document do you send to an employer listing your education and work experience?", "A CV", "Also: a curriculum vitae, a resume."),
    # Time and numbers
    _item("Time and numbers 22", "How many sides does a square have?", "Four"),
    _item("Time and numbers 23", "How many days does February have when it is not a leap year?", "Twenty-eight", "Also: 28."),
    _item("Time and numbers 24", "What is a quarter of one hundred?", "Twenty-five", "Also: 25."),
    _item("Time and numbers 25", "How many hours are there in two days?", "Forty-eight", "Also: 48."),
    # University
    _item("University 19", "What do we call a large room with rows of seats where lectures are given?", "A lecture theatre", "Also: a lecture hall."),
    _item("University 20", "What official document lists all the marks a student has received on a course?", "A transcript"),
    _item("University 21", "What is the week of welcome events for new students before term begins called?", "Freshers' week", "Also: orientation week, welcome week."),
    # Sport and arts
    _item("Sport 01", "How many players does each football team have on the pitch at the start of a match?", "Eleven", "Also: 11."),
    _item("Sport 02", "Which sport is played with rackets and a shuttlecock?", "Badminton"),
    _item("Arts 01", "What do we call a large group of musicians who play classical music together?", "An orchestra"),
]
