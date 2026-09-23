"""Generate original IELTS General Training practice content (migration 000071).

Writing Task 1 letters and General Training Reading section 1 and 2 texts.
All content is original Prepyo practice material, written to the public
IELTS General Training format; it is not official or recalled test content.
Section 3 of the General Training paper draws on the existing original
general-interest articles (000054, 000067), tagged for both modules in 000070.

The script checks its own answer keys before writing SQL: sentence-completion
answers must appear word for word in the paragraph the explanation cites,
ordered task types must follow the text, and every set must have the counts
the General Training blueprint deals. Output is deterministic and replay-safe
(ON CONFLICT DO NOTHING preserves later editorial changes).
"""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
VERSION = 'ielts-2026-01'

# ---------------------------------------------------------------------------
# Writing Task 1: letters
# ---------------------------------------------------------------------------

LETTER_TAIL = ('\n\nWrite at least 150 words.\n\nYou do NOT need to write any addresses.'
               '\n\nBegin your letter as follows:\n\n{opening}')

letters = [
    ('formal', 'A damaged delivery',
     'You recently bought a washing machine from a large electrical store, and it was delivered to your home yesterday. When you unpacked it, you found that it was damaged.',
     'the manager of the store',
     ['describe what you bought and when it was delivered', 'explain what is wrong with the washing machine', 'say what you would like the store to do'],
     'Dear Sir or Madam,'),
    ('formal', 'A neglected park',
     'There is a public park near your home, but in recent months it has fallen into poor condition.',
     'the local council',
     ['describe the park and how local people use it', 'explain what the problems are', 'suggest what the council could do to improve it'],
     'Dear Sir or Madam,'),
    ('formal', 'A weekend job at a museum',
     'You have seen an advertisement for a part-time weekend job at a local museum.',
     'the manager of the museum',
     ['say which job you are applying for and where you saw the advertisement', 'describe your relevant experience and skills', 'explain why you would like the job'],
     'Dear Sir or Madam,'),
    ('formal', 'Leaving a course early',
     'You are studying on an evening course at a college, but you now have to leave the course before it finishes.',
     'the course director',
     ['give details of the course you are taking', 'explain why you have to leave the course early', 'ask what options there are for completing the course later'],
     'Dear Sir or Madam,'),
    ('semi-formal', 'A noisy dog',
     "Your neighbour's dog has been barking loudly late at night for several weeks.",
     'your neighbour',
     ['describe the problem', 'explain how it is affecting you', 'suggest what your neighbour could do'],
     'Dear ..........,'),
    ('semi-formal', 'Changing your working hours',
     'You would like to change the hours that you work each week.',
     'your manager',
     ['explain why you would like to change your hours', 'say which hours you would prefer to work', 'explain how your work will still be completed'],
     'Dear ..........,'),
    ('semi-formal', 'A broken heating system',
     'You rent a flat, and the heating system has stopped working.',
     'your landlord',
     ['describe the problem with the heating', 'explain what you have already done about it', 'say when you would like it to be repaired and why'],
     'Dear ..........,'),
    ('semi-formal', 'A teacher who is retiring',
     'A teacher who helped you a great deal when you were younger is about to retire.',
     'the teacher',
     ['say why you are writing', 'describe how the teacher helped you', 'tell the teacher what you are doing now'],
     'Dear ..........,'),
    ('informal', 'A friend coming to stay',
     'A friend from another country is coming to stay with you for a week next month.',
     'your friend',
     ['suggest some things you could do together', 'explain the arrangements for when your friend arrives', 'say what your friend should bring'],
     'Dear ..........,'),
    ('informal', 'Missing a friend\'s wedding',
     'A close friend has invited you to their wedding, but you will not be able to go.',
     'your friend',
     ['thank your friend for the invitation', 'explain why you cannot go', 'suggest when you could meet to celebrate'],
     'Dear ..........,'),
    ('informal', 'Advice about a course',
     'A friend is trying to decide which course to study at college and has asked for your advice.',
     'your friend',
     ['say which course you think would suit your friend', 'explain why you think so', 'give some other advice about studying'],
     'Dear ..........,'),
    ('informal', 'Thanks for looking after your home',
     'A friend looked after your home while you were away on holiday.',
     'your friend',
     ['thank your friend for their help', 'tell your friend about your holiday', 'offer to do something for your friend in return'],
     'Dear ..........,'),
]

TONE_GUIDANCE = {
    'formal': 'Tone: formal, to someone you do not know. Open with the salutation given, state your purpose in the first paragraph, and close with "Yours faithfully" and your full name.',
    'semi-formal': 'Tone: semi-formal, polite but personal, to someone you know in a formal relationship. Use their title and surname, state your purpose early, and close with "Yours sincerely" and your full name.',
    'informal': 'Tone: informal and friendly, to someone you know well. Use their first name, write naturally, and close with an informal phrase such as "Best wishes" and your first name.',
}

