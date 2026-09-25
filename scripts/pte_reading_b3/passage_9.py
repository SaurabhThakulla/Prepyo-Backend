"""Passage 9: The Engineering of Alpine Funiculars."""

PASSAGE_9 = {
    "id": "rp-pr89-009",
    "title": "The Engineering of Alpine Funiculars",
    "subtitle": "The cable-driven mechanics and counterbalance systems of mountain railways",
    "topic": "Mechanical and Transportation Engineering",
    "paragraphs": [
        {
            "label": "A",
            "text": "Across the rugged precipices of the European Alps, nineteenth-century transport engineers confronted an intractable physical barrier: the adhesion limit of conventional steel-rail transit. Standard railways operate via wheel-rail friction, where tractive effort is generated entirely by the mechanical grip between smooth steel wheels and smooth steel rails. When gradients exceed six to eight percent (sixty to eighty per mille), this frictional interface fails; locomotives experience catastrophic wheel slip when climbing and lose braking adhesion when descending, creating grave risks of runaway derailments. To conquer sheer alpine cliffs with inclines exceeding sixty to one hundred percent, Victorian civil engineers revived and modernized an ancient mechanical concept: the funicular railway. By abandoning locomotive adhesion in favor of stationary cable traction, funiculars established a safe, dependable technology for surmounting the steepest mountain terrain."
        },
        {
            "label": "B",
            "text": "The core mechanical brilliance of a funicular railway lies in its elegant exploitation of gravitational counterbalance. Rather than propelling a self-powered vehicle uphill against gravity, a funicular links two matched passenger carriages permanently to opposite ends of a continuous steel haulage cable. This cable passes around a motorized drive sheave (pulley wheel) housed inside the upper mountain terminal station. As one carriage descends the mountain, its gravitational weight pulls the ascending carriage upward, acting as a massive counterweight. Consequently, the stationary drive motor does not need to lift the total physical mass of the train; it only supplies the relatively modest energy required to overcome mechanical friction and lift any weight differential between ascending and descending passenger loads."
        },
        {
            "label": "C",
            "text": "The earliest alpine systems operated with even greater mechanical economy by harnessing gravity-driven water ballast. On systems such as Switzerland's Giessbachbahn, inaugurated in 1879, no steam engine or electric motor was required. Instead, before departing the mountain summit, water diverted from an alpine stream was pumped into a massive ballast tank beneath the descending carriage. Once the descending car outweighed the ascending carriage and its passengers, the conductor released the track brakes, allowing gravity to pull the heavier carriage down while hoisting the lighter carriage to the top. Upon reaching the lower valley terminal, the water was dumped into a drainage channel, and the cycle repeated, providing a sustainable, passive transport system powered entirely by alpine hydrology."
        },
        {
            "label": "D",
            "text": "While early funiculars required two complete parallel tracks along their entire route, excavating double rights-of-way through sheer granite cliffs was prohibitively expensive and dangerous. In 1879, Swiss engineer Carl Roman Abt solved this infrastructural bottleneck by inventing the automatic bypass switch that bears his name. The Abt system enables two carriages to operate on a single central track that splits briefly into a passing loop midway along the route. Crucially, the Abt switch possesses no moving track points or frogs. Instead, each carriage is equipped with an asymmetric wheelset: wheels on one side feature double flanges that grip the outer continuous rail, guiding the car into its designated siding, while wheels on the opposing side are completely flat, gliding smoothly across rail gaps without derailing."
        },
        {
            "label": "E",
            "text": "Operating heavy carriages on precipitous mountain inclines demands fail-safe braking systems capable of arresting runaway loads without relying upon cable integrity. The haulage cable itself—composed of high-tensile steel strands wrapped around a synthetic core—is inspected daily for fatigue fractures. For emergency containment, every funicular carriage is outfitted with autonomous track brakes. Operated by spring-loaded hydraulic mechanisms and monitored by centrifugal overspeed governors, these emergency brakes do not act upon the wheels. If cable tension drops abruptly or carriage velocity exceeds safety limits, powerful hardened steel wedge clamps instantly grip the rail head directly, bringing the fully laden vehicle to an emergency halt within a few metres."
        },
        {
            "label": "F",
            "text": "In the contemporary era, funicular engineering has evolved from scenic surface railways into high-capacity subterranean transit arteries. In major ski regions, surface funiculars were frequently disrupted by winter blizzards, rockfalls, and destructive avalanches. Engineers responded by boring inclined tunnels through solid granite, enclosing funiculars entirely inside mountain bedrock (such as the Metro Alpin in Saas-Fee). Simultaneously, the technology has experienced a major urban renaissance. Hilly metropolitan cities, including Lisbon, Valparaíso, Naples, and Pittsburgh, utilize modernized funiculars equipped with regenerative electric drive motors. By feeding electrical energy back into municipal power grids during descent, cable-hauled urban transit provides a whisper-quiet, zero-emission mobility solution that navigates steep topography far more efficiently than diesel buses."
        }
    ],
    # 5 RWFIB items (Fill in the Blanks - Dropdown)
    "rwfib": [
        {
            "title": "Adhesion Railways: Gradient Limitations",
            "text": "Standard trains face severe physical obstacles when navigating mountainous regions. Beyond gradients of eight percent, friction between steel wheels and rails becomes [[b1]] to prevent wheel slipping. Locomotives can easily lose braking control during steep descents, creating serious risks of runaway [[b2]]. To surmount extreme alpine slopes safely, engineers abandoned locomotive adhesion and adopted [[b3]] funicular systems.",
            "blanks": [
                {"id": "b1", "options": ["abundant", "excessive", "insufficient", "superfluous"], "correctAnswer": "insufficient"},
                {"id": "b2", "options": ["accidents", "successes", "triumphes", "triumphs"], "correctAnswer": "accidents"},
                {"id": "b3", "options": ["cable-hauled", "diesel-fueled", "solar-driven", "wind-propelled"], "correctAnswer": "cable-hauled"}
            ],
            "explanation": "b1: 'insufficient' describes friction being too low on slopes ('abundant', 'excessive', and 'superfluous' mean having more than enough). b2: 'accidents' describes runaway hazards ('successes' and 'triumphs' are positive achievements). b3: 'cable-hauled' is the defining mechanical traction category ('wind-propelled', 'solar-driven', and 'diesel-fueled' are inapplicable propulsion types)."
        },
        {
            "title": "Funicular Counterbalance: Gravitational Mechanics",
            "text": "The mechanical design of a funicular railway relies on the principle of counterbalance. Two carriages are permanently linked to opposite ends of a continuous steel cable that passes around a drive [[b1]] at the upper station. As one car descends, its gravitational weight helps [[b2]] the ascending vehicle. Consequently, the stationary drive motor only needs to provide enough energy to overcome mechanical friction and [[b3]] passenger loads.",
            "blanks": [
                {"id": "b1", "options": ["boiler", "chimney", "piston", "sheave"], "correctAnswer": "sheave"},
                {"id": "b2", "options": ["anchor", "descend", "ground", "hoist"], "correctAnswer": "hoist"},
                {"id": "b3", "options": ["equal", "identical", "unbalanced", "uniform"], "correctAnswer": "unbalanced"}
            ],
            "explanation": "b1: 'sheave' is the grooved drive wheel at the summit ('piston', 'boiler', and 'chimney' are steam engine components). b2: 'hoist' means pulling up the ascending car ('descend' means going down; 'anchor' and 'ground' mean securing in place). b3: 'unbalanced' describes passenger weight differences between carriages ('equal', 'uniform', and 'identical' mean perfectly balanced)."
        },
        {
            "title": "Water Ballast: Gravity Propulsion",
            "text": "Early mountain funiculars operated without steam engines or electrical power. At the summit terminal, operators filled an underfloor ballast tank with mountain stream water to [[b1]] the car. Once the descending car outweighed the ascending train, the operator released the brakes, allowing gravity to [[b2]] the system. At the bottom station, the water was dumped into a drainage channel, completing a [[b3]] operational cycle.",
            "blanks": [
                {"id": "b1", "options": ["insulate", "lighten", "streamline", "weight"], "correctAnswer": "weight"},
                {"id": "b2", "options": ["arrest", "brake", "impede", "propel"], "correctAnswer": "propel"},
                {"id": "b3", "options": ["erratic", "intermittent", "sustainable", "wasteful"], "correctAnswer": "sustainable"}
            ],
            "explanation": "b1: 'weight' (as a transitive verb) means adding ballast weight to the car ('lighten' is opposite; 'streamline' and 'insulate' are irrelevant). b2: 'propel' means driving the system via gravity ('impede', 'brake', and 'arrest' mean stopping movement). b3: 'sustainable' matches continuous closed hydraulic cycles ('erratic', 'wasteful', and 'intermittent' mean unreliable or inefficient)."
        },
        {
            "title": "Abt Switch: Siding Geometry",
            "text": "Constructing two full tracks along a cliff was dangerous and costly. Swiss engineer Carl Roman Abt solved this problem by creating a passing loop that operates without moving track [[b1]]. Each carriage features asymmetric wheels with double flanges on the outer side that [[b2]] the vehicle onto its designated track. Meanwhile, wide flat wheels on the opposing side glide smoothly over rail gaps, completely [[b3]] the danger of derailment.",
            "blanks": [
                {"id": "b1", "options": ["buffers", "points", "signals", "sleepers"], "correctAnswer": "points"},
                {"id": "b2", "options": ["brake", "halt", "steer", "tilt"], "correctAnswer": "steer"},
                {"id": "b3", "options": ["courting", "eliminating", "fostering", "provoking"], "correctAnswer": "eliminating"}
            ],
            "explanation": "b1: 'points' (track switches) is the railway engineering term for movable rails ('buffers', 'signals', and 'sleepers' are stationary track elements). b2: 'steer' describes wheel flanges guiding the car ('brake', 'tilt', and 'halt' describe stopping or banking). b3: 'eliminating' describes removing derailment risks ('courting', 'fostering', and 'provoking' mean causing danger)."
        },
        {
            "title": "Track Clamps: Fail-Safe Brakes",
            "text": "Operating on steep inclines requires exceptional braking safety. Modern funicular carriages feature autonomous track brakes that do not depend on the haulage [[b1]]. If cable tension drops or an overspeed governor triggers, spring-loaded hydraulic clamps [[b2]] grip the rail head directly. This fail-safe mechanism brings the carriage to an immediate halt, ensuring passenger [[b3]] on extreme slopes.",
            "blanks": [
                {"id": "b1", "options": ["cable", "chassis", "cushion", "window"], "correctAnswer": "cable"},
                {"id": "b2", "options": ["forcefully", "gently", "hesitantly", "loosely"], "correctAnswer": "forcefully"},
                {"id": "b3", "options": ["hazard", "peril", "protection", "vulnerability"], "correctAnswer": "protection"}
            ],
            "explanation": "b1: 'cable' (haulage cable) is the primary traction link ('chassis', 'window', and 'cushion' are carriage components). b2: 'forcefully' describes emergency brake clamps gripping the rail ('gently', 'loosely', and 'hesitantly' mean weak braking that causes crashes). b3: 'protection' denotes passenger safety ('peril', 'vulnerability', and 'hazard' denote danger)."
        }
    ],
    # 5 RFIB items (Reading: Fill in the Blanks - Drag and Drop)
    "rfib": [
        {
            "title": "Steel Friction: Adhesion Limits",
            "text": "Traditional trains rely entirely on frictional contact between wheels and tracks to maintain movement. On steep inclines, smooth steel wheels lose their mechanical [[b1]] against rails. Without sufficient friction, locomotives cannot generate enough [[b2]] to haul cars uphill, making cable-assisted engineering a necessity on extreme mountain [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "grip"},
                {"id": "b2", "correctAnswer": "traction"},
                {"id": "b3", "correctAnswer": "gradients"}
            ],
            "pool": ['grip', 'traction', 'gradients', 'friction', 'adhesion', 'inclines'],
            "explanation": "Wheels lose mechanical grip (friction), cannot generate traction (forward pulling power), on mountain gradients (slopes)."
        },
        {
            "title": "Summit Terminals: Drive Pulleys",
            "text": "The mechanical powerhouse of a modern funicular sits at the top of the mountain. A massive motorized [[b1]] wheel turns the heavy steel cable that connects the ascending and descending trains. Because the two carriages counterbalance each other, the motor only needs to overcome mechanical [[b2]] and payload [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "pulley"},
                {"id": "b2", "correctAnswer": "friction"},
                {"id": "b3", "correctAnswer": "differences"}
            ],
            "pool": ['pulley', 'friction', 'differences', 'sheave', 'tension', 'variations'],
            "explanation": "Motor turns a pulley wheel (sheave), overcoming friction (resistance), and payload differences (weight variations)."
        },
        {
            "title": "Water Ballast Dynamics: Alpine Drainage",
            "text": "Water-powered funiculars offered an ingenious solution to alpine transit challenges. By storing mountain stream water in underfloor [[b1]], descending cars became heavy enough to lift passenger loads without burning coal. After releasing the water into valley [[b2]], the system relied on nature's gravity to maintain continuous [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "tanks"},
                {"id": "b2", "correctAnswer": "streams"},
                {"id": "b3", "correctAnswer": "service"}
            ],
            "pool": ['tanks', 'streams', 'service', 'reservoirs', 'torrents', 'operation'],
            "explanation": "Water was stored in underfloor tanks (reservoirs), discharged into valley streams (channels), maintaining continuous service (operation)."
        },
        {
            "title": "Tunnel Infrastructure: Subterranean Boring",
            "text": "Modern alpine resorts increasingly construct funiculars inside solid bedrock. Running tracks through inclined mountain [[b1]] shields transit operations from violent snowstorms and winter [[b2]]. This underground layout protects delicate alpine ecosystems while delivering thousands of passengers to mountain [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "tunnels"},
                {"id": "b2", "correctAnswer": "avalanches"},
                {"id": "b3", "correctAnswer": "summits"}
            ],
            "pool": ['tunnels', 'avalanches', 'summits', 'galleries', 'rockfalls', 'peaks'],
            "explanation": "Tracks run through tunnels (underground bores), shielded from avalanches (snow slides), delivering visitors to summits (peaks)."
        },
        {
            "title": "Urban Transit: Energy Regeneration",
            "text": "Cable railways are finding renewed utility in hilly metropolitan areas. Urban funiculars equipped with electric motors can capture energy during downhill [[b1]]. By converting braking friction into electricity, these modern lines feed power back into municipal [[b2]], offering a clean alternative to diesel-powered bus [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "trips"},
                {"id": "b2", "correctAnswer": "grids"},
                {"id": "b3", "correctAnswer": "fleets"}
            ],
            "pool": ['trips', 'grids', 'fleets', 'descents', 'networks', 'wagons'],
            "explanation": "Motors capture energy during downhill trips (journeys), feeding electricity to municipal grids (power networks), replacing bus fleets (vehicles)."
        }
    ],
    # 5 RMCSA items (Multiple Choice, Single Answer)
    "mcq_single": [
        {
            "title": "Alpine Railways: Adhesion Limit",
            "prompt": "According to paragraph A, why do conventional adhesion railways fail on steep mountain gradients exceeding eight percent?",
            "options": [
                {"id": "A", "text": "Frictional grip between smooth steel wheels and rails is insufficient to prevent slipping or runaway sliding"},
                {"id": "B", "text": "Locomotive steam engines cannot function in cold high-altitude mountain air"},
                {"id": "C", "text": "Steel tracks bend and melt under extreme atmospheric pressure on mountain peaks"},
                {"id": "D", "text": "Mountain wildlife frequently damages locomotive electrical contact wires"}
            ],
            "key": "A",
            "explanation": "Paragraph A states that standard railways operate via wheel-rail friction, and beyond gradients of 6-8%, this interface fails, causing wheel slip climbing and loss of braking adhesion descending."
        },
        {
            "title": "Mechanical Principles: Counterbalance Function",
            "prompt": "How does the principle of counterbalance reduce the energy required to operate a funicular railway (paragraph B)?",
            "options": [
                {"id": "A", "text": "The descending carriage acts as a counterweight that pulls the ascending carriage upward"},
                {"id": "B", "text": "Both carriages are equipped with solar panels that generate electrical propulsion power"},
                {"id": "C", "text": "The haulage cable floats on subterranean water channels to eliminate all friction"},
                {"id": "D", "text": "The drive motor stores passenger tickets to calculate track resistance"}
            ],
            "key": "A",
            "explanation": "Paragraph B explains that as one carriage descends, its gravitational weight pulls the ascending carriage upward, acting as a massive counterweight so the motor only overcomes friction and load differences."
        },
        {
            "title": "Water Ballast Systems: Giessbachbahn Mechanics",
            "prompt": "In paragraph C, how did the Giessbachbahn operate without an external steam engine or electric motor?",
            "options": [
                {"id": "A", "text": "Water diverted from an alpine stream filled an underfloor tank in the descending car until it outweighed the lower car"},
                {"id": "B", "text": "Carriages were towed upward by teams of trained alpine draft horses"},
                {"id": "C", "text": "Compressed natural gas was ignited inside hollow steel rail tubes"},
                {"id": "D", "text": "Passengers were required to push the carriages along level sections of track"}
            ],
            "key": "A",
            "explanation": "Paragraph C states that water diverted from an alpine stream was pumped into a ballast tank beneath the descending carriage until it outweighed the ascending car, with gravity pulling the heavier carriage down."
        },
        {
            "title": "Abt Switch Design: Pointless Operation",
            "prompt": "What major mechanical advantage of the Abt bypass switch is highlighted in paragraph D?",
            "options": [
                {"id": "A", "text": "It operates completely without moving track points or frogs, eliminating derailment risks"},
                {"id": "B", "text": "It allows funicular carriages to jump across open chasms without steel rails"},
                {"id": "C", "text": "It automatically converts descending carriages into passenger gliders"},
                {"id": "D", "text": "It requires only single-flanged wooden wheels on all four corners of the car"}
            ],
            "key": "A",
            "explanation": "Paragraph D highlights that the Abt switch possesses no moving track points or frogs, using asymmetric wheelsets to guide carriages into designated sidings without derailment risks."
        },
        {
            "title": "Modern Subterranean Funiculars: Weather Defense",
            "prompt": "Why did modern alpine ski resorts move funicular transit systems inside inclined tunnels bored through granite bedrock (paragraph F)?",
            "options": [
                {"id": "A", "text": "To protect transit operations from blizzards, rockfalls, and avalanches while preserving surface ecology"},
                {"id": "B", "text": "To prevent passengers from seeing the steep mountain precipices during transit"},
                {"id": "C", "text": "To mine valuable metallic gold and copper ore along the track alignment"},
                {"id": "D", "text": "Because steel haulage cables dissolve when exposed to open alpine sunlight"}
            ],
            "key": "A",
            "explanation": "Paragraph F explains that surface funiculars were frequently disrupted by blizzards, rockfalls, and avalanches, prompting engineers to bore inclined tunnels entirely inside mountain bedrock."
        }
    ],
    # 5 RMCMA items (Multiple Choice, Multiple Answers)
    "mcq_multiple": [
        {
            "title": "Funicular Components: Terminal Equipment",
            "prompt": "According to paragraph B, which TWO physical components are fundamental to the operation of a funicular railway?",
            "options": [
                {"id": "A", "text": "A continuous steel haulage cable linking two passenger carriages"},
                {"id": "B", "text": "A motorized drive sheave (pulley wheel) housed at the upper station"},
                {"id": "C", "text": "A coal-fired boiler mounted onboard each passenger carriage"},
                {"id": "D", "text": "Magnetic levitation coils embedded along the entire track bed"},
                {"id": "E", "text": "High-altitude passenger wings attached to carriage roofs"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph B identifies a continuous steel haulage cable linking matched carriages (A) and a motorized drive sheave housed in the upper terminal (B)."
        },
        {
            "title": "Water Ballast Systems: Operational Sequence",
            "prompt": "According to paragraph C, which TWO steps were part of the operating cycle of gravity-driven water-ballast funiculars?",
            "options": [
                {"id": "A", "text": "Pumping mountain stream water into an underfloor ballast tank at the summit"},
                {"id": "B", "text": "Dumping the ballast water into a drainage channel upon arriving at the valley terminal"},
                {"id": "C", "text": "Heating ballast water into high-pressure steam to drive locomotive pistons"},
                {"id": "D", "text": "Freezing ballast water into solid ice blocks to increase carriage aerodynamics"},
                {"id": "E", "text": "Boiling stream water to sterilize municipal drinking supplies"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph C describes pumping alpine water into a ballast tank at the summit (A) and dumping it into a drainage channel at the lower valley terminal (B)."
        },
        {
            "title": "Abt Switch Anatomy: Asymmetrical Wheelsets",
            "prompt": "According to paragraph D, which TWO features describe the wheelsets on funicular carriages using the Abt bypass system?",
            "options": [
                {"id": "A", "text": "Double-flanged wheels on the outer side that steer the carriage along the continuous rail"},
                {"id": "B", "text": "Wide, flat flangeless wheels on the inner side that glide across rail gaps"},
                {"id": "C", "text": "Pneumatic rubber tyres designed for driving on unpaved gravel roads"},
                {"id": "D", "text": "Spiked iron wheels that puncture wooden ties to prevent slipping"},
                {"id": "E", "text": "Caster wheels that swivel three hundred and sixty degrees in any direction"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph D states wheels on one side feature double flanges gripping the outer rail (A), while wheels on the opposing side are flat and flangeless, gliding over rail gaps (B)."
        },
        {
            "title": "Emergency Braking: Autonomous Clamping",
            "prompt": "According to paragraph E, which TWO characteristics describe emergency track brakes on funicular carriages?",
            "options": [
                {"id": "A", "text": "They deploy spring-loaded hydraulic wedge clamps that grip the rail head directly"},
                {"id": "B", "text": "They operate autonomously if cable tension drops or velocity exceeds safety thresholds"},
                {"id": "C", "text": "They reverse the direction of the rotating wheels to generate traction friction"},
                {"id": "D", "text": "They eject heavy concrete anchors into the surrounding soil using gunpowder"},
                {"id": "E", "text": "They detach the passenger cabin from the chassis and deploy parachutes"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph E states brakes feature spring-loaded hydraulic clamps gripping the rail head directly (A) and deploy autonomously if cable tension drops or speed limits are exceeded (B)."
        },
        {
            "title": "Modern Urban Transit: Sustainable Advantages",
            "prompt": "According to paragraph F, which TWO environmental and operational benefits are offered by modernized urban funiculars?",
            "options": [
                {"id": "A", "text": "Whisper-quiet, zero-emission mass transit across steep urban topography"},
                {"id": "B", "text": "Regenerative electric motors that feed power back into municipal grids during descent"},
                {"id": "C", "text": "Ability to fly over tall city skyscrapers without using fixed guide rails"},
                {"id": "D", "text": "Complete immunity to municipal electrical power outages"},
                {"id": "E", "text": "Elimination of all public ticketing and passenger boarding stations"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph F highlights quiet, zero-emission mobility across steep topography (A) and regenerative motors feeding energy back into municipal grids (B)."
        }
    ]
}
