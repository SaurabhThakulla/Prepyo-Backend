"""50 Original Respond to Situation (RTS) items for PTE Speaking Batch 3.

Constraints:
  - context_passage == audio_transcript (the situation text)
  - prep_time_seconds: 20, time_limit_seconds: 40
  - model_answer: direct spoken response to the person/entity
  - explanation: guidance on tone, politeness, clarity, and key points
  - Uncontroversial established situations
"""

RTS_ITEMS = [
    {
        "title": "A broken balance in the laboratory",
        "situation": "You are working in the university chemistry laboratory and the electronic analytical balance starts displaying error codes, preventing your group from weighing reagents. Speak to the laboratory technician, explain the problem, and ask if an alternative balance is available.",
        "model_answer": "Hello, excuse me. I'm working on the experiment at bench four, and our analytical balance keeps flashing an error code and won't tare properly. We need to weigh our reagents to continue. Could you please take a look at it, or is there an alternative balance we could use in the meantime?",
        "explanation": "Clearly state the lab bench location, describe the technical error, and politely inquire about a technician check or alternative equipment.",
    },
    {
        "title": "A reserved book is missing",
        "situation": "You reserved a required reference book through the library catalogue two days ago, but the holding shelf has no record of your reservation and the book is not there. Speak to the librarian at the circulation desk, explain your situation, and ask how you can access the text.",
        "model_answer": "Hi, good morning. I placed a reservation for the environmental economics textbook two days ago and received a confirmation email, but it wasn't on the holding shelf. Could you check the system to see if it's been misplaced, or let me know if a digital reserve copy is available?",
        "explanation": "Mention the confirmation email, describe the missing hold, and politely request a catalogue status check or digital reserve copy.",
    },
    {
        "title": "A car blocking your driveway",
        "situation": "You need to leave for an important appointment, but a neighbour's visitor has parked across your driveway and you cannot get your car out. Speak to your neighbour, explain the problem politely, and ask for the car to be moved.",
        "model_answer": "Hi, sorry to bother you. I think the silver car across my driveway belongs to one of your visitors. I need to leave for an appointment in about ten minutes, so could you ask them to move it, please? Thanks so much.",
        "explanation": "Identify the car, explain the urgency calmly, and make a polite, specific request without blaming anyone.",
    },
    {
        "title": "A broken radiator in your room",
        "situation": "The radiator in your dormitory bedroom has stopped working during a cold winter weekend, leaving your room uncomfortably cold. Visit the residential accommodation office, report the breakdown, and ask when a maintenance engineer can attend to it.",
        "model_answer": "Hello, I'm a resident in room 302 of Maple Hall. The radiator in my room completely stopped working yesterday, and it's become extremely cold. Could you please send a maintenance technician to inspect it today, or provide a temporary portable heater while it gets repaired?",
        "explanation": "Provide your room number and building, explain the lack of heating, and request urgent maintenance or a temporary heater.",
    },
    {
        "title": "The submission website crashes",
        "situation": "The online course portal crashed five minutes before the midnight deadline while you were uploading your final history essay. It is now the next morning. Speak to your course lecturer, explain what happened with the portal, and provide proof of your timely completion.",
        "model_answer": "Good morning Professor Miller. I wanted to speak with you regarding the history essay due last night. The university portal crashed while I was uploading my paper at 11:55 PM. I took a timestamped screenshot of the error and emailed my essay directly to you at that time. Could you please accept the submission without late penalties?",
        "explanation": "Politely describe the technical outage, reference your timestamped evidence, and confirm your completed paper was submitted.",
    },
    {
        "title": "Checking a meal for peanuts",
        "situation": "You have a severe peanut allergy and want to order lunch from the campus cafeteria hot food station, but none of the dishes have allergen labels. Speak to the cafeteria chef, explain your allergy, and ask which dishes are safe for you to eat.",
        "model_answer": "Excuse me, chef. I have a severe peanut allergy and noticed there are no allergen labels on the hot counter today. Could you please tell me which dishes are prepared completely free of peanuts and peanut oils, and whether there's any risk of cross-contamination?",
        "explanation": "Clearly state the severe peanut allergy, ask about ingredients and cross-contamination risks, and maintain a polite, safety-conscious tone.",
    },
    {
        "title": "Someone else is in your booked room",
        "situation": "You and your project group reserved a library study room for two o'clock, but when you arrive, another student group is already using the room and claims they booked it too. Speak to the group politely, show your reservation confirmation, and suggest a resolution.",
        "model_answer": "Hi there. Excuse me, but we actually have this room booked starting at two o'clock. Here is our booking confirmation email from the library portal. Did your reservation just end, or is there a double booking in the system? Perhaps we could check with the front desk together to sort it out.",
        "explanation": "Remain calm and polite, reference your confirmed booking, and suggest checking with library staff to resolve the overlap.",
    },
    {
        "title": "Asking for a day off for a wedding",
        "situation": "Your cousin is getting married next month on a Friday, when you are normally scheduled to work at a café. Speak to your manager, explain the situation, and ask whether you can have the day off.",
        "model_answer": "Hi, do you have a minute? My cousin is getting married on Friday the twelfth next month, and I'd really like to be there. Would it be possible to have that day off? I'm happy to work an extra shift another day or swap with someone.",
        "explanation": "Give plenty of notice, explain the reason briefly, ask politely, and offer a way to cover the shift.",
    },
    {
        "title": "Turning down a party invitation",
        "situation": "A friend has invited you to their birthday party on Saturday night, but you have an important exam early on Monday and need the weekend to study. Speak to your friend, thank them, and explain why you cannot come.",
        "model_answer": "Thanks so much for inviting me, I really appreciate it. Unfortunately I've got a big exam first thing on Monday and I need to spend the weekend revising. I'm sorry to miss it. Could we go out for a meal to celebrate after my exam?",
        "explanation": "Thank your friend, give the reason honestly, express regret, and suggest another way to celebrate.",
    },
    {
        "title": "A work shift clashes with a class",
        "situation": "You work part-time at the campus bookstore, and your manager has scheduled you to work during a newly scheduled compulsory exam review session next Thursday. Speak to your manager, explain the situation, and offer to swap shifts with a colleague.",
        "model_answer": "Hi Sarah, thank you for putting up the weekly schedule. I noticed I'm rostered for Thursday afternoon, but our professor just scheduled a mandatory exam review session during that exact time. I spoke with David, and he's happy to swap shifts with me so I can cover his Friday morning shift instead. Would that be okay with you?",
        "explanation": "Express gratitude for the roster, explain the academic obligation, and present a pre-arranged swap with a coworker to minimize inconvenience.",
    },
    {
        "title": "A leaking tap in your flat",
        "situation": "The kitchen tap in your rented flat has been dripping constantly for a week, and it is getting worse. Call your landlord, describe the problem, and ask when it can be repaired.",
        "model_answer": "Hello, it's the tenant at flat three. I'm calling because the kitchen tap has been dripping non-stop for about a week and it's getting worse, so it's wasting quite a lot of water. Could you arrange for someone to repair it? I'm at home most evenings if that helps.",
        "explanation": "Say who you are, describe the fault and how long it has lasted, explain why it matters, and suggest times for the repair.",
    },
    {
        "title": "Charged twice for textbooks",
        "situation": "You bought three textbooks at the campus bookstore and just noticed on your receipt that you were charged twice for the most expensive book. Return to the bookstore customer service counter, show your receipt, and request a refund for the duplicate charge.",
        "model_answer": "Hello, I just bought these textbooks a few minutes ago, but when I checked my receipt, I noticed that the biology textbook was scanned twice. I only purchased one copy. Here are the three books and my receipt. Could you please process a refund for the extra charge back to my card?",
        "explanation": "Politely show the receipt and items, clearly explain the duplicate scan error, and ask for the refund to be processed.",
    },
    {
        "title": "Asking a neighbour to take in a parcel",
        "situation": "You are expecting an important parcel tomorrow, but you will be at work all day. Ask your neighbour if they would be willing to accept the delivery for you.",
        "model_answer": "Hi, I'm sorry to ask a favour, but I'm expecting a parcel tomorrow and I'll be at work all day. Would you mind taking it in for me if you're at home? I'll come and collect it as soon as I'm back in the evening. Thank you so much.",
        "explanation": "Apologise for asking, explain when and why, make a clear request and say when you will collect it.",
    },
    {
        "title": "A bike locked to a handrail",
        "situation": "You locked your bicycle to a handrail near the lecture theatre because all the bike racks were full, and a campus security officer is about to impound it. Speak to the officer, apologize, explain the situation, and offer to move the bike immediately.",
        "model_answer": "Excuse me, officer. I'm so sorry, that's my bicycle. All the designated bike racks outside the hall were completely full and I was rushing to take an exam. I understand it shouldn't be locked to the handrail. If you give me just one moment, I will unlock it right now and move it to the secondary bike rack across the courtyard.",
        "explanation": "Acknowledge the violation politely, explain the full bike racks without being defensive, and offer immediate compliance.",
    },
    {
        "title": "Asking to swap seats on a train",
        "situation": "You are travelling on a busy train with your young child, but your reserved seats are not together. Ask the passenger sitting next to your child if they would mind changing seats with you.",
        "model_answer": "Excuse me, I'm sorry to trouble you. My daughter is sitting next to you and my reserved seat is two rows back, by the window. Would you mind swapping with me so I can sit with her? Thank you, that's very kind.",
        "explanation": "Explain the situation briefly, describe the other seat so it sounds fair, ask politely and thank the passenger.",
    },
    {
        "title": "A jammed printer",
        "situation": "The high-volume printer in the university computer lab has jammed with crumpled paper while you were printing your senior thesis, and several other students are waiting behind you. Speak to the lab assistant, describe the jam, and ask for technical assistance.",
        "model_answer": "Hi, excuse me. The main laser printer just had a paper jam while printing my document, and there's paper stuck inside tray two. A few other students are also waiting in line. Could you please help clear the jam and ensure our queued documents aren't canceled?",
        "explanation": "Explain the specific location of the paper jam, mention the waiting queue, and ask the technician to assist and check print queues.",
    },
    {
        "title": "Late for a lab because of a train",
        "situation": "Your morning commuter train was delayed by forty minutes due to signal failure, causing you to arrive late for a mandatory chemistry laboratory safety briefing. Approach the teaching assistant, explain the transit delay, and ask if you can join the lab.",
        "model_answer": "Good morning, I apologize for arriving late. My commuter train was halted for forty minutes due to a track signal failure. I know the safety briefing was scheduled for the first ten minutes. Could you please give me a quick safety briefing so I can participate in today's experiment safely?",
        "explanation": "Apologize sincerely, explain the train delay, acknowledge safety protocol importance, and ask for a brief overview to participate safely.",
    },
    {
        "title": "A forgotten locker code",
        "situation": "You forgot the combination to your rented gym locker in the campus sports complex and your student ID and car keys are locked inside. Speak to the sports center receptionist, prove your identity, and ask to have the locker opened.",
        "model_answer": "Hello, I'm having a problem with locker 45 in the men's changing room. My locker combination has slipped my mind, and my student ID and car keys are inside. I have a photo of my student card on my phone to verify that the locker is registered to my name. Could someone with the master key help me open it?",
        "explanation": "State the locker number, explain the forgotten code and trapped essentials, offer phone verification, and request master key help.",
    },
    {
        "title": "Needing a transcript urgently",
        "situation": "You need an official academic transcript stamped by the university registrar for a scholarship deadline tomorrow, but the standard processing time is five business days. Speak to the records officer, explain the urgent scholarship deadline, and ask if expedited processing is possible.",
        "model_answer": "Good morning. I urgently need an official stamped transcript for an external scholarship application with a strict deadline tomorrow afternoon. I realize standard processing takes five days, but is there any possibility of paying an expedited fee or having a copy certified today? This scholarship is critical for my tuition.",
        "explanation": "Politely describe the urgent scholarship deadline, acknowledge standard turnaround times, and ask about expedited or same-day processing.",
    },
    {
        "title": "A damaged camera kit",
        "situation": "While checking out a camera kit from the university media department, you notice the lens filter is cracked and the battery charger is missing. Speak to the equipment coordinator before leaving so you are not blamed for the damage.",
        "model_answer": "Hi, before I sign out this camera kit, I wanted to point out that the UV filter on the 50mm lens is already cracked, and the battery charger seems to be missing from the case. Could you please note this pre-existing damage on the checkout log so I'm not held responsible when returning it?",
        "explanation": "Point out the pre-existing damage proactively, verify missing parts, and ensure it is documented on the equipment log.",
    },
    {
        "title": "Cold air in a lecture hall",
        "situation": "Your regular lecture hall has an air conditioning system that is blowing freezing air directly onto students, making it impossible to concentrate during a two-hour lecture. Speak to the building facilities manager and ask if the thermostat can be adjusted.",
        "model_answer": "Hello, I'm a student in lecture hall B12. The air conditioning in that room is currently set extremely low and blowing freezing air directly over the seating area. Most students are shivering and finding it hard to take notes. Could you please adjust the thermostat to a more comfortable temperature?",
        "explanation": "Specify the exact room number, describe the discomfort affecting students' focus, and request a thermostat adjustment.",
    },
    {
        "title": "Changing your open day hours",
        "situation": "You volunteered to help staff the campus open day welcome booth from 9 AM to 3 PM, but you were just assigned a mandatory academic advising appointment at 1 PM. Speak to the volunteer coordinator, explain the conflict, and offer to work the morning shift.",
        "model_answer": "Hi Rebecca, regarding the open day welcome desk this Saturday: I was just scheduled for a mandatory academic advising appointment at one o'clock. Would it be alright if I work the morning shift from nine until twelve-thirty? That way I can still help welcome the largest rush of morning visitors.",
        "explanation": "Explain the newly scheduled academic appointment, propose a concrete morning shift commitment, and ensure the booth remains staffed.",
    },
    {
        "title": "The wrong size graduation gown",
        "situation": "You picked up your rented graduation cap and gown from the campus bookstore, but upon trying it on at home, you realized the gown is much too long and drags on the floor. Return to the distribution desk, explain the issue, and request a smaller size.",
        "model_answer": "Hello, I picked up this graduation gown yesterday, but when I tried it on, it turned out to be far too long and drags on the floor when I walk. Here is my rental receipt. Would it be possible to exchange it for a medium size that fits my height properly?",
        "explanation": "Politely present the gown and receipt, explain the sizing issue, and request an exchange for the correct fit.",
    },
    {
        "title": "Changing a restaurant booking",
        "situation": "You booked a table for four people at a restaurant for seven o'clock tonight, but your train is delayed and you will arrive later. Call the restaurant and ask to change the booking time.",
        "model_answer": "Hello, I have a booking for four people at seven o'clock tonight under the name Karki. Our train has been delayed, so we won't arrive until about half past eight. Would it be possible to move the booking to then? If not, is there another time you could offer us?",
        "explanation": "Give the booking details, explain the delay, ask for a specific new time and show flexibility.",
    },
    {
        "title": "A wallet left on the shuttle bus",
        "situation": "You accidentally left your leather wallet on the campus shuttle bus twenty minutes ago. Contact the university transit dispatch office, describe the wallet and the bus route you took, and ask if any driver has turned it in.",
        "model_answer": "Hello transit dispatch. I just got off the North Campus shuttle bus twenty minutes ago at the science building, and I believe I left my brown leather wallet on one of the middle seats. It has my driver's license and student ID inside. Has the driver reported finding any lost items on that route?",
        "explanation": "State the bus route, stop, and time accurately, provide an identifying description of the wallet and contents, and ask dispatch to check.",
    },
    {
        "title": "Recording lectures with a broken wrist",
        "situation": "You have a broken wrist and are temporarily unable to handwrite or type lecture notes fast enough to keep up. Speak to your economics professor after class, explain your physical impairment, and ask for permission to audio record lectures.",
        "model_answer": "Hello Professor Adams. As you can see, I recently fractured my right wrist and have it in a cast, which makes handwriting or typing notes very difficult right now. Would you mind if I place a small voice recorder on the podium to record your lectures solely for my personal study until my cast is removed?",
        "explanation": "Explain the physical injury, assure the professor the recording is strictly for personal study, and ask for permission respectfully.",
    },
    {
        "title": "A wrong charge on your meal plan",
        "situation": "You checked your campus meal plan balance online and noticed a charge of thirty-five dollars from the campus grill that you did not make. Speak to the dining services cashier supervisor, show your account statement, and ask for the error to be investigated.",
        "model_answer": "Hi, I'm checking my student meal account and noticed a charge of thirty-five dollars at the campus grill yesterday afternoon at three PM, but I was in a biology lecture across campus at that time. Could you please look into this transaction and see if another student's order was mistakenly charged to my student number?",
        "explanation": "Explain the incorrect charge, specify the date and time, mention your conflicting class, and politely ask for an account audit.",
    },
    {
        "title": "Inviting a classmate to a study group",
        "situation": "You notice a classmate who consistently asks insightful questions during complex organic chemistry lectures. Speak to them after class, compliment their contributions, and invite them to join your weekly study group.",
        "model_answer": "Hey David, I really liked the question you asked today about reaction mechanisms—it helped clear up a lot of confusion. A few of us from class meet every Wednesday evening at the library for a study group to work through problem sets. Would you be interested in joining us this week?",
        "explanation": "Pay a genuine compliment regarding their class contribution, explain the weekly study group format, and extend a friendly invitation.",
    },
    {
        "title": "A vending machine takes your money",
        "situation": "You put four dollars into the snack vending machine in the engineering building lobby, but the spiral coil jammed halfway, taking your money without releasing the snack. Speak to the building reception desk, report the machine number, and ask about a refund.",
        "model_answer": "Hello, I just tried to buy a snack from the vending machine in the lobby, machine number four, but the coil jammed halfway and didn't dispense the item, even though it took my four dollars. Is there a refund slip I can fill out, or a contact number for the vendor to get that resolved?",
        "explanation": "State the machine number, explain the jam and lost money, and inquire about the standard refund procedure.",
    },
    {
        "title": "Missing the field trip bus",
        "situation": "You arrived at the departmental field trip meeting point three minutes late due to a subway delay, and the main group bus has just departed. Phone the field trip organizer, apologize for being late, and ask if you can meet the group at the first field site.",
        "model_answer": "Hello Dr. Patel, this is Emily. I'm so sorry, my subway was delayed and I arrived at the campus pickup point just as your bus pulled away. I know the first stop is the geological quarry at ten AM. If I take the regional bus right now, I can meet you at the quarry gate by 9:50. Would that be acceptable?",
        "explanation": "Apologize for missing the departure, explain the subway delay, and present an immediate independent travel plan to rejoin the group on time.",
    },
    {
        "title": "A roommate's guest who stays for days",
        "situation": "Your roommate has had an overnight guest staying in your shared room for four consecutive days without asking you first, making it difficult to study and sleep. Speak to your roommate calmly, describe how it affects you, and suggest a fair guest policy.",
        "model_answer": "Hey Sam, do you have a few minutes to talk? Having your friend stay over for the past four nights has made it hard for me to study and get enough rest in our shared room. I don't mind having guests over occasionally, but could we agree to limit overnight guests to weekends and check with each other beforehand?",
        "explanation": "Remain calm and reasonable, express how extended guests impact privacy and study, and propose a sensible weekend guest rule.",
    },
    {
        "title": "A fine for a book you returned",
        "situation": "The library account shows a twenty-dollar overdue fine for a book that you returned into the exterior drop box three days before the due date. Speak to the library circulation desk supervisor, explain the timeline, and ask to have the fine waived.",
        "model_answer": "Good morning. My account shows an overdue fine for the modern history textbook, but I returned it into the library drop box on Friday evening, well before the Monday due date. It looks like it wasn't scanned in until Tuesday morning. Could you please check the return log and waive this fee?",
        "explanation": "Explain when the book was deposited in the drop box versus when it was checked in, and politely ask to waive the erroneous fine.",
    },
    {
        "title": "Asking which textbooks to buy",
        "situation": "You are trying to budget for course textbooks before the semester starts and want to know which books are mandatory. Email or speak to your upcoming literature professor, introduce yourself, and ask if the syllabus is available early.",
        "model_answer": "Hello Professor Higgins, my name is Alex and I'm enrolled in your Victorian Literature seminar this autumn. I'm currently organizing my semester budget and looking for second-hand textbooks. Would it be possible to see a preliminary copy of the reading list or syllabus so I can order the required novels in advance?",
        "explanation": "Introduce yourself politely, explain the practical reason for wanting the reading list early, and request a preliminary syllabus.",
    },
    {
        "title": "Following up a job application",
        "situation": "You submitted an application for a student assistant position at the university international office two weeks ago and haven't heard back. Visit the office, speak to the hiring coordinator, and politely inquire about the status of your application.",
        "model_answer": "Good afternoon. My name is Carlos and I submitted an application two weeks ago for the student assistant position in the international office. I wanted to follow up and see if the search committee has begun reviewing applications, and whether you need any additional documents from my end.",
        "explanation": "State your name and the position applied for, ask politely about the hiring timeline, and offer any extra information.",
    },
    {
        "title": "A broken window blind",
        "situation": "The roller blind in your dorm room snapped when you tried to pull it down last night, leaving your window unshaded from the streetlights. Report the issue to the residence hall maintenance portal supervisor and request a repair.",
        "model_answer": "Hi, I live in room 214 of West Hall. The cord on the window roller blind snapped last night while I was closing it, and now it's stuck completely open. The bright streetlights right outside make it very difficult to sleep. Could you please submit an urgent work order to have the blind repaired or replaced?",
        "explanation": "Give the room number, explain that the blind snapped and is stuck open, mention street light glare, and request maintenance.",
    },
    {
        "title": "Asking for more club funding",
        "situation": "The student council finance committee allocated only half of the requested funding for your environmental club's annual campus recycling festival. Meet with the treasurer, present your detailed itemized budget, and appeal for additional funding.",
        "model_answer": "Hello, thank you for meeting with me. Our environmental club received the allocation notice for the recycling festival, and we noticed our budget was reduced by half. We've prepared an itemized breakdown showing that venue rental and safety equipment make up our core costs. Could you explain the deduction and reconsider funding the safety portion?",
        "explanation": "Express appreciation for the meeting, reference your itemized cost breakdown, and politely appeal for essential funding categories.",
    },
    {
        "title": "Understanding essay feedback",
        "situation": "You received a lower grade than expected on your first philosophy essay and don't understand the written comments. Schedule office hours with your teaching assistant, express your desire to improve, and ask for specific guidance on future essays.",
        "model_answer": "Hello Mr. Davies, thank you for meeting with me. I read through your feedback on my first philosophy paper, but I'm having trouble understanding how to improve my counter-argument structure. I'm keen to do better on the next assignment. Could you walk me through where my analysis fell short and what you'd recommend focusing on?",
        "explanation": "Show openness to constructive criticism, mention a specific area of weakness, and ask for guidance to improve on future assignments.",
    },
    {
        "title": "Missing cutlery for a lunch event",
        "situation": "You are organizing a student symposium lunch and the catering company just delivered seventy boxed meals but forgot to include any forks, napkins, or drinking cups. Phone the catering manager immediately and request a prompt delivery of the missing supplies.",
        "model_answer": "Hi, this is Liam from the student symposium in the arts atrium. We just received the delivery of seventy boxed lunches, but the driver forgot all the cutlery, napkins, and cups. Our lunch break starts in twenty minutes. Could you please send someone over with the utensils right away so our guests can eat?",
        "explanation": "Identify the event and location, explain the missing utensils, emphasize the urgent 20-minute timeline, and request prompt delivery.",
    },
    {
        "title": "Cycling past a group on a shared path",
        "situation": "You are riding your bicycle along a shared pedestrian-cyclist campus path and a group of students wearing headphones is walking three abreast, blocking the path. Alert them politely, ring your bell, and pass safely.",
        "model_answer": "Excuse me, on your left! Sorry to interrupt, I'm just passing by on a bicycle. Thank you so much, have a great afternoon!",
        "explanation": "Give a clear, courteous directional warning ('on your left'), acknowledge the courtesy, and maintain safe cycling speed.",
    },
    {
        "title": "Asking a professor about research work",
        "situation": "You attended an inspiring guest lecture by a neurobiology professor and want to ask if she has any volunteer research assistant openings in her lab. Speak to her after the lecture, mention your interest in her topic, and ask about opportunities.",
        "model_answer": "Professor Zhang, thank you for such an inspiring lecture on synaptic plasticity. I'm a second-year neuroscience student, and I've read several of your lab's recent publications on glial signaling. I was wondering if you might have any volunteer openings for an undergraduate research assistant this semester?",
        "explanation": "Express genuine appreciation for the lecture, mention your academic background and interest, and ask politely about research openings.",
    },
    {
        "title": "Asking a guide to slow down",
        "situation": "You are on a walking tour of a city with your elderly grandfather, and the guide is walking too fast for him to keep up. Speak to the guide politely and ask whether the group could walk a little more slowly.",
        "model_answer": "Excuse me, I'm really enjoying the tour, but my grandfather is finding it hard to keep up with the pace. Would it be possible to walk a little more slowly, or to pause for a moment at each stop? That would make a big difference to him. Thank you.",
        "explanation": "Start positively, explain the difficulty tactfully, suggest a specific change and thank the guide.",
    },
    {
        "title": "No internet in a basement room",
        "situation": "You and your classmates frequently lose wireless internet connection in the new basement seminar room of the life sciences building. Report the issue to the campus IT help desk, give the room details, and ask if a signal booster can be added.",
        "model_answer": "Hello, I'd like to report a persistent WiFi dead zone. In room B08 of the Life Sciences building, students consistently lose internet connectivity during lectures, which makes accessing online course files impossible. Could the network team check the access points in that basement or install a signal booster?",
        "explanation": "Specify the exact building and basement room, describe the recurring connectivity failure, and request an IT technician check or booster.",
    },
    {
        "title": "Noise in the silent study area",
        "situation": "Two students are having an animated, loud social conversation in the designated silent study zone of the science library. Approach them politely, remind them of the zone's quiet rules, and suggest the ground floor cafe.",
        "model_answer": "Hi, excuse me. I'm really sorry to interrupt, but this whole floor is designated as a silent study zone and people are trying to prepare for exams. If you need to chat or take a break, there's plenty of open social seating downstairs in the cafe. Thank you for understanding.",
        "explanation": "Use a polite, respectful tone, point out the silent study designation, and suggest an alternative social location.",
    },
    {
        "title": "Asking to skip a required course",
        "situation": "You want to enroll in an advanced data science elective, but the system says you lack an introductory statistics prerequisite that you actually completed at your previous college. Speak to your academic advisor, show your transcript, and ask for a prerequisite waiver.",
        "model_answer": "Hello Dr. Kelly. I'm trying to enroll in Data Science 301, but the registration system flagged me for missing the introductory statistics prerequisite. I actually completed an equivalent course with an A grade at my previous university. Here is my certified transfer transcript. Could you please grant a prerequisite waiver?",
        "explanation": "Explain the registration block, show proof of the equivalent completed coursework on your transfer transcript, and ask for the prerequisite waiver.",
    },
    {
        "title": "Handing in car keys you found",
        "situation": "You found a set of car keys with a university gym tag lying on the grass outside the student union. Bring the keys to the campus security office, describe where you found them, and hand them over.",
        "model_answer": "Hello officer, I just found this set of car keys on the lawn right outside the north entrance of the student union about ten minutes ago. There's a campus gym barcode attached to the keyring, so you might be able to scan it and identify the student who lost them.",
        "explanation": "Describe the exact location and time the keys were found, highlight the gym barcode to assist identification, and hand them in.",
    },
    {
        "title": "Moving an appointment online",
        "situation": "You have a mandatory academic advising appointment scheduled for 10 AM tomorrow, but you have developed a severe cough and fever. Email or phone your advisor, apologize for the short notice, and ask to reschedule or hold the meeting via video call.",
        "model_answer": "Good morning Ms. Clark. I apologize for the short notice, but I woke up this morning with a high fever and a severe cough, and I want to avoid spreading illness. Could we possibly conduct our advising appointment via Zoom tomorrow, or reschedule for next week when I've recovered?",
        "explanation": "Apologize for the short notice, explain illness symptoms courteously, and offer an immediate video call option or future rescheduling.",
    },
    {
        "title": "A flooded pitch booking",
        "situation": "You booked the outdoor soccer pitch for your intramural team tonight, but torrential rain has flooded the field. Call the campus sports recreation desk, explain the weather conditions, and ask to cancel the booking without penalty.",
        "model_answer": "Hello, I booked the outdoor soccer field for seven o'clock tonight under the name Intramural Hawks. With the torrential rain today, the pitch is completely flooded and unplayable. Could we please cancel tonight's booking and receive a credit towards our next match without paying a cancellation fee?",
        "explanation": "Identify the booking name, describe the weather and unplayable field conditions, and politely ask for a penalty-free cancellation or credit.",
    },
    {
        "title": "Returning a gift without a receipt",
        "situation": "Someone gave you a jacket as a present, but it is too small and you do not have the receipt. Go to the shop and ask whether you can exchange it for a larger size.",
        "model_answer": "Hello, I was given this jacket as a gift, but unfortunately it's too small for me. I don't have the receipt, but the labels are still on it. Would it be possible to exchange it for a larger size, or for store credit?",
        "explanation": "Explain that it was a gift, mention its condition, ask about an exchange and offer an alternative such as store credit.",
    },
    {
        "title": "Storing luggage over the holidays",
        "situation": "You are an international student living in campus dorms and need a secure place to store your luggage over the three-month summer break. Speak to the housing facilities manager and ask if the university offers summer trunk storage.",
        "model_answer": "Hello, I'm an international student living in East Hall, and I'll be flying home for the summer vacation. Does the housing office offer secure summer luggage storage for international residents, or do you have a recommended local storage service that provides student discounts?",
        "explanation": "State your international resident status, explain the summer break timing, and inquire about campus storage or discounted partner options.",
    },
    {
        "title": "The wrong textbook delivered",
        "situation": "You ordered a required medical terminology textbook from the university bookstore's online service, but when the package arrived, it contained an introductory botany manual instead. Return to the bookstore dispatch counter, present the package, and ask for the correct book.",
        "model_answer": "Hi, I placed an online order for the Medical Terminology textbook, but when I opened the package this morning, it contained an Introductory Botany book instead. Here is my order confirmation and the packing slip. Could you please exchange this for the correct textbook I ordered?",
        "explanation": "Show the order confirmation, explain the packing mistake clearly, and request the immediate exchange for the ordered textbook.",
    },
]

assert len(RTS_ITEMS) == 50