rows = []
for index, (tone, title, situation, recipient, bullets, opening) in enumerate(letters, 1):
    prompt = (f'You should spend about 20 minutes on this task.\n\n{situation}\n\n'
              f'Write a letter to {recipient}. In your letter\n\n'
              + '\n'.join(f'- {b}' for b in bullets)
              + LETTER_TAIL.format(opening=opening))
    explanation = (TONE_GUIDANCE[tone] + ' Cover all three bullet points and develop each one with relevant detail; '
                   'a missing bullet point limits Task Achievement. Make the purpose of the letter clear, '
                   'organise it into paragraphs, and keep the tone consistent throughout. There is no single correct letter.')
    rows.append(dict(
        id=f'ielts-gt-letter-{index:03}', exam_version_id=VERSION, exam='IELTS', supported_exams=['IELTS'],
        skill='writing', type_id='ielts-writing-task1-letter', type_name='Write a letter',
        title=title, prompt=prompt, explanation=explanation, prep_time_seconds=0, time_limit_seconds=1200,
        points=20, difficulty='medium',
        tags=['Original practice', 'IELTS', 'Writing', 'Task 1', 'General Training', tone.capitalize()],
    ))

# ---------------------------------------------------------------------------
# Reading: General Training sections 1 and 2
# ---------------------------------------------------------------------------
#
# Each set is one "passage" whose lettered paragraphs are the separate short
# texts a General Training section is made of. Questions list the paragraph
# they are answered from (None for NOT GIVEN), which drives the explanation
# and the order checks.

TFNG = [{'id': 'TRUE', 'text': 'True'}, {'id': 'FALSE', 'text': 'False'}, {'id': 'NOT_GIVEN', 'text': 'Not Given'}]

