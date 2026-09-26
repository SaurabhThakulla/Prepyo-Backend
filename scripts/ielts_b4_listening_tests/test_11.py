"""IELTS Listening Practice Test 11. Original Prepyo practice material."""
from ielts_b4_listening_tests.common import ABC, NOTES_W1, form, matching, mcq

KIT = ABC('It is provided by the centre.', 'It must have the child\'s name on it.',
          'A used one is better than a new one.', 'It is sold at the school office.',
          'Children can decide whether to bring it.', 'It should be packed in the day bag.')
SOURCES = ABC('It contradicts the logbooks.', 'It is too damaged to use.', 'It shows the pupils\' point of view.',
              'It can only be seen online.', 'It includes photographs.', 'It covers only a short period.')

TEST = dict(
    key='t11', title='IELTS Listening Practice Test 11',
    parts=[
        dict(
            setting='You will hear a woman phoning a homestay agency about hosting an international student.',
            script=(
                "Adviser: Good morning, Norwich Homestay Service, Sophie speaking. "
                "Woman: Oh, hello. I'm ringing because I'm interested in having an international student to stay. A friend of mine does it and she's always saying how much she enjoys it. "
                "Adviser: That's great. We're always looking for new hosts, especially at the moment. Can I take some details and set up a registration form for you? "
                "Woman: Yes, of course. "
                "Adviser: Could I have your name first? "
                "Woman: It's Linda Carrington. "
                "Adviser: Is that C, A, double R? "
                "Woman: That's right. C, A, double R, I, N, G, T, O, N. "
                "Adviser: And your address? "
                "Woman: Fourteen... no, sorry, we've only been there a year and I still say it the wrong way round. Forty-one, Thorpe Road. "
                "Adviser: Forty-one Thorpe Road. And how many people live in the house? "
                "Woman: There's me, my husband and our son, who's twelve. Oh, and my father-in-law has been living with us since the spring, so that makes four. "
                "Adviser: Four, then. Any pets? Some students are allergic, so we always ask. "
                "Woman: We used to have two dogs, but we don't now. We've just got a parrot. He's very noisy, I should warn you, but he stays in the kitchen. "
                "Adviser: I'll put down parrot. Now, can you tell me a little about the room the student would have? "
                "Woman: It's a good-sized room at the back of the house, with a desk and a wardrobe. It hasn't got a TV, but it does have its own shower, so they wouldn't have to share the bathroom with our son. "
                "Adviser: Students love that. How far are you from the college? "
                "Woman: It's quite a long walk, forty minutes maybe, but there's a bus stop at the end of the road and the bus takes about twenty minutes. "
                "Adviser: That's fine. We usually ask hosts to take students for a minimum of four weeks. Is that all right? "
                "Woman: Actually, we'd rather have someone for longer, so they feel part of the family. Could we say at least twelve weeks? "
                "Adviser: Certainly. Some of our students stay for a whole term, so that's no problem. Do you have any house rules we should tell students about? "
                "Woman: Not really. We're quite relaxed about what time they come in, as long as they let us know. The only thing we're strict about is smoking. We don't allow it anywhere, not even in the garden. "
                "Adviser: Understood. Next, we need a reference from someone who's known you for at least two years. It can't be a relative. "
                "Woman: My neighbour's known me for ages... but perhaps it'd be better to ask my manager at work. I've been at the library for six years. "
                "Adviser: Your manager would be ideal. And the last thing is a home visit. One of us comes to see the house before we place a student. Could you do Tuesday the ninth of May? "
                "Woman: I'm working all day on Tuesday. What about the Thursday, the eleventh? I'm free all afternoon. "
                "Adviser: Thursday the eleventh is fine. I'll send you an email to confirm it. "
                "Woman: Lovely. Thanks for your help."),
            groups=[form('Homestay host registration form', [
                ('Surname: Linda ________', ['Carrington'], 'C, A, double R, I, N, G, T, O, N'),
                ('Address: ________ Thorpe Road', ['41', 'forty-one'], 'Forty-one, Thorpe Road'),
                ('Number of people in the household: ________', ['4', 'four'], 'so that makes four'),
                ('Pet: a ________', ['parrot'], "We've just got a parrot"),
                ("Student's room has its own ________", ['shower'], 'it does have its own shower'),
                ('Journey to college: about ________ minutes by bus', ['20', 'twenty'], 'the bus takes about twenty minutes'),
                ('Length of stay wanted: at least ________ weeks', ['12', 'twelve'], 'at least twelve weeks'),
                ('House rule: no ________', ['smoking'], "The only thing we're strict about is smoking"),
                ('Reference from: her ________', ['manager'], 'ask my manager at work'),
                ('Home visit: Thursday ________ May', ['11', '11th', 'eleventh'], 'Thursday the eleventh is fine'),
            ])],
        ),
        dict(
            setting='You will hear the head of Year 7 at a secondary school talking to parents about a residential trip.',
            script=(
                "Mr Hughes: Good evening, everyone, and thanks for coming out on such a wet night. For those who don't know me, I'm Mr Hughes, head of Year 7, and I'm going to tell you about the residential trip to Wales in June. I'll leave plenty of time for questions at the end. "
                "First, the dates. Some of you will have seen the letter that said the last week of June, and a few parents have asked me whether we'd moved it to get away from the end-of-year tests. That's not the reason. The centre we use is having its roof repaired that week, so they couldn't take us. We'll now be going from Monday the twelfth to Friday the sixteenth of June. "
                "The cost is the same as last year, two hundred and ten pounds, and that covers the coach, all meals and every activity. There's a small shop at the centre that sells sweets and postcards, so children can bring some pocket money, but please keep it to fifteen pounds at most. They look after it themselves; staff won't be collecting it in. "
                "I know mobile phones are a big question. In the past we've let children keep their phones and use them for an hour in the evenings, but we found a lot of them were up half the night messaging each other, and they were worn out by Wednesday. So this year phones are to stay at home. If you need to get in touch, ring the centre office and they'll pass a message on, and we'll put photos on the school website every evening. "
                "As for the activities, most of them will be familiar to anyone whose older child has been. The night walk is always the favourite, and that's staying, as are the canoeing and the climbing. The new thing this year is a morning on a working farm down the valley, where the children will help feed the animals. "
                "Now, paperwork. Nearly all of you have paid the deposit, thank you. What I really need by this Friday is the medical form. It's online this year, not on paper, and we can't take any child without one. "
                "Right, the kit list. I'll go through the items people ask about most. "
                "Waterproof trousers are essential, because the children will be in and out of rivers. Please don't go out and buy expensive ones. The school office sells a basic pair for eight pounds, and they're perfectly good. "
                "Last year's list said to pack a torch. You can cross that off. The centre hands out head torches for the night walk, and they're much better than anything the children would bring from home. "
                "Walking boots. It's tempting to buy a new pair specially, but new boots give blisters. A pair of trainers your child has been wearing for months is far better than boots straight out of the box. "
                "Sleeping bags. The centre has beds and pillows, but children need their own sleeping bag, and they all look the same when they're rolled up, so please write your child's name on a label and sew it inside. "
                "And finally, water bottles. The suitcases travel separately in a van and don't arrive until the evening, so the water bottle must go in the small day bag they carry on the coach, not in the suitcase. "
                "That's the kit. Now, does anyone have any questions?"),
            groups=[
                mcq([
                    ('Why has the date of the trip changed?', ABC('to avoid the end-of-year tests', 'because of building work at the centre', 'because the coach company was not available'), 'B', 'having its roof repaired'),
                    ('What does the speaker say about pocket money?', ABC('It should be no more than £15.', 'Children should not bring any.', 'It will be looked after by staff.'), 'A', 'please keep it to fifteen pounds at most'),
                    ('What is the rule about mobile phones this year?', ABC('Children may use them in the evening.', 'They must be handed in on the coach.', 'They must be left at home.'), 'C', 'this year phones are to stay at home'),
                    ('What is new on this year\'s trip?', ABC('a morning on a farm', 'the night walk', 'canoeing'), 'A', 'The new thing this year is a morning on a working farm'),
                    ('What must parents do by Friday?', ABC('pay the deposit', 'complete a medical form', 'return a paper consent form'), 'B', 'What I really need by this Friday is the medical form'),
                ]),
                matching('Kit list', 'What does the speaker say about each item on the kit list? Choose the correct letter, A-F.', [
                    ('waterproof trousers', KIT, 'D', 'The school office sells a basic pair'),
                    ('a torch', KIT, 'A', 'The centre hands out head torches'),
                    ('walking boots', KIT, 'C', 'far better than boots straight out of the box'),
                    ('a sleeping bag', KIT, 'B', "write your child's name on a label"),
                    ('a water bottle', KIT, 'F', 'must go in the small day bag'),
                ]),
            ],
        ),
        dict(
            setting='You will hear two history students, Isla and Kwame, discussing their project on Victorian school logbooks with their tutor.',
            script=(
                "Tutor: Come in, both of you. So, Isla, Kwame, you're looking at school logbooks. Remind me how you settled on that. "
                "Isla: Well, you did mention logbooks in one of your lectures, but that's not really what started it. Kwame spent last summer working at the county record office, cataloguing documents, and he kept telling me about them. "
                "Kwame: I was putting labels on boxes of them for six weeks, so I got quite curious. From 1862 every head teacher in a state-funded school had to keep one. It's a sort of daily diary of the school. "
                "Tutor: And what have you found so far? Anything unexpected? "
                "Kwame: We expected a lot about children missing school to help on farms, and there's plenty of that, especially at harvest time. But what really surprised us was the weather. Almost every entry starts with it. 'Heavy snow, only twenty present', that sort of thing. "
                "Isla: Which actually turns out to be useful, because it explains a lot of the attendance figures. "
                "Tutor: Good. Any practical problems? "
                "Isla: One big one. We'd assumed we could photograph the pages and work from the pictures at home, but we weren't allowed to take photographs, because some of the books are fragile. So everything has to be copied out by hand in the reading room. "
                "Tutor: That's slow work. Which brings me to something I wanted to raise. You've said you're going to cover twelve schools. "
                "Kwame: That's right, all the village schools in one district. "
                "Tutor: I think that's too many for the time you've got. I'd rather you chose four or five and looked at them in real depth. "
                "Isla: That would be a relief, to be honest. "
                "Tutor: And how are you going to show the attendance figures? You won't want pages of numbers. "
                "Kwame: We'd planned a table for each school, but you're right, it's a lot of numbers. So we're going to use a line graph for each school instead, showing attendance month by month, so the dips at harvest time stand out. "
                "Tutor: Much clearer. Now, you're also using other sources alongside the logbooks. How are they going? "
                "Isla: Mixed. The inspectors' reports are interesting, because they often don't agree with the logbooks. An inspector would write that attendance was satisfactory in a week when the head teacher had written that half the class was absent. "
                "Kwame: Perhaps the inspectors only came on good days. "
                "Tutor: Possibly. What else? "
                "Kwame: There's a memoir by a woman who went to one of the schools in the 1880s. It's the only source we've got that describes school from the pupils' side: what they ate, the games they played, how they felt about the teachers. "
                "Isla: Then the census. We'd hoped to see the original books at the record office, but they're only available online now, which is actually easier for us. "
                "Kwame: The local newspaper was disappointing. The library has copies, but for our area only a few years, from 1885 to 1889, have survived, so it doesn't cover most of our period. "
                "Isla: And the school board minute books. We were really looking forward to those, but they've been badly damaged by water at some point, and most of the pages can't be read at all. "
                "Tutor: That's a shame. Still, you've got plenty to work with."),
            groups=[
                mcq([
                    ('Why did the students choose this topic?', ABC('The tutor mentioned it in a lecture.', 'Kwame had worked with the logbooks.', 'Isla found a logbook at home.'), 'B', 'Kwame spent last summer working at the county record office'),
                    ('What surprised the students most about the logbooks?', ABC('how much they say about the weather', 'how often children missed school at harvest time', 'how short most entries are'), 'A', 'what really surprised us was the weather'),
                    ('What problem did the students have at the record office?', ABC('Some books were missing.', 'The reading room was often closed.', 'They could not take photographs.'), 'C', "we weren't allowed to take photographs"),
                    ('What does the tutor advise the students to do?', ABC('include schools in towns', 'study fewer schools', 'spend more time at the record office'), 'B', "I'd rather you chose four or five"),
                    ('How will the students present the attendance figures?', ABC('in a table', 'on a map', 'in a line graph'), 'C', 'use a line graph for each school'),
                ]),
                matching('Other sources', 'What do the students say about each source? Choose the correct letter, A-F.', [
                    ("the inspectors' reports", SOURCES, 'A', "they often don't agree with the logbooks"),
                    ("a former pupil's memoir", SOURCES, 'C', "describes school from the pupils' side"),
                    ('the census', SOURCES, 'D', "they're only available online now"),
                    ('the local newspaper', SOURCES, 'F', 'only a few years, from 1885 to 1889, have survived'),
                    ('the school board minute books', SOURCES, 'B', "most of the pages can't be read at all"),
                ]),
            ],
        ),
        dict(
            setting='You will hear part of a lecture about lichens.',
            script=(
                "Lecturer: This morning I want to look at a group of organisms that most people walk past every day without noticing: lichens. You'll have seen them as grey-green crusts on walls, or yellow patches on roofs. People often call them mosses, but they're not plants at all. A lichen is actually a partnership between two quite different organisms: a fungus and an alga, or in some species a type of bacterium that can use light in the same way. "
                "Each partner gets something from the arrangement. The fungus forms the body of the lichen, and protects its partner from strong sunlight and from drying out. The alga, living inside, produces sugars by photosynthesis, and shares them with the fungus, which can't make its own food. "
                "One of the most remarkable things about lichens is how well they cope with dry conditions. A lichen can lose almost all of its water and simply shut down, sometimes for months. Then, within minutes of rain, it starts working again as if nothing had happened. "
                "The price of this toughness is that they grow very slowly. Many species grow by less than a millimetre a year. Because of this, and because the growth rate of a particular species is fairly steady, researchers have measured lichens to estimate the age of old stone walls in the north of England, where no written records exist. "
                "Lichens are also very good at telling us about the air. They have no roots, so they take in everything they need, water and minerals, directly from the air and the rain. That means they absorb pollution too. They're especially sensitive to sulphur dioxide, which comes from burning coal, and in the nineteenth century lichens disappeared from the centres of most industrial cities in Britain. Since the Clean Air Acts they've been coming back, which is good news. "
                "However, there's a newer problem. Some bright orange lichens actually do well where there's a lot of nitrogen in the air, from vehicle exhaust and from farming. So if you see trees covered in orange lichen, that's often a sign of high levels of nitrogen, not of clean air. "
                "People have found many uses for lichens. In the far north, reindeer depend on them in winter, digging through the snow to reach them. In Scotland, lichens were traditionally used to dye wool, and some of the colours in Harris tweed came from lichens scraped off rocks. And if you did chemistry at school, you've already used a lichen product: litmus, the dye that turns red in acid and blue in alkali, was originally made from lichens. "
                "Finally, lichens are among the toughest living things we know of. In 2005, samples of two species were sent into space and left exposed outside a spacecraft for fifteen days. When they came back to Earth, they were still alive and able to photosynthesise normally."),
            groups=[form('Lichens', [
                ('A lichen is a partnership between a fungus and an ________', ['alga'], 'a fungus and an alga'),
                ('The alga produces sugars by ________', ['photosynthesis'], 'produces sugars by photosynthesis'),
                ('A dried-out lichen starts working again within minutes of ________', ['rain'], 'within minutes of rain'),
                ('Growth has been used to estimate the age of old stone ________', ['walls'], 'old stone walls'),
                ('Lichens are especially sensitive to sulphur ________', ['dioxide'], 'sensitive to sulphur dioxide'),
                ('Orange lichens on trees can show high levels of ________', ['nitrogen'], 'a sign of high levels of nitrogen'),
                ('In the far north, ________ feed on lichens in winter', ['reindeer'], 'reindeer depend on them in winter'),
                ('Lichens gave some of the colours in Harris ________', ['tweed'], 'Harris tweed'),
                ('Lichens were the original source of ________', ['litmus'], 'litmus, the dye'),
                ('In 2005 lichens survived fifteen days in ________', ['space'], 'sent into space'),
            ], instructions=NOTES_W1)],
        ),
    ],
)
