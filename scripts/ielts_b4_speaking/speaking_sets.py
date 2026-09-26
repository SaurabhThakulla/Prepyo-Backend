"""IELTS Speaking mock sets 11-20: complete three-part interviews.

Original Prepyo sets in the public IELTS Speaking format: Part 1 two familiar
topics of three questions, Part 2 a cue card with three prompts, an "explain"
prompt and a rounding-off question, Part 3 five discussion questions linked to
the Part 2 theme. Same JSON shape as sets sm-01 to sm-10 (migrations 000076
and 000085). Not official, recalled or copied test content.
"""

SETS = [
    ('sm-11', 'Shopping and money', {
        'part1': [
            {'topic': 'where you grew up', 'questions': ['Where did you grow up?', 'What was it like to live there as a child?', 'Has it changed much since then?']},
            {'topic': 'clothes', 'questions': ['What kind of clothes do you usually wear?', 'Do you enjoy shopping for clothes?', 'Have your tastes in clothes changed as you have got older?']},
        ],
        'part2': {'topic': 'a disappointing purchase', 'cue': 'Describe something you bought that you were not happy with.',
                  'points': ['what you bought', 'where and when you bought it', 'what was wrong with it'],
                  'explain': 'and explain what you did about it.',
                  'rounding': 'Do you usually take back things that you are not happy with?'},
        'part3': {'topic': 'shopping and consumers', 'questions': [
            'Why do people sometimes buy things they later regret?',
            'How can shoppers protect themselves from poor-quality products?',
            'Are online reviews useful when you are deciding what to buy?',
            'Should shops always accept goods that customers want to return?',
            'How have shopping habits in your country changed in recent years?']},
    }),
    ('sm-12', 'Science and discovery', {
        'part1': [
            {'topic': 'your course or job', 'questions': ['Are you studying or working at the moment?', 'Which part of your course or job do you find most interesting?', 'Is there anything you find difficult about it?']},
            {'topic': 'the sky', 'questions': ['Do you often look up at the sky?', 'Can you see the stars clearly where you live?', 'Did you learn anything about the planets when you were at school?']},
        ],
        'part2': {'topic': 'an interesting scientific discovery', 'cue': 'Describe a scientific discovery or a piece of science news that you found interesting.',
                  'points': ['what the discovery or news was', 'how you heard about it', 'what you know about it'],
                  'explain': 'and explain why you found it interesting.',
                  'rounding': 'Did you talk to anyone else about it?'},
        'part3': {'topic': 'science and society', 'questions': [
            'Why do some children lose interest in science as they get older?',
            'Is it more important to spend money on medical research or on space research?',
            'How can scientists explain their work to ordinary people?',
            'Should the results of scientific research be free for everyone to read?',
            'What do you think will be the most important scientific discovery of the next fifty years?']},
    }),
    ('sm-13', 'Languages and understanding', {
        'part1': [
            {'topic': 'where you live now', 'questions': ['Who do you live with at the moment?', 'What is the best thing about the area you live in?', 'How long do you plan to stay there?']},
            {'topic': 'films', 'questions': ['What kind of films do you enjoy?', 'Do you prefer watching films at home or at the cinema?', 'Do you ever watch films with subtitles?']},
        ],
        'part2': {'topic': 'communicating without a shared language', 'cue': 'Describe a time when you had to communicate with someone who did not speak your language.',
                  'points': ['who the person was', 'where you met them', 'how you managed to understand each other'],
                  'explain': 'and explain how you felt about the experience.',
                  'rounding': 'Has anything like this happened to you since?'},
        'part3': {'topic': 'language and communication', 'questions': [
            'Why do people learn foreign languages?',
            'Is it rude to visit a country without learning a few words of its language?',
            'How much can people communicate using gestures and body language?',
            'Will translation apps ever be as good as a human interpreter?',
            'Should English be the main language of international business?']},
    }),
    ('sm-14', 'Rules and responsibilities', {
        'part1': [
            {'topic': 'your job', 'questions': ['What job do you do?', 'What do you do on a typical working day?', 'Would you like to be doing the same job in ten years?']},
            {'topic': 'housework', 'questions': ['Which household jobs do you do?', 'Did you have to help at home when you were a child?', 'Is there any housework you really dislike?']},
        ],
        'part2': {'topic': 'a responsibility you have had', 'cue': 'Describe a responsibility you have had, at home, at school or at work.',
                  'points': ['what the responsibility was', 'when you had it', 'what you had to do'],
                  'explain': 'and explain how you felt about having this responsibility.',
                  'rounding': 'Do you still have this responsibility now?'},
        'part3': {'topic': 'responsibility', 'questions': [
            'At what age should young people be allowed to make their own decisions?',
            'Why do some people avoid taking responsibility at work?',
            'Should parents be held responsible for the behaviour of their children?',
            'What responsibilities do companies have towards their local community?',
            'Are people today more or less responsible than their grandparents were?']},
    }),
    ('sm-15', 'Sport and teams', {
        'part1': [
            {'topic': 'your street', 'questions': ['What is the street where you live like?', 'Do you know the people who live on your street?', 'Is your street busy or quiet?']},
            {'topic': 'photographs', 'questions': ['Do you take a lot of photographs?', 'What do you usually take photographs of?', 'Do you prefer to keep photographs on your phone or to print them?']},
        ],
        'part2': {'topic': 'a sports team', 'cue': 'Describe a sports team that you support or know about.',
                  'points': ['what the team is', 'how you learned about it', 'how well the team has done recently'],
                  'explain': 'and explain why you are interested in this team.',
                  'rounding': 'Have you ever seen this team play live?'},
        'part3': {'topic': 'team sports and competition', 'questions': [
            'What can children learn from playing team sports?',
            'Why do people feel so strongly about the teams they support?',
            'Should schools give more time to sport?',
            'Is winning the most important thing in sport?',
            'Why do some sports get more attention on television than others?']},
    }),
    ('sm-16', 'Weather and seasons', {
        'part1': [
            {'topic': 'your flat or house', 'questions': ['How many rooms are there in your home?', 'Which room do you spend most time in?', 'Is your home easy to keep warm in winter?']},
            {'topic': 'going out in the evening', 'questions': ['How often do you go out in the evening?', 'Where do people in your town go in the evening?', 'Do you prefer going out or staying in?']},
        ],
        'part2': {'topic': 'a day of perfect weather', 'cue': 'Describe a day when the weather was perfect for what you were doing.',
                  'points': ['when it was', 'what the weather was like', 'what you were doing'],
                  'explain': 'and explain why the weather made the day special.',
                  'rounding': 'Is weather like this common where you live?'},
        'part3': {'topic': 'weather and seasons', 'questions': [
            'Why do many people prefer one season to another?',
            'How do the seasons affect what people eat and wear?',
            'Why do some people move to countries with a warmer climate?',
            'Should schools close when the weather is very hot or very cold?',
            'How might the seasons in your country be different in fifty years?']},
    }),
    ('sm-17', 'Homes and living space', {
        'part1': [
            {'topic': 'your city', 'questions': ['Which city do you live in?', 'What do visitors usually want to see there?', 'What would you change about your city if you could?']},
            {'topic': 'shoes', 'questions': ['How many pairs of shoes do you own?', 'Do you prefer comfortable shoes or fashionable ones?', 'Where do you usually buy your shoes?']},
        ],
        'part2': {'topic': 'a room you like in another home', 'cue': "Describe a room in someone else's home that you like.",
                  'points': ['whose home it is', 'what the room looks like', 'what people do in it'],
                  'explain': 'and explain why you like this room.',
                  'rounding': 'Would you like to have a room like this in your own home?'},
        'part3': {'topic': 'homes and living space', 'questions': [
            'What makes a home comfortable?',
            'Why do people in some countries live in smaller homes than people elsewhere?',
            'Is it better to rent a home or to buy one?',
            'How have homes in your country changed over the last thirty years?',
            "Should young adults leave their parents' home as soon as they can afford to?"]},
    }),
    ('sm-18', 'Habits and change', {
        'part1': [
            {'topic': 'what you do', 'questions': ['Are you working or studying at the moment?', 'What part of it do you find most enjoyable?', 'Where would you like to be working in five years?']},
            {'topic': 'breakfast', 'questions': ['What did you have for breakfast today?', 'Do you eat breakfast at the same time every day?', 'Is breakfast at the weekend different from breakfast on weekdays?']},
        ],
        'part2': {'topic': 'a habit you would like to change', 'cue': 'Describe a habit you have that you would like to change.',
                  'points': ['what the habit is', 'when you started it', 'why you have not been able to change it yet'],
                  'explain': 'and explain why you would like to change it.',
                  'rounding': 'Do people close to you have the same habit?'},
        'part3': {'topic': 'habits and routines', 'questions': [
            'Why is it so hard for people to break bad habits?',
            'Do habits formed in childhood last for life?',
            'Can technology help people to build better habits?',
            'Are daily routines good for people, or do they make life dull?',
            'Is it easier to start a good habit or to stop a bad one?']},
    }),
    ('sm-19', 'Uniforms and dress', {
        'part1': [
            {'topic': 'your family', 'questions': ['How many people are there in your family?', 'What do you usually do together as a family?', 'Who are you most similar to in your family?']},
            {'topic': 'colours', 'questions': ['What colour do you wear most often?', 'Are there any colours you would never wear?', 'Do colours have special meanings in your country?']},
        ],
        'part2': {'topic': 'a uniform you have worn', 'cue': 'Describe a uniform you have had to wear, for example for school or a job.',
                  'points': ['what the uniform looked like', 'when you wore it', 'how you felt wearing it'],
                  'explain': 'and explain whether you think wearing it was a good idea.',
                  'rounding': 'Do you still have this uniform?'},
        'part3': {'topic': 'uniforms and dress codes', 'questions': [
            'Why do many schools make students wear a uniform?',
            'What do the clothes people wear to work say about them?',
            'Should employees be allowed to wear whatever they like to work?',
            'Why do some jobs require a uniform?',
            'Are people judged too much on their appearance?']},
    }),
    ('sm-20', 'Planning and organising', {
        'part1': [
            {'topic': 'your favourite place in your town', 'questions': ['Where is your favourite place in your town?', 'How often do you go there?', 'Do you usually go there alone or with other people?']},
            {'topic': 'writing lists', 'questions': ['Do you ever write shopping lists or lists of things to do?', 'Do you write lists on paper or on your phone?', 'Do you think making lists helps people to organise their lives?']},
        ],
        'part2': {'topic': 'an event you helped to organise', 'cue': 'Describe an event that you helped to organise.',
                  'points': ['what the event was', 'who you organised it with', 'what you did to prepare for it'],
                  'explain': 'and explain whether the event was a success.',
                  'rounding': 'Would you like to organise another event like this?'},
        'part3': {'topic': 'planning and organising', 'questions': [
            'Why are some people better at organising things than others?',
            'Is it better to plan every detail of a holiday or to leave some things to chance?',
            'What skills does a person need to organise a large event?',
            'Should schools teach young people how to plan their time?',
            'Do people plan their lives further ahead than they used to?']},
    }),
]