readings = [
    dict(
        key='gt1-millbrook', slot='GT Section 1', difficulty='easy', option_noun='Notice',
        title='Around Millbrook: Community Noticeboard', subtitle='Six notices from a town noticeboard',
        topic='Community notices',
        paragraphs={
            'A': "Millbrook Leisure Centre: pool timetable changes. From Monday 3 March, the main pool at Millbrook Leisure Centre will close at 8 p.m. on weekdays instead of 9.30 p.m., while essential repairs are carried out to the heating system. The learner pool is not affected and will keep its usual opening hours. Early-morning lane swimming, which starts at 6.30 a.m., will also continue as normal. Members who have booked lessons after 8 p.m. will be contacted by the centre and offered a place in a Saturday class at no extra charge. The work is expected to take about six weeks, and updates will be posted at reception and on the centre's website.",
            'B': "For sale: family bicycles. Two adult bicycles and one child's bicycle for sale, all in good working order. The adult bikes are three years old and have been serviced recently at a local cycle shop; the child's bike is suitable for ages six to nine and comes with stabilisers, which can easily be removed. We are moving abroad and must sell quickly, so the price for all three together is £180. We would prefer not to sell them separately. Helmets are included free of charge. Please call Maria after 6 p.m. on weekdays, or at any time at weekends. Buyers must collect the bicycles from Elm Road.",
            'C': "Evening classes at Millbrook College. Millbrook College is running ten-week evening courses in Spanish, Italian and Mandarin, starting in the week of 14 April. Classes last two hours and are held once a week. Beginners are welcome in all three languages, and an intermediate group will also run in Spanish if at least eight people enrol. The fee is £95 per course, with a reduced rate of £60 for people over sixty-five and for full-time students. Course books are not included in the fee. Places can be booked online or in person at the college office, which is open until 7 p.m. from Monday to Thursday.",
            'D': "Found at Millbrook station. Items left on trains or at Millbrook station are kept at the lost property office next to the ticket hall for one month. After that, anything unclaimed is given to local charities, except for passports and bank cards, which are handed to the police. To collect an item, you will need to describe it and show some form of identification. There is no charge for collecting lost property in person, but items that have to be sent by post will be charged at cost. The office is open from 8 a.m. to 4 p.m., Monday to Friday; it is closed at weekends and on public holidays.",
            'E': "The farmers' market is moving. Millbrook's monthly farmers' market, which has been held in the car park behind the town hall for the past twelve years, is moving to a new home. From May, the market will take place in Victoria Park, beside the bandstand, on the first Sunday of every month from 9 a.m. to 1 p.m. The move will make room for twenty more stalls, including several selling hot food. Parking will be available at the park's north entrance, but visitors are encouraged to walk or cycle if they can. Stallholders who wish to apply for a place should contact the market office by the end of March.",
            'F': "Volunteer drivers wanted. Millbrook Community Transport is looking for volunteer drivers to take older residents to hospital appointments. Volunteers use their own cars and are paid a mileage allowance to cover fuel costs. You will need to have held a full driving licence for at least three years, and you must be willing to give at least one morning or afternoon a fortnight. No previous experience of care work is needed, as all new drivers attend a half-day training session. For an application form, visit the Community Transport office on Station Road or telephone 01632 960 418.",
        },
        groups=[
            ('matching', 'reading-matching-information', 'Matching Information',
             'Look at the six notices, A-F. Which notice mentions the following information? Write the correct letter, A-F. You may use any letter more than once.', [
                ('a reduced price for older people', 'C', 'Notice C offers a reduced rate of £60 for people over sixty-five.'),
                ('equipment that is offered free with a purchase', 'B', 'Notice B: helmets are included free of charge with the bicycles.'),
                ('an event that will take place in a different location', 'E', 'Notice E: the market is moving from the town hall car park to Victoria Park.'),
                ('the need to prove who you are', 'D', 'Notice D: you must show some form of identification to collect an item.'),
                ('a deadline for applications', 'E', 'Notice E: stallholders should contact the market office by the end of March.'),
                ('a change that will only last for a limited period', 'A', 'Notice A: the earlier closing time lasts while repairs are carried out, about six weeks.'),
                ('a short course that all new helpers must attend', 'F', 'Notice F: all new drivers attend a half-day training session.'),
            ]),
            ('tfng', 'reading-true-false', 'True / False / Not Given',
             'Do the following statements agree with the information in the notices? Write TRUE if the statement agrees with the information, FALSE if the statement contradicts the information, or NOT GIVEN if there is no information on this.', [
                ('The learner pool will close early while the repairs take place.', 'FALSE', 'A', 'Notice A says the learner pool is not affected and keeps its usual opening hours.'),
                ('The adult bicycles have been checked by a cycle shop recently.', 'TRUE', 'B', 'Notice B says the adult bikes have been serviced recently at a local cycle shop.'),
                ('An intermediate Spanish group will definitely take place.', 'FALSE', 'C', 'Notice C says the intermediate group will run only if at least eight people enrol.'),
                ('Most people who take the college courses are beginners.', 'NOT_GIVEN', None, 'Notice C says beginners are welcome, but gives no information about how many learners are beginners.'),
                ('Unclaimed passports are passed to local charities.', 'FALSE', 'D', 'Notice D says passports and bank cards are handed to the police.'),
                ('There will be more stalls at the market in its new location.', 'TRUE', 'E', 'Notice E says the move will make room for twenty more stalls.'),
                ('Volunteer drivers receive money towards the cost of fuel.', 'TRUE', 'F', 'Notice F says volunteers are paid a mileage allowance to cover fuel costs.'),
            ]),
        ],
    ),
    dict(
        key='gt1-westfield', slot='GT Section 1', difficulty='easy', option_noun='Course',
        title='Westfield Adult Education: Autumn Short Courses', subtitle='Six course descriptions from a college brochure',
        topic='Adult education',
        paragraphs={
            'A': "First Aid for Parents. This one-day course covers the most common emergencies involving babies and young children, including choking, burns and high temperatures. It runs on Saturday 11 October from 10 a.m. to 4 p.m. and costs £40, which includes a light lunch. Everyone who completes the day receives a certificate, although this is not a professional qualification and cannot be used for work in childcare. Parents may not bring children to the session. Grandparents and other carers are also welcome to attend.",
            'B': "Introduction to Bookkeeping. Aimed at people who run a small business or are planning to start one, this eight-week course explains how to keep simple accounts, record expenses and prepare for a tax return. Sessions take place on Tuesday evenings from 6.30 to 8.30 p.m. Participants should bring their own laptop, as the college computer room is not available in the evenings. No previous knowledge of accounting is needed. The fee is £120, and payment can be made in two instalments if you prefer.",
            'C': "Photography with Your Phone. Learn to take better photographs using nothing more than the phone in your pocket. Over four Thursday afternoons, you will practise composition, lighting and simple editing. Two of the sessions are held outdoors in the town centre; if the weather is poor, these will be moved indoors and a different topic will be covered instead. The course is free for people who are currently unemployed, and costs £45 for everyone else. Please make sure your phone is fully charged before each session.",
            'D': "Cooking on a Budget. In six weekly sessions at the Riverside Community Centre kitchen, you will learn to plan meals, reduce food waste and cook healthy dishes for less than £2 a portion. All ingredients are provided and are included in the course fee of £75, and you will take home what you cook each week. Numbers are limited to ten people in each group because of the size of the kitchen, so early booking is advised. Please tell us about any food allergies when you enrol.",
            'E': "Job Interview Skills. Run in partnership with several local employers, this two-evening course helps you prepare for job interviews. On the first evening, managers from local firms explain what they look for in candidates. On the second, each participant takes part in a practice interview, which is filmed so that you can watch it back and receive feedback. The course is free, but places must be reserved in advance. It is particularly suitable for people who are returning to work after a long break.",
            'F': "Beginners' Yoga. This ten-week class introduces the basic positions and breathing techniques of yoga. It is suitable for all ages and fitness levels, and mats are provided, although you are welcome to bring your own. Wear loose, comfortable clothing and avoid eating a large meal in the two hours before a class. If you are pregnant or have a back injury, please bring a note from your doctor confirming that you can take part. The fee is £85 for the full course, and a free trial class is held in the first week.",
        },
        groups=[
            ('matching', 'reading-matching-information', 'Matching Information',
             'Look at the six course descriptions, A-F. Which course mentions the following information? Write the correct letter, A-F. You may use any letter more than once.', [
                ('a college facility that participants cannot use', 'B', 'Course B: the college computer room is not available in the evenings.'),
                ('taking home food that participants have prepared', 'D', 'Course D: you will take home what you cook each week.'),
                ('a change of content that depends on the weather', 'C', 'Course C: outdoor sessions move indoors with a different topic if the weather is poor.'),
                ('feedback based on a recording of the participant', 'E', 'Course E: the practice interview is filmed so participants can watch it back and receive feedback.'),
                ('permission from a doctor for some participants', 'F', 'Course F: pregnant people or those with a back injury need a note from their doctor.'),
                ('a document that cannot be used to get certain work', 'A', 'Course A: the certificate is not a professional qualification and cannot be used for work in childcare.'),
                ('the option of paying the fee in two parts', 'B', 'Course B: payment can be made in two instalments.'),
            ]),
            ('tfng', 'reading-true-false', 'True / False / Not Given',
             'Do the following statements agree with the information in the course descriptions? Write TRUE if the statement agrees with the information, FALSE if the statement contradicts the information, or NOT GIVEN if there is no information on this.', [
                ('Children can attend the first aid course with their parents.', 'FALSE', 'A', 'Course A says parents may not bring children to the session.'),
                ('The bookkeeping course is taught by a qualified accountant.', 'NOT_GIVEN', None, 'Course B does not say who teaches the course.'),
                ('Some photography sessions are held outside.', 'TRUE', 'C', 'Course C says two of the sessions are held outdoors in the town centre.'),
                ('The cooking course has been running for several years.', 'NOT_GIVEN', None, 'Course D gives no information about how long the course has been running.'),
                ('There is a limit on the size of each cooking group.', 'TRUE', 'D', 'Course D says numbers are limited to ten people in each group.'),
                ('The interview course is aimed mainly at school leavers.', 'FALSE', 'E', 'Course E says it is particularly suitable for people returning to work after a long break.'),
                ('Yoga participants must bring their own mats.', 'FALSE', 'F', 'Course F says mats are provided, although participants may bring their own.'),
            ]),
        ],
    ),
    dict(
        key='gt2-harlow', slot='GT Section 2', difficulty='medium',
        title='Harlow Logistics Staff Handbook: Annual Leave and Absence', subtitle='A section of an employee handbook',
        topic='Workplace policies',
        paragraphs={
            'A': "Annual leave entitlement. Full-time employees are entitled to 25 days of paid annual leave each year, in addition to public holidays. Part-time staff receive the same entitlement calculated in proportion to the hours they work; for example, someone who works three days a week receives 15 days. The leave year runs from 1 April to 31 March. Staff who join or leave the company part-way through the year receive a share of the full entitlement, which the payroll team will calculate. After five years of continuous service, employees receive two additional days of leave each year, and after ten years a further three days.",
            'B': "Booking leave. All leave must be requested through the online staff portal and approved by your line manager before you make any travel arrangements. Requests for periods of more than five working days should be submitted at least four weeks in advance, while shorter periods require one week's notice. Managers aim to respond to requests within three working days. Because the warehouse is especially busy in the weeks before major public holidays, leave requested for December is allocated in the order in which requests are received, and no more than a quarter of any team may be absent at the same time.",
            'C': "Carrying leave over. We encourage staff to take their full entitlement within the leave year, as regular breaks are important for health and concentration. However, up to five unused days may be carried over into the following year, provided they are taken before the end of June. Any days carried over that have not been used by that date will be lost. In exceptional circumstances, such as long-term illness, a manager may agree to extend this deadline, but such requests must be made in writing to the Human Resources department.",
            'D': "Reporting sickness. If you are unable to come to work because of illness, you must telephone your line manager, rather than send a text message or an email, before the start of your shift, or no later than one hour after it begins if you are too unwell to call earlier. You should explain the nature of the illness and say when you expect to return. If your manager is not available, you should call the duty supervisor instead. Failure to report absence correctly may result in the day being treated as unauthorised leave, which is unpaid.",
            'E': "Evidence and returning to work. For absences of up to seven calendar days, you will be asked to complete a self-certification form on your first day back. For longer absences, you must provide a fit note from a doctor or another qualified health professional. When you return after any absence of more than two weeks, you will have a short meeting with your manager to discuss whether temporary changes to your duties or hours would help you settle back in. These meetings are intended to be supportive and are not part of any disciplinary process.",
            'F': "Other types of leave. Employees may take up to two days of paid compassionate leave following the death of a close family member, and further unpaid leave may be granted at the manager's discretion. Staff who serve as volunteers with the emergency services are entitled to up to ten days of paid leave a year for training and call-outs. Requests for time off to attend medical appointments should, where possible, be made for the beginning or end of a shift, so that disruption to the team is kept to a minimum.",
        },
        groups=[
            ('completion', 'reading-sentence-completion', 'Sentence Completion',
             'Complete the sentences below. Choose ONE WORD ONLY from the text for each answer.', [
                ('Part-time staff receive leave in ________ to the hours they work.', 'proportion', 'A', 'Paragraph A: part-time entitlement is calculated in proportion to the hours worked.'),
                ('Leave must be approved before staff make any ________ arrangements.', 'travel', 'B', 'Paragraph B: leave must be approved before you make any travel arrangements.'),
                ('Unused days that are carried over must be taken before the end of ________.', 'June', 'C', 'Paragraph C: carried-over days must be taken before the end of June.'),
                ('Staff who are ill must ________ their line manager rather than send a message.', 'telephone', 'D', 'Paragraph D: you must telephone your line manager rather than send a text message or email.'),
                ('After a long absence, staff may discuss temporary changes to their duties or ________.', 'hours', 'E', 'Paragraph E: the meeting considers temporary changes to your duties or hours.'),
            ]),
            ('tfng', 'reading-true-false', 'True / False / Not Given',
             'Do the following statements agree with the information in the text? Write TRUE if the statement agrees with the information, FALSE if the statement contradicts the information, or NOT GIVEN if there is no information on this.', [
                ('Employees receive extra leave after five years of service.', 'TRUE', 'A', 'Paragraph A: after five years of continuous service, employees receive two additional days.'),
                ('Up to half of a team may be on leave at the same time in December.', 'FALSE', 'B', 'Paragraph B: no more than a quarter of any team may be absent at the same time.'),
                ('Requests to extend the carry-over deadline must be made in writing.', 'TRUE', 'C', 'Paragraph C: such requests must be made in writing to the Human Resources department.'),
                ('The company has more sickness absence than other warehouses.', 'NOT_GIVEN', None, 'Paragraph D explains how to report sickness but does not compare absence levels with other warehouses.'),
            ]),
            ('single', 'reading-mcq-single', 'Multiple Choice, Single Answer',
             'Choose the correct letter, A, B, C or D.', [
                ('How many days of annual leave does a member of staff who works three days a week receive?',
                 ['25 days', '15 days', '10 days', '5 days'], 'B', 'A', 'Paragraph A gives this example: someone who works three days a week receives 15 days.'),
                ('How is leave in December allocated?',
                 ['according to length of service', 'by the line manager\'s choice', 'in the order requests are received', 'by random selection'], 'C', 'B', 'Paragraph B: December leave is allocated in the order in which requests are received.'),
                ('What should an employee do if they are too ill to call before their shift?',
                 ['call within one hour of the shift starting', 'contact the Human Resources department', 'visit a doctor before calling', 'send an email by the end of the day'], 'A', 'D', 'Paragraph D: call no later than one hour after the shift begins if too unwell to call earlier.'),
                ('What can staff who volunteer with the emergency services take?',
                 ['two days of paid leave', 'up to ten days of paid leave a year', 'unlimited unpaid leave', 'leave only at the start of a shift'], 'B', 'F', 'Paragraph F: they are entitled to up to ten days of paid leave a year.'),
            ]),
        ],
    ),
    dict(
        key='gt2-riverside', slot='GT Section 2', difficulty='medium',
        title='Riverside Hotel Group: A Guide for New Front-of-House Staff', subtitle='Information for new employees',
        topic='Workplace induction',
        paragraphs={
            'A': "Welcome. This guide explains the main arrangements for new front-of-house staff, including receptionists, porters and restaurant hosts. It should be read together with your contract of employment, which sets out your hours and pay. During your first two weeks, you will be paired with an experienced colleague, known as your 'buddy', who can answer everyday questions about the hotel and its routines. More formal matters, such as holiday requests or changes to your contract, should always be discussed with your department manager rather than with your buddy.",
            'B': "Uniform. The hotel provides three sets of uniform, which remain the property of the company and must be returned when you leave. Uniforms are cleaned free of charge by the hotel laundry: simply leave them in the marked bags in the staff changing room before the end of your shift, and they will be ready within two days. Name badges must be worn at all times in public areas of the hotel. Staff are asked to wear plain black shoes, which are not supplied by the hotel, and to keep jewellery to a minimum.",
            'C': "Shift patterns. Front-of-house departments operate 24 hours a day, and most staff work a rota of early, late and occasional night shifts. Rotas for the following fortnight are published every Friday. If you wish to swap a shift with a colleague, both of you must agree, and the swap must also be approved by the duty manager at least 48 hours in advance. Staff working a night shift receive an additional payment for each hour worked between midnight and 6 a.m., and a free hot meal is provided during the break.",
            'D': "Training. All new staff complete a two-day induction, covering fire safety, customer service standards and the hotel's booking system. In addition, the company offers a range of optional courses, including languages and first aid. These courses are free of charge and take place during working hours, but places are limited and are allocated by department managers. Staff who complete a recognised qualification in hospitality may be eligible for a one-off bonus, and details can be obtained from the Human Resources office.",
            'E': "Staff facilities. A staff canteen on the lower ground floor serves breakfast, lunch and dinner at subsidised prices. Lockers are available in the changing rooms, and keys can be collected from the security office on your first day. A small deposit is taken, which is returned when the key is handed back. Staff are not permitted to use the guest gym or swimming pool, but they can join a local fitness centre at a discounted rate through the company's partnership scheme.",
            'F': "Guest feedback. Comments from guests are reviewed every week, and members of staff who are mentioned positively are recognised in the monthly staff newsletter. Complaints are handled by the duty manager, and staff should never try to resolve a serious complaint on their own. If a guest becomes aggressive, staff should remain calm, avoid arguing, and call security using the emergency number printed on the back of their name badge.",
        },
        groups=[
            ('completion', 'reading-sentence-completion', 'Sentence Completion',
             'Complete the sentences below. Choose ONE WORD ONLY from the text for each answer.', [
                ('New staff are paired with an experienced colleague for their first two ________.', 'weeks', 'A', 'Paragraph A: during your first two weeks you will be paired with a buddy.'),
                ('Staff must provide their own plain black ________.', 'shoes', 'B', 'Paragraph B: plain black shoes are not supplied by the hotel.'),
                ('Rotas for the following fortnight are published every ________.', 'Friday', 'C', 'Paragraph C: rotas are published every Friday.'),
                ('Optional courses are free and take place during working ________.', 'hours', 'D', 'Paragraph D: the courses take place during working hours.'),
                ('A small ________ is taken when staff collect a locker key.', 'deposit', 'E', 'Paragraph E: a small deposit is taken and returned when the key is handed back.'),
            ]),
            ('tfng', 'reading-true-false', 'True / False / Not Given',
             'Do the following statements agree with the information in the text? Write TRUE if the statement agrees with the information, FALSE if the statement contradicts the information, or NOT GIVEN if there is no information on this.', [
                ('New staff should ask their buddy about holiday requests.', 'FALSE', 'A', 'Paragraph A: holiday requests should be discussed with the department manager, not the buddy.'),
                ('Uniforms must be given back when an employee leaves the hotel.', 'TRUE', 'B', 'Paragraph B: uniforms remain the property of the company and must be returned when you leave.'),
                ('Night-shift staff are paid extra for hours worked after midnight.', 'TRUE', 'C', 'Paragraph C: an additional payment is made for each hour worked between midnight and 6 a.m.'),
                ('Most new staff choose to take the language courses.', 'NOT_GIVEN', None, 'Paragraph D lists languages among the optional courses but does not say how many staff take them.'),
            ]),
            ('single', 'reading-mcq-single', 'Multiple Choice, Single Answer',
             'Choose the correct letter, A, B, C or D.', [
                ('What do staff need in order to swap a shift?',
                 ['only their colleague\'s agreement', 'approval from Human Resources', 'their colleague\'s agreement and the duty manager\'s approval', 'one week\'s notice'], 'C', 'C', 'Paragraph C: both staff must agree and the duty manager must approve the swap.'),
                ('Who decides which staff get places on optional courses?',
                 ['the Human Resources office', 'department managers', 'the course trainers', 'the longest-serving staff'], 'B', 'D', 'Paragraph D: places are allocated by department managers.'),
                ('What can staff who want to exercise do?',
                 ['use the guest gym after their shift', 'use the pool at quiet times', 'join a local fitness centre at a reduced price', 'get free membership of a gym'], 'C', 'E', 'Paragraph E: staff can join a local fitness centre at a discounted rate.'),
                ('What should staff do if a guest becomes aggressive?',
                 ['try to resolve the complaint quickly', 'call security', 'ask the guest to leave', 'contact the Human Resources office'], 'B', 'F', 'Paragraph F: staff should remain calm, avoid arguing and call security.'),
            ]),
        ],
    ),
]


