"""Passage 6: The Mechanics of Insect Flight."""

PASSAGE_6 = {
    "id": "rp-pr89-006",
    "title": "The Mechanics of Insect Flight",
    "subtitle": "How asynchronous flight muscles and leading-edge vortices generate lift",
    "topic": "Biomechanics and Aerodynamics",
    "paragraphs": [
        {
            "label": "A",
            "text": "For much of the twentieth century, the biomechanics of insect flight remained an enduring aerodynamic paradox. According to standard aeronautical theory, which governs the steady-state performance of commercial passenger airliners and gliders, a bumblebee or hoverfly should be physically incapable of leaving the ground. Conventional aerodynamic equations assume smooth, uninterrupted airflow across rigid, stationary aerofoils operating at high Reynolds numbers. In the miniature realm of insects, however, wings operate at low Reynolds numbers—a physical domain where air behaves not as a thin gas, but as a thick, viscous fluid resembling light syrup. Furthermore, insect wings do not remain stationary; they flap, twist, and oscillate through complex three-dimensional figure-eight trajectories. Resolving how insects generate sufficient lift required aerodynamicists to discard steady-state assumptions and embrace the chaotic, highly dynamic physics of unsteady aerodynamics."
        },
        {
            "label": "B",
            "text": "The foundation of insect flight lies in the internal muscular architecture of the thorax. Over evolutionary time, winged insects evolved two fundamentally distinct physiological systems to drive wing locomotion. Basal orders, such as dragonflies and damselflies (Odonata), utilize direct flight musculature in which separate muscle bundles attach directly to the hardened base of each wing. This primitive arrangement allows exquisite independent steering of forewings and hindwings, but imposes a mechanical ceiling on wingbeat frequency. In contrast, more advanced orders—including flies (Diptera), bees (Hymenoptera), and beetles (Coleoptera)—possess indirect flight musculature. In these species, the flight muscles do not attach to the wings at all. Instead, massive blocks of longitudinal and dorsoventral muscles fill the thoracic cavity, rhythmically deforming the elastic chitinous exoskeleton. When the thorax flexes and snaps back like an oscillating mechanical spring, internal lever hinges amplify this minute skeletal deformation into sweeping wing strokes."
        },
        {
            "label": "C",
            "text": "In tiny insects such as midges and mosquitoes, wings beat at astonishing frequencies exceeding eight hundred to one thousand cycles per second. At these extreme velocities, standard neuromuscular physiology breaks down: a motor neuron cannot biochemically fire action potentials, release neurotransmitters, and reabsorb calcium ions fast enough to signal individual contractions. To circumvent this neurological limitation, advanced insects evolved fibrillar asynchronous muscle tissue. In asynchronous flight systems, muscle contraction is decoupled from individual nervous impulses. Instead, the muscle becomes stretch-activated: a single initial nerve impulse primes the system, after which the contraction of one muscle group deforms the thorax, mechanically stretching the opposing antagonistic muscle group and triggering its immediate contraction in response. This myogenic resonance sustains rapid, self-perpetuating oscillations with minimal neural expenditure."
        },
        {
            "label": "D",
            "text": "While asynchronous thoracic muscles supply the necessary mechanical power, the primary aerodynamic mechanism generating lift occurs along the wing surface itself: the leading-edge vortex (LEV). In conventional aviation, tilting an aircraft wing beyond an angle of fifteen degrees causes the airflow to detach completely, inducing a catastrophic loss of lift known as a stall. Insect wings, however, routinely operate at extreme angles of attack between thirty and forty-five degrees. As the wing sweeps through its downstroke, air separating over the sharp leading edge curls into a tight, spiraling whirlpool of air that runs parallel to the wing margin. This leading-edge vortex generates a localized pocket of intense low atmospheric pressure above the wing, producing an upward suction force that generates up to three times more lift than conventional steady-state aeronautics would predict."
        },
        {
            "label": "E",
            "text": "The complexity of insect kinematics is further enhanced by dynamic stroke reversals at the end of each half-stroke. As the wing completes its forward sweep, it does not simply halt; instead, internal steering muscles rapidly rotate the blade by nearly one hundred and eighty degrees along its longitudinal axis, flipping its orientation before initiating the backstroke. This rapid rotation generates rotational circulation that boosts lift immediately before reversal. Crucially, as the wing accelerates into the subsequent stroke, it sweeps directly back through the swirling, energetic wake shed during the preceding motion. Known as wake capture, this maneuver allows the insect to recover kinetic energy from the disturbed air, transforming what would otherwise be wasted aerodynamic turbulence into an additional burst of positive upward thrust."
        },
        {
            "label": "F",
            "text": "Today, insights gleaned from insect aerodynamics are driving a technological revolution in the development of biomimetic micro air vehicles (MAVs). Traditional rotary quadcopters suffer severe aerodynamic efficiency losses when scaled down to insect dimensions because miniature propellers operate poorly in viscous, low-Reynolds-number airflows. Robotics engineers at institutions such as Harvard have developed millimeter-scale robotic insects, such as the RoboBee, that utilize piezoelectric ceramic actuators and laser-machined carbon-fiber flexures to mimic true flapping kinematics. Weighing less than a tenth of a gram, these autonomous artificial insects are envisioned for hazardous operations where larger drones cannot venture, including navigating collapsed rubble during earthquake search-and-rescue missions, conducting environmental pollution monitoring inside industrial conduits, and performing targeted agricultural pollination."
        }
    ],
    # 5 RWFIB items (Fill in the Blanks - Dropdown)
    "rwfib": [
        {
            "title": "Insect Aerodynamics: Steady-State Paradox",
            "text": "For decades, classical aeronautical theory struggled to explain insect flight. Standard aerodynamic equations, which model passenger aircraft, assumed that air flows smoothly over rigid wings. At miniature scales, however, air behaves as a [[b1]] fluid, and insect wings undergo rapid oscillations. By discarding stationary wing models, scientists recognized that insects rely on [[b2]] aerodynamics. This dynamic approach finally [[b3]] the long-standing flight paradox.",
            "blanks": [
                {"id": "b1", "options": ["frictionless", "porous", "rarefied", "viscous"], "correctAnswer": "viscous"},
                {"id": "b2", "options": ["laminar", "quiescent", "static", "unsteady"], "correctAnswer": "unsteady"},
                {"id": "b3", "options": ["exacerbated", "obfuscated", "postponed", "resolved"], "correctAnswer": "resolved"}
            ],
            "explanation": "b1: 'viscous' matches air behaving like thick fluid at micro-scales ('porous' means permeable; 'frictionless' is opposite; 'rarefied' means thin). b2: 'unsteady' describes dynamic non-steady aerodynamics ('static', 'laminar', and 'quiescent' mean steady or still). b3: 'resolved' means solving the paradox ('obfuscated' means confusing; 'exacerbated' means worsening; 'postponed' means delaying)."
        },
        {
            "title": "Flight Musculature: Thoracic Mechanics",
            "text": "Insects possess two distinct muscular designs for driving wing motion. Dragonflies feature direct flight muscles that attach directly to the wing bases, enabling precise individual [[b1]]. In contrast, advanced species such as flies use indirect musculature that deforms the elastic [[b2]]. When thoracic muscles flex, internal hinges amplify these small movements into [[b3]] wing strokes.",
            "blanks": [
                {"id": "b1", "options": ["hesitation", "paralysis", "stagnation", "steering"], "correctAnswer": "steering"},
                {"id": "b2", "options": ["exoskeleton", "marrow", "skeleton", "tendon"], "correctAnswer": "exoskeleton"},
                {"id": "b3", "options": ["freezing", "pausing", "stumbling", "sweeping"], "correctAnswer": "sweeping"}
            ],
            "explanation": "b1: 'steering' describes flight control adjustments ('stagnation', 'paralysis', and 'hesitation' denote lack of movement). b2: 'exoskeleton' is the correct anatomical term for insect outer cuticle ('marrow', 'skeleton', and 'tendon' apply to vertebrate biology). b3: 'sweeping' describes broad wing stroke arcs ('pausing', 'stumbling', and 'freezing' mean stopping)."
        },
        {
            "title": "Asynchronous Physiology: Stretch Activation",
            "text": "In rapidly flapping insects, wingbeat frequencies can exceed one thousand beats per second. Because nerve impulses cannot fire fast enough to signal individual contractions, advanced fliers evolved [[b1]] muscles. In this system, muscle contraction is triggered by mechanical [[b2]] rather than neural commands. A single nerve impulse initiates an oscillating cycle that [[b3]] rapid wing beats with minimal brain overhead.",
            "blanks": [
                {"id": "b1", "options": ["asynchronous", "consecutive", "coordinated", "synchronized"], "correctAnswer": "asynchronous"},
                {"id": "b2", "options": ["compression", "recoil", "slack", "stretch"], "correctAnswer": "stretch"},
                {"id": "b3", "options": ["halts", "interrupts", "sabotages", "sustains"], "correctAnswer": "sustains"}
            ],
            "explanation": "b1: 'asynchronous' describes muscle oscillations decoupled from direct nerve impulses ('synchronized', 'consecutive', and 'coordinated' mean one-to-one firing). b2: 'stretch' triggers delayed contraction ('slack' and 'compression' mean looseness or squashing; 'recoil' is rebound). b3: 'sustains' describes maintaining continuous high frequency ('halts', 'interrupts', and 'sabotages' mean stopping)."
        },
        {
            "title": "Vortex Aerodynamics: Leading-Edge Suction",
            "text": "Insect wings achieve remarkable lift by operating at high angles of attack. When an insect sweeps its wing forward, the air separating over the front margin curls into a [[b1]] vortex. This stable vortex creates an area of intense low [[b2]] directly above the wing surface. The resulting suction force produces substantially more lift than could be generated by [[b3]] wings.",
            "blanks": [
                {"id": "b1", "options": ["level", "parallel", "spiral", "straight"], "correctAnswer": "spiral"},
                {"id": "b2", "options": ["drag", "friction", "gravity", "pressure"], "correctAnswer": "pressure"},
                {"id": "b3", "options": ["chaotic", "mobile", "stationary", "turbulent"], "correctAnswer": "stationary"}
            ],
            "explanation": "b1: 'spiral' describes the leading-edge vortex vortex core ('straight', 'parallel', and 'level' are non-vortical geometry). b2: 'pressure' (specifically low pressure suction) creates lift ('friction', 'gravity', and 'drag' do not create lift). b3: 'stationary' matches vortex remaining attached without shedding ('mobile', 'turbulent', and 'chaotic' describe shedding and detachment)."
        },
        {
            "title": "Biomimetic Drones: Micro Aerial Robotics",
            "text": "Principles of insect aerodynamics are inspiring the design of miniature robots. Traditional propellers lose efficiency at small scales because air viscosity [[b1]] performance. Consequently, engineers are developing micro aerial vehicles with flapping wings driven by [[b2]] actuators. These compact devices are ideal for navigating hazardous collapsed buildings during [[b3]] operations.",
            "blanks": [
                {"id": "b1", "options": ["degrades", "endures", "improves", "thrives"], "correctAnswer": "degrades"},
                {"id": "b2", "options": ["combustible", "hydraulic", "piezoelectric", "pneumatic"], "correctAnswer": "piezoelectric"},
                {"id": "b3", "options": ["negligence", "obsolescence", "reconnaissance", "retirement"], "correctAnswer": "reconnaissance"}
            ],
            "explanation": "b1: 'degrades' describes micro-motor efficiency falling at tiny scales ('thrives', 'improves', and 'endures' mean remaining good). b2: 'piezoelectric' is the precise solid-state actuation technology ('combustible' means flammable; 'pneumatic' and 'hydraulic' are bulky fluid systems). b3: 'reconnaissance' denotes micro-drone surveillance missions ('negligence', 'obsolescence', and 'retirement' describe failure or discard)."
        }
    ],
    # 5 RFIB items (Reading: Fill in the Blanks - Drag and Drop)
    "rfib": [
        {
            "title": "Low Reynolds Physics: Viscous Airflow",
            "text": "Operating at millimeter scales fundamentally changes the physics of movement. In the miniature realm of small insects, air acts as a relatively [[b1]] fluid rather than a frictionless gas. Because inertia is low relative to surface [[b2]], wings must actively generate vortex patterns to maintain aerodynamic [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "viscous"},
                {"id": "b2", "correctAnswer": "drag"},
                {"id": "b3", "correctAnswer": "support"}
            ],
            "pool": ['viscous', 'drag', 'support', 'turbulent', 'friction', 'buoyancy'],
            "explanation": "Air acts as a viscous fluid (thick/resistant), inertia is low relative to drag (friction), requiring vortices for support (lift)."
        },
        {
            "title": "Indirect Muscles: Thoracic Oscillation",
            "text": "The evolution of indirect flight musculature provided major aerodynamic advantages. Instead of connecting directly to wing joints, massive muscles deform the chitinous [[b1]] of the insect body. The thoracic skeleton functions as a mechanical [[b2]], storing and releasing elastic energy with each wing [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "shell"},
                {"id": "b2", "correctAnswer": "spring"},
                {"id": "b3", "correctAnswer": "stroke"}
            ],
            "pool": ['shell', 'spring', 'stroke', 'skeleton', 'hinge', 'motion'],
            "explanation": "Muscles deform the outer shell (exoskeleton), which functions as an elastic spring (energy store), with each stroke (flapping cycle)."
        },
        {
            "title": "Dynamic Wing Twisting: Stroke Reversal",
            "text": "At the conclusion of each half-stroke, insect wings undergo rapid repositioning. Specialized muscles twist the blade to alter its angle before initiating the [[b1]] motion. This rapid rotation generates additional [[b2]], allowing the wing to smoothly cut back through the turbulent [[b3]] left by the previous sweep.",
            "blanks": [
                {"id": "b1", "correctAnswer": "return"},
                {"id": "b2", "correctAnswer": "circulation"},
                {"id": "b3", "correctAnswer": "wake"}
            ],
            "pool": ['return', 'circulation', 'wake', 'recovery', 'vorticity', 'plume'],
            "explanation": "The wing initiates the return motion (backstroke), generating circulation (fluid spin), cutting through the wake (disturbed air)."
        },
        {
            "title": "Wake Capture: Energy Recovery",
            "text": "Insects utilize sophisticated kinematic maneuvers to conserve energy in flight. When reversing stroke direction, the wing intercepts the swirling vortex [[b1]] behind by the previous movement. By capturing this kinetic energy, the insect produces a powerful pulse of upward [[b2]] while reducing overall muscular [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "shed"},
                {"id": "b2", "correctAnswer": "thrust"},
                {"id": "b3", "correctAnswer": "strain"}
            ],
            "pool": ['shed', 'thrust', 'strain', 'formed', 'propulsion', 'pressure'],
            "explanation": "The wing intercepts vortices shed previously (released), producing upward thrust (force), reducing muscular strain (effort)."
        },
        {
            "title": "Robotic Kinematics: Micro Fabrication",
            "text": "Constructing functional robotic insects requires unconventional manufacturing strategies. Traditional mechanical gears and rotary joints cannot easily be [[b1]] at millimeter dimensions. Instead, micro air vehicles rely on flexible carbon-fiber [[b2]] that bend elastically, replicating the resonant movements of biological [[b3]].",
            "blanks": [
                {"id": "b1", "correctAnswer": "machined"},
                {"id": "b2", "correctAnswer": "joints"},
                {"id": "b3", "correctAnswer": "wings"}
            ],
            "pool": ['machined', 'joints', 'wings', 'fabricated', 'hinges', 'blades'],
            "explanation": "Gears cannot be easily machined at micro scales (manufactured), relying on flexible joints (hinges), mimicking wings (appendages)."
        }
    ],
    # 5 RMCSA items (Multiple Choice, Single Answer)
    "mcq_single": [
        {
            "title": "Insect Aerodynamics: Steady-State Failure",
            "prompt": "According to paragraph A, why did conventional steady-state aeronautical theory predict that bumblebees could not fly?",
            "options": [
                {"id": "A", "text": "It evaluated wings as static aerofoils, ignoring dynamic unsteady vortex phenomena at low Reynolds numbers"},
                {"id": "B", "text": "It assumed insects were too heavy to generate internal metabolic heat"},
                {"id": "C", "text": "It failed to recognize that insect flight muscles are powered by hydraulic liquid pressure"},
                {"id": "D", "text": "It assumed bumblebee wings were made of porous limestone rather than chitin"}
            ],
            "key": "A",
            "explanation": "Paragraph A states conventional theory assumed smooth, uninterrupted airflow across rigid, stationary aerofoils, whereas insects operate at low Reynolds numbers in an unsteady aerodynamic regime."
        },
        {
            "title": "Thoracic Anatomy: Indirect Musculature",
            "prompt": "How do indirect flight muscles move the wings in advanced insects (paragraph B)?",
            "options": [
                {"id": "A", "text": "They rhythmically deform the elastic thoracic exoskeleton, which transfers motion to wings through internal lever hinges"},
                {"id": "B", "text": "They pull directly on individual tendons anchored to the wing tips"},
                {"id": "C", "text": "They pump pressurized blood into hollow wing veins to force downward flapping"},
                {"id": "D", "text": "They rotate the insect's head to deflect incoming atmospheric air currents"}
            ],
            "key": "A",
            "explanation": "Paragraph B explains that muscles do not attach to wings at all, but rhythmically deform the elastic chitinous exoskeleton, with internal hinges amplifying the deformation into sweeping strokes."
        },
        {
            "title": "Muscle Physiology: Asynchronous Contraction",
            "prompt": "According to paragraph C, what enables asynchronous flight muscles to beat faster than the rate of neural firing?",
            "options": [
                {"id": "A", "text": "The muscles are stretch-activated, maintaining an oscillating myogenic rhythm triggered by skeletal deformation"},
                {"id": "B", "text": "Each wing possesses its own miniature secondary brain that generates electrical signals"},
                {"id": "C", "text": "The insect's nervous system relies on optical laser pulses rather than chemical neurotransmitters"},
                {"id": "D", "text": "Nerve fibers are insulated with liquid petroleum gel to accelerate electrical signals"}
            ],
            "key": "A",
            "explanation": "Paragraph C explains that muscle becomes stretch-activated, where contraction of one group mechanically stretches the opposing group, sustaining rapid oscillations with minimal neural expenditure."
        },
        {
            "title": "Aerodynamic Mechanisms: Leading-Edge Vortex",
            "prompt": "In paragraph D, what is the primary role of the leading-edge vortex (LEV) during an insect's wing downstroke?",
            "options": [
                {"id": "A", "text": "It creates a pocket of intense low pressure above the wing that generates powerful upward suction lift"},
                {"id": "B", "text": "It warms the surrounding air to prevent ice accumulation on delicate wing membranes"},
                {"id": "C", "text": "It deflects sound waves to prevent nocturnal predators from hearing the flapping sound"},
                {"id": "D", "text": "It forces air backward to act as a miniature jet propulsion exhaust"}
            ],
            "key": "A",
            "explanation": "Paragraph D states that the leading-edge vortex generates a localized pocket of intense low atmospheric pressure above the wing, producing an upward suction force that generates substantial lift."
        },
        {
            "title": "Biomimetic Robotics: Drone Miniaturization",
            "prompt": "According to paragraph F, why are flapping wings preferred over rotary propellers for millimeter-scale micro air vehicles?",
            "options": [
                {"id": "A", "text": "Miniature propellers suffer severe efficiency losses in viscous, low-Reynolds-number airflows"},
                {"id": "B", "text": "Rotary motors cannot operate when exposed to ordinary atmospheric oxygen"},
                {"id": "C", "text": "Flapping wings are immune to damage from electrical power surges"},
                {"id": "D", "text": "International aviation regulations prohibit propeller-driven drones below one centimetre"}
            ],
            "key": "A",
            "explanation": "Paragraph F notes that traditional rotary quadcopters suffer severe aerodynamic efficiency losses when scaled down because miniature propellers operate poorly in viscous, low-Reynolds-number airflows."
        }
    ],
    # 5 RMCMA items (Multiple Choice, Multiple Answers)
    "mcq_multiple": [
        {
            "title": "Aerodynamic Constraints: Low Reynolds Flight",
            "prompt": "According to paragraph A, which TWO factors characterize the physical regime of insect flight?",
            "options": [
                {"id": "A", "text": "Low Reynolds numbers where air behaves as a relatively thick, viscous fluid"},
                {"id": "B", "text": "Unsteady aerodynamic conditions caused by complex oscillating wing trajectories"},
                {"id": "C", "text": "Supersonic velocities that generate atmospheric sonic booms"},
                {"id": "D", "text": "Steady-state laminar airflow across perfectly rigid stationary wings"},
                {"id": "E", "text": "Complete absence of atmospheric air resistance at ground level"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph A identifies low Reynolds numbers where air behaves as a thick, viscous fluid (A) and dynamic unsteady aerodynamics with oscillating trajectories (B)."
        },
        {
            "title": "Insect Orders: Muscular Classification",
            "prompt": "According to paragraph B, which TWO insect orders rely on indirect flight musculature?",
            "options": [
                {"id": "A", "text": "Diptera (true flies)"},
                {"id": "B", "text": "Hymenoptera (bees and wasps)"},
                {"id": "C", "text": "Odonata (dragonflies and damselflies)"},
                {"id": "D", "text": "Araneae (spiders)"},
                {"id": "E", "text": "Decapoda (crabs and lobsters)"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph B identifies advanced orders including flies (Diptera) and bees (Hymenoptera) as possessing indirect flight musculature (A and B)."
        },
        {
            "title": "High-Frequency Flight: Neurological Overcoming",
            "prompt": "According to paragraph C, which TWO physiological features describe asynchronous flight systems?",
            "options": [
                {"id": "A", "text": "Contractions are decoupled from individual motor neuron action potentials"},
                {"id": "B", "text": "Mechanical stretching of opposing muscle blocks triggers sequential contractions"},
                {"id": "C", "text": "Each wing contraction requires three separate neural chemical signals"},
                {"id": "D", "text": "Flight muscles consume zero metabolic energy during continuous hovering"},
                {"id": "E", "text": "Nerves physically contract and pull on wing joints without muscles"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph C notes contraction is decoupled from individual nervous impulses (A) and muscles are stretch-activated by opposing contractions (B)."
        },
        {
            "title": "Kinematic Maneuvers: Dynamic Stroke Mechanisms",
            "prompt": "According to paragraph E, which TWO aerodynamic phenomena occur at the reversal point of an insect's wing stroke?",
            "options": [
                {"id": "A", "text": "Rapid wing rotation along the longitudinal axis generating rotational circulation lift"},
                {"id": "B", "text": "Wake capture where the wing extracts kinetic energy from previously shed turbulence"},
                {"id": "C", "text": "Temporary cessation of all aerodynamic forces allowing the insect to free-fall"},
                {"id": "D", "text": "Ejection of microscopic water droplets to reduce insect body mass"},
                {"id": "E", "text": "Sudden locking of the thoracic hinges to halt muscle oscillation"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph E identifies rapid rotation along the longitudinal axis generating rotational circulation (A) and wake capture recovering energy from disturbed air (B)."
        },
        {
            "title": "Biomimetic Robotics: MAV Applications",
            "prompt": "According to paragraph F, which TWO real-world applications are envisioned for insect-scale flapping-wing robots?",
            "options": [
                {"id": "A", "text": "Navigating through collapsed rubble during post-disaster search-and-rescue operations"},
                {"id": "B", "text": "Performing targeted agricultural pollination where natural pollinators are absent"},
                {"id": "C", "text": "Transporting heavy cargo freight across transoceanic commercial shipping lanes"},
                {"id": "D", "text": "Dredging deep maritime navigation channels in silting harbour basins"},
                {"id": "E", "text": "Generating commercial electrical power from high-altitude jet stream winds"}
            ],
            "keys": ["A", "B"],
            "explanation": "Paragraph F envisions navigating collapsed rubble in search-and-rescue (A) and performing targeted agricultural pollination (B)."
        }
    ]
}
