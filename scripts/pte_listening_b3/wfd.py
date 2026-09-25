"""PTE Listening Bank 3: Write from Dictation (50 items)
Canonical type_id: 'write-from-dictation'
Live type_name: 'Write from Dictation'
Constraints:
- Sentence length: 8-14 words
- Everyday academic sentences a student would hear on campus or in a lecture
- Key matches transcript exactly
"""

WFD_RAW = [
    ("Overdue loans", "Overdue books must be returned before new loans can be approved."),
    ("Tutorial registration", "Please register for your tutorial group by the end of this week."),
    ("Draft deadline", "The first draft of your essay is due on Friday afternoon."),
    ("Individual marks", "Each member of the group will receive an individual mark."),
    ("Course texts", "The reading list for this course is available on the website."),
    ("Exam timetable", "The final examination timetable will be published next month."),
    ("Protective glasses", "Protective glasses must be worn at all times in the laboratory."),
    ("Office hours", "The lecturer holds office hours every Tuesday and Thursday morning."),
    ("Print top-ups", "Students can add printing credit at the machines near the entrance."),
    ("Field trip", "The field trip was postponed because of heavy rain in the valley."),
    ("Referencing style", "All sources should be cited using the same referencing style."),
    ("Evening classes", "The survey results suggest that most students prefer evening classes."),
    ("Careers fair", "Several major employers will attend the careers fair in March."),
    ("Accommodation office", "The accommodation office can help you find a room near campus."),
    ("Peer feedback", "Honest feedback from classmates can improve the quality of your writing."),
    ("Scholarship rules", "Scholarship holders must maintain good grades throughout the academic year."),
    ("A class outing", "Our class will visit the history museum on Wednesday morning."),
    ("Data privacy", "Personal information collected in the study will be kept confidential."),
    ("Study abroad", "Students who study abroad often return with greater confidence."),
    ("Guest lecture", "The guest lecture has been moved to the main hall."),
    ("Staying hydrated", "Drinking enough water helps the body regulate its temperature."),
    ("Desert merchants", "Merchants carried spices and textiles across the desert for centuries."),
    ("Renewable targets", "The government has set ambitious targets for renewable electricity."),
    ("Sample size", "A larger sample would make the conclusions of the study more reliable."),
    ("Presentation slides", "Keep your presentation slides simple and avoid long paragraphs of text."),
    ("Assignment questions", "Questions about the assignment can be posted on the online forum."),
    ("Conversation evenings", "The language club meets every Monday evening in the student centre."),
    ("Cycling to campus", "More students are cycling to campus since the new paths opened."),
    ("Consumer spending", "Economic growth slowed last year as consumer spending began to fall."),
    ("Healthy breakfast", "Researchers found that a healthy breakfast improved concentration in young children."),
    ("Recycling bins", "Paper and plastic should be placed in separate recycling bins."),
    ("Thesis supervisor", "Your thesis supervisor must approve the topic before you begin."),
    ("Ethics approval", "Research involving human participants requires approval from the ethics committee."),
    ("Late submissions", "Late submissions will lose five percent of the available marks each day."),
    ("Sports centre", "The sports centre offers free fitness classes during the first week."),
    ("A warming trend", "Long-term climate records show a clear warming trend over the century."),
    ("Representing students", "The student union represents the interests of all enrolled students."),
    ("Graduation ceremony", "Tickets for the graduation ceremony are limited to two guests."),
    ("Cheaper fares", "Cheaper public transport could reduce traffic in the city centre."),
    ("Comparing accounts", "Historians compare several sources before accepting any single account."),
    ("Marine reserve", "The marine reserve has helped fish populations recover along the coast."),
    ("Writing centre", "The writing centre offers one-to-one support with academic essays."),
    ("Timetable clash", "Contact the faculty office if two of your classes clash."),
    ("Getting there first", "Birds that migrate earlier in spring may find more food."),
    ("Budget report", "The annual budget report will be discussed at the next meeting."),
    ("Computer lab", "The computer lab on the third floor is open until midnight."),
    ("Work placements", "Students may receive academic credit for an approved summer internship."),
    ("Translation project", "The translation project brought together volunteers from twelve different countries."),
    ("Nutrition labels", "Clear nutrition labels help shoppers compare products more easily."),
    ("Final reminder", "Remember to switch off your phones before the examination begins."),
]

WFD_ITEMS = []
for idx, (title, sentence) in enumerate(WFD_RAW, 1):
    qid = f"q-pl90-wfd-{idx:02d}"
    WFD_ITEMS.append({
        "id": qid,
        "type_id": "write-from-dictation",
        "type_name": "Write from Dictation",
        "title": title,
        "sentence": sentence,
        "prompt": "You will hear a sentence. Type the sentence in the box below exactly as you hear it. Write as much of the sentence as you can. You will hear the sentence only once.",
        "audio_transcript": sentence,
        "correct_answers": [sentence],
        "time_limit_seconds": 90,
        "prep_time_seconds": 7,
        "points": 10,
        "explanation": f"Reference sentence: {sentence}"
    })