# Further detail for each text, appended to its paragraphs, so that a General
# Training paper reaches the official 2,150-2,375 words with its section 3
# article. None of it answers or contradicts a question: the key checks below
# run on the merged text.
EXTRA = {
    'gt1-millbrook': {
        'B': 'The bicycles can be viewed before you buy by arrangement.',
        'C': 'Each class has a maximum of fifteen students, so that everyone has a chance to speak.',
        'D': 'Umbrellas, which make up the largest group of items handed in, are kept for only two weeks.',
        'E': 'Dogs on leads are welcome, and a shuttle bus will run from the railway station on market days.',
        'F': 'Journeys usually take place on weekdays between 9 a.m. and 5 p.m.',
    },
    'gt1-westfield': {
        'A': 'The course is taught by a paramedic with many years of experience. Places are limited to twenty and usually fill quickly, so we recommend booking early.',
        'B': 'By the end of the course, you should be able to set up a simple spreadsheet to track the money coming into and going out of your business.',
        'C': "You do not need an expensive phone, as the techniques work on most models from the last five years. At the end of the course, a selection of participants' photographs will be displayed in the college library.",
        'D': 'Sessions are led by a professional chef who has worked in several local restaurants, and aprons are provided.',
        'E': 'You will also receive advice on writing a covering letter and on answering difficult questions, such as why you left your last job. Please bring a copy of your current CV to the first session, even if it is not yet complete, as it will be used during the practice interview.',
        'F': 'Classes are held on Monday evenings in the main hall, which has changing facilities and showers. The tutor has taught yoga at the college for over ten years and will suggest simple exercises that you can practise at home between classes.',
    },
    'gt2-harlow': {
        'A': 'Your remaining entitlement is shown on the online staff portal and is updated at the start of every month.',
        'B': 'If a request is refused, your manager will explain the reason and suggest alternative dates where possible. Leave that has been approved will only be cancelled by the company in an emergency, and you will be compensated for any costs you cannot recover.',
        'C': 'Staff who are close to the carry-over limit will receive a reminder from the payroll team in January.',
        'D': 'If you become ill while you are at work, tell your supervisor before leaving the site so that your absence can be recorded correctly.',
        'E': "Self-certification forms are available from the staff portal or from reception. If you are absent frequently, your manager may suggest a referral to the company's occupational health adviser.",
        'F': 'Parents may also request unpaid parental leave, which should be discussed with the Human Resources department at least four weeks in advance.',
    },
    'gt2-riverside': {
        'A': 'Your buddy will also show you around the building on your first day, including the fire exits and assembly points. At the end of your third month, you will have a review meeting with your department manager.',
        'B': 'Hair should be neat and tidy, and long hair must be tied back when working in the restaurant. Replacement uniform items can be ordered from the housekeeping department if yours become damaged, although the cost of any items lost through carelessness may be taken from your pay.',
        'C': 'Rotas are planned so that no one works more than two night shifts in a row, and anyone who has difficulty with a particular shift pattern should speak to their department manager.',
        'D': 'The induction also includes a tour of each department, so that you understand how your role fits with the work of others. Language courses are offered in French, Spanish and Mandarin, reflecting the countries our guests most often come from.',
        'E': 'The canteen is open from 6 a.m. until 11 p.m., and a vending machine is available at other times. A quiet room is also provided for staff who need to rest during their breaks.',
        'F': 'Staff are encouraged to pass on suggestions for improving the guest experience, and the best idea each quarter receives a small prize. Any comments that mention health or safety concerns are passed immediately to the duty manager, who will decide what action needs to be taken.',
    },
}
for reading in readings:
    for label, text in EXTRA.get(reading['key'], {}).items():
        reading['paragraphs'][label] += ' ' + text


