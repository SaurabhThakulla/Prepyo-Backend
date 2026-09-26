"""Passage 3: How Satellite Positioning Works."""

PASSAGE_3 = {
    "id": "rp-pr102-003",
    "title": "How Satellite Positioning Works",
    "subtitle": "Timing signals, relativity and the correction of errors in systems such as GPS",
    "topic": "Technology and Space Science",
    "paragraphs": [
        {
            "label": "A",
            "text": "A satellite positioning receiver, whether in a car, an aircraft or a mobile phone, does not send anything into space. It works out its location by listening. The American Global Positioning System, the first such system to be fully operational, uses a group of around thirty satellites orbiting about 20,200 kilometres above the Earth, each circling the planet roughly twice a day. Every satellite carries atomic clocks and continuously broadcasts a signal stating the exact time at which the signal was sent and where the satellite was at that moment. The receiver notes when the signal arrives, and because radio waves travel at the speed of light, the delay tells it how far away the satellite is. The timing must be extremely precise: an error of just one millionth of a second corresponds to a distance of about 300 metres."
        },
        {
            "label": "B",
            "text": "A single distance measurement places the receiver somewhere on the surface of an imaginary sphere centred on the satellite. A second measurement narrows the possibilities to the circle where two spheres meet, and a third reduces them to two points, one of which is usually far out in space and can be rejected. In theory, then, three satellites should be enough. In practice a fourth is needed, because the clock inside an ordinary receiver is an inexpensive quartz device that is nowhere near as accurate as the atomic clocks on the satellites. By taking measurements from at least four satellites at once, the receiver can calculate its own clock error as well as its three coordinates. This is why a phone can find its position to within a few metres without containing an expensive clock of its own."
        },
        {
            "label": "C",
            "text": "Designers of the system also had to allow for the effects described by Einstein's theories of relativity. Because the satellites travel at almost four kilometres per second, their clocks run slightly slow compared with clocks on the ground, losing about seven millionths of a second each day. At the same time, the satellites are farther from the Earth's centre, where gravity is weaker, and this makes their clocks run faster by about forty-five millionths of a second a day. The combined result is a gain of roughly thirty-eight millionths of a second daily. The difference sounds tiny, but if it were ignored, calculated positions would drift by around ten kilometres every day. To prevent this, the satellite clocks are deliberately set to tick at a slightly lower rate before launch."
        },
        {
            "label": "D",
            "text": "Even with perfect clocks, several factors reduce accuracy. The most important is the ionosphere, a layer of electrically charged gas in the upper atmosphere that slows radio signals by an amount that varies with the time of day and the activity of the Sun. Water vapour in the lower atmosphere causes smaller delays. Close to the ground, signals may bounce off buildings or hillsides before reaching the receiver, so that they travel further than the direct route; this problem, known as multipath, is especially common in city streets lined with tall buildings. Accuracy also depends on the positions of the satellites in the sky. When the visible satellites are spread widely, their measurements cross at sharp angles and fix the position well, but when they are bunched together, small errors in each distance produce a much larger error in the result."
        },
        {
            "label": "E",
            "text": "Engineers have developed several ways of removing these errors. Because the delay caused by the ionosphere depends on the frequency of the signal, receivers that pick up two different frequencies can compare them and calculate the delay directly. Another approach relies on reference stations placed at points whose positions have been surveyed very precisely. Each station compares its known position with the one it calculates from the satellites and broadcasts the difference as a correction to nearby users. Systems of this kind, such as EGNOS in Europe, are used to guide aircraft during their approach to airports. The most demanding users, including surveyors and operators of farm machinery, use a technique that tracks the waves of the signal itself rather than only the timing information it carries, which can give positions accurate to a centimetre or two."
        },
        {
            "label": "F",
            "text": "For its first years of full operation, the accuracy of GPS available to civilian users was deliberately reduced for security reasons. When this restriction was switched off in May 2000, the accuracy that ordinary receivers could achieve improved roughly tenfold overnight, and a wave of new applications followed. Today GPS is one of several global systems, alongside Russia's GLONASS, the European Union's Galileo and China's BeiDou, and most modern receivers use satellites from more than one of them. This makes a good fix more likely in difficult places such as narrow streets and deep valleys. Positioning is not the only service these systems provide. Because they deliver extremely accurate time to any receiver, they are used to synchronise mobile phone networks, electricity grids and financial transactions, uses that most people are completely unaware of."
        }
    ],
    # 5 RWFIB items (Fill in the Blanks - Dropdown)
    "rwfib": [
        {
            "title": "Satellite Positioning: Listening for Signals",
            "text": "A positioning receiver finds its location without [[b1]] any signal of its own. Each satellite broadcasts the time at which its signal left, and the receiver measures how long the signal took to arrive. Since radio waves travel at the speed of light, even a very small timing error [[b2]] a large error in distance. A mistake of one millionth of a second, [[b3]] instance, puts the position out by about 300 metres.",
            "blanks": [
                {"id": "b1", "options": ["transmitting", "transmitted", "transmits", "transmission"], "correctAnswer": "transmitting"},
                {"id": "b2", "options": ["leads to", "results from", "refers to", "depends on"], "correctAnswer": "leads to"},
                {"id": "b3", "options": ["for", "in", "by", "as"], "correctAnswer": "for"}
            ],
            "explanation": "b1: 'without' is a preposition, so it is followed by an -ing form: 'without transmitting'. b2: the timing error causes the distance error, so 'leads to' is right; 'results from' reverses cause and effect, and 'refers to' and 'depends on' give the wrong meaning. b3: the fixed phrase is 'for instance'."
        },
        {
            "title": "Satellite Positioning: The Fourth Satellite",
            "text": "Measurements from three satellites should in theory be [[b1]] to fix a position. However, the clock in an ordinary receiver is far less accurate than the atomic clocks the satellites carry. A fourth measurement allows the receiver to work out its own clock error, [[b2]] is why a phone can find its position without an expensive clock. In other words, the extra satellite [[b3]] for the weakness of the receiver.",
            "blanks": [
                {"id": "b1", "options": ["sufficient", "sufficiently", "sufficiency", "suffice"], "correctAnswer": "sufficient"},
                {"id": "b2", "options": ["which", "what", "that", "who"], "correctAnswer": "which"},
                {"id": "b3", "options": ["compensates", "complains", "competes", "complies"], "correctAnswer": "compensates"}
            ],
            "explanation": "b1: after 'be' an adjective is needed, so 'sufficient'. b2: a non-defining relative clause referring back to the whole previous idea begins with 'which'; 'that' cannot follow a comma in this way, 'what' does not refer back, and 'who' is for people. b3: 'compensates for' means 'makes up for', which matches the meaning; 'complains', 'competes' and 'complies' do not."
        },
        {
            "title": "Satellite Positioning: Clocks and Relativity",
            "text": "Because the satellites move fast, their clocks run slightly slow [[b1]] clocks on the ground. On the other hand, the weaker gravity at their height makes the clocks run faster. Overall, the satellite clocks gain about thirty-eight millionths of a second a day. If this were not corrected, calculated positions [[b2]] by around ten kilometres daily, so the clocks are [[b3]] to tick slightly more slowly before launch.",
            "blanks": [
                {"id": "b1", "options": ["compared with", "apart from", "instead of", "owing to"], "correctAnswer": "compared with"},
                {"id": "b2", "options": ["would drift", "drifted", "will drift", "have drifted"], "correctAnswer": "would drift"},
                {"id": "b3", "options": ["adjusted", "adjoined", "admitted", "adopted"], "correctAnswer": "adjusted"}
            ],
            "explanation": "b1: the clocks are slow 'compared with' clocks on the ground; the other phrases do not express a comparison. b2: 'If this were not corrected' is a second conditional, so the result clause needs 'would drift'. b3: the clocks are 'adjusted' (changed slightly) to tick more slowly; 'adjoined', 'admitted' and 'adopted' do not make sense."
        },
        {
            "title": "Satellite Positioning: Sources of Error",
            "text": "The ionosphere, a charged layer of the upper atmosphere, slows radio signals by an amount that [[b1]] with the time of day. Near the ground, signals may bounce off buildings before reaching the receiver, [[b2]] that they travel further than the direct route. When the visible satellites are bunched together in the sky, small errors in each distance produce a much [[b3]] error in the final position.",
            "blanks": [
                {"id": "b1", "options": ["varies", "various", "variety", "variably"], "correctAnswer": "varies"},
                {"id": "b2", "options": ["so", "such", "even", "unless"], "correctAnswer": "so"},
                {"id": "b3", "options": ["larger", "large", "largest", "largely"], "correctAnswer": "larger"}
            ],
            "explanation": "b1: the relative clause 'that ___ with the time of day' needs a verb, so 'varies'. b2: 'so that' introduces the result of the bouncing; 'such that' needs a noun before it, and 'even' and 'unless' do not fit. b3: 'much' before an adjective is used with a comparative, so 'much larger'."
        },
        {
            "title": "Satellite Positioning: Correcting the Signal",
            "text": "Receivers that pick up two frequencies can calculate the delay caused by the ionosphere [[b1]], because that delay depends on frequency. Reference stations at precisely surveyed points compare their known positions with those [[b2]] from the satellites and broadcast corrections. For surveying and farming, a technique that tracks the waves of the signal can give positions [[b3]] to within a centimetre or two.",
            "blanks": [
                {"id": "b1", "options": ["directly", "direct", "direction", "directed"], "correctAnswer": "directly"},
                {"id": "b2", "options": ["calculated", "calculating", "calculation", "calculates"], "correctAnswer": "calculated"},
                {"id": "b3", "options": ["accurate", "accuracy", "accurately", "accurateness"], "correctAnswer": "accurate"}
            ],
            "explanation": "b1: an adverb is needed to describe how the delay is calculated, so 'directly'. b2: 'those ___ from the satellites' is a reduced passive ('those that are calculated'), so the past participle 'calculated' fits. b3: 'give positions accurate to within...' needs an adjective describing 'positions', so 'accurate'."
        }
    ],
    # 5 RFIB items (Reading: Fill in the Blanks - Drag and Drop)
    "rfib": [
        {
            "title": "Satellite Positioning: Spheres of Distance",
            "text": "One distance measurement places a receiver somewhere on an imaginary sphere around the satellite. A second measurement narrows the [[b1]] to a circle, and a third to just two points. One of these points usually lies far out in space, so it can be [[b2]]. The remaining point [[b3]] the position of the receiver.",
            "blanks": [
                {"id": "b1", "correctAnswer": "possibilities"},
                {"id": "b2", "correctAnswer": "rejected"},
                {"id": "b3", "correctAnswer": "gives"}
            ],
            "pool": ["possibilities", "rejected", "gives", "invited", "celebrates", "furniture"],
            "explanation": "A second measurement narrows the 'possibilities' to a circle. A point far out in space cannot be the receiver's position, so it is 'rejected'. The other point 'gives' the position. 'Invited', 'celebrates' and 'furniture' make no sense in any gap."
        },
        {
            "title": "Satellite Positioning: Einstein in Orbit",
            "text": "The speed of the satellites makes their clocks lose about seven millionths of a second a day, while the [[b1]] gravity at their height makes them gain about forty-five. The overall effect is a small daily gain. Although the difference seems [[b2]], it would cause positions to drift by kilometres, so engineers set the clocks to run slightly [[b3]] before launch.",
            "blanks": [
                {"id": "b1", "correctAnswer": "weaker"},
                {"id": "b2", "correctAnswer": "trivial"},
                {"id": "b3", "correctAnswer": "slower"}
            ],
            "pool": ["weaker", "trivial", "slower", "enormous", "wetter", "happily"],
            "explanation": "Gravity is 'weaker' further from the Earth's centre, and 'wetter' makes no sense. 'Although' signals a contrast: the difference seems 'trivial' but has large effects; 'enormous' would remove the contrast. To cancel a gain, the clocks must run 'slower'. 'Happily' fits no gap."
        },
        {
            "title": "Satellite Positioning: Signals in the City",
            "text": "In city streets lined with tall buildings, satellite signals often bounce off walls before they [[b1]] a receiver. Because the reflected signal travels further than the direct one, the receiver [[b2]] the distance to the satellite. This problem is known as multipath, and it is one reason why positions are often less [[b3]] in town centres than in open country.",
            "blanks": [
                {"id": "b1", "correctAnswer": "reach"},
                {"id": "b2", "correctAnswer": "overestimates"},
                {"id": "b3", "correctAnswer": "precise"}
            ],
            "pool": ["reach", "overestimates", "precise", "underestimates", "decorate", "crowded"],
            "explanation": "Signals bounce before they 'reach' a receiver. A longer path makes the receiver think the satellite is further away, so it 'overestimates' the distance; 'underestimates' is the opposite. Positions are less 'precise' in town centres. 'Decorate' fits no gap, and 'less crowded in town centres' contradicts the idea of tall, busy streets and does not follow from multipath."
        },
        {
            "title": "Satellite Positioning: Help for Aircraft",
            "text": "A reference station sits at a point whose position is already known with great [[b1]]. It compares this position with the one it works out from the satellites, and the difference is [[b2]] to nearby receivers as a correction. Systems like this are used to [[b3]] aircraft as they approach an airport.",
            "blanks": [
                {"id": "b1", "correctAnswer": "accuracy"},
                {"id": "b2", "correctAnswer": "broadcast"},
                {"id": "b3", "correctAnswer": "guide"}
            ],
            "pool": ["accuracy", "broadcast", "guide", "confusion", "swallowed", "noisy"],
            "explanation": "The position is known 'with great accuracy'; 'with great confusion' contradicts the purpose of a reference station. The correction is 'broadcast' to receivers. The corrections 'guide' aircraft. 'Swallowed' and 'noisy' fit no gap."
        },
        {
            "title": "Satellite Positioning: More Than a Map",
            "text": "When the deliberate reduction in civilian accuracy was switched off in 2000, ordinary receivers became about ten times more accurate almost [[b1]]. Today several global systems operate side by side, and receivers that combine them are more likely to [[b2]] a good fix in narrow streets. These systems also supply very accurate time, which is used to [[b3]] mobile phone networks and electricity grids.",
            "blanks": [
                {"id": "b1", "correctAnswer": "immediately"},
                {"id": "b2", "correctAnswer": "obtain"},
                {"id": "b3", "correctAnswer": "synchronise"}
            ],
            "pool": ["immediately", "obtain", "synchronise", "reluctantly", "abandon", "painting"],
            "explanation": "Accuracy improved 'almost immediately' once the restriction was removed. Combining systems makes receivers more likely to 'obtain' (get) a good fix; 'abandon' contradicts this. Accurate time is used to 'synchronise' networks. 'Reluctantly' and 'painting' fit no gap."
        }
    ],
    # 5 RMCSA items (Multiple Choice, Single Answer)
    "mcq_single": [
        {
            "title": "Satellite Positioning: How Distance Is Found",
            "prompt": "According to paragraph A, how does a receiver work out its distance from a satellite?",
            "options": [
                {"id": "A", "text": "By measuring how long the satellite's signal takes to arrive"},
                {"id": "B", "text": "By sending a signal to the satellite and timing the reply"},
                {"id": "C", "text": "By measuring the strength of the satellite's signal"},
                {"id": "D", "text": "By comparing the satellite's position with a stored map"}
            ],
            "key": "A",
            "explanation": "Paragraph A explains that the receiver notes when the signal arrives and 'the delay tells it how far away the satellite is' (A). The receiver 'does not send anything into space', and signal strength and stored maps are not mentioned."
        },
        {
            "title": "Satellite Positioning: Why Four Are Needed",
            "prompt": "Why does a receiver need signals from at least four satellites rather than three?",
            "options": [
                {"id": "A", "text": "Its own clock is not accurate enough, so the clock error must also be calculated."},
                {"id": "B", "text": "One of the satellites is always hidden behind the Earth."},
                {"id": "C", "text": "Each satellite can only give a rough estimate of its own position."},
                {"id": "D", "text": "The atomic clocks on the satellites often disagree with one another."}
            ],
            "key": "A",
            "explanation": "Paragraph B states that a receiver's quartz clock is far less accurate than the satellites' atomic clocks, so a fourth measurement lets it 'calculate its own clock error as well as its three coordinates' (A). The other reasons are not given in the passage."
        },
        {
            "title": "Satellite Positioning: The Net Effect of Relativity",
            "prompt": "According to paragraph C, what is the overall effect of relativity on the satellite clocks?",
            "options": [
                {"id": "A", "text": "They gain about thirty-eight millionths of a second a day."},
                {"id": "B", "text": "They lose about seven millionths of a second a day."},
                {"id": "C", "text": "They gain about forty-five millionths of a second a day."},
                {"id": "D", "text": "The two effects cancel each other out exactly."}
            ],
            "key": "A",
            "explanation": "The loss of seven and the gain of forty-five millionths of a second give 'a gain of roughly thirty-eight millionths of a second daily' (A). The figures of seven and forty-five describe only one effect each, and the effects do not cancel out."
        },
        {
            "title": "Satellite Positioning: The Main Source of Error",
            "prompt": "Which factor does paragraph D describe as the most important cause of reduced accuracy?",
            "options": [
                {"id": "A", "text": "The ionosphere slowing radio signals"},
                {"id": "B", "text": "Water vapour in the lower atmosphere"},
                {"id": "C", "text": "Signals reflecting off buildings and hillsides"},
                {"id": "D", "text": "Satellites being bunched together in the sky"}
            ],
            "key": "A",
            "explanation": "Paragraph D begins: 'The most important is the ionosphere' (A). Water vapour is said to cause 'smaller delays', and multipath and satellite positions are presented as further factors, not the most important one."
        },
        {
            "title": "Satellite Positioning: A Hidden Use",
            "prompt": "What does paragraph F suggest about the timing service provided by satellite systems?",
            "options": [
                {"id": "A", "text": "It is widely used but most people do not know about it."},
                {"id": "B", "text": "It is less accurate than the positioning service."},
                {"id": "C", "text": "It is available only to users of the Galileo system."},
                {"id": "D", "text": "It was switched off in May 2000 for security reasons."}
            ],
            "key": "A",
            "explanation": "Paragraph F says accurate time is used to synchronise phone networks, grids and financial transactions, 'uses that most people are completely unaware of' (A). The time is described as 'extremely accurate', all the systems provide it, and what was switched off in 2000 was a reduction in accuracy."
        }
    ],
    # 5 RMCMA items (Multiple Choice, Multiple Answers)
    "mcq_multiple": [
        {
            "title": "Satellite Positioning: What Each Satellite Sends",
            "prompt": "According to paragraph A, which TWO pieces of information does each satellite broadcast?",
            "options": [
                {"id": "A", "text": "The time at which the signal was sent"},
                {"id": "B", "text": "The distance to the nearest receiver"},
                {"id": "C", "text": "The position of the satellite when the signal was sent"},
                {"id": "D", "text": "The current state of the ionosphere"},
                {"id": "E", "text": "The positions of the other satellites in the system"}
            ],
            "keys": ["A", "C"],
            "explanation": "Each satellite broadcasts 'the exact time at which the signal was sent' (A) and 'where the satellite was at that moment' (C). The satellite cannot know the receiver's distance, and the other items are not mentioned."
        },
        {
            "title": "Satellite Positioning: Two Relativity Effects",
            "prompt": "According to paragraph C, which TWO factors affect the rate at which the satellite clocks run?",
            "options": [
                {"id": "A", "text": "The temperature of space around the satellites"},
                {"id": "B", "text": "The speed at which the satellites travel"},
                {"id": "C", "text": "The age of the atomic clocks"},
                {"id": "D", "text": "The weaker gravity at the height of the satellites"},
                {"id": "E", "text": "The number of receivers using the signals"}
            ],
            "keys": ["B", "D"],
            "explanation": "The satellites' speed makes their clocks run slow (B), and the weaker gravity far from the Earth's centre makes them run fast (D). Temperature, clock age and the number of users are not mentioned."
        },
        {
            "title": "Satellite Positioning: Trouble in Town",
            "prompt": "Based on paragraph D, which TWO conditions would be likely to make a position fix less accurate?",
            "options": [
                {"id": "A", "text": "Standing in open country with a clear view of the sky"},
                {"id": "B", "text": "Using a receiver with a quartz clock"},
                {"id": "C", "text": "Receiving signals that have bounced off tall buildings"},
                {"id": "D", "text": "Having visible satellites spread widely across the sky"},
                {"id": "E", "text": "Having all the visible satellites bunched close together"}
            ],
            "keys": ["C", "E"],
            "explanation": "Paragraph D explains that reflected signals (multipath) travel further than the direct route (C), and that bunched satellites turn small errors into a much larger one (E). Widely spread satellites 'fix the position well', open country avoids multipath, and the quartz clock error is removed using a fourth satellite, as paragraph B explains."
        },
        {
            "title": "Satellite Positioning: Removing Errors",
            "prompt": "Which TWO methods of improving accuracy are described in paragraph E?",
            "options": [
                {"id": "A", "text": "Comparing signals on two different frequencies"},
                {"id": "B", "text": "Launching satellites into lower orbits"},
                {"id": "C", "text": "Placing atomic clocks inside ordinary receivers"},
                {"id": "D", "text": "Broadcasting corrections from stations at precisely known points"},
                {"id": "E", "text": "Using signals only at night when the ionosphere is inactive"}
            ],
            "keys": ["A", "D"],
            "explanation": "Paragraph E describes dual-frequency receivers that 'compare them and calculate the delay directly' (A) and reference stations that broadcast 'the difference as a correction' (D). Lower orbits, receiver atomic clocks and night-only use are not mentioned."
        },
        {
            "title": "Satellite Positioning: Several Systems at Once",
            "prompt": "According to paragraph F, which TWO statements are true?",
            "options": [
                {"id": "A", "text": "GPS is now the only global positioning system in operation."},
                {"id": "B", "text": "Civilian accuracy improved sharply after a restriction was removed in 2000."},
                {"id": "C", "text": "Receivers that use more than one system are more likely to get a good fix in difficult places."},
                {"id": "D", "text": "Galileo was built by Russia to replace GLONASS."},
                {"id": "E", "text": "Most modern receivers can use only one system at a time."}
            ],
            "keys": ["B", "C"],
            "explanation": "Accuracy 'improved roughly tenfold overnight' when the restriction was switched off (B), and using several systems 'makes a good fix more likely in difficult places' (C). GPS is one of several systems, Galileo belongs to the European Union, and most receivers use more than one system."
        }
    ]
}
