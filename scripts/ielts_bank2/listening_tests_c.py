"""IELTS Listening Practice Tests 8-10. Same structure and checks as Test 1.
Original Prepyo practice material; not official or recalled test content.
"""
from ielts_bank2.listening_lectures_v2 import PART4
from ielts_bank2.common import ABC, MAP_OPTIONS

T8_ORGS = ABC('will reply quickly', 'wants to see the results', 'is not relevant', 'is difficult to schedule', 'can provide wider contacts')
T9_ADVERTS = ABC('has an effective story', 'may upset some viewers', 'does not show the product clearly', 'contains too much information', 'may discourage young people')
T10_FIXES = ABC('added a layer of gravel', 'added charcoal', 'used thicker clay', 'fitted a rubber seal', 'used pictures', 'changed the sand')

TESTS_C = [dict(
    key='t8', title='IELTS Listening Practice Test 8',
    parts=[
        dict(
            setting='You will hear a woman phoning a company to hire camping equipment.',
            script=(
                "Assistant: Outdoor Hire, can I help you? "
                "Woman: Yes, I'm going camping with some friends and we'd like to hire some equipment. "
                "Assistant: Sure. When's the trip? "
                "Woman: The last weekend in August, Friday to Monday. "
                "Assistant: So three nights. What do you need? "
                "Woman: A tent for four people, and four sleeping bags. "
                "Assistant: We have a four-person tent, but if you want more room, the six-person tent is only a little more. "
                "Woman: The four-person one is fine; we don't have much space in the car. "
                "Assistant: OK. The tent is twenty pounds a night and sleeping bags are six pounds each per night. "
                "Woman: Do you have cooking equipment? "
                "Assistant: Yes, a camping stove and pans, for fifteen pounds for the whole trip. "
                "Woman: We'll take those. "
                "Assistant: Would you like us to deliver? It's free if you live within ten miles. "
                "Woman: We live in Hartley, which is about eight miles away. "
                "Assistant: Then delivery is free. We'll deliver on Thursday evening. We need a deposit of fifty pounds. "
                "Woman: Can I pay by bank transfer? "
                "Assistant: Of course. Please check the tent when it arrives, and make sure it's dry before you return it; if it's returned wet, there's a cleaning charge."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Outdoor Hire: order form',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Number of nights: ________', ['3', 'three'], 'So three nights'),
                    ('Tent size: ________ people', ['4', 'four'], 'The four-person one is fine'),
                    ('Tent cost per night: £________', ['20', 'twenty'], 'twenty pounds a night'),
                    ('Sleeping bags: £________ each per night', ['6', 'six'], 'six pounds each per night'),
                    ('Also hiring: a stove and ________', ['pans'], 'a camping stove and pans'),
                    ('Customer lives in: ________', ['Hartley'], 'We live in Hartley'),
                    ('Delivery day: ________', ['Thursday'], 'deliver on Thursday evening'),
                    ('Deposit: £________', ['50', 'fifty'], 'a deposit of fifty pounds'),
                    ('Payment by: bank ________', ['transfer'], 'by bank transfer'),
                    ('Cleaning charge if the tent is returned ________', ['wet'], "if it's returned wet"),
                ])],
        ),
        dict(
            setting='You will hear a council officer talking on local radio about changes to bus services.',
            script=(
                "Officer: Good morning. I'm here to explain the changes to Redbridge bus services, which come into effect on the first of October. "
                "The main reason for the changes is not to save money, as some newspapers have suggested, but to reflect how people now travel. "
                "More people work from home, so there are fewer passengers at rush hour and more in the middle of the day. "
                "The biggest change is to route 5, which will now run every ten minutes instead of every fifteen, because it serves the hospital. "
                "Route 12 will no longer go through the town centre; instead it will go round the ring road, which should cut the journey time. "
                "Route 20, which was very little used in the evenings, will stop at eight p.m. "
                "And we're introducing a completely new route, the 30, which will connect the railway station with the new business park. "
                "Tickets are also changing. The single fare will stay at two pounds, but we're introducing a weekly ticket for twelve pounds, which can be bought on your phone. "
                "Paper tickets will still be available from drivers, but only for cash. "
                "We know that some people are unhappy about the changes, particularly older residents. "
                "So we're holding a public meeting at the library on the fifteenth of September, where you can ask questions. "
                "And from October, anyone over seventy can travel free at any time of day."),
            groups=[
                dict(type_id='ielts-listening-completion', heading='Changes to bus routes',
                     instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                     questions=[
                         ('Route 5: every ________ minutes', ['10', 'ten'], 'every ten minutes'),
                         ('Route 5 serves the ________', ['hospital'], 'it serves the hospital'),
                         ('Route 12: goes round the ________ road', ['ring'], 'go round the ring road'),
                         ('Route 20: stops at ________ p.m.', ['8', 'eight'], 'stop at eight p.m.'),
                         ('New route 30: railway station to the new business ________', ['park'], 'the new business park'),
                     ]),
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('How much will a weekly ticket cost?', ABC('£2', '£10', '£12'), 'C', 'a weekly ticket for twelve pounds'),
                         ('How can paper tickets be bought?', ABC('online', 'with cash from drivers', 'at the library'), 'B', 'only for cash'),
                         ('Where will the public meeting be held?', ABC('at the library', 'at the town hall', 'at the station'), 'A', 'a public meeting at the library'),
                         ('When is the public meeting?', ABC('1 September', '15 September', '1 October'), 'B', 'on the fifteenth of September'),
                         ('Who will be able to travel free from October?', ABC('children', 'people over seventy', 'students'), 'B', 'anyone over seventy can travel free'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two sociology students, Ella and Sam, planning a project on volunteering.',
            script=(
                "Ella: Sam, we need to finalise our project on volunteering among young people. "
                "Sam: Yes. Did you read the report I sent you? "
                "Ella: I did. I was surprised that the number of young volunteers has actually increased, because the news always says young people don't care about their communities. "
                "Sam: Exactly, and that's why I think it's a good topic. "
                "Ella: The report said the main reason young people volunteer is to gain experience for their careers. "
                "Sam: Right, although when we did our own quick survey, most people said they wanted to help others. "
                "Ella: That's a difference we could explore. What should our main research question be? "
                "Sam: I suggest we focus on why some young people stop volunteering after a short time. "
                "Ella: Good, that's more original than just asking why they start. For methods, I think interviews would give us richer data than another survey. "
                "Sam: Agreed, but we'll need at least fifteen to make it convincing. "
                "Ella: Now, who should we contact? "
                "Sam: The food bank would be a good start; they have lots of student volunteers. "
                "Ella: And they're used to researchers, so they'll respond quickly. "
                "Sam: The animal shelter? "
                "Ella: They've had problems keeping volunteers, so they'd be really interested in our findings. "
                "Sam: What about the hospital? "
                "Ella: Their volunteers are mostly retired people, so they're not relevant to our topic. "
                "Sam: The youth football club? "
                "Ella: They're relevant, but they only meet at weekends, so interviews would be hard to arrange. "
                "Sam: And the environmental group? "
                "Ella: They offered to put us in touch with members in other cities, which would widen our sample."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What surprised Ella in the report?', ABC('Volunteering among young people has increased.', 'Young people prefer paid work.', 'Most volunteers are students.'), 'A', 'the number of young volunteers has actually increased'),
                         ('According to the report, why do most young people volunteer?', ABC('to help others', 'to gain career experience', 'to make friends'), 'B', 'to gain experience for their careers'),
                         ('What will be the focus of their research?', ABC('why young people start volunteering', 'why young people stop volunteering', 'where young people volunteer'), 'B', 'why some young people stop volunteering'),
                         ('Which research method do they choose?', ABC('a survey', 'interviews', 'observation'), 'B', 'interviews would give us richer data'),
                         ('How many interviews do they think they need?', ABC('at least 10', 'at least 15', 'at least 20'), 'B', 'at least fifteen'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Organisations to contact',
                     instructions='What does Ella say about each organisation? Choose the correct letter, A-E.',
                     questions=[
                         ('the food bank', T8_ORGS, 'A', "they'll respond quickly"),
                         ('the animal shelter', T8_ORGS, 'B', 'really interested in our findings'),
                         ('the hospital', T8_ORGS, 'C', 'not relevant to our topic'),
                         ('the youth football club', T8_ORGS, 'D', 'interviews would be hard to arrange'),
                         ('the environmental group', T8_ORGS, 'E', 'widen our sample'),
                     ]),
            ],
        ),
        PART4['t8'],
    ],
), dict(
    key='t9', title='IELTS Listening Practice Test 9',
    parts=[
        dict(
            setting='You will hear a man phoning a hotel to book a room for his family.',
            script=(
                "Receptionist: Good evening, Seacliff Hotel. How can I help? "
                "Man: Hello, I'd like to book a room for my family in October. "
                "Receptionist: Certainly. What dates? "
                "Man: From the nineteenth to the twenty-second. "
                "Receptionist: Three nights. How many guests? "
                "Man: Two adults and two children. "
                "Receptionist: We have family rooms with a double bed and two single beds. Would you like a sea view? "
                "Man: How much extra is that? "
                "Receptionist: Fifteen pounds a night. "
                "Man: Yes, let's have the sea view. "
                "Receptionist: The family room with a sea view is a hundred and forty pounds a night, including breakfast. "
                "Man: That's fine. Do you have parking? "
                "Receptionist: Yes, it's free, but spaces can't be reserved. "
                "Man: OK. My younger son is allergic to feathers. "
                "Receptionist: I'll make sure the room has non-feather pillows. "
                "Man: Thank you. Is there anything for children to do? "
                "Receptionist: There's an indoor pool, which is open until eight in the evening, and a games room. "
                "Man: Great. Could I book dinner for our first night? "
                "Receptionist: Of course. At what time? "
                "Man: Half past six, please; the children eat early. "
                "Receptionist: And your name, please? "
                "Man: Martin Kowalczyk. K, O, W, A, L, C, Z, Y, K. "
                "Receptionist: Thank you. We'll need a card to guarantee the booking."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Seacliff Hotel: reservation',
                instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Arrival date: ________ October', ['19', '19th', 'nineteenth'], 'From the nineteenth'),
                    ('Number of nights: ________', ['3', 'three'], 'Three nights'),
                    ('Room: family room with a ________ view', ['sea'], "let's have the sea view"),
                    ('Cost per night: £________', ['140'], 'a hundred and forty pounds a night'),
                    ('Price includes: ________', ['breakfast'], 'including breakfast'),
                    ('Parking: ________', ['free'], "it's free, but spaces"),
                    ('Allergy: ________', ['feathers'], 'allergic to feathers'),
                    ('Pool open until: ________ p.m.', ['8', 'eight'], 'open until eight in the evening'),
                    ('Dinner booked for: ________ p.m.', ['6.30', '6:30'], 'Half past six, please'),
                    ('Surname: ________', ['Kowalczyk'], 'K, O, W, A, L, C, Z, Y, K'),
                ])],
        ),
        dict(
            setting='You will hear a guide talking to visitors at a castle heritage centre.',
            script=(
                "Guide: Welcome to Greystone Heritage Centre, which tells the story of Greystone Castle. "
                "The castle itself was first built in the twelfth century, but what you see today dates mostly from the fifteenth, when it was rebuilt after a fire. "
                "For three hundred years it belonged to the Ashford family, who sold it to the state in 1952. "
                "Before you set off, a few things to know. The castle walls are open to visitors, but the north tower is closed at the moment for repairs to the stairs. "
                "Our audio guides are free, and they're available in nine languages. "
                "On Saturdays, there's a special event: actors in costume show what daily life was like in the castle. "
                "If you're short of time, the thing I'd recommend most is the view from the east wall, over the river. "
                "Now, have a look at the plan of the heritage centre. The main entrance is at the bottom. "
                "If you haven't bought your ticket online, you'll need to go to the ticket office, the first room on your right as you come in. "
                "The armour gallery is the room on your right just after the cross corridor, next to the chapel. "
                "Our newest display is the kitchen exhibition, the last room on your left, at the far end, opposite the chapel. "
                "The toilets are in the first room on your left as you come in. "
                "And families should head for the children's activity room, the room on your right just before the cross corridor, which is opposite the shop."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('When was most of the castle that visitors see today built?', ABC('in the twelfth century', 'in the fifteenth century', 'in 1952'), 'B', 'dates mostly from the fifteenth'),
                         ('Why is the north tower closed?', ABC('for repairs to the stairs', 'because of bad weather', 'for a new exhibition'), 'A', 'for repairs to the stairs'),
                         ('What is true about the audio guides?', ABC('They are free.', 'They cost £9.', 'They are only in English.'), 'A', 'Our audio guides are free'),
                         ('What happens on Saturdays?', ABC('extra guided tours', 'actors show daily life in the castle', 'children’s workshops'), 'B', 'actors in costume show what daily life was like'),
                         ('What does the guide recommend most?', ABC('the armour gallery', 'the view from the east wall', 'the chapel'), 'B', 'the view from the east wall'),
                     ]),
                dict(type_id='ielts-listening-map', heading='Greystone Heritage Centre',
                     instructions='Label the plan below. Choose the correct letter, A-F.',
                     image=('Greystone Heritage Centre', {('W', 2): 'Shop', ('E', 4): 'Chapel'},
                            [('W', 1), ('E', 1), ('E', 2), ('W', 3), ('E', 3), ('W', 4)]),
                     questions=[
                         ('Ticket office', MAP_OPTIONS, 'B', 'the first room on your right as you come in'),
                         ('Armour gallery', MAP_OPTIONS, 'E', 'the room on your right just after the cross corridor'),
                         ('Kitchen exhibition', MAP_OPTIONS, 'F', 'the last room on your left, at the far end'),
                         ('Toilets', MAP_OPTIONS, 'A', 'the first room on your left as you come in'),
                         ('Children’s activity room', MAP_OPTIONS, 'C', 'the room on your right just before the cross corridor'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two marketing students, Nadia and Ben, discussing an assignment about advertising.',
            script=(
                "Ben: Nadia, have you looked at the adverts we collected for the marketing assignment? "
                "Nadia: Yes. I thought we'd find that most of them used celebrities, but only three out of twenty did. "
                "Ben: I noticed that too. Humour was much more common. "
                "Nadia: Right. So what's our argument going to be? "
                "Ben: I think we should argue that emotional adverts work better than ones that just give information. "
                "Nadia: I agree, but we need evidence, not just our opinion. "
                "Ben: The tutor recommended the textbook by Carver, which has a chapter on emotional appeals. "
                "Nadia: I've read it, but it's quite old. The journal articles are more useful. "
                "Ben: Fair enough. How shall we present the adverts? "
                "Nadia: We can't include all twenty, so let's choose five that represent different approaches. "
                "Ben: Good idea. And the deadline is the fourteenth, so we should have a draft ready a week earlier. "
                "Nadia: That gives the tutor time to comment. "
                "Ben: Let's decide what we'll say about each of the five adverts. The car advert? "
                "Nadia: The music is clever, but it doesn't really show the product. "
                "Ben: The bank advert? "
                "Nadia: That one's too serious; it would put young people off. "
                "Ben: The coffee advert? "
                "Nadia: That's my favourite; it tells a story in thirty seconds. "
                "Ben: The phone advert? "
                "Nadia: Technically brilliant, but it gives too much information, so it's confusing. "
                "Ben: And the charity advert? "
                "Nadia: It's very moving, but some viewers might find it upsetting."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What did Nadia expect to find about the adverts?', ABC('Most would use celebrities.', 'Most would use humour.', 'Most would be for food.'), 'A', 'most of them used celebrities'),
                         ('What will the students argue?', ABC('Humour is the most effective approach.', 'Emotional adverts are more effective.', 'Information matters more than emotion.'), 'B', 'emotional adverts work better'),
                         ('What does Nadia say about the textbook by Carver?', ABC('It is out of date.', 'It is too difficult.', 'It is very useful.'), 'A', "it's quite old"),
                         ('How many adverts will they present?', ABC('three', 'five', 'twenty'), 'B', 'choose five'),
                         ('When will they complete a draft?', ABC('a week before the deadline', 'on the day of the deadline', 'after the tutor’s comments'), 'A', 'a draft ready a week earlier'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='The five adverts',
                     instructions='What does Nadia say about each advert? Choose the correct letter, A-E.',
                     questions=[
                         ('the car advert', T9_ADVERTS, 'C', "it doesn't really show the product"),
                         ('the bank advert', T9_ADVERTS, 'E', 'would put young people off'),
                         ('the coffee advert', T9_ADVERTS, 'A', 'it tells a story in thirty seconds'),
                         ('the phone advert', T9_ADVERTS, 'D', 'gives too much information'),
                         ('the charity advert', T9_ADVERTS, 'B', 'might find it upsetting'),
                     ]),
            ],
        ),
        PART4['t9'],
    ],
), dict(
    key='t10', title='IELTS Listening Practice Test 10',
    parts=[
        dict(
            setting='You will hear a woman phoning a trampoline park to book a birthday party.',
            script=(
                "Assistant: Jump Zone, Kelly speaking. "
                "Mother: Hi, I'd like to book a birthday party for my daughter. "
                "Assistant: Lovely. How old will she be? "
                "Mother: Seven. "
                "Assistant: And what date? "
                "Mother: Saturday the ninth of November. "
                "Assistant: Let me check. The morning is booked, but we have a slot at two o'clock. "
                "Mother: Two is fine. "
                "Assistant: How many children? "
                "Mother: About fifteen. "
                "Assistant: Our standard package covers up to twenty children and costs a hundred and eighty pounds. That includes an hour of jumping and forty-five minutes in the party room. "
                "Mother: What about food? "
                "Assistant: We provide pizza and drinks. If you want a cake, you'll need to bring your own. "
                "Mother: That's fine. Do the children need anything? "
                "Assistant: They must wear our special grip socks, which cost two pounds a pair, and every child needs a form signed by a parent. "
                "Mother: Can parents send the forms in advance? "
                "Assistant: Yes, they can complete them on our website. "
                "Mother: Great. "
                "Assistant: Could I have your name? "
                "Mother: Alison Greaves. G, R, E, A, V, E, S. "
                "Assistant: We'll need a deposit of fifty pounds to confirm."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Jump Zone: party booking',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Daughter’s age: ________', ['7', 'seven'], 'Mother: Seven.'),
                    ('Date: Saturday ________ November', ['9', '9th', 'ninth'], 'the ninth of November'),
                    ('Time: ________ p.m.', ['2', 'two'], 'Two is fine'),
                    ('Package price: £________', ['180'], 'a hundred and eighty pounds'),
                    ('Time in the party room: ________ minutes', ['45', 'forty-five'], 'forty-five minutes in the party room'),
                    ('Food provided: pizza and ________', ['drinks'], 'pizza and drinks'),
                    ('Parents bring: a ________', ['cake'], 'If you want a cake'),
                    ('Children must wear grip ________', ['socks'], 'grip socks'),
                    ('Forms can be completed on the ________', ['website'], 'on our website'),
                    ('Name: Alison ________', ['Greaves'], 'G, R, E, A, V, E, S'),
                ])],
        ),
        dict(
            setting='You will hear a radio interview about a new city cycle hire scheme.',
            script=(
                "Presenter: Now, with me is Laura Kent, who manages the city's new cycle hire scheme, CityRide. "
                "Laura: Thanks. The scheme launches next Monday, with five hundred bikes at sixty docking stations across the city. "
                "Half of the bikes are electric, which should help people with the hills in the north of the city. "
                "Using the scheme is simple. You download our app, and you can unlock a bike by scanning a code on the handlebars. "
                "The first thirty minutes of every ride are free for members, and after that it costs one pound for each additional half hour. "
                "Annual membership is ninety pounds, but students pay half price. "
                "When you finish, you must return the bike to a docking station; if you leave it anywhere else, you'll be charged a fine of twenty pounds. "
                "Presenter: What about safety? "
                "Laura: Every bike has lights that come on automatically, and we strongly recommend wearing a helmet, although it isn't compulsory. "
                "We'll also be running free cycling lessons for adults at weekends. "
                "Presenter: Where will the stations be? "
                "Laura: The largest station is outside the central railway station, with fifty bikes. There are also stations at all three universities and at the main hospital. "
                "We did consider putting one at the airport, but it's too far out, so we decided against it. "
                "Presenter: And how will you know if the scheme is working? "
                "Laura: We'll judge it mainly by how many car journeys it replaces, which we'll measure through a user survey."),
            groups=[
                dict(type_id='ielts-listening-completion', heading='CityRide cycle hire',
                     instructions='Complete the notes below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                     questions=[
                         ('Number of bikes: ________', ['500'], 'five hundred bikes'),
                         ('Proportion that are electric: ________', ['half'], 'Half of the bikes are electric'),
                         ('Unlock a bike by scanning a ________', ['code'], 'scanning a code'),
                         ('Free period for members: ________ minutes', ['30', 'thirty'], 'The first thirty minutes'),
                         ('Annual membership: £________', ['90', 'ninety'], 'Annual membership is ninety pounds'),
                     ]),
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What happens if a bike is not returned to a docking station?', ABC('The user pays a fine.', 'The account is closed.', 'The bike is locked.'), 'A', 'a fine of twenty pounds'),
                         ('What does Laura say about helmets?', ABC('They are compulsory.', 'They are recommended.', 'They are provided.'), 'B', 'we strongly recommend wearing a helmet'),
                         ('What will be offered at weekends?', ABC('free rides', 'free lessons', 'guided tours'), 'B', 'free cycling lessons for adults'),
                         ('Where is the largest docking station?', ABC('at a university', 'at the hospital', 'at the railway station'), 'C', 'outside the central railway station'),
                         ('How will the success of the scheme mainly be judged?', ABC('by the number of members', 'by the number of car journeys replaced', 'by its income'), 'B', 'how many car journeys it replaces'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two engineering students, Lucy and Omar, discussing a design competition.',
            script=(
                "Lucy: Omar, the competition deadline is getting close. How's the water filter design coming along? "
                "Omar: The prototype works, but it's slower than we hoped. It takes about two hours to filter a litre. "
                "Lucy: The judges said the most important thing was cost, not speed, didn't they? "
                "Omar: That's right. They want something families can afford, so our target is under five dollars per filter. "
                "Lucy: And are we within that? "
                "Omar: Just. The most expensive part is the ceramic container. "
                "Lucy: I wonder if we could use a plastic bottle instead. "
                "Omar: I thought of that, but plastic can release chemicals in hot weather. "
                "Lucy: OK, we'll keep the ceramic. What did the tests show about the water quality? "
                "Omar: It removed almost all the bacteria, which is the main thing. "
                "Lucy: Brilliant. How are we presenting it to the judges? "
                "Omar: We have ten minutes, and they want a live demonstration, so we'll need to bring dirty water with us. "
                "Lucy: Let's go through the problems we had and what we did about each. The first was the sand clogging up. "
                "Omar: We solved that by adding a layer of gravel on top. "
                "Lucy: Then the bad taste. "
                "Omar: We added charcoal, which fixed it. "
                "Lucy: The container cracking? "
                "Omar: We switched to a thicker clay. "
                "Lucy: And the leaking lid? "
                "Omar: We fitted a rubber seal. "
                "Lucy: And finally, the instructions being confusing. "
                "Omar: We replaced the text with pictures, so they work in any language."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='', instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What is the problem with the prototype?', ABC('It is too expensive.', 'It is too slow.', 'It is too heavy.'), 'B', "it's slower than we hoped"),
                         ('What did the judges say was most important?', ABC('speed', 'cost', 'appearance'), 'B', 'the most important thing was cost'),
                         ('Why will they not use a plastic bottle?', ABC('It could release chemicals.', 'It is too fragile.', 'It is more expensive.'), 'A', 'plastic can release chemicals'),
                         ('What did the tests show?', ABC('The filter removed most bacteria.', 'The filter improved the taste.', 'The filter removed chemicals.'), 'A', 'removed almost all the bacteria'),
                         ('What must they do at the presentation?', ABC('show a video', 'give a live demonstration', 'hand in a written report'), 'B', 'they want a live demonstration'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Problems and solutions',
                     instructions='How did the students solve each problem? Choose the correct letter, A-F.',
                     questions=[
                         ('the sand clogging up', T10_FIXES, 'A', 'adding a layer of gravel'),
                         ('the bad taste', T10_FIXES, 'B', 'We added charcoal'),
                         ('the container cracking', T10_FIXES, 'C', 'a thicker clay'),
                         ('the leaking lid', T10_FIXES, 'D', 'a rubber seal'),
                         ('the confusing instructions', T10_FIXES, 'E', 'replaced the text with pictures'),
                     ]),
            ],
        ),
        PART4['t10'],
    ],
)]