def words(text):
    return re.findall(r"[A-Za-z0-9'’£.-]+", text)


def check(reading):
    paras = reading['paragraphs']
    for key, type_id, _, _, questions in reading['groups']:
        where = []
        for q in questions:
            if type_id == 'reading-matching-information':
                prompt, answer, _ = q
                assert answer in paras, (reading['key'], prompt)
            elif type_id == 'reading-true-false':
                prompt, answer, para, _ = q
                assert answer in ('TRUE', 'FALSE', 'NOT_GIVEN')
                assert (para is None) == (answer == 'NOT_GIVEN'), (reading['key'], prompt)
                where.append(para)
            elif type_id == 'reading-sentence-completion':
                prompt, answer, para, _ = q
                assert '________' in prompt and len(answer.split()) == 1, prompt
                assert re.search(r'\b' + re.escape(answer) + r'\b', paras[para]), (reading['key'], answer)
                where.append(para)
            elif type_id == 'reading-mcq-single':
                prompt, options, answer, para, _ = q
                assert len(options) == 4 and answer in 'ABCD'
                where.append(para)
        located = [w for w in where if w]
        # Ordered tasks follow the text.
        assert located == sorted(located), (reading['key'], key, located)


for reading in readings:
    check(reading)
counts = {}
for reading in readings:
    for key, type_id, _, _, questions in reading['groups']:
        counts.setdefault((reading['slot'], type_id), set()).add(len(questions))
