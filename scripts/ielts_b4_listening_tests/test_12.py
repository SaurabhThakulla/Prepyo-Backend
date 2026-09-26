"""IELTS Listening Practice Test 12. Original Prepyo practice material."""
from ielts_b4_listening_tests.common import ABC, MAP_OPTIONS, NOTES_WN, PLAN, form, matching, mcq

WALKS = ABC('It is suitable for young children.', 'You will need a map and compass.', 'Part of the route is closed.',
            'It includes a steep climb.', 'It starts from the hostel.', 'It can be combined with a boat trip.',
            'It is best done early in the morning.')
WORDS = ABC('It is mainly used by older people.', 'It has spread through social media.', 'Teachers often correct it.',
            'It is used in only one of the two towns.', 'None of the participants knew it.', 'Its meaning has changed.')

TEST = dict(
    key='t12', title='IELTS Listening Practice Test 12',
    parts=[
        dict(
            setting='You will hear a tenant phoning a letting agency to report problems in his flat.',
            script=(
                "Agent: Good morning, Harwood Lettings, Grace speaking. "
                "Man: Oh, hi. I'm one of your tenants, and I've got a problem in my flat that I need to report. "
                "Agent: Sorry to hear that. I'll fill in a repair request for you. Can I start with your name? "
                "Man: Yes, it's Adam Fenwick. "
                "Agent: How do you spell your surname? "
                "Man: F, E, N, W, I, C, K. "
                "Agent: Thanks. And the address? "
                "Man: Flat 6, Albion Court. It's the block on Victoria Road. "
                "Agent: Albion Court, yes, we look after several flats there. So what's the problem? "
                "Man: It's the boiler, I think. The heating's working fine, the radiators are warm, but it's the hot water. There just isn't any. I've been boiling the kettle to wash. "
                "Agent: Oh dear. When did it start? "
                "Man: It went off on Saturday... no, hang on, it was Sunday morning, because I'd just come back from a run and wanted a shower. "
                "Agent: Have you tried resetting the boiler? "
                "Man: Twice. There's a little red light flashing on the front. "
                "Agent: I'll note that down, it helps the engineer. Is there anything else while I've got you? "
                "Man: Actually, yes. There's a small leak under the kitchen sink. Not the bathroom, that's fine. It's only a drip, but the floor of the cupboard is getting soft. "
                "Agent: We'll get that looked at on the same visit. Anything else? "
                "Man: The blind in the bedroom is broken. The cord snapped and it won't go up. It's not urgent, though. "
                "Agent: I'll add it anyway. Now, the engineer will need to get in. When are you usually at home? "
                "Man: I work until three, and it takes me about forty minutes to get back, so any time after four would be fine. "
                "Agent: And if you're not in, does anyone else have a key? "
                "Man: My neighbour has one, but she's away at the moment. The caretaker has a spare key, though. He's in the building on weekdays. "
                "Agent: That's useful. What about parking? Engineers carry a lot of tools. "
                "Man: There's no parking on the road at the front, it's all yellow lines. But there's a yard behind the building, and he can park in the yard there. You go in through the gate next to the newsagent's. "
                "Agent: Great. And how would you like us to contact you about the appointment? By email? "
                "Man: I don't really check emails during the day. A text would be better. "
                "Agent: No problem. We'll send you a time this afternoon."),
            groups=[form('Harwood Lettings: repair request', [
                ("Tenant's name: Adam ________", ['Fenwick'], 'F, E, N, W, I, C, K'),
                ('Address: Flat 6, ________ Court', ['Albion'], 'Flat 6, Albion Court'),
                ('Main problem: no hot ________', ['water'], "it's the hot water"),
                ('Problem started: ________ morning', ['Sunday'], 'it was Sunday morning'),
                ('Also: a leak under the kitchen ________', ['sink'], 'under the kitchen sink'),
                ('Also: a broken ________ in the bedroom', ['blind'], 'The blind in the bedroom'),
                ('Best time for a visit: after ________ pm', ['4', 'four'], 'any time after four'),
                ('Spare key: ask the ________', ['caretaker'], 'The caretaker has a spare key'),
                ('Engineer can park in the ________ behind the building', ['yard'], 'park in the yard'),
                ('Contact tenant by: ________', ['text'], 'A text would be better'),
            ])],
        ),
        dict(
            setting='You will hear the manager of a youth hostel in the Lake District talking to guests on their first evening.',
            script=(
                "Manager: Evening, everyone. I'm Rob, and I run the hostel with my partner, Jen. I'll keep this short because I know you'll want your dinner. First, a few walks that people always ask about, and then I'll take you round the building using the plan you've been given. "
                "Right, walks. The lake shore path is the one most people do first. It looks easy on the map, and it is flat, but it's ten miles right round, so it's not one for small children. The nice thing is that you don't have to do the whole thing: you can walk half way and catch the launch back into town from one of the jetties. "
                "Catbells is probably the best-known hill round here, and you'll hear it described as a family walk. Don't be fooled. It's short, but near the top there's a steep, rocky section where you'll need your hands as well as your feet. "
                "If you'd like something gentler, Castlerigg Stone Circle is lovely in the evening light. You don't need to drive; you can walk there straight from our front door in about forty minutes. "
                "Walla Crag is usually a favourite, but I should warn you that the path up beside the stream is closed at the moment, because part of the bank slipped in the winter. You can still get to the top, but you'll have to go the longer way round through the woods. "
                "And Skiddaw. It's a big mountain, over nine hundred metres, and the weather at the top can be completely different from down here. The cloud comes down fast, so don't go up without a map and compass, and make sure you know how to use them. Phones are no good up there. "
                "OK, now the building. If you look at the plan, the entrance is at the bottom, which is south. There's a main corridor running north, and another corridor crossing it in the middle. The stairs and the lounge are already marked for you. "
                "We moved a few things round during the building work last year, so even if you've stayed with us before, please have a look. The self-catering kitchen used to be by the entrance, but it's now the room in the north-west corner, beside the stairs. "
                "When you come back soaked, and you will, the drying room is the room on the east side just south of the cross corridor. Leave boots and wet coats in there, not in your bedroom, please. The room opposite, on the other side of the main corridor, is our office, and that's private, I'm afraid. "
                "The games room, with table tennis and a shelf of board games, can be found in the room in the north-east corner, next to the lounge, so it can get a bit noisy round there in the evenings. "
                "A lot of people expect the laundry to be upstairs, but it's down here: the laundry's the room in the south-west corner, on your left as you come through the door. The machines take coins or cards. "
                "And if you've come by bike, the bike store is the room in the south-east corner. The door has a code, which is printed on your key card. Right, that's everything. Dinner's in ten minutes."),
            groups=[
                matching('Local walks', 'What does the manager say about each walk? Choose the correct letter, A-G.', [
                    ('the lake shore path', WALKS, 'F', 'catch the launch back into town'),
                    ('Catbells', WALKS, 'D', 'a steep, rocky section'),
                    ('Castlerigg Stone Circle', WALKS, 'E', 'straight from our front door'),
                    ('Walla Crag', WALKS, 'C', 'is closed at the moment'),
                    ('Skiddaw', WALKS, 'B', "don't go up without a map and compass"),
                ]),
                dict(type_id='ielts-listening-map', heading='Hostel: ground floor', instructions=PLAN,
                     image=('Hostel: ground floor', {('W', 3): 'Stairs', ('E', 3): 'Lounge'},
                            [('W', 1), ('E', 1), ('W', 2), ('E', 2), ('W', 4), ('E', 4)]),
                     questions=[
                         ('Self-catering kitchen', MAP_OPTIONS, 'E', "it's now the room in the north-west corner"),
                         ('Drying room', MAP_OPTIONS, 'D', 'the drying room is the room on the east side just south of the cross corridor'),
                         ('Games room', MAP_OPTIONS, 'F', 'can be found in the room in the north-east corner'),
                         ('Laundry', MAP_OPTIONS, 'A', "the laundry's the room in the south-west corner"),
                         ('Bike store', MAP_OPTIONS, 'B', 'the bike store is the room in the south-east corner'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two linguistics students, Chloe and Arjun, discussing their project on dialect words with their tutor.',
            script=(
                "Tutor: Hello, Chloe, Arjun. Take a seat. I've read your draft, so let's talk it through. Can you start by reminding me what the project set out to do? "
                "Chloe: Sure. There's quite a lot of research on the words older people use in the north of England, and on accents, but we wanted to find out whether teenagers still use local dialect words at all, or whether they're disappearing. "
                "Arjun: We chose two towns, Bradford and Burnley, because they're only about twenty-five miles apart but on different sides of the Pennines. "
                "Tutor: And how did you collect your data? I remember you were planning interviews. "
                "Arjun: We were, and we tried a few. But as soon as the recorder was on, people got self-conscious and started speaking very correctly. So in the end we showed them pictures, an alley, a bread roll, someone shivering, and asked what they'd call each thing. "
                "Tutor: That's a sensible change. What came out of it? "
                "Chloe: Well, one example is 'ginnel', which is an old word for a narrow passage between houses. Nearly all the teenagers knew what it meant, but most of them said they'd never actually say it themselves. They'd just say 'alley'. "
                "Tutor: That's a common pattern. Now, I do have one worry, and it isn't the size of the sample. Sixty participants is fine for a project like this. It's that all of them came from the same school in each town. "
                "Arjun: We did wonder about that. "
                "Tutor: It means you can't tell whether you're seeing a difference between towns or between schools. You'll need to go to at least one more school in each town. "
                "Chloe: OK. Can we just arrange that with the schools directly? "
                "Tutor: Not yet. Before you visit anywhere new, you'll have to go back to the ethics committee for permission, because the approval you've got names the two schools. The schools themselves will deal with the consent forms for parents, as before. "
                "Arjun: Right. We'll send that off this week. "
                "Tutor: Good. Now, talk me through the individual words in your results section. "
                "Chloe: The most interesting one is 'mardy', meaning bad-tempered. We expected it to be dying out, but it's everywhere. It's been in a lot of videos online, and teenagers in both towns use it, even ones whose families aren't from the north. "
                "Arjun: We wondered whether its meaning had changed too, but no, it means exactly what it always did. "
                "Tutor: Interesting. What about 'nesh'? "
                "Arjun: That means someone who feels the cold easily. Teenagers recognised it, but only because their grandparents say it. Nearly everyone who uses it is over sixty. "
                "Chloe: Then there's 'barm', a bread roll. That's a clear example of the Pennines as a boundary. Every teenager in Burnley used it, but in Bradford nobody did. They say 'bread cake'. "
                "Arjun: We also included 'spice', which is an old Yorkshire word for sweets. We thought a few people would know it. In fact none of the sixty had ever heard it used like that. "
                "Chloe: And the last one is 'while', used to mean 'until', as in 'wait while four o'clock'. Lots of teenagers in Bradford say it, but several told us their teachers correct it every time they use it in class. "
                "Tutor: That's a lovely detail. Put that in your discussion."),
            groups=[
                mcq([
                    ('What was the main aim of the students\' project?', ABC('to record the words older people use', 'to find out if teenagers use local dialect words', 'to compare accents on each side of the Pennines'), 'B', 'whether teenagers still use local dialect words'),
                    ('How did the students finally collect their data?', ABC('recorded interviews', 'a written questionnaire', 'a picture-naming task'), 'C', 'we showed them pictures'),
                    ('What did the students find about the word "ginnel"?', ABC('Most teenagers knew it but did not use it.', 'It was used more in Bradford than in Burnley.', 'Only a few teenagers understood it.'), 'A', "they'd never actually say it themselves"),
                    ('What is the tutor\'s main concern about the study?', ABC('the number of participants', 'the choice of schools', 'the choice of words'), 'B', 'all of them came from the same school'),
                    ('What must the students do before visiting another school?', ABC('apply to the ethics committee', 'collect consent forms from parents', 'ask the tutor to contact the schools'), 'A', 'go back to the ethics committee'),
                ]),
                matching('Dialect words', 'What do the students say about each word? Choose the correct letter, A-F.', [
                    ('mardy', WORDS, 'B', "It's been in a lot of videos online"),
                    ('nesh', WORDS, 'A', 'Nearly everyone who uses it is over sixty'),
                    ('barm', WORDS, 'D', 'in Bradford nobody did'),
                    ('spice', WORDS, 'E', 'none of the sixty had ever heard it'),
                    ('while', WORDS, 'C', 'their teachers correct it every time'),
                ]),
            ],
        ),
        dict(
            setting='You will hear part of a lecture about the early history of weather forecasting.',
            script=(
                "Lecturer: Today we're going to look at how weather forecasting began, and I want to start with a man called Robert FitzRoy. Some of you may know the name already, because as a young naval officer he was captain of the Beagle, the ship that took Charles Darwin round the world. But it's his later work that concerns us today. "
                "In 1854 FitzRoy was put in charge of a small new government department, set up to collect weather observations from ships at sea. The idea at first was simply to gather data. FitzRoy, however, wanted to go further and actually predict the weather, and a severe storm in 1859, which wrecked many ships around the British coast, convinced him that warnings could save lives. "
                "What made this possible was a new technology: the electric telegraph. For the first time, reports of wind and air pressure from stations around the coast could reach London within an hour or so, fast enough to be useful. "
                "FitzRoy also needed a name for what he was doing. He didn't like the word 'prophecy', which sounded unscientific, so he chose the word 'forecast', which we still use today. "
                "His first service was storm warnings for shipping. When a storm was expected, a large canvas cone was raised on a mast at the harbour, pointing up for a gale from the north and down for one from the south, so that crews could see it before they set out. "
                "Then, in August 1861, the first public weather forecasts appeared in The Times newspaper. They were very popular with readers, but they were often wrong, and FitzRoy was frequently mocked in the press when rain fell on a day he'd promised would be fine. "
                "In 1866 a committee of scientists examined FitzRoy's methods and concluded that there was no sound scientific basis for them, and the daily forecasts were stopped. The storm warnings were stopped as well. But that decision was unpopular, and the warnings were brought back in 1867 after complaints from fishermen and ship owners, who had found them genuinely useful. "
                "Forecasting developed slowly after that. One important step was the shipping forecast, which has been broadcast on radio since the 1920s, and which many people in Britain still listen to, even if they never go to sea. "
                "The real change came with mathematics. In 1922 a British scientist, Lewis Fry Richardson, published a method for calculating the weather using equations. He tested it himself, working by hand, and it took him about six weeks to produce a forecast for just six hours ahead. The forecast was wrong, but the method was right. Richardson imagined a huge hall with sixty-four thousand people doing calculations together, fast enough to keep up with the weather. "
                "In the end, that hall was replaced by a machine. The first forecast made by computer was produced in 1950, in the United States. It took about twenty-four hours to calculate a forecast for twenty-four hours ahead, so it wasn't much use in practice, but it showed what was possible. "
                "Since then, forecasts have steadily improved. Today a four-day forecast is about as accurate as a one-day forecast was in the 1980s."),
            groups=[form('The beginnings of weather forecasting', [
                ('Reports reached London quickly by the electric ________', ['telegraph'], 'the electric telegraph'),
                ('FitzRoy rejected the word "________" as unscientific', ['prophecy'], "He didn't like the word 'prophecy'"),
                ('Storm warning: a ________ raised on a mast at the harbour', ['cone'], 'a large canvas cone was raised'),
                ('First public forecasts in a newspaper: August ________', ['1861'], 'in August 1861'),
                ('1866: a committee of ________ found no sound basis for the methods', ['scientists'], 'a committee of scientists'),
                ('Warnings returned after complaints from fishermen and ship ________', ['owners'], 'fishermen and ship owners'),
                ('The shipping forecast has been broadcast on ________ since the 1920s', ['radio'], 'broadcast on radio'),
                ("Richardson's hand calculation took about six ________", ['weeks'], 'about six weeks'),
                ('The first forecast by ________ was made in 1950', ['computer'], 'The first forecast made by computer'),
                ('Today a ________-day forecast is as accurate as a one-day forecast in the 1980s', ['4', 'four'], 'Today a four-day forecast'),
            ], instructions=NOTES_WN)],
        ),
    ],
)
