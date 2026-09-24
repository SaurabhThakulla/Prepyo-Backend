"""IELTS Listening Practice Tests 2-4. Same structure and checks as Test 1
(scripts/generate_ielts_listening_tests.py): four parts of ten questions, each
question carrying the exact words in the script that give its answer.
Original Prepyo practice material; not official or recalled test content.
"""
from ielts_bank2.listening_lectures_v2 import PART4
from ielts_bank2.common import ABC, MAP_OPTIONS, WN, W1

T2_GROUPS = ABC('free for a limited period', 'must be booked online', 'needs a letter from a doctor',
                'run by youth workers', 'offered at a reduced price', 'only at weekends')
T2_FARMS = ABC('It brings people together.', 'It uses a great deal of energy.', 'It may lose its land.',
               'It costs little to operate.', 'It can be expensive to establish.', 'It needs a lot of water.')

TESTS_A = [dict(
    key='t2', title='IELTS Listening Practice Test 2',
    parts=[
        dict(
            setting='You will hear a woman phoning a cycle tour company to book a tour for her colleagues.',
            script=(
                "Agent: Good afternoon, Greenway Cycle Tours, Paul speaking. How can I help? "
                "Woman: Hello. I'm interested in one of your guided tours for a group from my office. "
                "Agent: Lovely. How many people would it be? "
                "Woman: There are twelve of us. Well, it was going to be twelve, but two people have dropped out, so ten. "
                "Agent: Ten, fine. And which tour were you thinking of? We have the river tour, the castle tour and the vineyard tour. "
                "Woman: The castle one sounds interesting, but some of my colleagues aren't very fit, and I've heard it's quite hilly. "
                "Agent: It is. The river tour is completely flat. "
                "Woman: Then the river tour, please. "
                "Agent: When would you like to go? "
                "Woman: We were thinking of a Friday in June. "
                "Agent: Let me check. The sixth is full, but the thirteenth is free. "
                "Woman: The thirteenth is perfect. "
                "Agent: The tour starts at ten, and it's about thirty kilometres in total, with a stop for lunch at a pub by the water. "
                "Woman: And what does it cost? "
                "Agent: It's forty-five pounds per person, which includes the bike and a helmet. Lunch is extra. "
                "Woman: OK. Do people need to bring anything? "
                "Agent: Just comfortable shoes and a waterproof jacket. We supply water bottles. "
                "Woman: Where do we meet? "
                "Agent: At our shop, which is on Mill Lane, opposite the station. "
                "Woman: Great. And could I have a contact number for the guide on the day? "
                "Agent: The guide will be Karen, and I'll text you her number the day before. Can I take your name? "
                "Woman: It's Susan Ashworth. A, S, H, W, O, R, T, H. "
                "Agent: Thank you, Ms Ashworth. I'll send you a confirmation email."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Greenway Cycle Tours: booking form',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Number of people: ________', ['10', 'ten'], 'so ten'),
                    ('Tour chosen: the ________ tour', ['river'], 'Then the river tour'),
                    ('Date: ________ June', ['13', '13th', 'thirteenth'], 'the thirteenth is free'),
                    ('Length of tour: about ________ km', ['30', 'thirty'], 'about thirty kilometres'),
                    ('Lunch stop: a ________ by the water', ['pub'], 'a pub by the water'),
                    ('Price per person: £________', ['45', 'forty-five'], 'forty-five pounds per person'),
                    ('Bring: comfortable shoes and a waterproof ________', ['jacket'], 'a waterproof jacket'),
                    ('Meeting place: the shop on ________ Lane', ['Mill'], 'on Mill Lane'),
                    ('Name of guide: ________', ['Karen'], 'The guide will be Karen'),
                    ('Customer surname: ________', ['Ashworth'], 'A, S, H, W, O, R, T, H'),
                ])],
        ),
        dict(
            setting='You will hear the manager of a new leisure centre talking to local residents.',
            script=(
                "Manager: Good evening, everyone, and thank you for coming to hear about the new Westfield Leisure Centre, which opens next month. "
                "I know many of you have been waiting a long time for this. The project was delayed by almost a year. People often assume that was because of money, "
                "but in fact the funding was agreed early on. The delay happened because the builders discovered old mine tunnels under the site, and the foundations had to be redesigned. "
                "The centre has been designed with families in mind. The main pool is twenty-five metres long, and there's a separate learner pool. "
                "What we're most proud of, though, is the climbing wall, which is the tallest in the region. "
                "Opening hours will be from six in the morning until ten at night on weekdays. At weekends we'll open at eight, and close at the same time as on weekdays. "
                "Membership is available in several forms. We expected most people to choose monthly membership, but in fact the most popular option so far is the family pass, "
                "which covers two adults and up to three children. "
                "Parking has been a concern for some residents. There are only eighty spaces, so we're encouraging people to cycle, and we've installed covered racks for two hundred bicycles. "
                "Buses also stop right outside. "
                "Now let me tell you about some of the activities for different groups. For older residents, we'll run gentle exercise classes every weekday morning, "
                "and these will be free of charge for the first three months. For teenagers, the sports hall will be open on Friday nights with basketball and music, run by youth workers. "
                "Parents with babies can join our aqua classes, but these need to be booked online because spaces are limited. "
                "For people recovering from injuries, we'll offer sessions with a physiotherapist, which require a letter from your doctor. "
                "And finally, for local schools, we'll provide swimming lessons during the school day, and schools will receive a discount of thirty per cent."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('Why was the opening of the centre delayed?', ABC('a lack of funding', 'problems under the ground', 'bad weather'), 'B', 'discovered old mine tunnels'),
                         ('What is the manager most proud of?', ABC('the main pool', 'the learner pool', 'the climbing wall'), 'C', 'is the climbing wall'),
                         ('At what time will the centre open at weekends?', ABC('6 a.m.', '8 a.m.', '10 a.m.'), 'B', "At weekends we'll open at eight"),
                         ('Which type of membership is most popular so far?', ABC('monthly', 'annual', 'family'), 'C', 'the most popular option so far is the family pass'),
                         ('How are visitors encouraged to travel to the centre?', ABC('by car', 'by bicycle', 'on foot'), 'B', "we're encouraging people to cycle"),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Activities for different groups',
                     instructions='What does the manager say about the activity for each group? Choose the correct letter, A-F.',
                     questions=[
                         ('older residents', T2_GROUPS, 'A', 'free of charge for the first three months'),
                         ('teenagers', T2_GROUPS, 'D', 'run by youth workers'),
                         ('parents with babies', T2_GROUPS, 'B', 'need to be booked online'),
                         ('people recovering from injuries', T2_GROUPS, 'C', 'require a letter from your doctor'),
                         ('local schools', T2_GROUPS, 'E', 'receive a discount of thirty per cent'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two students, Leo and Priya, discussing their presentation on urban farming with their tutor.',
            script=(
                "Tutor: Come in, Leo and Priya. So, you're presenting on urban farming next week. How are you getting on? "
                "Leo: Pretty well. We chose the topic because Priya visited a rooftop farm in the city centre last summer. "
                "Priya: Yes, I'd read about vertical farms, but actually seeing the rooftop farm made me want to find out more. "
                "Tutor: Good. What do you see as the main benefit of urban farming? "
                "Leo: Well, people often say it reduces transport costs, and it does, a little. But the research we read suggests the main benefit is educational: it reconnects city people with where their food comes from. "
                "Tutor: Interesting. What was the hardest part of the research? "
                "Priya: Finding reliable figures. Lots of websites give numbers for how much food city farms produce, but they often don't say where the figures come from. We ended up relying on two academic studies. "
                "Tutor: Sensible. And how are you going to structure the presentation? "
                "Leo: We thought we'd start with a short video of the rooftop farm, but the file is too large for the classroom computer. So we'll begin with a photograph and a question for the audience instead. "
                "Tutor: That could work well. How long is it? "
                "Priya: We have fifteen minutes. We'd planned twenty, but we cut the section on history. "
                "Tutor: Now, you've looked at several types of urban farm. What did you conclude about each? "
                "Leo: Rooftop farms were the most successful in the studies we read, because they get plenty of sunlight, but they're expensive to set up, as buildings often need strengthening. "
                "Priya: Community gardens were different. They produce relatively little food, but they're excellent for bringing neighbours together. "
                "Leo: Vertical farms, the indoor ones with stacked shelves and artificial light, use very little water, but their energy costs are enormous. "
                "Priya: Then there are school gardens. Their main value is teaching children, and they cost almost nothing to run. "
                "Leo: And finally, farms on unused land, like old car parks. The problem there is that the land often isn't guaranteed; the owners can take it back for building at any time. "
                "Tutor: That's a nice range. I look forward to it."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('Why did the students choose urban farming as their topic?', ABC('They had read about vertical farms.', 'Priya had visited a rooftop farm.', 'Their tutor suggested it.'), 'B', 'Priya visited a rooftop farm'),
                         ('According to their research, what is the main benefit of urban farming?', ABC('lower transport costs', 'education', 'fresher food'), 'B', 'the main benefit is educational'),
                         ('What was the students’ main difficulty in their research?', ABC('finding reliable figures', 'understanding academic studies', 'contacting farmers'), 'A', 'Finding reliable figures'),
                         ('How will the presentation begin?', ABC('with a video', 'with a photograph', 'with some history'), 'B', "we'll begin with a photograph"),
                         ('How long will the presentation last?', ABC('10 minutes', '15 minutes', '20 minutes'), 'B', 'We have fifteen minutes'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Types of urban farm',
                     instructions='What do the students say about each type of urban farm? Choose the correct letter, A-F.',
                     questions=[
                         ('rooftop farms', T2_FARMS, 'E', "they're expensive to set up"),
                         ('community gardens', T2_FARMS, 'A', 'excellent for bringing neighbours together'),
                         ('vertical farms', T2_FARMS, 'B', 'their energy costs are enormous'),
                         ('school gardens', T2_FARMS, 'D', 'they cost almost nothing to run'),
                         ('farms on unused land', T2_FARMS, 'C', 'the owners can take it back'),
                     ]),
            ],
        ),
        PART4['t2'],
    ],
), dict(
    key='t3', title='IELTS Listening Practice Test 3',
    parts=[
        dict(
            setting='You will hear a man phoning a festival office to volunteer.',
            script=(
                "Coordinator: Hello, Brookfield Summer Festival volunteers' office. "
                "Man: Hi, I saw your advert for volunteers and I'd like to sign up. "
                "Coordinator: Great. Let me take some details. What's your full name? "
                "Man: Daniel Okafor. That's O, K, A, F, O, R. "
                "Coordinator: Thanks. And how old are you, Daniel? We need volunteers to be over eighteen. "
                "Man: I'm twenty-three. "
                "Coordinator: Lovely. What's your address? "
                "Man: Flat 4, 27 Seaton Road. "
                "Coordinator: And which area would you like to work in? We need people for car parking, the information tent and litter collection. "
                "Man: I'd prefer the information tent. I'm good with people. "
                "Coordinator: Perfect. The festival runs from Thursday to Sunday. Which days can you do? "
                "Man: I work on Thursdays, so Friday to Sunday. "
                "Coordinator: That's fine. Shifts are six hours long. Have you volunteered before? "
                "Man: Yes, I helped at a marathon last year. "
                "Coordinator: Excellent. All volunteers get a free T-shirt. What size are you? "
                "Man: Large, please. "
                "Coordinator: And do you have any dietary requirements? We provide meals for volunteers. "
                "Man: I'm vegetarian. "
                "Coordinator: Noted. Finally, you'll need to attend a training session. There's one on the twelfth of July and another on the nineteenth. "
                "Man: The twelfth is better for me. "
                "Coordinator: Great. It's at the town hall at seven in the evening."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Brookfield Summer Festival: volunteer form',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Surname: ________', ['Okafor'], 'O, K, A, F, O, R'),
                    ('Age: ________', ['23', 'twenty-three'], "I'm twenty-three"),
                    ('Address: Flat 4, 27 ________ Road', ['Seaton'], 'Flat 4, 27 Seaton Road'),
                    ('Preferred area: the information ________', ['tent'], "I'd prefer the information tent"),
                    ('Available: Friday to ________', ['Sunday'], 'so Friday to Sunday'),
                    ('Length of shifts: ________ hours', ['6', 'six'], 'Shifts are six hours long'),
                    ('Previous experience: helped at a ________', ['marathon'], 'helped at a marathon'),
                    ('T-shirt size: ________', ['large'], 'Large, please'),
                    ('Diet: ________', ['vegetarian'], "I'm vegetarian"),
                    ('Training date: ________ July', ['12', '12th', 'twelfth'], 'The twelfth is better for me'),
                ])],
        ),
        dict(
            setting='You will hear an instructor welcoming a group to an activity centre.',
            script=(
                "Instructor: Hello, everyone, and welcome to Ashcombe Activity Centre. I'm Jo, and I'll be looking after you this week. "
                "Before we start, a few things about the centre. It was originally built as a farm school in the nineteen sixties. "
                "Most people think it was a hospital, because of the long corridors, but it was always used for education. "
                "The centre is run by a charity, and most of our instructors are volunteers, although I'm one of the few paid staff. "
                "During your stay, the most important rule is about the lake. You may be keen to swim, but nobody may go in the water without an instructor, even strong swimmers. "
                "Meals are served at set times. Breakfast is at half past seven, and please don't be late, because the kitchen closes at half past eight. "
                "If you have a problem during the night, don't go to the office; call the number on your door, and the duty instructor will come to you. "
                "Now, look at the plan of the main building. You come in through the main entrance at the bottom, and reception is on your left. "
                "You can leave your valuables in the lockers, in the first room on your right as you come in. "
                "Carry on along the main corridor. The climbing wall is in the room on your left just after the cross corridor, opposite the café. "
                "At the end of the building, the equipment store is the last room on your right, at the far end, and that's where you'll collect helmets and life jackets. "
                "The changing rooms are in the room on your left just before the cross corridor, next to reception. "
                "And if it rains in the evening, the games room is the last room on your left, at the far end, with table tennis and board games."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What was the building originally used as?', ABC('a farm school', 'a hospital', 'a hotel'), 'A', 'originally built as a farm school'),
                         ('Who are most of the instructors at the centre?', ABC('paid staff', 'volunteers', 'students'), 'B', 'most of our instructors are volunteers'),
                         ('What is the most important rule?', ABC('Stay out of the kitchen.', 'Only swim with an instructor.', 'Do not go outside at night.'), 'B', 'nobody may go in the water without an instructor'),
                         ('What time does the kitchen close after breakfast?', ABC('7.30', '8.00', '8.30'), 'C', 'the kitchen closes at half past eight'),
                         ('What should guests do if there is a problem at night?', ABC('go to the office', 'call a number', 'wake another guest'), 'B', 'call the number on your door'),
                     ]),
                dict(type_id='ielts-listening-map', heading='Ashcombe Activity Centre',
                     instructions='Label the plan below. Choose the correct letter, A-F.',
                     image=('Ashcombe Activity Centre', {('W', 1): 'Reception', ('E', 3): 'Café'},
                            [('E', 1), ('W', 2), ('E', 2), ('W', 3), ('W', 4), ('E', 4)]),
                     questions=[
                         ('Lockers', MAP_OPTIONS, 'A', 'in the first room on your right as you come in'),
                         ('Climbing wall', MAP_OPTIONS, 'D', 'in the room on your left just after the cross corridor'),
                         ('Equipment store', MAP_OPTIONS, 'F', 'the last room on your right, at the far end'),
                         ('Changing rooms', MAP_OPTIONS, 'B', 'in the room on your left just before the cross corridor'),
                         ('Games room', MAP_OPTIONS, 'E', 'the last room on your left, at the far end'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two students, Hannah and Marcus, discussing a report on their museum work placement.',
            script=(
                "Hannah: Hi Marcus. Have you got a minute to talk about our museum placement report? "
                "Marcus: Sure. How did you find the placement overall? "
                "Hannah: I really enjoyed it. I thought I'd spend most of my time in the archives, but actually I was mainly working with visitors, which I preferred. "
                "Marcus: Same for me. The thing that surprised me most was how few staff there are. The museum relies on just six full-time employees. "
                "Hannah: I know. And the director told me their biggest challenge isn't money, it's attracting teenagers, who hardly ever visit. "
                "Marcus: That's what our report should focus on, then. What did you think of the new digital guide? "
                "Hannah: The technology was impressive, but visitors found the headphones uncomfortable, so a lot of them gave up after ten minutes. "
                "Marcus: That's a good point to include. Now, our tutor wants the report by the end of the month. Should we aim for five thousand words? "
                "Hannah: She said four thousand is the minimum, so let's aim for about four and a half thousand. "
                "Marcus: OK. Let's divide up the work. Who's going to write the introduction? "
                "Hannah: I'll do that, since I've already drafted some notes. "
                "Marcus: Fine. I'll collect the visitor statistics; I've got the contact at the museum. "
                "Hannah: What about the interviews with staff? There are quite a lot. "
                "Marcus: Let's share those; you take the curators and I'll take the education team. "
                "Hannah: Good. And the recommendations? "
                "Marcus: You're better at that kind of thing. You write them and I'll check them. "
                "Hannah: OK. Then the photographs. "
                "Marcus: I took most of them, so I'll choose which ones to use."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What did Hannah spend most of her placement doing?', ABC('working in the archives', 'working with visitors', 'preparing exhibitions'), 'B', 'I was mainly working with visitors'),
                         ('What surprised Marcus most about the museum?', ABC('the small number of staff', 'the lack of money', 'the number of visitors'), 'A', 'how few staff there are'),
                         ('According to the director, what is the museum’s biggest challenge?', ABC('funding', 'attracting teenagers', 'finding volunteers'), 'B', "it's attracting teenagers"),
                         ('What was the problem with the digital guide?', ABC('It was too expensive.', 'The technology did not work.', 'The headphones were uncomfortable.'), 'C', 'visitors found the headphones uncomfortable'),
                         ('How long will the report be?', ABC('4,000 words', '4,500 words', '5,000 words'), 'B', 'aim for about four and a half thousand'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Responsibilities for the report',
                     instructions='Who will be responsible for each part of the report? Choose the correct letter, A, B or C. '
                                  'A: Hannah. B: Marcus. C: both Hannah and Marcus.',
                     questions=[
                         ('the introduction', ABC('Hannah', 'Marcus', 'both'), 'A', "I'll do that, since I've already drafted"),
                         ('the visitor statistics', ABC('Hannah', 'Marcus', 'both'), 'B', "I'll collect the visitor statistics"),
                         ('the staff interviews', ABC('Hannah', 'Marcus', 'both'), 'C', "Let's share those"),
                         ('the recommendations', ABC('Hannah', 'Marcus', 'both'), 'A', 'You write them'),
                         ('the photographs', ABC('Hannah', 'Marcus', 'both'), 'B', "so I'll choose which ones to use"),
                     ]),
            ],
        ),
        PART4['t3'],
    ],
), dict(
    key='t4', title='IELTS Listening Practice Test 4',
    parts=[
        dict(
            setting='You will hear a woman phoning a pet care service to arrange for someone to look after her cat.',
            script=(
                "Assistant: Happy Paws Pet Care, this is Tom. "
                "Woman: Hello, I'm going away next month and I need someone to look after my cat. "
                "Assistant: Of course. Can I start with your name? "
                "Woman: Yes, it's Rachel Lindqvist. That's L, I, N, D, Q, V, I, S, T. "
                "Assistant: Thank you. And your address? "
                "Woman: 15 Orchard Close, in Fenwick. "
                "Assistant: Great, that's within our area. What's your cat's name? "
                "Woman: She's called Pepper. "
                "Assistant: How old is she? "
                "Woman: She's nine. Actually, no, she had her birthday in May, so she's ten now. "
                "Assistant: Ten. And the dates you need? "
                "Woman: From the third to the seventeenth of August. "
                "Assistant: Two weeks. We usually visit twice a day. Would that suit you? "
                "Woman: She's quite independent, so once a day would be enough. "
                "Assistant: Once a day, then, in the morning or the evening? "
                "Woman: The evening, please. That's when she eats. "
                "Assistant: Does she need any medicine? "
                "Woman: Yes, she takes a tablet for her heart, crushed into her food. "
                "Assistant: I'll make a note. Any other instructions? "
                "Woman: Please could you water my plants as well? "
                "Assistant: Certainly, there's no extra charge for that. How would you like to leave the key? "
                "Woman: I'll leave it with my neighbour at number 17. "
                "Assistant: Perfect. Our rate is twelve pounds per visit. "
                "Woman: That's fine."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Happy Paws Pet Care: booking',
                instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Surname: ________', ['Lindqvist'], 'L, I, N, D, Q, V, I, S, T'),
                    ('Address: 15 ________ Close, Fenwick', ['Orchard'], '15 Orchard Close'),
                    ('Cat’s name: ________', ['Pepper'], "She's called Pepper"),
                    ('Cat’s age: ________', ['10', 'ten'], "so she's ten now"),
                    ('Dates: 3 to ________ August', ['17', '17th', 'seventeenth'], 'the seventeenth of August'),
                    ('Visits: once a day, in the ________', ['evening'], 'The evening, please'),
                    ('Medicine: a tablet for her ________', ['heart'], 'a tablet for her heart'),
                    ('Also: water the ________', ['plants'], 'water my plants'),
                    ('Key: with the neighbour at number ________', ['17', 'seventeen'], 'my neighbour at number 17'),
                    ('Cost per visit: £________', ['12', 'twelve'], 'twelve pounds per visit'),
                ])],
        ),
        dict(
            setting='You will hear a radio interview with the organiser of a food festival.',
            script=(
                "Presenter: And now for our guest, Maria Costa, who's organising this year's Harbourside Food Festival. "
                "Maria: Thanks for having me. This is the festival's tenth year. It started as a small market with just a dozen stalls, and this year we'll have over a hundred and twenty. "
                "Presenter: That's a big change. What's new this year? "
                "Maria: Well, we've always had cookery demonstrations, and the children's area has been running for years. The new thing is a zero-waste zone, where all the food is served without any plastic. "
                "Presenter: Is the festival free? "
                "Maria: Entry is free on Friday, but on Saturday and Sunday there's a small charge of five pounds, which goes to the lifeboat charity. "
                "Presenter: And what's the best way to get there? "
                "Maria: Please don't drive; there'll be no parking near the harbour. We're running a free boat service from the marina, which is the most relaxing way to arrive. "
                "Presenter: What about the weather? "
                "Maria: Everything is under cover this year, so rain won't stop anything. "
                "Presenter: Can you tell us about some of the highlights? "
                "Maria: Of course. The festival opens on Friday evening with a street food night. On Saturday morning, local schools are holding a baking competition, which is always popular. "
                "Then on Saturday afternoon we have a talk by the chef Sam Oduya about cooking with seasonal vegetables. "
                "The famous oyster-opening contest takes place on Sunday morning, and the festival closes on Sunday evening with fireworks over the water. "
                "Oh, and I nearly forgot: the farmers' breakfast, which used to be on Sunday, has moved to Friday morning, before the festival officially opens."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('How many stalls were there when the festival started?', ABC('12', '100', '120'), 'A', 'just a dozen stalls'),
                         ('What is new at the festival this year?', ABC('cookery demonstrations', 'a children’s area', 'a zero-waste zone'), 'C', 'The new thing is a zero-waste zone'),
                         ('On which day is entry to the festival free?', ABC('Friday', 'Saturday', 'Sunday'), 'A', 'Entry is free on Friday'),
                         ('What is the recommended way to travel to the festival?', ABC('by car', 'by boat', 'by bus'), 'B', 'free boat service from the marina'),
                         ('What does Maria say about the weather?', ABC('Some events may be cancelled if it rains.', 'All events are under cover.', 'The forecast is good.'), 'B', 'Everything is under cover this year'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Festival events',
                     instructions='When does each event take place? Choose the correct letter, A, B or C. A: Friday. B: Saturday. C: Sunday.',
                     questions=[
                         ('street food night', ABC('Friday', 'Saturday', 'Sunday'), 'A', 'Friday evening with a street food night'),
                         ('schools’ baking competition', ABC('Friday', 'Saturday', 'Sunday'), 'B', 'local schools are holding a baking competition'),
                         ('talk on seasonal vegetables', ABC('Friday', 'Saturday', 'Sunday'), 'B', 'a talk by the chef'),
                         ('oyster-opening contest', ABC('Friday', 'Saturday', 'Sunday'), 'C', 'oyster-opening contest takes place on Sunday'),
                         ('farmers’ breakfast', ABC('Friday', 'Saturday', 'Sunday'), 'A', 'has moved to Friday morning'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two psychology students, Clara and Ahmed, discussing an experiment they carried out.',
            script=(
                "Clara: So, Ahmed, shall we go over the results of our experiment before we write it up? "
                "Ahmed: Good idea. Just to remind ourselves, we tested whether listening to music helps people remember a list of words. "
                "Clara: Right. We originally wanted sixty participants, but in the end we only managed forty, because it was exam season. "
                "Ahmed: Forty is still enough for a pilot study, our supervisor said. "
                "Clara: True. And the main finding? "
                "Ahmed: That's the interesting part. We expected the group with classical music to do best, but in fact the silent group remembered the most words. "
                "Clara: And the group with pop music with lyrics did worst, which makes sense, because the words in the songs compete with the words in the list. "
                "Ahmed: I think our biggest weakness was the room. It was next to the cafeteria, so it was sometimes noisy. "
                "Clara: Yes, we should mention that. What about the time of day? "
                "Ahmed: We tested everyone in the afternoon, so that isn't a problem. "
                "Clara: For the write-up, I think we should add a graph showing the average scores for each group. "
                "Ahmed: Agreed. And in the discussion, we need to compare our results with the study by Petersen, which found the opposite. "
                "Clara: If we did it again, I'd like to use participants of different ages, not just students. "
                "Ahmed: And I'd test people again after a week, to see whether the music affects long-term memory. "
                "Clara: Good. Our supervisor also suggested we could present the study at the student conference in May."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What was the aim of the experiment?', ABC('to test whether music helps memory', 'to compare types of music', 'to measure concentration'), 'A', 'whether listening to music helps people remember'),
                         ('Why did they have fewer participants than planned?', ABC('The supervisor said forty was enough.', 'It was exam season.', 'Some people dropped out.'), 'B', 'because it was exam season'),
                         ('Which group remembered the most words?', ABC('the classical music group', 'the pop music group', 'the silent group'), 'C', 'the silent group remembered the most words'),
                         ('Why did the pop music group do worst?', ABC('The lyrics interfered.', 'The music was too loud.', 'They were tired.'), 'A', 'the words in the songs compete'),
                         ('What was the main weakness of the experiment?', ABC('the time of day', 'noise near the room', 'the choice of words'), 'B', 'It was next to the cafeteria'),
                     ]),
                dict(type_id='ielts-listening-completion', heading='Plans for the write-up and future research',
                     instructions='Complete the sentences below. Write ONE WORD ONLY for each answer.', limit=1,
                     questions=[
                         ('They will add a ________ of the average scores.', ['graph'], 'add a graph'),
                         ('They will compare their results with a study by ________.', ['Petersen'], 'the study by Petersen'),
                         ('In future, they would use participants of different ________.', ['ages'], 'participants of different ages'),
                         ('They would test people again after a ________.', ['week'], 'after a week'),
                         ('They may present the study at a student ________ in May.', ['conference'], 'the student conference in May'),
                     ]),
            ],
        ),
        PART4['t4'],
    ],
)]