# The General Training blueprint (000070) deals these counts.
assert counts[('GT Section 1', 'reading-matching-information')] == {7}
assert counts[('GT Section 1', 'reading-true-false')] == {7}
assert counts[('GT Section 2', 'reading-sentence-completion')] == {5}
assert counts[('GT Section 2', 'reading-true-false')] == {4}
assert counts[('GT Section 2', 'reading-mcq-single')] == {4}
assert len(rows) == 12

# ---------------------------------------------------------------------------
# SQL
# ---------------------------------------------------------------------------


def lit(value):
    if value is None:
        return 'NULL'
    if isinstance(value, bool):
        return 'TRUE' if value else 'FALSE'
    if isinstance(value, int):
        return str(value)
    if isinstance(value, (list, dict)):
        return "'" + json.dumps(value, ensure_ascii=False).replace("'", "''") + "'::jsonb"
    return "'" + str(value).replace("'", "''") + "'"


def array(values):
    return 'ARRAY[' + ', '.join(lit(v) for v in values) + ']'


out = []
out.append('-- Generated by scripts/generate_ielts_general_training.py. Original practice content.')
out.append('-- IELTS General Training: 12 Writing Task 1 letters and four reading texts for')
out.append('-- General Training Reading sections 1 and 2 (27 or 26 questions each set of two).')
out.append('-- Not official or recalled IELTS material. Replay preserves editorial changes.')
out.append('')
out.append('INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, explanation, prep_time_seconds, time_limit_seconds, points, difficulty, tags) VALUES')
out.append(',\n'.join(
    '(' + ', '.join([lit(r['id']), lit(r['exam_version_id']), lit(r['exam']), array(r['supported_exams']),
                     lit(r['skill']), lit(r['type_id']), lit(r['type_name']), lit(r['title']), lit(r['prompt']),
                     lit(r['explanation']), lit(r['prep_time_seconds']), lit(r['time_limit_seconds']),
                     lit(r['points']), lit(r['difficulty']), array(r['tags'])]) + ')'
    for r in rows))
