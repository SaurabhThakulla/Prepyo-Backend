"""IELTS Listening practice, batch 4: note and form completion drills.

Original Prepyo scripts in the public IELTS Listening completion format. The
answers come in the same order as the notes, and most scripts give a first
answer that the speaker then corrects, as the test does. Settings: sports and
leisure centres, health services, shopping and repairs, community events,
transport and travel bookings, museum and gallery talks, and short academic
talks. Not official, recalled or copied test content.

Each item: (title, script, notes, limit). `notes` is a list of (text with
____ for the gap, [answer, accepted variants...]).
"""
from ielts_b4_listening.common import W1, WN, W2

FILL = [
    # ------------------------------------------------ sports and leisure
    ('Climbing wall induction',
     "Receptionist: Castle Street Leisure Centre, how can I help? "
     "Caller: Hi, I'd like to book an induction for the climbing wall. I've climbed before, but never at your centre. "
     "Receptionist: OK. Everyone has to do the induction first, even experienced climbers, I'm afraid. We run them on Tuesday evenings and Saturday mornings. "
     "Caller: Saturday would suit me better. "
     "Receptionist: Right. It lasts about an hour and a half, and it costs eighteen pounds. "
     "Caller: Do I need to bring my own equipment? "
     "Receptionist: We lend you everything except shoes. You can hire those here for three pounds, but most regular climbers buy their own. "
     "And it's worth arriving a few minutes early, because there's a health form to fill in at the desk before you go up.",
     [('Induction day: ____', ['Saturday']), ('Cost: £____', ['18', 'eighteen']),
      ('Shoe hire: £____', ['3', 'three']), ('Before starting, complete a health ____', ['form'])], WN),

    ('Badminton court booking',
     "Assistant: Sports hall bookings, Jamie speaking. "
     "Caller: Hello, I'd like to book a badminton court for next Thursday, please. "
     "Assistant: What time were you thinking of? "
     "Caller: About seven? "
     "Assistant: Seven's gone, I'm afraid. I could do half past eight. "
     "Caller: That's a bit late, but never mind, we'll take it. "
     "Assistant: Is that singles or doubles? "
     "Caller: Doubles, so there'll be four of us. "
     "Assistant: Then I'll put you on court three, which is the biggest. Do you need to borrow rackets? "
     "Caller: No, we've got our own, but could we have some shuttlecocks? "
     "Assistant: We don't lend those any more. You can buy a tube from the machine by the changing rooms. And could I take a surname for the booking? "
     "Caller: Yes, it's Pritchard. P-R-I-T-C-H-A-R-D.",
     [('Time: ____ pm', ['8.30', '8:30']), ('Court number: ____', ['3', 'three']),
      ('Buy shuttlecocks from the ____', ['machine']), ('Name: ____', ['Pritchard'])], WN),

    ('Junior tennis summer camp',
     "Coach: So the summer camp runs for two weeks in August, Monday to Friday. "
     "Parent: And what ages is it for? "
     "Coach: Eight to fourteen. We put the children into groups by ability rather than age, so your son might be with some slightly older ones. "
     "Parent: That's fine. What are the times? "
     "Coach: Drop-off is at nine, and pick-up is at four. The website still says half past four, but that's been changed. "
     "Parent: Does he need to bring lunch? "
     "Coach: Yes, a packed lunch and plenty of water. And suncream, please, because they're outdoors most of the day. "
     "The price for the fortnight is a hundred and sixty pounds, and there's ten per cent off if you book a brother or sister as well.",
     [('Groups are based on: ____', ['ability']), ('Pick-up time: ____ pm', ['4', 'four']),
      ('Bring: packed lunch, water and ____', ['suncream']), ('Cost for two weeks: £____', ['160'])], WN),

    ('Aqua aerobics enquiry',
     "Woman: Hello, I'm ringing about the aqua aerobics classes. My doctor suggested it for my knees. "
     "Receptionist: Good idea, it's very gentle on the joints. We've got a class on Wednesday mornings at ten, and one on Friday afternoons. "
     "Woman: The Wednesday one, I think. Is it in the main pool? "
     "Receptionist: No, it's in the learner pool, which is shallower and a bit warmer. "
     "Woman: Lovely. And do I have to be a member? "
     "Receptionist: Not at all. Non-members pay six pounds twenty a class. Members pay four fifty. "
     "Woman: Is there anything I should bring? "
     "Receptionist: Just a costume and a towel. Oh, and if you've got any, a pair of water shoes, because the steps can be slippery.",
     [('Day chosen: ____', ['Wednesday']), ('Held in the ____ pool', ['learner']),
      ('Price for non-members: £____', ['6.20']), ('Useful to bring: water ____', ['shoes'])], WN),

    ('Walking football group',
     "Organiser: The walking football group meets every Monday at the sports ground on Mill Lane. We used to play on the grass, but now we use the artificial pitch, so it doesn't matter if it's been raining. "
     "Man: And there's no running at all? "
     "Organiser: None. If you run, the referee gives a free kick to the other side. Most of our players are over fifty, and one of them is seventy-eight. "
     "Man: What does it cost? "
     "Organiser: Three pounds a session, and that pays for the pitch. Just wear trainers, not football boots, and bring a drink. "
     "We finish at half past twelve, and a lot of us go to the café afterwards.",
     [('Playing surface: the artificial ____', ['pitch']), ('Running leads to a free ____', ['kick']),
      ('Session fee: £____', ['3', 'three']), ('Footwear: ____', ['trainers'])], WN),

    ('Ice rink party booking',
     "Assistant: Ice rink, party bookings. "
     "Caller: Hi, I'd like to book a party for my daughter. She'll be nine. "
     "Assistant: Lovely. Our parties are on Saturday or Sunday afternoons. "
     "Caller: Sunday, the fourteenth, please. There'll be twelve children. Actually, no, two more have just said yes, so fourteen. "
     "Assistant: That's fine, the maximum is twenty. It's forty-five minutes on the ice with a coach, then food in the party room. "
     "Caller: What food do you do? "
     "Assistant: Pizza or hot dogs. Most people go for pizza. "
     "Caller: Pizza, then. "
     "Assistant: And all the children must wear gloves on the ice. We supply helmets, but not gloves.",
     [('Day: ____', ['Sunday']), ('Number of children: ____', ['14', 'fourteen']),
      ('Food chosen: ____', ['pizza']), ('Children must bring: ____', ['gloves'])], WN),

    ('Bowls club open day',
     "Secretary: Our open day is on the first Saturday in May, from ten until three. Anyone can come and have a go; you don't need any experience. "
     "The one rule is that you must wear flat shoes on the green, so no heels, and no trainers with a heavy tread either. "
     "We'll lend you a set of bowls, of course. Coaching is free on the day, and if you decide to join, the first year's membership is half price, which comes to forty-two pounds. "
     "Tea and cakes will be served in the pavilion all day.",
     [('Month: ____', ['May']), ('Shoes must be: ____', ['flat']),
      ('First-year membership: £____', ['42', 'forty-two']), ('Refreshments in the: ____', ['pavilion'])], WN),

    # ----------------------------------------------------- health services
    ('Registering with a GP surgery',
     "Receptionist: Good morning, Tarrant Road Surgery. "
     "Caller: Hi, I've just moved to the area and I'd like to register. "
     "Receptionist: Of course. Can I check your post code first, to make sure you're in our area? "
     "Caller: It's LS6 2QT. "
     "Receptionist: Yes, that's fine. You can fill in the form online, or collect a paper one from us. "
     "Caller: Online, I think. "
     "Receptionist: OK. When you've done that, we'll ask you to come in for a health check with the practice nurse, not the doctor. It takes about twenty minutes. "
     "Caller: Do I need to bring anything? "
     "Receptionist: Photo ID, such as a driving licence, and a list of any medicines you take.",
     [('Post code: ____', ['LS6 2QT']), ('First appointment with: the practice ____', ['nurse']),
      ('Length: about ____ minutes', ['20', 'twenty']), ('Bring: a list of ____', ['medicines'])], WN),

    ('Physiotherapy appointment',
     "Receptionist: Physiotherapy department. "
     "Patient: Hello, my GP has referred me for my shoulder. I had a letter asking me to phone. "
     "Receptionist: Yes, I can see the referral. Your first appointment will be with Ms Okafor. That's O-K-A-F-O-R. "
     "Patient: And when would that be? "
     "Receptionist: The earliest is Monday the twelfth at twenty to ten. "
     "Patient: That's fine. Where do I go? "
     "Receptionist: We're in the outpatients' building, on the ground floor, not the first floor as the old signs say. "
     "Please wear a vest top or loose clothing so she can look at the shoulder properly, and allow about an hour.",
     [('Physiotherapist: Ms ____', ['Okafor']), ('Time: ____ am', ['9.40', '9:40']),
      ('Department is on the ____ floor', ['ground']), ('Wear: a vest top or loose ____', ['clothing'])], WN),

    ('Flu vaccination clinic',
     "Nurse: This year's flu clinics will be held in the church hall on Victoria Street, because our waiting room is too small for the numbers. "
     "The first clinic is on Saturday the fourth of October. You don't need an appointment, but please bring your NHS number if you have it. "
     "It's free for anyone over sixty-five and for people with certain health conditions. Everyone else can still have the vaccine at a pharmacy, where it costs around fifteen pounds. "
     "And please wear a short-sleeved top, because it's given in the upper arm.",
     [('Place: the church ____', ['hall']), ('First clinic: 4 ____', ['October']),
      ('Free for people over: ____', ['65', 'sixty-five']), ('Wear: a short-sleeved ____', ['top'])], WN),

    ('Booking an eye test',
     "Optician: Hello, Clarke and Webb Opticians. "
     "Caller: Hi, I'd like to book an eye test. I've been getting headaches when I read. "
     "Optician: OK. When did you last have one? "
     "Caller: About three years ago. "
     "Optician: Right, so you're due. We've got Friday at eleven fifteen. "
     "Caller: Could it be a bit later? Say after lunch? "
     "Optician: Friday at two, then. The test is free if you're a student. "
     "Caller: I am, yes. "
     "Optician: Good. Then please bring your student card. And if you drive, don't come by car, because we may put drops in your eyes and you won't be able to drive for a few hours.",
     [('Symptom: ____', ['headaches']), ('Appointment: Friday at ____', ['2', 'two']),
      ('Bring: student ____', ['card']), ('Do not travel by ____', ['car'])], WN),

    ('Collecting a prescription',
     "Pharmacist: Your prescription's ready, but we could only give you part of it today. We're short of the cream, so you've got the tablets now and the cream will be in on Wednesday. "
     "Customer: OK. Do I need to come back in person? "
     "Pharmacist: Yes, but just give your name at the counter. Take one tablet twice a day, with food. They may make you feel a bit sleepy, so no alcohol while you're taking them. "
     "And if the rash hasn't cleared in ten days, go back and see your doctor.",
     [('Item not yet in stock: the ____', ['cream']), ('Available from: ____', ['Wednesday']),
      ('Take tablets with: ____', ['food']), ('See doctor if no better after ____ days', ['10', 'ten'])], WN),

    ('Hearing test at the clinic',
     "Receptionist: Audiology. "
     "Caller: Hello, I'm phoning for my father. He's been referred for a hearing test. "
     "Receptionist: Can I have his date of birth? "
     "Caller: The sixth of March, nineteen forty-eight. "
     "Receptionist: Thank you. I've got him down for Tuesday at half past two, in Room 14. "
     "Caller: Is that in the main hospital? "
     "Receptionist: No, it's in the Beech Building, which is behind the multi-storey car park. "
     "The test takes about forty-five minutes. He'll need to avoid loud noise for a whole day before the test, so no concerts or power tools. "
     "Caller: I'll tell him. And can I come in with him? "
     "Receptionist: Yes, a relative can stay in the room.",
     [('Room number: ____', ['14', 'fourteen']), ('Building: the ____ Building', ['Beech']),
      ('Test length: ____ minutes', ['45', 'forty-five']), ('Day before: avoid loud ____', ['noise'])], WN),

    # --------------------------------------------------- shopping and repairs
    ('Washing machine repair',
     "Engineer: Appliance repairs, how can I help? "
     "Customer: My washing machine won't drain. It fills up and washes, but the water just stays in the drum. "
     "Engineer: It's often the pump. What make is it? "
     "Customer: It's a Hoover, about six years old. "
     "Engineer: OK. There's a call-out charge of thirty-five pounds, and then parts on top. I can come on Thursday, but only in the morning. "
     "Customer: Thursday morning's fine. "
     "Engineer: Before I come, could you clean the filter? It's behind a little door at the bottom on the front. If that doesn't fix it, I'll take a look. "
     "And make sure there's a towel on the floor, because a lot of water comes out.",
     [('Problem: water does not ____', ['drain']), ('Call-out charge: £____', ['35', 'thirty-five']),
      ('Visit: Thursday ____', ['morning']), ('Customer should clean the ____ first', ['filter'])], WN),

    ('Returning a faulty kettle',
     "Assistant: Customer services. "
     "Customer: Hi, I bought this kettle three weeks ago and the switch has stopped working. It doesn't turn off when it boils. "
     "Assistant: Oh dear, that's not safe. Have you got the receipt? "
     "Customer: Not the paper one, but I paid by card. "
     "Assistant: That's fine, I can find it from your card. Would you like a refund or a replacement? "
     "Customer: A replacement, please, but can it be a different model? "
     "Assistant: Yes. The one up to ten pounds more is the Delta. You'd just pay the difference, which is four pounds. "
     "Customer: OK, I'll do that. "
     "Assistant: Right, let me just fetch one from the stockroom.",
     [('Fault: the ____ does not work', ['switch']), ('Proof of purchase: paid by ____', ['card']),
      ('Replacement model: the ____', ['Delta']), ('Customer pays: £____', ['4', 'four'])], WN),

    ('Watch repair',
     "Jeweller: So what's the problem with it? "
     "Customer: It keeps stopping. I thought it was the battery, so I had that changed last month, but it's still happening. "
     "Jeweller: Then it may need a full service. It's quite an old watch, so the movement probably needs cleaning. We send those to our workshop in Birmingham. "
     "Customer: How long does that take? "
     "Jeweller: About three weeks. The service is ninety-five pounds. While it's there, would you like a new strap? This one's quite worn. "
     "Customer: Yes, a brown leather one, please. "
     "Jeweller: That's another eighteen pounds. I'll give you a ticket; you'll need to show it when you collect the watch.",
     [('Already replaced: the ____', ['battery']), ('Workshop in: ____', ['Birmingham']),
      ('Service cost: £____', ['95', 'ninety-five']), ('Show a ____ when collecting', ['ticket'])], WN),

    ('Laptop screen repair',
     "Technician: So the screen's cracked in the corner, but everything else works? "
     "Customer: Yes, I dropped it on the stairs. "
     "Technician: We'll need to order a screen. For this model it's a hundred and twenty pounds, including fitting. "
     "Customer: How long will it take? "
     "Technician: The part usually arrives in two days, and fitting takes about an hour, so you could have it back on Friday. "
     "Customer: Do I need to leave the charger? "
     "Technician: No, keep that. But please write your password on this form, so we can check it's all working afterwards. "
     "Customer: I'd rather not. "
     "Technician: That's fine. Then you'll have to come in and test it with us before you take it home.",
     [('Damage: cracked in the ____', ['corner']), ('Price including fitting: £____', ['120']),
      ('Ready on: ____', ['Friday']), ('Customer keeps the ____', ['charger'])], WN),

    ('Shoe repair and key cutting',
     "Assistant: Morning. What can I do for you? "
     "Customer: Two things. These boots need new heels, and I need a spare key for my front door. "
     "Assistant: Heels on boots like these are fourteen pounds, and they'll be ready tomorrow afternoon. The key I can do now. "
     "Customer: Great. Can you make two copies instead of one? "
     "Assistant: Yes, three pounds fifty each. Oh, actually, this is a security key, so I'll need to see the card that came with the lock. Without it, I can't copy it. "
     "Customer: I think it's at home. "
     "Assistant: Bring it with you tomorrow, then, when you pick up the boots.",
     [('Repair for boots: new ____', ['heels']), ('Boots ready: tomorrow ____', ['afternoon']),
      ('Number of keys: ____', ['2', 'two']), ('Needed to copy the key: the lock ____', ['card'])], WN),

    ('Piano tuning appointment',
     "Tuner: Hello, piano tuning. "
     "Customer: Hi, our piano hasn't been tuned for years. It's an upright. "
     "Tuner: If it's been a long time, it might need two visits: one to bring it up to pitch and one to fine-tune. "
     "Customer: How much is that? "
     "Tuner: A single tuning is sixty-five pounds. If it needs the second visit, that's half price. "
     "Customer: OK. Can you come on the twentieth? "
     "Tuner: The twentieth's fine. Please make sure the room isn't too hot. Keep the piano away from the radiator if you can, and clear anything off the top so I can open the lid.",
     [('Type of piano: ____', ['upright']), ('Single tuning: £____', ['65', 'sixty-five']),
      ('Keep piano away from the ____', ['radiator']), ('Clear the top so the ____ can open', ['lid'])], WN),

    ('Carpet fitting quote',
     "Salesman: So it's the living room and the hall? "
     "Customer: Just the living room, actually. We've decided to leave the hall for now. "
     "Salesman: OK. And you liked the grey wool? "
     "Customer: We did, but the dog would ruin it. We'll go for the nylon one. "
     "Salesman: Sensible. That's twenty-two pounds a square metre, and the room is about eighteen square metres. Fitting is free if you spend over three hundred. "
     "Customer: And the old carpet? "
     "Salesman: We'll take it away for a charge of twenty-five pounds. We could fit it on the ninth of June. We start at eight, so the furniture needs to be moved out before then.",
     [('Room to be fitted: the ____ room', ['living']), ('Material: ____', ['nylon']),
      ('Price per square metre: £____', ['22', 'twenty-two']), ('Fitting date: 9 ____', ['June'])], WN),

    ('Bike shop service',
     "Mechanic: What does it need? "
     "Customer: The gears keep slipping, and the brakes are squeaking. "
     "Mechanic: The brakes probably just need new pads. The gears are likely to be the chain. It looks quite stretched. "
     "Customer: Can you do both? "
     "Mechanic: Yes. Our standard service is fifty pounds, and that includes pads, but the chain is extra, about twenty. I'll text you before I do anything else. "
     "Customer: My number's on the form. "
     "Mechanic: Good. It'll be ready by Saturday. We close at five on Saturdays, not half five, so don't be late.",
     [('Brakes need new ____', ['pads']), ('Gear problem probably caused by the ____', ['chain']),
      ('Standard service: £____', ['50', 'fifty']), ('Saturday closing time: ____ pm', ['5', 'five'])], WN),

    # ------------------------------------------------ community events
    ('Village summer fete stall',
     "Organiser: So you'd like a stall at the fete? What will you be selling? "
     "Woman: Homemade jam, mostly, and some chutney. "
     "Organiser: Lovely. The fete is on the recreation ground, on the nineteenth of July, and it opens at midday. Stalls are twelve pounds each. "
     "Woman: Do you provide tables? "
     "Organiser: We provide a table and a gazebo, in case it rains. You'll need to bring a float, because a lot of people pay with cash. "
     "And please label everything with the ingredients, because of allergies. "
     "Woman: Of course. "
     "Organiser: Setting up starts at ten. Park behind the scout hut, not on the grass.",
     [('Selling: homemade ____ and chutney', ['jam']), ('Stall price: £____', ['12', 'twelve']),
      ('Provided: a table and a ____', ['gazebo']), ('Park behind the scout ____', ['hut'])], WN),

    ('Charity quiz night',
     "Organiser: The quiz is on the last Friday of the month, at the Red Lion. It starts at half seven, so aim to be there by quarter past. "
     "Teams can have up to six people. It's five pounds each, and all the money goes to the local hospice. "
     "There are eight rounds, including a picture round and a music round. "
     "There's a prize for the winning team, and last year it was a meal for six, but this year it's a hamper donated by the farm shop. "
     "And we're asking people to leave their phones on the table, face down, so nobody can look up the answers.",
     [('Maximum team size: ____', ['6', 'six']), ('Money goes to the local ____', ['hospice']),
      ('Number of rounds: ____', ['8', 'eight']), ('Prize: a ____', ['hamper'])], WN),

    ('Beach litter-pick volunteers',
     "Coordinator: Thanks for signing up for Saturday's litter-pick. We'll meet by the lifeboat station at nine. "
     "We provide gloves, bags and litter pickers, but bring your own water bottle, because there's no café open that early. "
     "Last time we collected mostly plastic bottles, but the biggest problem now is fishing line, which is dangerous for birds. "
     "Please don't pick up anything sharp, like broken glass. Tell one of the team leaders, who'll be wearing orange bibs. "
     "We'll finish at twelve, and weigh everything we've collected.",
     [('Meeting point: the lifeboat ____', ['station']), ('Bring your own water ____', ['bottle']),
      ('Main problem now: fishing ____', ['line']), ('Team leaders wear: ____ bibs', ['orange'])], W1),

    ('Street party application',
     "Council officer: For a street party, you need to apply to close the road at least six weeks in advance. "
     "Resident: We're planning one for the first weekend of June. "
     "Officer: That's fine if you apply by the middle of April. It's free to apply. You'll need to show that everyone on the street has been told, so we ask for a letter to be put through every door. "
     "Resident: Do we need insurance? "
     "Officer: We recommend it, but it isn't compulsory for a small event. What you must do is keep a clear route for fire engines. "
     "Resident: And the signs? "
     "Officer: We lend you the road-closed signs. You collect them from the depot on Station Road the Friday before.",
     [('Apply at least ____ weeks before', ['6', 'six']), ('Tell residents with a ____', ['letter']),
      ('Keep a route clear for fire ____', ['engines']), ('Collect signs from the ____', ['depot'])], WN),

    ('Christmas lights switch-on',
     "Announcer: This year's Christmas lights switch-on is on Thursday the twenty-seventh of November in the Market Square. "
     "The lights will be switched on at six o'clock, not half past five as printed in the local paper. "
     "Before that, there'll be carols from the primary school choir and a brass band. "
     "Hot chocolate will be on sale from the town council's stall, and all profits go towards next year's display. "
     "Some roads will be closed from four, so if you're coming by car, please use the long-stay car park at the leisure centre. "
     "There's no parking in the square itself.",
     [('Place: ____ Square', ['Market']), ('Time of switch-on: ____ pm', ['6', 'six']),
      ('Music from a school choir and a brass ____', ['band']), ('Use car park at the ____ centre', ['leisure'])], WN),

    ('Repair café volunteers',
     "Organiser: The repair café runs on the second Sunday of every month, in the community centre. People bring in broken things, and our volunteers try to fix them for free. "
     "Man: I'm quite good with electrical things. "
     "Organiser: That's great, because that's where we're busiest. Toasters, lamps, radios. We do ask electrical volunteers to do a short safety course first. It's online and takes about two hours. "
     "Man: That's fine. Do I bring my own tools? "
     "Organiser: Please do, if you have them. We have some, but they're shared. Visitors are asked for a donation, and the money pays for the room hire. "
     "Oh, and we'd like everyone to wear a name badge, which we'll give you.",
     [('Held on the second ____ of the month', ['Sunday']), ('Busiest area: ____ items', ['electrical']),
      ('Online safety course: about ____ hours', ['2', 'two']), ('Donations pay for room ____', ['hire'])], WN),

    ('Local history society talk',
     "Secretary: Our next talk is on the history of the town's canal, given by Dr Ellis from the university. It's on Wednesday the eighth, in the library, not the town hall as usual, because the hall is being redecorated. "
     "Doors open at seven. Members come free, and visitors pay three pounds. "
     "Dr Ellis will be bringing a collection of old photographs of the canal and the boats that used it, and she's asked if anyone has photos of their own to bring along. "
     "Her book will be on sale afterwards for twelve pounds.",
     [('Topic: the town’s ____', ['canal']), ('Venue: the ____', ['library']),
      ('Visitors pay: £____', ['3', 'three']), ('Audience asked to bring their own ____', ['photos', 'photographs'])], WN),

    ('Dog show at the fete',
     "Organiser: The dog show starts at two, in the ring next to the bandstand. There are six classes, and you can enter on the day at the white tent for a pound a class. "
     "The most popular classes are 'waggiest tail' and 'best rescue dog'. The judge this year is a local vet, Mrs Hughes. "
     "All dogs must be on a lead at all times, and we'd ask owners to bring water for their dogs, because it can be very hot in the ring. "
     "Every dog that enters gets a rosette.",
     [('Ring is next to the ____', ['bandstand']), ('Enter at the white ____', ['tent']),
      ('Judge: Mrs ____', ['Hughes']), ('Every dog receives a ____', ['rosette'])], W1),

    # ---------------------------------------------- transport and travel
    ('Coach ticket to Edinburgh',
     "Agent: Coach bookings, how can I help? "
     "Caller: I need a single to Edinburgh on the fifteenth. "
     "Agent: From Newcastle? "
     "Caller: Yes. "
     "Agent: There's one at eight ten, arriving at twenty past eleven, and one at two in the afternoon. "
     "Caller: The morning one, please. "
     "Agent: That's nineteen pounds, or seventeen if you have a coach card. "
     "Caller: I don't. "
     "Agent: Nineteen, then. You're allowed two suitcases and one small bag. The coach leaves from bay G, at the back of the bus station. And your ticket will come by email, so there's nothing to print.",
     [('Departure time: ____ am', ['8.10', '8:10']), ('Fare: £____', ['19', 'nineteen']),
      ('Leaves from bay ____', ['G']), ('Ticket sent by ____', ['email'])], WN),

    ('Airport parking booking',
     "Agent: Airport parking. "
     "Caller: Hi, I'd like to book parking for a week. I fly out on the third of August. "
     "Agent: And back on? "
     "Caller: The tenth. "
     "Agent: The short-stay car park would be a hundred and forty pounds for that. The long-stay is only sixty-two, but you need to take a shuttle bus to the terminal. "
     "Caller: The long-stay's fine. "
     "Agent: Buses run every fifteen minutes. And I'll need your registration number. "
     "Caller: It's SK19 PWF. "
     "Agent: Thanks. The barrier will read your number plate, so you won't need a ticket.",
     [('Car park chosen: ____', ['long-stay']), ('Price: £____', ['62', 'sixty-two']),
      ('Shuttle every ____ minutes', ['15', 'fifteen']), ('Registration: ____', ['SK19 PWF'])], WN),

    ('Ferry crossing booking',
     "Agent: Ferry bookings. "
     "Caller: I'd like to take my car across to the Isle of Wight on the Saturday morning and come back on the Monday. "
     "Agent: How many passengers? "
     "Caller: Two adults and one child. "
     "Agent: The nine o'clock is full, I'm afraid. The next is at ten thirty. "
     "Caller: That's fine. "
     "Agent: And the return on Monday? "
     "Caller: Late afternoon. "
     "Agent: I've got the five fifteen. That's a hundred and eight pounds in total. You need to check in thirty minutes before, and the child will need a booster seat if they're under twelve.",
     [('Outward sailing: ____ am', ['10.30', '10:30']), ('Return sailing: ____ pm', ['5.15', '5:15']),
      ('Total cost: £____', ['108']), ('Check in ____ minutes before', ['30', 'thirty'])], WN),

    ('Applying for a railcard',
     "Clerk: So you want a railcard? Are you a student? "
     "Customer: I'm twenty-three, so I think I qualify for the young person's one. "
     "Clerk: Yes, it's for sixteen to twenty-five. It costs thirty pounds a year and gives you a third off most fares. "
     "Customer: Can I use it at any time? "
     "Clerk: Before ten in the morning there's a minimum fare of twelve pounds, except in July and August. "
     "Customer: How do I apply? "
     "Clerk: You can do it online, or here. Here you'll need a passport photo and ID, like your passport. Online you just upload a photo, and it goes onto an app on your phone.",
     [('Age range: 16 to ____', ['25', 'twenty-five']), ('Discount: a ____ off', ['third']),
      ('Minimum fare before 10 am: £____', ['12', 'twelve']), ('Online card goes onto an ____', ['app'])], WN),

    ('Car hire in Inverness',
     "Agent: So that's a car from Inverness airport, from the twenty-first to the twenty-eighth? "
     "Customer: Yes. Something small, but with enough room for two suitcases. "
     "Agent: I'd suggest a small estate. It's two hundred and ten pounds for the week. "
     "Customer: OK. Do I need extra insurance? "
     "Agent: It's included, but there's an excess of a thousand pounds, unless you pay nine pounds a day to reduce it. "
     "Customer: I'll leave that. "
     "Agent: The car will have a full tank, and you need to return it full. The office is in the car park opposite the terminal, and it's open until ten.",
     [('Type of car: small ____', ['estate']), ('Weekly price: £____', ['210']),
      ('Return the car with a full ____', ['tank']), ('Office: in the car park opposite the ____', ['terminal'])], WN),

    ('Sleeper train booking',
     "Agent: Sleeper reservations. "
     "Caller: I'd like to book a cabin from London to Fort William, for the night of the second of May. "
     "Agent: Just for you? "
     "Caller: For me and my sister. "
     "Agent: Then you'd want a twin cabin, with bunk beds. The price is two hundred and ninety pounds for the two of you, and that includes breakfast. "
     "Caller: What time does it leave? "
     "Agent: Nine fifteen from Euston, though you can board from half past eight. It arrives just before ten the next morning. "
     "Caller: Can we take our bikes? "
     "Agent: Yes, but you must book a space for them. There's no charge.",
     [('Cabin type: ____', ['twin']), ('Price includes: ____', ['breakfast']),
      ('Departs from: ____', ['Euston']), ('Must book a space for: ____', ['bikes'])], W1),

    ('Minibus hire for a club trip',
     "Assistant: Minibus hire. "
     "Caller: Hi, I'm organising a trip for our walking club. There'll be about fourteen of us. "
     "Assistant: Our sixteen-seater would be best. Is it with a driver? "
     "Caller: No, one of our members will drive. "
     "Assistant: Then the driver needs to be over twenty-five, and they'll need a D1 category on their licence. Most people who passed their test before 1997 have that. "
     "Caller: He's sixty, so that's fine. "
     "Assistant: Then it's ninety-five pounds for the day. We'll need a deposit of two hundred, which you get back if there's no damage. The bus has to come back by eight in the evening.",
     [('Number of seats: ____', ['16', 'sixteen']), ('Driver must be over: ____', ['25', 'twenty-five']),
      ('Day rate: £____', ['95', 'ninety-five']), ('Deposit: £____', ['200'])], WN),

    # ------------------------------------------- museums and galleries
    ('Gallery late opening',
     "Guide: On the first Friday of every month the gallery stays open until nine in the evening. "
     "These late openings are free, although there's a small charge for the special exhibition, which at the moment is on Scottish landscape painting. "
     "There's usually live music in the entrance hall, and the café becomes a bar. "
     "This month, one of the curators is giving a short talk at seven about how paintings are cleaned and repaired. "
     "Please note that the top floor is closed during late openings, and large bags must be left in the lockers downstairs.",
     [('Closing time on late nights: ____ pm', ['9', 'nine']), ('Current exhibition: ____ landscape painting', ['Scottish']),
      ('Talk: how paintings are cleaned and ____', ['repaired']), ('Large bags go in the ____', ['lockers'])], WN),

    ('Museum family trail',
     "Guide: Before you go, here's the family trail. Children collect a stamp in each gallery on the trail. "
     "Start in the Egyptian room, then go up to the costume gallery, and finish in the clock room, where the big clock strikes on the hour. "
     "When they've got all six stamps, they bring the card back to the shop and choose a badge. "
     "The trail is for children aged five to eleven. It takes about forty minutes. "
     "And please remind children that they can't touch anything in the cases, but in the costume gallery there's a box of hats they can try on.",
     [('Children collect a ____ in each gallery', ['stamp']), ('Finish in the ____ room', ['clock']),
      ('Reward: a ____', ['badge']), ('Can try on: ____', ['hats'])], W1),

    ('Textile exhibition talk',
     "Curator: This exhibition covers two hundred years of weaving in the north of England. "
     "The oldest piece is the blanket in the first case, made in about 1790, and it's wool, not cotton as people often assume. "
     "Cotton only arrived in this valley later, in the 1830s, when the big mills were built. "
     "The pattern books in the middle of the room belonged to a single family firm, and they contain more than four thousand samples. "
     "At the end, you'll see a working loom, and a demonstration is given every day at half past two.",
     [('Oldest piece: a ____', ['blanket']), ('Made of: ____', ['wool']),
      ('Pattern books contain more than ____ samples', ['4000', '4,000']), ('Daily demonstration of a working ____', ['loom'])], WN),

    ('Sculpture park audio guide',
     "Assistant: The audio guide costs four pounds. It covers thirty sculptures, and you type in the number on the sign next to each one. "
     "The walk around the whole park is about three kilometres, so give yourself a couple of hours. "
     "The path by the lake can be muddy, so if you're not wearing boots, follow the upper path instead. "
     "Please return the guide here to the visitor centre, not the café, before five, and we'll give you your deposit back.",
     [('Number of sculptures: ____', ['30', 'thirty']), ('Length of walk: about ____ km', ['3', 'three']),
      ('Without boots, take the ____ path', ['upper']), ('Return the guide to the visitor ____', ['centre'])], WN),

    ('Railway museum volunteers',
     "Manager: Most of our volunteers work as guides in the engine shed, but we also need people in the archive, sorting old tickets and timetables. "
     "Volunteers need to give at least one day a fortnight. We pay travel expenses up to ten pounds a day, and you get free lunch in the canteen. "
     "Everyone does an induction on a Tuesday, and you'll be given a uniform, which is a green fleece. "
     "If you'd like to drive the miniature railway, you'll need extra training, which takes place in the spring.",
     [('Extra help needed in the ____', ['archive']), ('Travel expenses up to: £____', ['10', 'ten']),
      ('Uniform: a green ____', ['fleece']), ('Driver training takes place in ____', ['spring'])], WN),

    ('Photography exhibition tour',
     "Guide: The photographs in this room were all taken on the same street in Glasgow, but fifty years apart. "
     "On the left are prints from the 1970s; on the right, the same views photographed last year by students from the art school. "
     "The earlier ones were taken by a bus driver, who photographed the street on his days off. He never showed them to anyone, and they were found in his attic after he retired. "
     "Notice how many of the shops have changed, but the church and the school are still exactly the same.",
     [('All photos show the same ____', ['street']), ('Recent photos by: art school ____', ['students']),
      ('Early photographer’s job: bus ____', ['driver']), ('Found in his ____', ['attic'])], W1),

    # ------------------------------------------------- academic talks
    ('Lecture: Where shops place their products',
     "Lecturer: Walk into almost any large shop and the layout has been planned in detail. "
     "Products placed at eye level sell best, so that's where manufacturers pay to be. Cheaper own-brand goods tend to be placed lower down. "
     "Everyday items such as milk and bread are usually at the back of the store, so customers walk past other goods on their way. "
     "Near the tills you'll find small, cheap items like sweets, which people buy on impulse while they're waiting. "
     "Recently some chains have removed sweets from the checkout, after pressure from parents.",
     [('Best-selling position: eye ____', ['level']), ('Milk and bread: at the ____ of the store', ['back']),
      ('Checkout items are bought on ____', ['impulse']), ('Pressure to remove sweets came from ____', ['parents'])], W1),

    ('Lecture: How franchising works',
     "Lecturer: In a franchise, a business owner, the franchisor, sells other people the right to run a branch under its name. "
     "The person who buys it, the franchisee, pays an initial fee and then, usually, a percentage of sales every month. This is called a royalty. "
     "In return, they get a known brand, training and advice. "
     "The main disadvantage for the franchisee is control. They can't change the menu, the prices or even the colour of the walls without permission. "
     "Research suggests franchises are less likely to fail in their first three years than independent businesses.",
     [('Monthly payment based on sales: a ____', ['royalty']), ('Franchisee receives brand, ____ and advice', ['training']),
      ('Main disadvantage: loss of ____', ['control']), ('Less likely to fail in first ____ years', ['3', 'three'])], WN),

    ('Lecture: Price anchoring',
     "Lecturer: When we decide whether a price is fair, we compare it with the first figure we saw. Psychologists call this anchoring. "
     "Restaurants use it by putting one very expensive dish at the top of the menu. Few people order it, but the other dishes then look reasonable. "
     "In one experiment, students were asked to write down the last two digits of their phone number before bidding for a bottle of wine. "
     "Those with high numbers bid significantly more, even though the number had nothing to do with the wine. "
     "Experts are affected too: estate agents given a higher asking price valued the same house more highly.",
     [('Menus: one very expensive ____ at the top', ['dish']), ('Students wrote down digits of their phone ____', ['number']),
      ('Item bid for: a bottle of ____', ['wine']), ('Experts affected: estate ____', ['agents'])], W1),

    ('Lecture: Remembering faces and names',
     "Lecturer: Most of us find names harder to remember than faces. One reason is that a name tells us very little. "
     "If I say a man is a baker, you can picture flour and bread. If I say his name is Baker, there's nothing to picture. "
     "Researchers call this the baker-Baker effect. "
     "Repeating a name aloud straight after hearing it helps a little. What helps more is linking the name to something you already know, such as a friend with the same name. "
     "And names are forgotten faster when we're anxious, which is why we so often forget them at parties and interviews.",
     [('A job gives us something to ____', ['picture']), ('Better method: link the name to a ____ with the same name', ['friend']),
      ('Names are forgotten faster when we feel ____', ['anxious']), ('Common situations: parties and ____', ['interviews'])], W1),

    ('Lecture: Why bread goes stale',
     "Lecturer: Most people think bread goes stale because it dries out. In fact, the main cause is a change in the starch. "
     "When bread is baked, the starch absorbs water and softens. As it cools, the starch slowly forms crystals again, and the bread becomes firm. "
     "Surprisingly, this happens fastest at fridge temperature, so keeping bread in the fridge makes it go stale more quickly. "
     "Freezing stops the process almost completely. And warming stale bread in the oven reverses it for a short time, which is why a stale loaf can be improved by heating.",
     [('Main cause: a change in the ____', ['starch']), ('On cooling, the starch forms ____', ['crystals']),
      ('Staling is fastest in the ____', ['fridge']), ('Heating in the ____ reverses it for a short time', ['oven'])], W1),

    ('Lecture: Making ice cream smooth',
     "Lecturer: The texture of ice cream depends on the size of its ice crystals. Small crystals feel smooth; large ones feel grainy. "
     "So manufacturers freeze the mixture quickly while stirring it constantly. "
     "Air is also important. Cheap ice cream can be half air, which is why a litre tub of it weighs so little. "
     "Sugar lowers the freezing point, so ice cream stays soft enough to scoop. "
     "The main enemy of texture is the home freezer. Every time the door is opened, the ice cream warms slightly, and the crystals grow larger.",
     [('Small crystals feel ____', ['smooth']), ('Cheap ice cream can be half ____', ['air']),
      ('Sugar lowers the freezing ____', ['point']), ('Crystals grow when the freezer ____ is opened', ['door'])], W1),

    ('Lecture: Victorian terraced houses',
     "Lecturer: Terraced houses were built in huge numbers in British towns in the nineteenth century, mostly for factory workers. "
     "Building in a row was cheap, because each house shared its side walls with its neighbours. "
     "Many were built as back-to-backs, with no garden or back door, but these were banned in most cities after 1909 because of poor ventilation. "
     "Toilets were outside, in the yard. "
     "Today these houses are popular again, partly because they're close to town centres, and many have been extended into the loft.",
     [('Built mostly for factory ____', ['workers']), ('Cheap because houses shared side ____', ['walls']),
      ('Back-to-backs banned because of poor ____', ['ventilation']), ('Common extension: into the ____', ['loft'])], W1),

    ('Lecture: Thatched roofs',
     "Lecturer: Thatch is one of the oldest roofing materials in Britain. The most common type today is water reed, which can last fifty years or more. "
     "Wheat straw is cheaper, but it needs replacing more often. "
     "A thatched roof is steep, so rain runs off before it can soak in. "
     "The top layer, along the ridge, wears out first and needs to be redone every ten to fifteen years. "
     "The main worry for owners is fire, and insurance for thatched houses is usually more expensive.",
     [('Most common type: water ____', ['reed']), ('Cheaper material: wheat ____', ['straw']),
      ('Part that wears out first: the ____', ['ridge']), ('Owners’ main worry: ____', ['fire'])], W1),

    ('Lecture: How rivers meander',
     "Lecturer: Rivers on flat ground rarely flow in a straight line. They form loops, called meanders. "
     "On the outside of a bend, the water flows faster and wears away the bank, forming a small cliff. "
     "On the inside, the water is slower, so sand and gravel are dropped, forming a gentle slope called a point bar. "
     "Over time the loops become larger. Eventually, during a flood, the river may cut straight across the narrow neck of a loop, "
     "leaving the old loop cut off as a curved lake, known as an oxbow lake.",
     [('Outside of the bend: water flows ____', ['faster']), ('Result on the outside: a small ____', ['cliff']),
      ('Inside of the bend: a point ____', ['bar']), ('River cuts across the loop during a ____', ['flood'])], W1),

    ('Lecture: Coastal spits',
     "Lecturer: A spit is a long, narrow ridge of sand or shingle that grows out from the coast into the sea. "
     "It forms where waves hit the beach at an angle and move material along the shore. This is called longshore drift. "
     "When the coastline changes direction, for example at a river mouth, the material keeps going and builds up in open water. "
     "The end of a spit is often curved, because of the wind. "
     "Behind the spit, in sheltered water, salt marshes often develop, which are important for wading birds.",
     [('Made of sand or ____', ['shingle']), ('Waves hit the beach at an ____', ['angle']),
      ('Often forms at a river ____', ['mouth']), ('End is curved because of the ____', ['wind'])], W1),
]
