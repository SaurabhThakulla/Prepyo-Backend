"""IELTS Listening practice, batch 4: multiple-choice drills.

Original Prepyo scripts in the public IELTS Listening multiple-choice format
(choose ONE letter, A, B or C). Every script mentions the wrong options as
well, as the test does, and `evidence` is the exact words that give the
answer. Not official, recalled or copied test content.

Each item: (title, script, question, [A, B, C], key, evidence).
"""

MCQ = [
    # ------------------------------------------------ sports and leisure
    ('Why the pool closes early on Tuesday',
     "Receptionist: I'm afraid the pool closes at six this Tuesday instead of nine. It's not for cleaning, that happens on Sunday nights, and the lifeguards aren't on strike or anything like that. "
     "It's because the county swimming gala is being held here, and they need the evening to set up the timing equipment.",
     'Why will the pool close early on Tuesday?', ['for cleaning', 'because of a shortage of lifeguards', 'to prepare for a competition'],
     'C', "the county swimming gala is being held here"),
    ('Choosing a leisure centre membership',
     "Man: The gold membership looked good, with the sauna and everything, but I'd never use it enough. And the off-peak one only lets you in before four, which is no good because I work until five. "
     "So I've gone for the standard membership. It's a bit more than off-peak, but I can go any time.",
     'Which membership has the man chosen?', ['standard', 'gold', 'off-peak'], 'A', "I've gone for the standard membership"),
    ('Best class for a returning runner',
     "Coach: If you haven't run for a few years, I wouldn't start with the ten-kilometre group, even though you've done that distance before. And the track sessions are really for people training for races. "
     "What I'd suggest is our Couch to 5K course. It builds up slowly over nine weeks.",
     'What does the coach suggest?', ['a course that builds up slowly', 'the ten-kilometre group', 'track sessions'],
     'A', "What I'd suggest is our Couch to 5K course"),
    ('Problem with the new sports hall floor',
     "Manager: The new floor was laid on time, and it's the right colour, so no complaints there. But when the basketball club started using it, they found the ball didn't bounce evenly in one corner. "
     "It turns out there's a damp patch underneath, so that section will have to come up.",
     'What is wrong with the new floor?', ['It was finished late.', 'Part of it is uneven because of damp.', 'It is the wrong colour.'],
     'B', "there's a damp patch underneath"),
    ('New meeting point for the running club',
     "Club secretary: Just a reminder that from next week we're not meeting in the pub car park any more. The landlord was fine about it, but it's too dark in winter. "
     "Someone suggested the library steps, but they're locked after six. So we'll meet outside the leisure centre, where there's good lighting and we can use the toilets.",
     'Where will the running club meet from next week?', ['the pub car park', 'the library steps', 'outside the leisure centre'],
     'C', "we'll meet outside the leisure centre"),
    ('Golf lessons for beginners',
     "Instructor: Beginners always want to go straight out onto the course, and some ask to start on the driving range. "
     "But actually, for the first two lessons we'll stay on the putting green. Putting is half the game, and it builds your confidence.",
     'Where will the first lessons take place?', ['on the course', 'on the putting green', 'on the driving range'],
     'B', "for the first two lessons we'll stay on the putting green"),

    # ----------------------------------------------------- health services
    ('Reason for a missed appointment',
     "Patient: I'm sorry I missed my appointment yesterday. I didn't forget, and I had the right time. The problem was that the letter gave the old address, "
     "so I went to the surgery on Park Lane, and by the time I got here it was too late.",
     'Why did the patient miss the appointment?', ['He went to the wrong building.', 'He forgot about it.', 'He was given the wrong time.'],
     'A', "the letter gave the old address"),
    ('Advice from the pharmacist',
     "Pharmacist: You could try a nasal spray, but I wouldn't use it for more than a week. Antibiotics won't help with a cold, so there's no point asking the doctor for those. "
     "Honestly, the best thing is plenty of fluids and rest. It should clear up in a few days.",
     'What does the pharmacist recommend?', ['a nasal spray', 'antibiotics', 'fluids and rest'], 'C', "the best thing is plenty of fluids and rest"),
    ('Getting blood test results',
     "Nurse: Your results should be back in about a week. We don't phone patients with results, I'm afraid, and we can't give them out at reception. "
     "You'll be able to see them on the NHS app, and if the doctor needs to talk to you, she'll send you a text.",
     'How will the patient get the results?', ['by phone', 'at reception', 'on an app'], 'C', "You'll be able to see them on the NHS app"),
    ('New service at the health centre',
     "Practice manager: We've always had a nurse-led asthma clinic, and the physiotherapist has been here on Mondays for years. "
     "What's new from this month is a pharmacist based at the surgery, who can review your medicines with you.",
     'What is new at the health centre?', ['an asthma clinic', 'a pharmacist', 'a physiotherapist'], 'B', "What's new from this month is a pharmacist"),
    ('Choosing a dentist appointment',
     "Receptionist: I can offer you Wednesday at nine, or Thursday at four. "
     "Patient: Thursday's no good, I pick up the children then. And Wednesday at nine I'm at work. Is there anything on Saturday morning? "
     "Receptionist: Yes, we've just had a cancellation. Saturday at ten thirty. "
     "Patient: Perfect, I'll take it.",
     'When will the patient see the dentist?', ['Wednesday', 'Thursday', 'Saturday'], 'C', "Saturday at ten thirty. Patient: Perfect, I'll take it"),
    ('Walk-in centre waiting times',
     "Receptionist: It's usually busiest first thing, when people come in before work, and again straight after the schools finish. "
     "If you can, come in the middle of the afternoon, between one and three. That's when the wait is shortest.",
     'When is the waiting time shortest?', ['early morning', 'early afternoon', 'late afternoon'], 'B', "between one and three. That's when the wait is shortest"),

    # --------------------------------------------------- shopping and repairs
    ('Why the phone repair will take longer',
     "Technician: I know we said Tuesday. It's not that we're busy, and the screen itself was easy to fix. But when we opened the phone, we found some water damage inside, "
     "and we need to order a part from the manufacturer for that. So it'll be Friday at the earliest.",
     'Why will the repair take longer?', ['The shop is very busy.', 'A part must be ordered.', 'The screen was hard to fix.'], 'B', "we need to order a part from the manufacturer"),
    ('Refund or exchange',
     "Assistant: Because it's been more than thirty days, I can't give you your money back, I'm afraid. And we don't have that jacket in the larger size any more. "
     "What I can do is give you a credit note for the full price, which you can spend at any of our branches.",
     'What does the assistant offer?', ['a credit note', 'a refund', 'a larger jacket'], 'A', "What I can do is give you a credit note"),
    ('Cost of a boiler repair',
     "Engineer: The call-out's included in your cover, so that's free. The part itself is forty pounds, and normally labour would be another sixty on top of that. "
     "But because it's still under guarantee, you'll only pay for the part.",
     'How much will the customer pay?', ['£40', '£60', '£100'], 'A', "you'll only pay for the part"),
    ('Fault with the dishwasher',
     "Customer: It fills up with water fine, and it drains fine. And the door closes properly. The trouble is the dishes come out cold and still dirty, "
     "so I think the heater must have gone.",
     'What does the customer think is wrong?', ['The water does not drain.', 'The door does not close.', 'The water does not heat up.'], 'C', "I think the heater must have gone"),
    ('Complaint about a damaged parcel',
     "Customer: The parcel arrived on the right day, and it was the right item. But the box was completely crushed, and two of the plates inside were broken. "
     "I'd like replacements, please, rather than my money back.",
     'What does the customer want?', ['a refund', 'new plates', 'a different delivery day'], 'B', "I'd like replacements, please"),
    ('Choosing a new mattress',
     "Woman: The memory foam one was comfortable in the shop, but I've read that it gets very hot at night. The cheapest one felt too soft. "
     "So we went for the one with pocket springs. It was the most expensive, but it has a ten-year guarantee.",
     'Which mattress did the woman buy?', ['memory foam', 'the cheapest one', 'pocket springs'], 'C', "we went for the one with pocket springs"),

    # ------------------------------------------------ community events
    ('Change to the fete programme',
     "Organiser: A quick change to the programme. The dog show will still be at two, and the tug of war stays at four. "
     "But the children's races have been moved from the morning to half past three, because the field won't be dry enough earlier.",
     'Which event has been moved?', ['the dog show', 'the children’s races', 'the tug of war'], 'B', "the children's races have been moved"),
    ('Why the fireworks moved to the park',
     "Councillor: The fireworks have always been on the beach, and we've had no complaints about noise. It isn't about cost either. "
     "The simple fact is that high tides this year would leave too little space on the beach for the crowd, so we've moved to Victoria Park.",
     'Why has the fireworks display moved?', ['There is not enough space on the beach.', 'People complained about the noise.', 'The beach is too expensive.'],
     'A', "high tides this year would leave too little space on the beach"),
    ('Jobs at the river clean-up',
     "Organiser: We've got enough people for litter-picking along the banks, and the canoe club are pulling rubbish out of the water. "
     "What we need more of is people to sort what we collect into recycling and general waste.",
     'What kind of volunteer is most needed?', ['litter-pickers', 'people to sort rubbish', 'people with canoes'], 'B', "What we need more of is people to sort what we collect"),
    ('What the fun day money will pay for',
     "Head teacher: Last year the money from the fun day went on new library books, and the year before on a minibus. "
     "This year we're raising money to cover the playground, so the children can play outside when it rains.",
     'What will this year’s money pay for?', ['library books', 'a minibus', 'a covered playground'], 'C', "we're raising money to cover the playground"),
    ('Buying tickets for the quiz',
     "Organiser: You can't pay on the door this year, because we sold out so quickly last time and people were turned away. "
     "And we're not selling them online. Tickets are only available from the post office in the high street, and they must be bought by Thursday.",
     'Where can tickets be bought?', ['at the door', 'online', 'at the post office'], 'C', "Tickets are only available from the post office"),
    ('New venue for choir practice',
     "Choir leader: We can't use the church hall on Wednesdays any more, as it's been booked by a dance class. The school offered us their music room, but it's far too small. "
     "So from next month we'll be rehearsing in the library's meeting room, which has a piano.",
     'Where will the choir rehearse from next month?', ['the church hall', 'the library', 'the school'], 'B', "we'll be rehearsing in the library's meeting room"),
    ('Parking for the carnival',
     "Announcer: The car park at the station will be closed all day for the carnival, and there's no parking on the parade route. "
     "Visitors are asked to use the park-and-ride at the retail park. The buses are free and run every ten minutes.",
     'Where should visitors park?', ['at the retail park', 'at the station', 'on the parade route'], 'A', "use the park-and-ride at the retail park"),

    # ---------------------------------------------- transport and travel
    ('Cheapest rail ticket to York',
     "Clerk: An anytime return is eighty-four pounds. An off-peak return is fifty-two, but you can't travel before half nine. "
     "The cheapest option is two advance singles at nineteen pounds each, but you'll have to travel on the trains you book.",
     'What is the cheapest option?', ['an anytime return', 'an off-peak return', 'two advance singles'], 'C', "The cheapest option is two advance singles"),
    ('Why the coach was delayed',
     "Driver: Sorry for the delay, everyone. The traffic on the motorway was actually fine, and it's nothing wrong with the coach. "
     "We had to wait at the last stop for a passenger whose train was late, and the company asked me to hold on for her.",
     'Why was the coach late?', ['There was heavy traffic.', 'There was a mechanical problem.', 'The driver waited for a passenger.'], 'C', "We had to wait at the last stop for a passenger"),
    ('Seats on the overnight ferry',
     "Man: We looked at a cabin, but it was over a hundred pounds, and the reclining seats looked really uncomfortable. "
     "In the end we just booked the sleeping pods. They're cheap, and you can lie almost flat.",
     'What did they book?', ['a cabin', 'reclining seats', 'sleeping pods'], 'C', "we just booked the sleeping pods"),
    ('Luggage on a budget flight',
     "Agent: Your ticket includes one small bag that fits under the seat. A cabin bag in the overhead locker is twelve pounds extra, and a hold bag is twenty-five. "
     "You said you're only going for two nights, so I'd say the small bag is all you need.",
     'What does the agent advise?', ['paying for a cabin bag', 'taking only the small bag', 'paying for a hold bag'], 'B', "the small bag is all you need"),
    ('Parking at the railway station',
     "Clerk: The station car park is usually full by seven. There's a multi-storey in the town centre, but it's a fifteen-minute walk. "
     "Most commuters use the new overflow car park behind the station, which only opened in May. It's cheaper, too.",
     'Where do most commuters park?', ['in the station car park', 'in the overflow car park', 'in the multi-storey'], 'B', "Most commuters use the new overflow car park"),
    ('Replacement bus service',
     "Announcer: Because of engineering work, there are no trains between Exeter and Plymouth this weekend. Replacement buses will run instead. "
     "Please note they don't leave from the station forecourt as usual, but from the bus station, which is a five-minute walk. Allow an extra hour for your journey.",
     'Where do the replacement buses leave from?', ['the station forecourt', 'the bus station', 'the car park'], 'B', "but from the bus station"),
    ('Extra cover for a hire car',
     "Agent: Would you like to add a second driver? "
     "Customer: No, just me. But I'll take the extra insurance, because the excess is so high. "
     "Agent: And a sat nav? "
     "Customer: No, I'll use my phone. And I don't need a child seat.",
     'What extra does the customer take?', ['a second driver', 'extra insurance', 'a sat nav'], 'B', "I'll take the extra insurance"),

    # ------------------------------------------- museums and galleries
    ('Why the portrait was moved',
     "Guide: Many visitors ask why the portrait of the Duchess is no longer in the main hall. It hasn't been sold, and it isn't being cleaned. "
     "It was moved because the sunlight coming through the new windows was fading the paint, so it now hangs in the east gallery, where there's no direct light.",
     'Why was the portrait moved?', ['It was being cleaned.', 'Light was damaging it.', 'It had been sold.'], 'B', "the sunlight coming through the new windows was fading the paint"),
    ('The new gallery at the museum',
     "Curator: People expected the new gallery to be about the Romans, because of the excavation last year, and some people hoped for dinosaurs. "
     "But in fact it tells the story of the town's shoe factories, which employed half the population a hundred years ago.",
     'What is the new gallery about?', ['the Romans', 'dinosaurs', 'local industry'], 'C', "it tells the story of the town's shoe factories"),
    ('Photography in the museum',
     "Guide: You're welcome to take photos in most of the galleries, though we ask you not to use flash. You can't use a tripod anywhere. "
     "And in the special exhibition, no photography is allowed at all, because the objects are on loan from other museums.",
     'What is not allowed anywhere in the museum?', ['taking photos', 'using flash', 'using a tripod'], 'C', "You can't use a tripod anywhere"),
    ('Highlight of the ceramics collection',
     "Curator: Visitors usually head straight for the Chinese vases, and the tea sets are very popular with schools. "
     "But if I had to choose one piece, it would be the small blue jug by the window. It was made in this city in 1760, and it's the oldest piece of local pottery we have.",
     'Which piece does the curator choose?', ['a jug', 'a tea set', 'a vase'], 'A', "it would be the small blue jug by the window"),
    ('How the museum got the clock',
     "Guide: The clock wasn't bought at auction, as some people think, and the museum didn't commission it. "
     "It was left to us by a local clockmaker's granddaughter in her will, together with all his tools.",
     'How did the museum get the clock?', ['It was bought at auction.', 'It was made for the museum.', 'It was given in a will.'], 'C', "It was left to us by a local clockmaker's granddaughter"),
    ('Starting time for the gallery tour',
     "Assistant: The free tour usually starts at eleven. Today, though, the guide is running a school visit first, so it'll start at half past twelve. "
     "The two o'clock tour is the same as usual.",
     'When does the first tour start today?', ['11.00', '12.30', '2.00'], 'B', "it'll start at half past twelve"),

    # ------------------------------------------------- academic talks
    ('Lecture: Loss aversion',
     "Lecturer: Psychologists have found that losing ten pounds feels roughly twice as bad as winning ten pounds feels good. "
     "This isn't about how rich or poor people are, and it isn't simply that people dislike risk. "
     "It's that losses have a stronger emotional effect than equal gains, which is known as loss aversion.",
     'According to the lecturer, what explains loss aversion?', ['how much money people have', 'a general dislike of risk', 'the stronger feeling caused by losses'],
     'C', "losses have a stronger emotional effect than equal gains"),
    ('Lecture: First impressions at interviews',
     "Lecturer: A study of job interviews found that the interviewers' final decision was often very close to the judgement they made in the first ten seconds. "
     "Qualifications and answers to questions mattered less than expected. The strongest influence was the handshake and eye contact at the very start.",
     'What had the strongest influence on interviewers?', ['qualifications', 'answers to questions', 'behaviour at the start'], 'C', "The strongest influence was the handshake and eye contact"),
    ('Lecture: Why small businesses fail',
     "Lecturer: New owners often blame competition or a lack of customers when a business fails. Those are factors, of course. "
     "But studies of small firms consistently find the most common cause is cash flow: the business is profitable on paper, but runs out of money to pay its bills.",
     'What is the most common cause of small business failure?', ['cash flow problems', 'too much competition', 'too few customers'], 'A', "the most common cause is cash flow"),
    ('Lecture: Prices ending in 99',
     "Lecturer: Why are so many prices £4.99 rather than £5? It isn't because of tax, and shops don't really need the extra coins for change. "
     "The reason is that we read from left to right, so we notice the four first, and the price seems closer to four pounds than five.",
     'Why do prices end in 99?', ['Because of tax rules.', 'To make items seem cheaper.', 'So that shops can give change.'], 'B', "the price seems closer to four pounds than five"),
    ('Lecture: Family businesses',
     "Lecturer: Family firms tend to be good at keeping staff, and they often think long-term. "
     "Their biggest weakness, though, is handing over to the next generation. Only about a third survive into the second generation, often because no clear plan was made.",
     'What is the biggest weakness of family businesses?', ['keeping staff', 'planning for the long term', 'passing the business on'], 'C', "Their biggest weakness, though, is handing over to the next generation"),
    ('Lecture: Why toast turns brown',
     "Lecturer: Toast doesn't turn brown because it burns, at least not at first, and it isn't simply drying out. "
     "The browning comes from a reaction between sugars and proteins at high temperatures, called the Maillard reaction. This also produces the smell and flavour of toast.",
     'What causes toast to turn brown?', ['burning', 'a reaction between sugars and proteins', 'the bread drying out'], 'B', "The browning comes from a reaction between sugars and proteins"),
    ('Lecture: Ripening bananas',
     "Lecturer: Bananas are picked green and shipped in cool containers, which slows ripening. They don't ripen on the ship because of the temperature. "
     "When they arrive, they're placed in special rooms and given a gas called ethylene, which the fruit also produces naturally. That's what starts the ripening.",
     'What starts bananas ripening after they arrive?', ['sunlight', 'warm containers', 'ethylene gas'], 'C', "given a gas called ethylene"),
    ('Lecture: Frozen vegetables',
     "Lecturer: Frozen peas often contain more vitamin C than fresh peas from a shop. That isn't because anything is added. "
     "It's because they're frozen within a few hours of picking, while fresh peas may spend days in lorries and on shelves, losing vitamins all the time.",
     'Why can frozen peas contain more vitamin C?', ['Vitamins are added to them.', 'They are frozen soon after picking.', 'They are grown differently.'], 'B', "they're frozen within a few hours of picking"),
    ('Lecture: Courtyards in hot countries',
     "Lecturer: Traditional houses in hot, dry countries were often built around a courtyard. It gave families privacy, and space for a garden. "
     "But its main purpose was cooling. Cool night air collected in the courtyard and stayed there for much of the next day, keeping the rooms around it cooler.",
     'What was the main purpose of the courtyard?', ['to keep the house cool', 'to give families privacy', 'to provide space for a garden'], 'A', "its main purpose was cooling"),
    ('Lecture: The window tax',
     "Lecturer: From 1696, houses in England were taxed according to the number of windows. That's why you sometimes see old houses with windows filled in with brick. "
     "Owners did it to pay less tax, not for insulation or security.",
     'Why did owners fill in windows?', ['to keep the house warmer', 'to make the house more secure', 'to reduce their taxes'], 'C', "Owners did it to pay less tax"),
    ('Lecture: Green belts around cities',
     "Lecturer: Britain's green belts were not mainly created to protect farming or wildlife, although they do both. "
     "The chief aim, set out in the 1950s, was to stop towns spreading into each other, so that each kept its own identity.",
     'What was the main aim of green belts?', ['to stop towns joining together', 'to protect farmland', 'to protect wildlife'], 'A', "The chief aim, set out in the 1950s, was to stop towns spreading"),
    ('Lecture: Why towns grew at river crossings',
     "Lecturer: Many old market towns sit where a river could be crossed, at a ford or a bridge. The land there wasn't especially fertile, and defence wasn't usually the reason. "
     "Roads met at the crossing, so traders passed through, and markets grew up to serve them.",
     'Why did towns grow at river crossings?', ['The land was fertile.', 'They were easy to defend.', 'Traders passed through them.'], 'C', "Roads met at the crossing, so traders passed through"),
    ('Lecture: Frost hollows',
     "Lecturer: On clear, still nights, cold air flows downhill, because it's heavier than warm air, and collects in the lowest ground. These places are called frost hollows. "
     "Gardeners notice them because plants at the bottom of a slope are damaged by frost, while those halfway up are fine.",
     'Where do frost hollows form?', ['on hilltops', 'halfway up slopes', 'in low ground'], 'C', "collects in the lowest ground"),
    ('Lecture: Out-of-town retail parks',
     "Lecturer: Retail parks grew in the 1980s, mainly because of the car. Rents were lower than in town centres, but that wasn't the main attraction for shoppers. "
     "What shoppers wanted was free parking right outside the shop.",
     'What attracted shoppers to retail parks?', ['lower prices', 'free parking', 'a wider range of shops'], 'B', "What shoppers wanted was free parking"),
]