out.append('ON CONFLICT (id) DO NOTHING;')

for reading in readings:
    pid = 'rp-' + reading['key']
    paragraphs = [{'label': label, 'text': text} for label, text in reading['paragraphs'].items()]
    word_count = sum(len(words(p['text'])) for p in paragraphs)
    reading['word_count'] = word_count
    tags = ['IELTS Reading', 'General Training', reading['slot'], 'Original practice']
    out.append('')
    out.append(f"-- {reading['title']} ({reading['slot']}, {word_count} words)")
    out.append('INSERT INTO reading_passages (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot, modules) VALUES')
    out.append('(' + ', '.join([lit(pid), lit(VERSION), lit(reading['title']), lit(reading['subtitle']), lit(paragraphs),
                                lit(word_count), lit(reading['difficulty']), lit(reading['topic']), array(tags),
                                lit('custom'), array(['general_training'])]) + ')')
    out.append('ON CONFLICT (id) DO NOTHING;')

    group_rows, question_rows = [], []
    for position, (key, type_id, type_name, instructions, questions) in enumerate(reading['groups'], 1):
        gid = f"g-{reading['key']}-{key}"
        group_rows.append('(' + ', '.join([lit(gid), lit(pid), lit(position), lit(type_id), lit(type_name),
                                           lit(instructions), "'[]'::jsonb", lit('full'), 'FALSE', '0']) + ')')
        for index, q in enumerate(questions, 1):
            qid = f"q-{reading['key']}-{key}-{index}"
            if type_id == 'reading-matching-information':
                prompt, answer, explanation = q
                noun = reading.get('option_noun', 'Paragraph')
                options = [{'id': label, 'text': f'{noun} {label}'} for label in reading['paragraphs']]
                correct, sources = [answer], [answer]
            elif type_id == 'reading-true-false':
                prompt, answer, para, explanation = q
                options, correct, sources = TFNG, [answer], ([para] if para else [])
            elif type_id == 'reading-sentence-completion':
                prompt, answer, para, explanation = q
                options, correct, sources = [], [answer], [para]
            else:
                prompt, texts, answer, para, explanation = q
                options = [{'id': 'ABCD'[i], 'text': t} for i, t in enumerate(texts)]
                correct, sources = [answer], [para]
            question_rows.append('(' + ', '.join([
                lit(qid), lit(VERSION), lit('IELTS'), array(['IELTS']), lit('reading'), lit(type_id), lit(type_name),
                lit(reading['title']), lit(prompt), lit(options), lit(correct), lit(explanation), '1',
                lit(reading['difficulty']), array(['IELTS Reading', 'General Training', 'Original practice']),
                lit(pid), lit(gid), lit(index), 'TRUE', '0', array(sources) + '::text[]']) + ')')
    out.append('INSERT INTO reading_question_groups (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds) VALUES')
    out.append(',\n'.join(group_rows))
    out.append('ON CONFLICT (id) DO NOTHING;')
    out.append('INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, options, correct_answers, explanation, points, difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds, source_paragraphs) VALUES')
    out.append(',\n'.join(question_rows))
    out.append('ON CONFLICT (id) DO NOTHING;')

(ROOT / 'migrations/000071_ielts_general_training_content.up.sql').write_text('\n'.join(out) + '\n', encoding='utf-8', newline='\n')
(ROOT / 'migrations/000071_ielts_general_training_content.down.sql').write_text(
    '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n', encoding='utf-8', newline='\n')
print('letters:', len(rows))
for reading in readings:
    print(reading['key'], reading['slot'], reading['word_count'], 'words')
