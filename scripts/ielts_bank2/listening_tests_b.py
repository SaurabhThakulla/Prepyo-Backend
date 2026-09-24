"""IELTS Listening Practice Tests 5-7. Same structure and checks as Test 1.
Original Prepyo practice material; not official or recalled test content.
"""
from ielts_bank2.listening_lectures_v2 import PART4
from ielts_bank2.common import ABC, MAP_OPTIONS

T5_SOURCES = ABC('should definitely be used', 'is too general', 'is out of date', 'may be biased', 'is useful for comparison')
T6_SECTIONS = ABC('checking', 'more recent data', 'shortening', 'starting', 'more supporting evidence')
T7_TEAMS = ABC('animal care', 'reception', 'fundraising')

TESTS_B = [dict(
    key='t5', title='IELTS Listening Practice Test 5',
    parts=[
        dict(
            setting='You will hear a man phoning a car hire company.',
            script=(
                "Agent: Good morning, Coastline Car Hire. How can I help? "
                "Man: Hi, I'd like to hire a car for a week in September. "
                "Agent: Certainly. Where will you be collecting it? We have offices at the airport and in the city centre. "
                "Man: The airport, please, because I'm flying in. "
                "Agent: And the date? "
                "Man: The eleventh of September. I land at about two in the afternoon. "
                "Agent: What kind of car are you looking for? "
                "Man: Something small would be fine, but there'll be four of us with luggage, so maybe a medium one. Actually, let's say an estate car, for the extra space. "
                "Agent: An estate. That's sixty-two pounds a day. "
                "Man: That's fine. "
                "Agent: Would you like to add a second driver? "
                "Man: Yes, my wife will drive too. "
                "Agent: That's an extra five pounds a day. Can I have her name? "
                "Man: Ingrid Hollis. "
                "Agent: Do you need any extras, such as a child seat or satellite navigation? "
                "Man: No child seat, but satellite navigation would be useful. "
                "Agent: It's included free with estate cars. "
                "Man: Oh, good. "
                "Agent: The car will come with a full tank, and you need to return it full. "
                "Man: Understood. "
                "Agent: And for the insurance, you'll need to show your driving licence and a credit card. "
                "Man: No problem."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Coastline Car Hire: booking details',
                instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Collection point: the ________', ['airport'], 'The airport, please'),
                    ('Date: ________ September', ['11', '11th', 'eleventh'], 'The eleventh of September'),
                    ('Arrival time: about ________ p.m.', ['2', 'two'], 'about two in the afternoon'),
                    ('Type of car: ________', ['estate'], "let's say an estate car"),
                    ('Daily rate: £________', ['62', 'sixty-two'], 'sixty-two pounds a day'),
                    ('Second driver: Ingrid ________', ['Hollis'], 'Ingrid Hollis'),
                    ('Extra: satellite ________', ['navigation'], 'satellite navigation would be useful'),
                    ('Cost of this extra: ________', ['free', 'nothing'], 'included free with estate cars'),
                    ('Fuel: return the car ________', ['full'], 'you need to return it full'),
                    ('Bring: driving licence and a ________ card', ['credit'], 'a credit card'),
                ])],
        ),
        dict(
            setting='You will hear a guide welcoming visitors to a botanical garden.',
            script=(
                "Guide: Good morning and welcome to Lindley Botanical Gardens. The gardens were founded in 1846 by a doctor who wanted to grow plants for medicine, "
                "and that collection of medicinal plants is still here today, near the main gate. "
                "Today the gardens cover sixteen hectares and contain over twelve thousand species. Our most famous plant is the giant water lily in the tropical house. "
                "Its leaves can grow up to three metres across, and they're strong enough to support the weight of a small child, although of course we don't allow that. "
                "If you're interested in birds, the best place to see them is the lake at the northern end, where herons often nest. "
                "The gardens rely heavily on volunteers, who do most of the weeding and guide many of the tours. "
                "A few practical points. You're welcome to take photographs anywhere except in the orchid house, because the flash damages the flowers. "
                "The café serves hot food until two o'clock, and after that only sandwiches and cakes. There are free guided walks every day at eleven. "
                "And if you'd like to support us, the best way is to become a Friend of the Gardens, which gives you free entry all year for a small annual fee. "
                "Before you go, do visit the shop, which sells seeds collected here in the gardens."),
            groups=[
                dict(type_id='ielts-listening-completion', heading='Lindley Botanical Gardens',
                     instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                     questions=[
                         ('Founded in 1846 by a ________', ['doctor'], 'by a doctor'),
                         ('Size: ________ hectares', ['16', 'sixteen'], 'cover sixteen hectares'),
                         ('Giant water lily leaves: up to ________ metres across', ['3', 'three'], 'up to three metres across'),
                         ('Best place for birds: the ________', ['lake'], 'is the lake at the northern end'),
                         ('Volunteers do most of the ________', ['weeding'], 'do most of the weeding'),
                     ]),
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('Where are photographs not allowed?', ABC('in the tropical house', 'in the orchid house', 'near the lake'), 'B', 'except in the orchid house'),
                         ('What can visitors get at the café after two o’clock?', ABC('hot food', 'drinks only', 'sandwiches and cakes'), 'C', 'only sandwiches and cakes'),
                         ('When are the free guided walks?', ABC('every day at 11', 'at weekends only', 'every day at 2'), 'A', 'free guided walks every day at eleven'),
                         ('What is the best way to support the gardens?', ABC('make a donation', 'become a Friend', 'volunteer'), 'B', 'become a Friend of the Gardens'),
                         ('What does the shop sell?', ABC('seeds from the gardens', 'books about plants', 'garden tools'), 'A', 'sells seeds collected here'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two geography students, Grace and Oliver, discussing their field trip report.',
            script=(
                "Grace: Oliver, have you started the field trip report yet? "
                "Oliver: I've done the introduction, but I'm stuck on the methods section. "
                "Grace: What's the problem? "
                "Oliver: I can't remember exactly how we measured the beach profile. "
                "Grace: We used ranging poles every five metres from the cliff to the sea. "
                "Oliver: Of course. And the main finding was that the beach had got narrower since the old survey? "
                "Grace: Yes, by about eight metres, which is more than we expected. The sea wall seems to be pushing the erosion further along the coast. "
                "Oliver: That's what the local council said too. I was surprised how honest the council officer was about it. "
                "Grace: Me too. What did you think was the most useful part of the trip? "
                "Oliver: The interview with the fisherman, definitely. He's watched the coast change for forty years. "
                "Grace: I agree, although the museum visit gave us good background. "
                "Oliver: Now, we need to decide which sources to use in the report. What did you think of the government report on coastal defence? "
                "Grace: It's very detailed, but it was published in 1998, so it's rather out of date. "
                "Oliver: True. What about the article by Dr Hayes? "
                "Grace: That's excellent; it's exactly on our topic, so we should definitely use it. "
                "Oliver: And the textbook chapter on erosion? "
                "Grace: Useful for explaining basic processes, but too general for our case. "
                "Oliver: The website of the coastal protection group? "
                "Grace: I'm worried it's one-sided. They campaign against the sea wall, so it's not neutral. "
                "Oliver: And the old maps from the library? "
                "Grace: They show the coastline in 1900, so they're perfect for comparison."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What is Oliver having difficulty with?', ABC('the introduction', 'the methods section', 'the conclusion'), 'B', "I'm stuck on the methods section"),
                         ('How did the students measure the beach profile?', ABC('with poles at regular intervals', 'with a laser', 'from photographs'), 'A', 'ranging poles every five metres'),
                         ('What did the students find about the beach?', ABC('It was wider than before.', 'It was narrower than before.', 'It had not changed.'), 'B', 'by about eight metres'),
                         ('What surprised Oliver about the council officer?', ABC('his knowledge', 'his honesty', 'his age'), 'B', 'how honest the council officer was'),
                         ('Which part of the trip did Oliver find most useful?', ABC('the interview', 'the museum visit', 'the beach survey'), 'A', 'The interview with the fisherman'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Sources for the report',
                     instructions='What does Grace say about each source? Choose the correct letter, A-E.',
                     questions=[
                         ('the government report', T5_SOURCES, 'C', "so it's rather out of date"),
                         ('the article by Dr Hayes', T5_SOURCES, 'A', 'we should definitely use it'),
                         ('the textbook chapter', T5_SOURCES, 'B', 'too general for our case'),
                         ('the campaign group’s website', T5_SOURCES, 'D', "it's one-sided"),
                         ('the old maps', T5_SOURCES, 'E', 'perfect for comparison'),
                     ]),
            ],
        ),
        PART4['t5'],
    ],
), dict(
    key='t6', title='IELTS Listening Practice Test 6',
    parts=[
        dict(
            setting='You will hear a father phoning an arts centre to book a summer camp for his son.',
            script=(
                "Receptionist: Good morning, Kingsway Arts Centre. "
                "Father: Hello, I'm calling about the summer art camp for children. I'd like to book a place for my son. "
                "Receptionist: Of course. How old is he? "
                "Father: He's eight. "
                "Receptionist: Then he'd be in the junior group, which is for ages seven to ten. "
                "Father: Good. What does the camp involve? "
                "Receptionist: Each week has a theme. The first week is painting, the second is sculpture, and the third is animation. "
                "Father: He's mad about cartoons, so the third week, please. "
                "Receptionist: Animation, then. That week runs from the twenty-first to the twenty-fifth of July, from nine until three each day. "
                "Father: What does it cost? "
                "Receptionist: It's a hundred and ten pounds for the week, but there's a ten per cent discount if you book before the end of May. "
                "Father: Great, I'll book now. Does he need to bring anything? "
                "Receptionist: Just a packed lunch and an old shirt to protect his clothes. "
                "Father: And who's the teacher? "
                "Receptionist: The animation week is taught by Anna Friedman. She's worked for a film studio. "
                "Father: Wonderful. "
                "Receptionist: Can I take your son's name? "
                "Father: It's Leo Barrington. "
                "Receptionist: And is there anything we should know about his health? "
                "Father: He has asthma, so he carries an inhaler. "
                "Receptionist: Thanks, I'll note that. Finally, how did you hear about us? "
                "Father: From a poster at his school."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Kingsway Arts Centre: summer camp booking',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Child’s age: ________', ['8', 'eight'], "He's eight"),
                    ('Group: ________', ['junior'], 'the junior group'),
                    ('Week chosen: ________', ['animation'], 'Animation, then'),
                    ('Dates: 21 to ________ July', ['25', '25th', 'twenty-fifth'], 'to the twenty-fifth of July'),
                    ('Price: £________ for the week', ['110'], 'a hundred and ten pounds'),
                    ('Bring: packed lunch and an old ________', ['shirt'], 'an old shirt'),
                    ('Teacher: Anna ________', ['Friedman'], 'Anna Friedman'),
                    ('Child’s surname: ________', ['Barrington'], 'Leo Barrington'),
                    ('Health: has ________', ['asthma'], 'He has asthma'),
                    ('Heard about the camp from: a ________', ['poster'], 'a poster at his school'),
                ])],
        ),
        dict(
            setting='You will hear a librarian giving new students an introduction to the university library.',
            script=(
                "Librarian: Hello, everyone. Welcome to Westmoor University Library. I'm going to give you a quick introduction before you explore on your own. "
                "The library is open twenty-four hours a day during term, but outside term time it closes at nine in the evening. You'll need your student card to get in after six. "
                "You can borrow up to fifteen books at a time. Most books can be kept for three weeks, but books in high demand, which have a red label, can only be borrowed for two days. "
                "If you want a book that's out on loan, you can reserve it online, and we'll email you when it comes back. "
                "Please don't eat anywhere except in the café, although you can bring drinks in bottles with lids. "
                "Many students ask about printing. Printing is paid for with your student card, and it's cheaper in black and white. "
                "Now, have a look at the plan. You come in at the main entrance, at the bottom of the plan, and the help desk is straight ahead on the right. "
                "The group study area is the room on the east side just north of the cross corridor. You can book it for up to three hours. "
                "The printing room is the room in the south-west corner, on your left as you come in. "
                "The silent reading room is the room in the north-east corner, as far as possible from the entrance. "
                "The newspaper archive is the room on the west side just north of the cross corridor, opposite the group study area. "
                "And finally, the café is the room in the south-east corner, so you can get a coffee as soon as you arrive."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('When does the library close outside term time?', ABC('6 p.m.', '9 p.m.', 'It never closes.'), 'B', 'it closes at nine in the evening'),
                         ('How long can books with a red label be borrowed for?', ABC('two days', 'one week', 'three weeks'), 'A', 'can only be borrowed for two days'),
                         ('How will students know that a reserved book has been returned?', ABC('by text message', 'by email', 'by letter'), 'B', "we'll email you when it comes back"),
                         ('What are students allowed to bring into the library?', ABC('hot food', 'drinks with lids', 'snacks'), 'B', 'drinks in bottles with lids'),
                         ('What is true about printing?', ABC('It is free.', 'Colour printing is cheaper.', 'It is paid for with a student card.'), 'C', 'Printing is paid for with your student card'),
                     ]),
                dict(type_id='ielts-listening-map', heading='Westmoor University Library',
                     instructions='Label the plan below. Choose the correct letter, A-F.',
                     image=('Westmoor University Library', {('E', 2): 'Help desk', ('W', 4): 'Stairs'},
                            [('W', 1), ('E', 1), ('W', 2), ('W', 3), ('E', 3), ('E', 4)]),
                     questions=[
                         ('Group study area', MAP_OPTIONS, 'E', 'the room on the east side just north of the cross corridor'),
                         ('Printing room', MAP_OPTIONS, 'A', 'the room in the south-west corner'),
                         ('Silent reading room', MAP_OPTIONS, 'F', 'the room in the north-east corner'),
                         ('Newspaper archive', MAP_OPTIONS, 'D', 'the room on the west side just north of the cross corridor'),
                         ('Café', MAP_OPTIONS, 'B', 'the room in the south-east corner'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two business students, Mia and Jake, discussing their case study with their tutor.',
            script=(
                "Tutor: So, Mia and Jake, how's the case study on Barton's Bakery going? "
                "Mia: Really well. The owner, Mrs Barton, was very helpful. "
                "Jake: We were worried she wouldn't want to share financial information, but she gave us everything. "
                "Tutor: Excellent. What did you find was the main reason for the bakery's success? "
                "Mia: We assumed it was the location, because it's on the high street, but actually it's the loyalty of the customers. Most of them have been going there for years. "
                "Jake: And the quality of the bread, of course. "
                "Tutor: What challenges is the business facing? "
                "Jake: The biggest one is the cost of energy. The ovens run all night, and her bills have doubled. "
                "Mia: Staffing is difficult too, but she said energy is the thing that keeps her awake at night. "
                "Tutor: Did she have plans for the future? "
                "Mia: She's thinking of opening a small café inside the shop, rather than a second shop, which is what we expected. "
                "Tutor: Good. And what are you going to recommend? "
                "Jake: That's where we disagree a bit. I think she should start selling online, with delivery to local offices. "
                "Mia: I'm not sure she has the staff for that. I'd recommend running baking classes in the evenings, when the kitchen isn't being used. "
                "Tutor: Both good ideas. Now let's think about the sections of your report and what each still needs. The introduction? "
                "Mia: It's finished; it just needs checking. "
                "Tutor: The financial analysis? "
                "Jake: We need more up-to-date figures; the ones we have are from two years ago. "
                "Tutor: The interview summary? "
                "Mia: That's too long. We need to cut it by half. "
                "Tutor: The comparison with competitors? "
                "Jake: We haven't started that yet. "
                "Tutor: And the recommendations? "
                "Mia: They need more evidence to support them."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What had the students been worried about?', ABC('that the owner would not have time', 'that the owner would not share financial information', 'that the bakery would close'), 'B', "she wouldn't want to share financial information"),
                         ('What is the main reason for the bakery’s success?', ABC('its location', 'loyal customers', 'low prices'), 'B', "it's the loyalty of the customers"),
                         ('What is the biggest challenge for the business?', ABC('energy costs', 'finding staff', 'competition'), 'A', 'the cost of energy'),
                         ('What is the owner planning?', ABC('to open a second shop', 'to open a café in the shop', 'to sell the business'), 'B', 'opening a small café inside the shop'),
                         ('What does Mia want to recommend?', ABC('online sales', 'evening classes', 'office deliveries'), 'B', 'running baking classes in the evenings'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Sections of the report',
                     instructions='What does each section of the report still need? Choose the correct letter, A-E.',
                     questions=[
                         ('the introduction', T6_SECTIONS, 'A', 'it just needs checking'),
                         ('the financial analysis', T6_SECTIONS, 'B', 'more up-to-date figures'),
                         ('the interview summary', T6_SECTIONS, 'C', 'cut it by half'),
                         ('the comparison with competitors', T6_SECTIONS, 'D', "We haven't started that yet"),
                         ('the recommendations', T6_SECTIONS, 'E', 'They need more evidence'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear part of a lecture about seed banks.',
            script=(
                "Lecturer: In this lecture I'll be looking at seed banks, which are collections of seeds stored to protect the variety of the world's plants. "
                "Why are they needed? Over the last century, farmers have increasingly grown a small number of high-yielding crop varieties, and thousands of traditional varieties have been lost. "
                "Seed banks aim to preserve this genetic diversity, which may be vital in the future, for example to breed crops that can resist new diseases or survive droughts. "
                "The best-known example is the Svalbard Global Seed Vault, on a remote Norwegian island in the Arctic. It opened in 2008 and is built deep inside a mountain. "
                "The site was chosen partly because the permafrost keeps the rock naturally cold, so the seeds would remain frozen even if the electricity failed. "
                "Inside, the seeds are kept at minus eighteen degrees Celsius. "
                "Svalbard acts as a back-up for other seed banks around the world. Countries and institutions send duplicate samples, which remain their property; the vault simply keeps them safe. "
                "It has been described as an insurance policy for the world's food supply. "
                "Storing seeds is not as simple as it sounds. Before freezing, seeds must be dried carefully, because ice crystals can damage them. "
                "And even frozen seeds do not last for ever, so samples must be tested regularly to check that they will still germinate. "
                "When their quality declines, they must be planted and grown so that fresh seeds can be collected. "
                "Some plants cannot be stored this way at all. Many tropical species, including cocoa and avocado, have seeds that die when they are dried. "
                "For these, scientists are developing other methods, such as freezing small pieces of plant tissue in liquid nitrogen."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Seed banks',
                instructions='Complete the notes below. Write ONE WORD ONLY for each answer.', limit=1,
                questions=[
                    ('Farmers now grow a small number of high-________ varieties', ['yielding'], 'high-yielding crop varieties'),
                    ('Seed banks preserve genetic ________', ['diversity'], 'preserve this genetic diversity'),
                    ('New crops may need to survive ________', ['droughts'], 'survive droughts'),
                    ('The Svalbard vault is built inside a ________', ['mountain'], 'deep inside a mountain'),
                    ('The ________ keeps the rock cold', ['permafrost'], 'the permafrost keeps the rock naturally cold'),
                    ('Countries send ________ samples', ['duplicate'], 'send duplicate samples'),
                    ('Seeds must be ________ before freezing', ['dried'], 'seeds must be dried carefully'),
                    ('Samples are tested to check they will still ________', ['germinate'], 'will still germinate'),
                    ('Seeds of cocoa and ________ die when dried', ['avocado'], 'including cocoa and avocado'),
                    ('Plant tissue can be frozen in liquid ________', ['nitrogen'], 'in liquid nitrogen'),
                ])],
        ),
    ],
), dict(
    key='t7', title='IELTS Listening Practice Test 7',
    parts=[
        dict(
            setting='You will hear a student registering at a language school.',
            script=(
                "Secretary: Welcome to Lindon Language School. Are you here to register? "
                "Student: Yes, for the English course. "
                "Secretary: Lovely. Can I have your first name? "
                "Student: It's Yusuf. "
                "Secretary: And your family name? "
                "Student: Demir. D, E, M, I, R. "
                "Secretary: Thank you. Where are you from? "
                "Student: Turkey, from Izmir. "
                "Secretary: And what's your date of birth? "
                "Student: The fourth of March, 1999. "
                "Secretary: What's your current level of English? "
                "Student: I think intermediate. My teacher at home said upper intermediate, but I'm not so confident. "
                "Secretary: You'll take a placement test anyway. Which course are you interested in? We have general English, business English and exam preparation. "
                "Student: Exam preparation. I need a good score for university. "
                "Secretary: Morning or afternoon classes? "
                "Student: Mornings, because I work in a hotel in the afternoons. "
                "Secretary: How long would you like to study? "
                "Student: Twelve weeks. "
                "Secretary: Where are you staying? "
                "Student: With a host family in Clifton. "
                "Secretary: And who should we contact in an emergency? "
                "Student: My cousin, Selin. "
                "Secretary: Great. Finally, what are your interests? We organise social activities. "
                "Student: I love photography. "
                "Secretary: Then you'll enjoy our photography club on Wednesdays."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Lindon Language School: registration form',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Family name: ________', ['Demir'], 'D, E, M, I, R'),
                    ('City: ________', ['Izmir'], 'from Izmir'),
                    ('Date of birth: 4 ________ 1999', ['March'], 'The fourth of March, 1999'),
                    ('Level: ________', ['intermediate'], 'I think intermediate'),
                    ('Course: ________ preparation', ['exam'], 'Exam preparation. I need'),
                    ('Classes: ________', ['mornings', 'morning'], 'Mornings, because I work'),
                    ('Length of study: ________ weeks', ['12', 'twelve'], 'Twelve weeks'),
                    ('Accommodation: host family in ________', ['Clifton'], 'host family in Clifton'),
                    ('Emergency contact: cousin, ________', ['Selin'], 'My cousin, Selin'),
                    ('Interest: ________', ['photography'], 'I love photography'),
                ])],
        ),
        dict(
            setting='You will hear the manager of a wildlife rescue centre talking to new volunteers.',
            script=(
                "Manager: Good morning and welcome to Oakwood Wildlife Rescue. Thank you all for giving up your time. "
                "We treat around three thousand animals a year, and the number has doubled in the last five years. "
                "Many people think we mainly look after hedgehogs, and we do see a lot of them, but birds are by far the largest group, especially young ones that have fallen from nests in spring. "
                "Our aim is always to return animals to the wild. Animals that can't survive in the wild are placed in sanctuaries; we don't keep any here permanently. "
                "There are a few important safety rules. Always wear gloves when handling animals, even small ones, and never feed an animal unless the vet has written instructions on its card. "
                "And please don't give the animals names, because it makes it harder to let them go. "
                "We're open every day, including public holidays, and new volunteers should start with shifts of four hours. "
                "Now, let me tell you about the different roles. If you work in animal care, you'll be cleaning cages and preparing food. "
                "The reception team answer calls from the public, and they also do the first check of animals brought in. "
                "Fundraising volunteers organise events such as our summer fair. "
                "The animal care team are also the people who release animals back into the wild. "
                "Reception volunteers will be trained to use our database. And the fundraising team run our social media accounts."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('How many animals does the centre treat each year?', ABC('about 1,500', 'about 3,000', 'about 6,000'), 'B', 'around three thousand animals a year'),
                         ('Which animals does the centre see most of?', ABC('hedgehogs', 'birds', 'foxes'), 'B', 'birds are by far the largest group'),
                         ('What happens to animals that cannot return to the wild?', ABC('They stay at the centre.', 'They go to sanctuaries.', 'They are given to zoos.'), 'B', 'are placed in sanctuaries'),
                         ('What should volunteers never do?', ABC('handle small animals', 'feed animals without written instructions', 'work on public holidays'), 'B', 'never feed an animal unless the vet has written instructions'),
                         ('How long should new volunteers’ shifts be?', ABC('two hours', 'four hours', 'six hours'), 'B', 'shifts of four hours'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Volunteer teams',
                     instructions='Which team is responsible for each activity? Choose the correct letter, A, B or C.',
                     questions=[
                         ('preparing food', T7_TEAMS, 'A', 'cleaning cages and preparing food'),
                         ('checking animals when they arrive', T7_TEAMS, 'B', 'the first check of animals brought in'),
                         ('organising events', T7_TEAMS, 'C', 'organise events such as our summer fair'),
                         ('releasing animals', T7_TEAMS, 'A', 'release animals back into the wild'),
                         ('managing social media', T7_TEAMS, 'C', 'run our social media accounts'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two students, Raj and Sophie, discussing a questionnaire for their research project.',
            script=(
                "Raj: Sophie, have you got the draft of our questionnaire? "
                "Sophie: Yes, here it is. I've tried to keep it short. "
                "Raj: Good, because the tutor said response rates drop sharply if a survey takes more than ten minutes. "
                "Sophie: This one should take about six. "
                "Raj: I see you've changed the first question. "
                "Sophie: Yes, originally we asked how often people use the gym, but that assumes everyone uses it. Now the first question asks whether they've used it at all this year. "
                "Raj: Much better. How are we going to distribute it? "
                "Sophie: I thought about handing out paper copies in the canteen, but the students' union has agreed to send it by email to all first-year students. "
                "Raj: That's great. What about the question on cost? "
                "Sophie: I've kept it, but I've made it multiple choice, because people didn't like writing numbers in the pilot. "
                "Raj: Did the pilot show anything else? "
                "Sophie: Yes, several people misunderstood the word 'facilities', so I've added examples. "
                "Raj: Now, the analysis. I think we should start by grouping the answers by year of study. "
                "Sophie: Agreed. And we should compare men and women too. "
                "Raj: For the open questions, we'll need to identify common themes. "
                "Sophie: We could use a spreadsheet for that. "
                "Raj: When shall we send it out? "
                "Sophie: Not before the holidays; let's wait until the start of next term. "
                "Raj: And we'll need a short covering message explaining that answers are anonymous. "
                "Sophie: Yes, that's essential."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('According to the tutor, what reduces the number of people who complete a survey?', ABC('its length', 'difficult questions', 'paper copies'), 'A', 'takes more than ten minutes'),
                         ('Why did Sophie change the first question?', ABC('It was too personal.', 'It assumed everyone used the gym.', 'It was too long.'), 'B', 'that assumes everyone uses it'),
                         ('How will the questionnaire be distributed?', ABC('on paper in the canteen', 'by email', 'on social media'), 'B', 'send it by email to all first-year students'),
                         ('Why did Sophie make the question about cost multiple choice?', ABC('People disliked writing numbers.', 'It was quicker to analyse.', 'The tutor suggested it.'), 'A', "people didn't like writing numbers"),
                         ('What problem did the pilot reveal?', ABC('A word was misunderstood.', 'Some questions were repeated.', 'It was too short.'), 'A', "misunderstood the word 'facilities'"),
                     ]),
                dict(type_id='ielts-listening-completion', heading='Plans for the analysis',
                     instructions='Complete the notes below. Write ONE WORD ONLY for each answer.', limit=1,
                     questions=[
                         ('Group the answers by ________ of study', ['year'], 'by year of study'),
                         ('Compare men and ________', ['women'], 'compare men and women'),
                         ('Open questions: identify common ________', ['themes'], 'identify common themes'),
                         ('Tool: a ________', ['spreadsheet'], 'use a spreadsheet'),
                         ('Covering message: say that answers are ________', ['anonymous'], 'answers are anonymous'),
                     ]),
            ],
        ),
        PART4['t7'],
    ],
)]
