"""10 original PTE Core Respond to a Situation items.

Format from Pearson's PTE Core test format (Speaking & Writing) and Score Guide
(January 2024): an everyday situation of up to 60 words, read and heard, 20
seconds to prepare and 40 seconds to respond, rated on Appropriacy,
Pronunciation and Fluency.
"""

RTS_ITEMS = [
    {
        "title": "A prescription that is not ready",
        "situation": "You have come to the chemist to collect a prescription for your daughter's ear infection. The assistant tells you it will not be ready until tomorrow, but your daughter needs to start the medicine today. Speak to the assistant.",
        "model_answer": "I'm sorry, but my daughter really needs to start these antibiotics today. The doctor said the infection could get worse if we wait. Is there any way you could prepare it this afternoon? I'm happy to wait here or come back later. If that isn't possible, could you tell me whether another chemist nearby might have it in stock?",
        "explanation": "Stay polite but explain why it is urgent, then ask for a solution and offer a fallback. A calm, cooperative tone works better than a complaint.",
    },
    {
        "title": "Turning down a wedding invitation",
        "situation": "A close friend has invited you to her wedding in another city. You have just found out that you must travel abroad for work that weekend and cannot go. You call your friend to tell her.",
        "model_answer": "Hi Emma, it's me. First of all, thank you so much for the invitation. I was really looking forward to it. Unfortunately, my company has just told me I have to go to a conference in Germany that weekend, and I can't change it. I'm so sorry to miss your big day. Can I take you both out for dinner when you're back from your honeymoon?",
        "explanation": "With a friend, an informal register is appropriate. Thank her, give the reason, say you are sorry and suggest another way to celebrate.",
    },
    {
        "title": "No heating in a hotel room",
        "situation": "You have checked into a hotel for three nights. It is a cold evening, and the heater in your room does not work. You go to the reception desk to speak to the receptionist.",
        "model_answer": "Good evening. I'm staying in room 214, and I'm afraid the heater isn't working. I've tried the controls several times, but the room is still very cold. Could you send someone to look at it tonight? If it can't be fixed quickly, would it be possible to move to another room? I'm here for three nights, so I'd really like to get it sorted.",
        "explanation": "Give your room number, describe the problem, and make a clear request with an alternative. A polite, formal register suits a hotel reception.",
    },
    {
        "title": "Breaking a colleague's umbrella",
        "situation": "This morning a colleague lent you her umbrella because it was raining heavily. On your way back to the office, the wind blew it inside out and broke it. Speak to your colleague.",
        "model_answer": "Hi Rachel, I'm really sorry, but I've got some bad news about your umbrella. The wind was so strong on the way back that it blew it inside out, and two of the spokes snapped. I feel terrible because you were kind enough to lend it to me. I'd like to buy you a new one. Could you tell me where you got it, or would you like me to choose one?",
        "explanation": "Apologise at the start, explain what happened honestly and offer to put it right. A friendly but sincere tone suits a colleague.",
    },
    {
        "title": "Wrong change at the supermarket",
        "situation": "You have just paid for your shopping with a twenty-pound note. When you check your change, you notice the cashier has given you five pounds too little. Speak to the cashier.",
        "model_answer": "Excuse me, I'm sorry to bother you, but I think there's been a small mistake with my change. My shopping came to eleven pounds forty, and I paid with a twenty-pound note, so I should have eight pounds sixty. But you've given me three pounds sixty. Could you check it for me, please? I know it's busy, so it's easy to do.",
        "explanation": "Point out the error politely without blaming the cashier. Give the exact numbers so the problem is clear, and ask them to check.",
    },
    {
        "title": "Running late for a hair appointment",
        "situation": "You have an appointment at the hairdresser's at 2 pm. Your bus has broken down, and you will be about twenty minutes late. You phone the salon to speak to the receptionist.",
        "model_answer": "Hello, my name is Kate Morgan, and I have an appointment at two o'clock with Louise. I'm really sorry, but my bus has broken down, and I'm going to be about twenty minutes late. Is it still possible for Louise to see me if I arrive at around twenty past? If not, I completely understand, and I'd be happy to book another time later this week.",
        "explanation": "Say who you are and which appointment it is, explain the delay briefly, and ask whether they can still fit you in, offering to rebook if not.",
    },
    {
        "title": "Welcoming new neighbours",
        "situation": "A family has just moved into the house next door to you. You see them carrying boxes from a van and decide to go and introduce yourself.",
        "model_answer": "Hi there, welcome to the street! I'm Ben, and I live next door at number twelve with my wife and our two boys. It looks like you've got a lot to carry. Do you need a hand with any of the heavier boxes? Also, the bins are collected on Tuesday mornings, and there's a good bakery just around the corner. Let us know if you need anything at all.",
        "explanation": "A friendly, informal greeting suits this situation. Introduce yourself, offer practical help and give some useful local information.",
    },
    {
        "title": "Checking the bus route",
        "situation": "You need to get to the city hospital for an appointment, but you are not sure which bus to take. A bus has just stopped, and you speak to the driver before getting on.",
        "model_answer": "Excuse me, does this bus go to the city hospital? I've got an appointment there at half past ten, and I'm not sure which bus I need. If it does, could you tell me which stop is closest to the main entrance, please? And if it doesn't, could you tell me which number I should take instead? Thank you very much.",
        "explanation": "Ask a direct question first, then give the reason and follow-up questions. Keep it short and polite, as the driver is working.",
    },
    {
        "title": "Asked to work at the weekend",
        "situation": "Your manager asks you to work this Saturday because a colleague is ill. You have already booked tickets to take your parents to a concert that day. Reply to your manager.",
        "model_answer": "I'd really like to help, and I'm sorry to hear about Jenny. The problem is that I've already booked concert tickets for my parents on Saturday as a birthday present, and I can't change the date. But I could stay late on Friday to finish anything urgent, or I could work on Sunday instead if that would help. Would either of those work for you?",
        "explanation": "Show that you want to help, explain your reason briefly and offer alternatives. This keeps a polite, professional tone while still saying no.",
    },
    {
        "title": "A car park machine that takes no cards",
        "situation": "You have parked at a shopping centre car park. The ticket machine is not accepting your card, and you have no cash. You go to the car park office to speak to the attendant.",
        "model_answer": "Hello, I'm trying to pay for my parking, but the machine on level two won't accept my card. I've tried it three times, and I also tried a different card, but it keeps saying 'payment declined'. I don't have any cash with me today. Is there another way I can pay, perhaps here at the office or by phone? I don't want to get a fine.",
        "explanation": "Explain what you tried, why you cannot pay another way, and ask for a practical solution. A clear, polite explanation helps the attendant help you.",
    },
]
