"""Passage 3: the passenger lift."""

from scripts.ielts_b4_reading_a._build import opts, paras

PASSAGE = {
    "id_slug": "passengerlift",
    "title": "Going Up: The Passenger Lift",
    "subtitle": "How a safety brake changed the shape of cities",
    "topic": "History of technology and architecture",
    "paragraphs": paras(
        "For most of history, the height of ordinary buildings was limited less by the strength of materials "
        "than by the willingness of people to climb stairs. Towers and church spires had risen to great heights "
        "for centuries, but houses, offices and hotels rarely exceeded five or six storeys. In the cities of the "
        "early nineteenth century, the higher floors of a building were the least desirable: rents fell with "
        "every flight of stairs, and the attic rooms at the top were typically let to the poorest tenants or "
        "occupied by servants. Hoisting machines had existed since ancient times, and by the 1800s steam-powered "
        "hoists were common in factories and warehouses, where they lifted goods between floors. Few people were "
        "prepared to ride in them, however, since if the rope broke, the platform and anyone on it would fall to "
        "the bottom of the shaft.",

        "The breakthrough is usually credited to Elisha Otis, a mechanic from Vermont who was working at a "
        "bedstead factory near New York in the early 1850s. Otis designed a hoist with a strong spring attached "
        "to the top of the platform and a toothed metal strip, or ratchet, running up each side of the shaft. As "
        "long as the rope was under tension, the spring was held back; if the rope failed, the spring was "
        "released and pushed catches into the teeth of the ratchets, stopping the platform almost at once. In "
        "1854, at an exhibition in New York, Otis demonstrated the device in dramatic fashion. He stood on a "
        "raised platform in front of a crowd and ordered an assistant to cut the only rope holding it up. The "
        "platform dropped a few centimetres and stopped.",

        "Otis did not invent the lift, and his device was a brake rather than a new way of moving people, but it "
        "changed public attitudes. The first passenger lift fitted with his safety brake was installed in 1857 in "
        "a five-storey department store in New York, and it moved at about 12 metres per minute, slowly enough "
        "for customers to feel at ease. Steam and then hydraulic power were used to drive early lifts, and in "
        "London a number of hotels and blocks of flats installed hydraulic lifts connected to a public network "
        "of high-pressure water pipes laid beneath the streets from the 1880s. The electric lift, first shown by "
        "the German engineer Werner von Siemens in 1880, eventually replaced these systems, because it needed no "
        "bulky machinery at the base of the building and could serve much taller structures.",

        "The effect on the value of floor space was striking. Once people could reach the top of a building "
        "without effort, the upper floors, with their better light, cleaner air and distance from the noise of "
        "the street, became the most desirable rather than the least. The penthouse flat and the top-floor "
        "restaurant are both products of this reversal. At the same time, the lift made the tall office building "
        "commercially possible. Steel frames allowed architects in Chicago and New York to design buildings of "
        "ten, twenty and eventually over a hundred storeys, but these would have been of little use if the "
        "people working in them had to walk up. It is no exaggeration to say that the skyscraper owes as much to "
        "the lift as it does to the steel frame, even though the frame usually receives far more attention in "
        "histories of architecture.",

        "As buildings grew taller, however, the lift created problems of its own. Each lift needs a shaft, and in "
        "a very tall tower the shafts can take up a large share of the lower storeys, reducing the floor area "
        "that can be let. Engineers have developed several responses. In some towers, express lifts carry "
        "passengers to 'sky lobbies' on intermediate floors, where they change to local lifts serving a smaller "
        "group of floors, so that several lifts can share the same shaft at different heights. More recently, "
        "some manufacturers have introduced systems in which two independent cars travel in a single shaft, one "
        "above the other. The height a single lift can travel has also been limited by the weight of its steel "
        "cables, which in very long shafts become so heavy that they are largely supporting themselves; new "
        "cables made from carbon fibre are designed to reduce this weight considerably.",

        "A less visible change has taken place in the way lifts are controlled. In many modern office buildings, "
        "passengers no longer press a button to call a lift and then choose their floor inside it. Instead they "
        "enter their destination on a panel in the lobby, and a computer assigns them to a particular car, "
        "grouping people who are travelling to the same or nearby floors. These 'destination control' systems "
        "are claimed to cut waiting and travelling times considerably at busy periods, although some users find "
        "them confusing at first, particularly as there are no floor buttons inside the car. Whether they are "
        "worth the expense in smaller buildings is doubtful. In a hotel or a block of flats with only a few "
        "lifts and light traffic, I suspect the traditional system works perfectly well, and passengers may "
        "value the reassurance of a button they can press.",
    ),
    "mcq_single": [
        {
            "prompt": "What does the writer say about the top floors of city buildings in the early nineteenth century?",
            "options": opts(
                "They were usually kept for the owners' families.",
                "They were used mainly for storing goods.",
                "They were the first to be served by hoists.",
                "They were the cheapest parts of the building to rent.",
            ),
            "key": "D",
            "explanation": "Paragraph A says rents fell with every flight of stairs, and the attic rooms were let to the poorest tenants.",
        },
        {
            "prompt": "What was the purpose of the spring in Otis's design?",
            "options": opts(
                "to force catches into the ratchets if the rope failed",
                "to keep the rope under constant tension",
                "to soften the landing at the bottom of the shaft",
                "to help raise the platform more quickly",
            ),
            "key": "A",
            "explanation": "Paragraph B says that if the rope failed, the spring was released and pushed catches into the teeth of the ratchets.",
        },
        {
            "prompt": "What does the writer say about Otis's achievement?",
            "options": opts(
                "He built the first lift ever to carry people.",
                "His brake made the public willing to use lifts.",
                "He was the first to drive a lift by electricity.",
                "He designed London's network of water pipes.",
            ),
            "key": "B",
            "explanation": "Paragraph C says Otis did not invent the lift and his device was a brake, but it changed public attitudes.",
        },
        {
            "prompt": "Why did electric lifts eventually replace hydraulic ones?",
            "options": opts(
                "Electricity was cheaper than high-pressure water.",
                "Electric lifts were quieter for hotel guests.",
                "They needed no bulky machinery and could serve taller buildings.",
                "The public water network was closed down.",
            ),
            "key": "C",
            "explanation": "Paragraph C says the electric lift needed no bulky machinery at the base of the building and could serve much taller structures.",
        },
        {
            "prompt": "What is the purpose of a 'sky lobby'?",
            "options": opts(
                "to give visitors a view over the city",
                "to house the machinery that drives the lifts",
                "to provide a waiting area during busy periods",
                "to allow several lifts to share one shaft at different heights",
            ),
            "key": "D",
            "explanation": "Paragraph E says passengers change to local lifts at sky lobbies, so that several lifts can share the same shaft at different heights.",
        },
    ],
    "mcq_multiple": [
        {
            "prompt": "Which TWO facts about buildings before the passenger lift are given?",
            "options": opts(
                "Church spires could not be built very high.",
                "Houses rarely had more than five or six storeys.",
                "Stairs were usually built outside the building.",
                "Servants refused to live in the attic rooms.",
                "Steam-powered hoists were used to move goods.",
            ),
            "keys": ["B", "E"],
            "explanation": "Paragraph A says houses, offices and hotels rarely exceeded five or six storeys (B) and that steam-powered hoists lifted goods in factories and warehouses (E).",
        },
        {
            "prompt": "Which TWO statements about the first passenger lift with Otis's brake are correct?",
            "options": opts(
                "It was installed in 1857.",
                "It was powered by electricity.",
                "It was in a department store.",
                "It served a building of twelve storeys.",
                "It travelled at 12 metres per second.",
            ),
            "keys": ["A", "C"],
            "explanation": "Paragraph C says it was installed in 1857 (A) in a five-storey department store (C), and moved at about 12 metres per minute, not per second.",
        },
        {
            "prompt": "Which TWO advantages of the upper floors of a building are mentioned?",
            "options": opts(
                "larger rooms",
                "better light",
                "lower heating costs",
                "distance from street noise",
                "direct access to the roof",
            ),
            "keys": ["B", "D"],
            "explanation": "Paragraph D mentions the better light (B), cleaner air and distance from the noise of the street (D).",
        },
        {
            "prompt": "Which TWO recent developments in lift design are described?",
            "options": opts(
                "lifts running up the outside of buildings",
                "two separate cars travelling in one shaft",
                "stairs replacing lifts on the lowest floors",
                "cables made from carbon fibre",
                "doors that open and close more quickly",
            ),
            "keys": ["B", "D"],
            "explanation": "Paragraph E describes systems with two independent cars in a single shaft (B) and new cables made from carbon fibre (D).",
        },
        {
            "prompt": "Which TWO features of destination control are mentioned?",
            "options": opts(
                "Passengers choose their floor on a panel in the lobby.",
                "Lifts travel at a higher speed than before.",
                "It costs less to install than a traditional system.",
                "It is found mainly in hotels.",
                "There are no floor buttons inside the car.",
            ),
            "keys": ["A", "E"],
            "explanation": "Paragraph F says passengers enter their destination on a panel in the lobby (A) and that there are no floor buttons inside the car (E).",
        },
    ],
    "tfng": [
        {
            "statement": "Hoisting machines were unknown before the nineteenth century.",
            "key": "FALSE",
            "explanation": "Paragraph A says hoisting machines had existed since ancient times.",
        },
        {
            "statement": "Otis received a prize for his demonstration at the New York exhibition.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph B describes the demonstration but does not mention any prize.",
        },
        {
            "statement": "The first passenger lift with a safety brake was designed to move slowly enough to reassure its users.",
            "key": "TRUE",
            "explanation": "Paragraph C says it moved at about 12 metres per minute, slowly enough for customers to feel at ease.",
        },
        {
            "statement": "Some London lifts were powered by water supplied through pipes under the streets.",
            "key": "TRUE",
            "explanation": "Paragraph C says hydraulic lifts in London were connected to a public network of high-pressure water pipes laid beneath the streets.",
        },
        {
            "statement": "Carbon fibre cables are heavier than steel ones.",
            "key": "FALSE",
            "explanation": "Paragraph E says carbon fibre cables are designed to reduce the weight of steel cables considerably.",
        },
    ],
    "ynng": [
        {
            "statement": "Histories of architecture have given the lift less credit than it deserves.",
            "key": "YES",
            "explanation": "Paragraph D says the skyscraper owes as much to the lift as to the steel frame, even though the frame receives far more attention.",
        },
        {
            "statement": "Very tall office buildings would have been useful even without lifts.",
            "key": "NO",
            "explanation": "Paragraph D says such buildings would have been of little use if the people working in them had to walk up.",
        },
        {
            "statement": "Sky lobbies are the best way of dealing with the space taken up by lift shafts.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph E describes sky lobbies as one of several responses but does not say which is best.",
        },
        {
            "statement": "Destination control is probably worth installing in small buildings.",
            "key": "NO",
            "explanation": "Paragraph F says whether it is worth the expense in smaller buildings is doubtful.",
        },
        {
            "statement": "Destination control will soon be standard in new blocks of flats.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph F gives the writer's doubts about smaller buildings but makes no prediction about what will become standard.",
        },
    ],
    "matching": [
        {
            "prompt": "a reason why people were unwilling to travel on early hoists",
            "key": "A",
            "evidence": "if the rope broke, the platform and anyone on it would fall to the bottom of the shaft",
            "explanation": "Paragraph A explains the danger of riding on early hoists.",
        },
        {
            "prompt": "an account of a public test of a safety device",
            "key": "B",
            "evidence": "ordered an assistant to cut the only rope holding it up",
            "explanation": "Paragraph B describes Otis's demonstration in 1854.",
        },
        {
            "prompt": "the name of an engineer who presented an electric lift",
            "key": "C",
            "evidence": "first shown by the German engineer Werner von Siemens in 1880",
            "explanation": "Paragraph C names Werner von Siemens.",
        },
        {
            "prompt": "a reversal in the way that floors of a building were valued",
            "key": "D",
            "evidence": "became the most desirable rather than the least",
            "explanation": "Paragraph D says upper floors became the most desirable rather than the least.",
        },
        {
            "prompt": "a method of putting passengers together according to where they are going",
            "key": "F",
            "evidence": "grouping people who are travelling to the same or nearby floors",
            "explanation": "Paragraph F describes how destination control groups passengers.",
        },
    ],
    "completion": [
        {
            "prompt": "Before lifts, the attic rooms of a building were often used by ________.",
            "answer": "servants",
            "explanation": "Paragraph A says the attic rooms were let to the poorest tenants or occupied by servants.",
        },
        {
            "prompt": "Otis's hoist had a toothed metal strip called a ________ on each side of the shaft.",
            "answer": "ratchet",
            "explanation": "Paragraph B describes a toothed metal strip, or ratchet, running up each side of the shaft.",
        },
        {
            "prompt": "In London, some early lifts were driven by ________ power.",
            "answer": "hydraulic",
            "explanation": "Paragraph C says hotels and blocks of flats in London installed hydraulic lifts.",
        },
        {
            "prompt": "Tall office buildings became possible thanks to the lift and the steel ________.",
            "answer": "frame",
            "explanation": "Paragraph D says the skyscraper owes as much to the lift as it does to the steel frame.",
        },
        {
            "prompt": "Systems in which passengers are assigned to a car by computer are known as ________ control.",
            "answer": "destination",
            "explanation": "Paragraph F calls these 'destination control' systems.",
        },
    ],
}
