"""Replacement Part 4 lectures for Listening Tests 2-5 and 7-10, and
replacement short lecture drills.

The first versions repeated topics already live as reading passages
(bamboo, honeybee dances, urban heat islands, sleep in animals, mangroves,
paper money, octopus intelligence, lighthouses; and for the drills beekeeping,
the bicycle, sleep and memory, coral reefs, salt, bird migration, chocolate,
and glass-making, which batch 3 reading now covers). A learner who had read
the passage could answer the listening questions without listening.
scripts/live_topics.py now checks every heading against the live titles.
Original Prepyo practice material; not official or recalled test content.
"""

W1_NOTES = 'Complete the notes below. Write ONE WORD ONLY for each answer.'


def lecture(setting, heading, script, questions):
    return dict(setting=setting, script=script, groups=[dict(
        type_id='ielts-listening-completion', heading=heading,
        instructions=W1_NOTES, limit=1, questions=questions)])


PART4 = {
    't2': lecture(
        'You will hear part of a lecture about the history of preserving food in sealed containers.',
        'The history of canning',
        "Lecturer: Today I'm going to talk about a technology that most of us use every day without thinking about it: the preservation of food in sealed containers. "
        "At the end of the eighteenth century, armies and navies faced a serious problem. Fresh food spoiled quickly, and soldiers on long campaigns often suffered from poor diets. "
        "In 1795 the French government offered a prize of twelve thousand francs to anyone who could find a reliable way of preserving food. "
        "The prize was eventually won by Nicolas Appert, a confectioner from Paris. Appert's method was to seal food inside glass bottles, close them with cork, and then heat the bottles in boiling water. "
        "He did not know why the method worked. That explanation came decades later, when Louis Pasteur showed that heat kills the microorganisms that cause food to decay. "
        "Glass had one obvious weakness: it broke easily. In 1810 an Englishman, Peter Durand, was granted a patent for preserving food in containers made of tin-coated iron, "
        "and within a few years a factory in London was supplying canned food to the Royal Navy. "
        "Early cans were heavy and extremely difficult to open. The instructions on some of them advised using a hammer and chisel. "
        "Remarkably, a practical can opener was not invented until almost fifty years after the first cans appeared. "
        "Canned food changed the way people ate. It allowed explorers to travel for years at a time, and by the end of the nineteenth century it had made fruit, vegetables and fish "
        "available all year round in cities far from where they were produced. "
        "There were problems, however. Some early cans were sealed with lead solder, which could contaminate the food, and it took many years before safer methods were adopted. "
        "Today most cans are made of steel or aluminium, and they are among the most widely recycled packaging materials in the world.",
        [
            ('Soldiers on long campaigns often had poor ________', ['diets', 'diet'], 'suffered from poor diets'),
            ('The prize was offered by the French ________', ['government'], 'the French government offered a prize'),
            ('Appert worked as a ________', ['confectioner'], 'a confectioner from Paris'),
            ('Appert closed the bottles with ________', ['cork'], 'close them with cork'),
            ('Pasteur showed that heat kills ________', ['microorganisms'], 'kills the microorganisms'),
            ('In 1810 Durand was granted a ________', ['patent'], 'granted a patent'),
            ('Some early cans had to be opened with a hammer and ________', ['chisel'], 'a hammer and chisel'),
            ('Canned food allowed ________ to travel for years', ['explorers'], 'allowed explorers to travel'),
            ('Some early cans were sealed with ________ solder', ['lead'], 'sealed with lead solder'),
            ('Today cans are among the most widely ________ packaging materials', ['recycled'], 'most widely recycled'),
        ]),
    't3': lecture(
        'You will hear part of a lecture about the axolotl.',
        'The axolotl',
        "Lecturer: This morning I'd like to introduce you to one of the most studied animals in biology: the axolotl, a salamander that comes from a single lake system near Mexico City. "
        "Axolotls are unusual in several ways. Most salamanders begin life in water, breathing through gills, and later change into land-living adults with lungs. "
        "Axolotls, however, usually never complete this change. They keep their feathery external gills and remain in the water throughout their lives, a condition known as neoteny. "
        "They can nevertheless reproduce, just like the adults of other species. "
        "What has made axolotls famous in laboratories is their power of regeneration. If an axolotl loses a leg, it can grow a complete new one, including bones, muscles and nerves, "
        "usually within a few weeks, and without leaving a scar. Even more remarkably, they can repair damage to the spinal cord, and even to parts of the heart and brain. "
        "How do they do it? After an injury, cells at the wound form a structure called a blastema, a mass of cells that can develop into the different tissues needed to rebuild the missing part. "
        "Researchers hope that understanding this process may one day help doctors to improve healing in humans, although that goal is still a long way off. "
        "Axolotls also have a very large genome, around ten times the size of the human genome, which made it difficult for scientists to decode. It was fully sequenced only in 2018. "
        "Sadly, while axolotls are common in laboratories and as pets, they are critically endangered in the wild. "
        "Their only natural habitat, the canals of Xochimilco, has been damaged by pollution and the growth of the city, and introduced fish such as carp and tilapia eat their eggs and young. "
        "Conservation projects now work with local farmers to create refuges where the water is filtered and protected.",
        [
            ('Most salamanders become land-living adults with ________', ['lungs'], 'adults with lungs'),
            ('Axolotls keep their feathery external ________', ['gills'], 'feathery external gills'),
            ('Staying in this young form is called ________', ['neoteny'], 'known as neoteny'),
            ('A new leg grows without leaving a ________', ['scar'], 'without leaving a scar'),
            ('They can repair damage to the spinal ________', ['cord'], 'the spinal cord'),
            ('After injury, cells form a structure called a ________', ['blastema'], 'called a blastema'),
            ('Their genome is about ten times the size of the ________ genome', ['human'], 'the human genome'),
            ('The genome was fully sequenced in ________', ['2018'], 'only in 2018'),
            ('Their habitat has been damaged by ________ and the growth of the city', ['pollution'], 'damaged by pollution'),
            ('Conservation projects work with local ________', ['farmers'], 'local farmers'),
        ]),
    't4': lecture(
        'You will hear part of a lecture about how scientists monitor volcanoes.',
        'Monitoring volcanoes',
        "Lecturer: Today I'm going to look at how scientists monitor volcanoes and try to forecast eruptions. "
        "About eight hundred million people live within a hundred kilometres of an active volcano, so this work matters enormously. "
        "The most important tool is the seismometer, which records small earthquakes. As molten rock, or magma, forces its way upwards, it cracks the surrounding rock, "
        "producing swarms of tiny earthquakes, often too weak for people to feel. A sudden increase in this activity is one of the clearest warning signs. "
        "A second method is to measure changes in the shape of the ground. When magma collects beneath a volcano, the surface can swell, sometimes by several metres. "
        "Scientists detect this deformation with GPS receivers placed on the volcano, and with satellites that use radar to compare images taken weeks or months apart. "
        "Gases provide a third clue. As magma rises, gases dissolved in it escape, rather like the bubbles in a bottle of fizzy drink when it is opened. "
        "Instruments that measure sulphur dioxide in the air above a volcano can show whether fresh magma is approaching the surface. "
        "Even with all these tools, forecasting is difficult. Many volcanoes show signs of unrest that never lead to an eruption, and each volcano behaves differently, "
        "so scientists rely heavily on the history of the particular volcano they are studying. "
        "A well-known success came in 1991 at Mount Pinatubo in the Philippines. Monitoring allowed scientists to warn the authorities, "
        "and tens of thousands of people were evacuated before one of the largest eruptions of the twentieth century. Without that warning, the death toll would almost certainly have been far higher. "
        "Finally, communication matters as much as science. Warnings are only useful if local people trust them, so many observatories now use simple colour-coded alert levels that everyone can understand.",
        [
            ('The most important tool: the ________', ['seismometer'], 'The most important tool is the seismometer'),
            ('Rising magma produces swarms of tiny ________', ['earthquakes'], 'swarms of tiny earthquakes'),
            ('As magma collects, the surface can ________', ['swell'], 'the surface can swell'),
            ('Deformation is detected by GPS receivers and by ________', ['satellites'], 'with satellites'),
            ('Satellites use ________ to compare images', ['radar'], 'use radar to compare'),
            ('Instruments measure sulphur ________ above the volcano', ['dioxide'], 'measure sulphur dioxide'),
            ('Scientists rely heavily on the ________ of each volcano', ['history'], 'rely heavily on the history'),
            ('Mount Pinatubo is in the ________', ['Philippines'], 'Mount Pinatubo in the Philippines'),
            ('In 1991 tens of thousands of people were ________', ['evacuated'], 'were evacuated'),
            ('Observatories use simple colour-coded alert ________', ['levels'], 'colour-coded alert levels'),
        ]),
    't5': lecture(
        'You will hear part of a lecture about why leaves change colour in autumn.',
        'Autumn colours in trees',
        "Lecturer: In today's lecture I want to explain one of the most spectacular events in the natural calendar: the change in the colour of leaves in autumn. "
        "During spring and summer, leaves are green because they contain large amounts of chlorophyll, the pigment that captures energy from sunlight for photosynthesis. "
        "Chlorophyll is constantly broken down and replaced while the leaf is active. "
        "As autumn approaches, the days become shorter and temperatures fall. Trees respond by stopping the production of chlorophyll. "
        "The chlorophyll that remains breaks down, and the tree recovers valuable nutrients such as nitrogen from it before the leaf falls. "
        "Once the green disappears, other pigments that were present all along become visible. These are the carotenoids, which produce yellow and orange colours. "
        "The same pigments give carrots their colour. "
        "Red and purple colours have a different origin. They come from pigments called anthocyanins, which in many species are not present in summer but are actually produced in autumn. "
        "Why a tree should spend energy making new pigments just before losing its leaves is still debated. One theory is that they act as a kind of sunscreen, "
        "protecting the leaf while nutrients are being withdrawn. "
        "The weather has a strong influence on how brilliant the colours are. The best displays usually follow a season of warm, sunny days and cool nights, but not freezing ones. "
        "A very early frost can kill the leaves before the colours develop. "
        "Finally, the leaf must be released. At the base of each leaf stalk, the tree forms a special layer of cells, called the abscission layer, which gradually cuts the leaf off. "
        "Evergreen trees, by contrast, keep most of their leaves through the winter, protected by a thick waxy coating.",
        [
            ('Leaves are green because they contain ________', ['chlorophyll'], 'large amounts of chlorophyll'),
            ('In autumn days become shorter and ________ fall', ['temperatures'], 'temperatures fall'),
            ('The tree recovers nutrients such as ________', ['nitrogen'], 'nutrients such as nitrogen'),
            ('Yellow and orange colours come from ________', ['carotenoids'], 'These are the carotenoids'),
            ('The same pigments give ________ their colour', ['carrots'], 'give carrots their colour'),
            ('Red and purple colours come from ________', ['anthocyanins'], 'pigments called anthocyanins'),
            ('One theory: red pigments act as a kind of ________', ['sunscreen'], 'a kind of sunscreen'),
            ('The best displays follow sunny days and cool ________', ['nights'], 'cool nights'),
            ('A very early ________ can kill the leaves', ['frost'], 'A very early frost'),
            ('Evergreen leaves are protected by a thick waxy ________', ['coating'], 'waxy coating'),
        ]),
    't7': lecture(
        'You will hear part of a lecture about how camels survive in the desert.',
        'How camels survive in the desert',
        "Lecturer: Today we're looking at one of the best-adapted animals on Earth: the camel. Let's start with a common misunderstanding. "
        "Many people believe that a camel's hump is full of water. In fact, the hump stores fat, which the camel can break down for energy when food is scarce. "
        "Storing fat in one place, rather than all over the body, also helps the camel to lose heat through the rest of its skin. "
        "Camels are extremely good at conserving water. Their kidneys produce very concentrated urine, and their droppings are so dry that they can be used as fuel almost immediately. "
        "Their nostrils also recover moisture from the air they breathe out. "
        "Perhaps their most remarkable ability is their tolerance of dehydration. A camel can lose around a quarter of its body weight in water and survive, a loss that would kill most mammals. "
        "When water is available again, a thirsty camel can drink more than a hundred litres in a few minutes. "
        "Camels also deal with heat in an unusual way. Instead of sweating to keep a constant temperature, a camel allows its body temperature to rise by several degrees during the day, "
        "and then loses the extra heat during the cool desert night. This saves a great deal of water that would otherwise be lost as sweat. "
        "Other features protect them from the desert environment. They have two rows of long eyelashes to keep out sand, and they can close their nostrils during sandstorms. "
        "Their wide, padded feet spread their weight, so they do not sink into soft ground. "
        "Finally, camels have been important to people for thousands of years. They were domesticated around three thousand years ago, "
        "and they still provide milk, meat, wool and transport in many dry regions. Camel milk contains more vitamin C than cow's milk, which makes it valuable where fresh fruit is rare.",
        [
            ("A camel's hump stores ________", ['fat'], 'the hump stores fat'),
            ('Dry droppings can be used as ________', ['fuel'], 'used as fuel'),
            ('Their nostrils recover ________ from their breath', ['moisture'], 'recover moisture'),
            ('A camel can lose about a ________ of its body weight in water', ['quarter'], 'around a quarter'),
            ('A thirsty camel can drink over a hundred ________ in minutes', ['litres', 'liters'], 'more than a hundred litres'),
            ('Its temperature rises instead of ________', ['sweating'], 'Instead of sweating'),
            ('Two rows of long ________ keep out sand', ['eyelashes'], 'long eyelashes'),
            ('Wide, padded ________ spread their weight', ['feet'], 'padded feet'),
            ('Camels were domesticated around three ________ years ago', ['thousand'], 'around three thousand years ago'),
            ('Camel milk is valuable where fresh ________ is rare', ['fruit'], 'fresh fruit is rare'),
        ]),
    't8': lecture(
        'You will hear part of a lecture about the history of the pencil.',
        'The history of the pencil',
        "Lecturer: Today I'm going to tell the story of an everyday object: the pencil. The key material in a pencil is graphite, a soft, dark form of carbon. "
        "In the sixteenth century, a large deposit of unusually pure graphite was discovered at Borrowdale, in the north of England. "
        "Local people found it useful for marking sheep, and it was soon being cut into sticks for writing and drawing. "
        "At first, nobody knew what the substance was. Because it left a grey mark rather like lead, people called it 'black lead', "
        "and that is why we still talk about the lead in a pencil, even though pencils have never contained lead. "
        "The Borrowdale graphite was so valuable that the mine was guarded, and at times it was opened for only a few weeks a year to prevent the price from falling. "
        "Sticks of graphite were brittle and dirty to hold, so at first they were wrapped in string, and later inserted into wooden holders. "
        "The modern pencil owes a great deal to a French engineer, Nicolas-Jacques Conte. In 1795, when war cut France off from English graphite, "
        "Conte developed a method of mixing powdered graphite with clay and baking the mixture in a kiln. This had two advantages. "
        "Poorer-quality graphite could be used, and by changing the proportion of clay, manufacturers could make pencils that were harder or softer. "
        "The grading still printed on pencils today, with letters such as H for hard and B for black, depends on this development. "
        "Later innovations included machines that mass-produced wooden casings and, in the nineteenth century, the attachment of an eraser to the end of the pencil. "
        "Pencils remain popular for good reasons. They are cheap, they work upside down and in freezing cold, and their marks can be removed when a mistake is made.",
        [
            ('Graphite is a form of ________', ['carbon'], 'a soft, dark form of carbon'),
            ('Local people first used it for marking ________', ['sheep'], 'marking sheep'),
            ('People called it black ________', ['lead'], "called it 'black lead'"),
            ('The Borrowdale mine was ________', ['guarded'], 'the mine was guarded'),
            ('Early sticks were wrapped in ________', ['string'], 'wrapped in string'),
            ('Conte was a ________ engineer', ['French'], 'a French engineer'),
            ('Powdered graphite was mixed with ________', ['clay'], 'powdered graphite with clay'),
            ('The mixture was baked in a ________', ['kiln'], 'baking the mixture in a kiln'),
            ('On a pencil, the letter H stands for ________', ['hard'], 'H for hard'),
            ('In the nineteenth century an ________ was attached to the end', ['eraser'], 'attachment of an eraser'),
        ]),
    't9': lecture(
        'You will hear part of a lecture about how rainbows form.',
        'How rainbows form',
        "Lecturer: Today's topic is the rainbow, and the physics that creates it. A rainbow appears when sunlight shines on raindrops in the air. "
        "To see one, you must have the sun behind you and the rain in front of you, which is why rainbows are most common in the early morning or late afternoon, when the sun is low in the sky. "
        "When a ray of sunlight enters a raindrop, it slows down and bends. This bending is called refraction. "
        "White sunlight is actually a mixture of colours, and each colour bends by a slightly different amount, so the light is split into a spectrum. "
        "This separation of colours is known as dispersion. "
        "The light then reflects off the back of the drop and bends again as it leaves. The result is that each colour leaves the drop at a particular angle. "
        "Red light comes out at about forty-two degrees to the direction of the incoming sunlight, and violet at about forty degrees. "
        "That is why a rainbow is always part of a circle, and why red is on the outside of the bow and violet on the inside. "
        "Sometimes a second, fainter rainbow appears outside the first. This happens when light reflects twice inside each drop. "
        "Because of the extra reflection, the order of the colours in this secondary bow is reversed. "
        "The sky between the two bows often looks darker, an effect named after the Greek philosopher Alexander of Aphrodisias, who described it nearly two thousand years ago. "
        "An interesting consequence of the geometry is that no two people see exactly the same rainbow, because each observer sees light from a different set of drops. "
        "And from an aircraft, where there is no ground in the way, a rainbow can sometimes be seen as a complete circle.",
        [
            ('To see a rainbow, the sun must be ________ you', ['behind'], 'the sun behind you'),
            ('Rainbows are common when the sun is ________ in the sky', ['low'], 'the sun is low'),
            ('The bending of light is called ________', ['refraction'], 'called refraction'),
            ('The separation of colours is called ________', ['dispersion'], 'known as dispersion'),
            ('Light reflects off the ________ of the drop', ['back'], 'the back of the drop'),
            ('Red light leaves at about forty-two ________', ['degrees'], 'forty-two degrees'),
            ('Red appears on the ________ of the bow', ['outside'], 'red is on the outside'),
            ('In a secondary bow, light reflects ________ in each drop', ['twice'], 'reflects twice'),
            ('The dark band is named after a Greek ________', ['philosopher'], 'Greek philosopher'),
            ('From an aircraft a rainbow may be a complete ________', ['circle'], 'a complete circle'),
        ]),
    't10': lecture(
        'You will hear part of a lecture about the first balloon flights.',
        'The first balloon flights',
        "Lecturer: Today I'd like to talk about the first human flights, which were made not in aeroplanes but in balloons. "
        "The story begins in France in 1783 with two brothers, Joseph and Etienne Montgolfier, who ran a paper-making business. "
        "They noticed that heated air could lift light objects, and began experimenting with bags made of paper and cloth. "
        "The principle is simple. Hot air is less dense than the cooler air around it, so a large bag of hot air rises, rather as a cork rises in water. "
        "The Montgolfiers believed at first that the smoke itself was responsible, and they burned straw and wool to produce as much smoke as possible. "
        "Their first public demonstration took place in June 1783. In September, at the royal palace of Versailles, they sent up a balloon carrying three animals: "
        "a sheep, a duck and a rooster. The animals landed safely after about eight minutes, which reassured people that it was possible to survive high above the ground. "
        "The first free flight with people on board followed in November 1783, when two men flew over Paris for about twenty-five minutes. "
        "Only ten days later, a rival balloon filled with hydrogen, a gas much lighter than hot air, made its own flight. "
        "Hydrogen balloons could stay in the air far longer, but hydrogen is highly flammable, which made them dangerous. "
        "For most of the following century, balloons were used mainly for science and entertainment. Scientists used them to study the upper atmosphere, "
        "and during the siege of Paris in 1870 balloons carried letters and passengers out of the surrounded city. "
        "Modern hot-air ballooning became popular only in the 1960s, when new materials such as nylon and burners fuelled by propane made balloons cheaper and safer. "
        "Today most balloon flights are for tourism, especially in places with calm and predictable winds.",
        [
            ('The Montgolfiers ran a ________-making business', ['paper'], 'a paper-making business'),
            ('Hot air is less ________ than cooler air', ['dense'], 'less dense'),
            ('At first they thought the ________ lifted the balloon', ['smoke'], 'the smoke itself was responsible'),
            ('Animals were sent up at the palace of ________', ['Versailles'], 'palace of Versailles'),
            ('The animals were a sheep, a duck and a ________', ['rooster'], 'and a rooster'),
            ('The first flight with people lasted about ________ minutes', ['25', 'twenty-five'], 'about twenty-five minutes'),
            ('A rival balloon was filled with ________', ['hydrogen'], 'filled with hydrogen'),
            ('Hydrogen is highly ________', ['flammable'], 'highly flammable'),
            ('In 1870 balloons carried ________ and passengers out of Paris', ['letters'], 'carried letters'),
            ('Modern burners are fuelled by ________', ['propane'], 'fuelled by propane'),
        ]),
}

W1 = 'ONE WORD ONLY'

# Replacement drills, by the title of the drill each replaces.
DRILLS = {
    'Lecture: Urban beekeeping': ('Lecture: Traffic signals',
        "Lecturer: The first traffic signal was installed outside the Houses of Parliament in London in 1868. It was operated by a police officer and lit by gas at night. Unfortunately, it exploded a few weeks later, injuring the officer. Electric traffic lights appeared in the United States in the early twentieth century.",
        [('The first signal stood outside the Houses of ____', ['Parliament']), ('At night it was lit by ____', ['gas']), ('A few weeks later it ____', ['exploded'])], W1),
    'Lecture: History of the bicycle': ('Lecture: How penguins keep warm',
        "Lecturer: Emperor penguins survive Antarctic winters in which temperatures fall below minus forty degrees. Their dense feathers trap a layer of air, and a thick layer of fat lies beneath the skin. In the worst storms, thousands of birds huddle together, taking turns to move from the cold edge of the group to the warm centre.",
        [('Their feathers trap a layer of ____', ['air']), ('Beneath the skin is a thick layer of ____', ['fat']), ('Birds take turns moving to the warm ____', ['centre', 'center'])], W1),
    'Lecture: Sleep and memory': ('Lecture: The zip fastener',
        "Lecturer: Early designs for a sliding fastener were unreliable and often came apart. In 1913 the engineer Gideon Sundback produced a much better design, using interlocking teeth. The name zipper was first used for rubber boots that had the fastener, and only much later did zips become common on clothing.",
        [('Early fasteners often came ____', ['apart']), ("Sundback's design used interlocking ____", ['teeth']), ('The name was first used for rubber ____', ['boots'])], W1),
    'Lecture: Coral reefs': ('Lecture: Earthworms and soil',
        "Lecturer: Earthworms are among the most important animals for healthy soil. As they burrow, they create tunnels that let air and water reach plant roots. They also eat dead leaves and produce casts that are rich in nutrients. Charles Darwin was so fascinated by them that he spent decades studying their behaviour.",
        [('Tunnels let air and ____ reach plant roots', ['water']), ('Worms produce casts rich in ____', ['nutrients']), ('Darwin studied them for ____', ['decades'])], W1),
    'Lecture: Salt in history': ('Lecture: The kiwi',
        "Lecturer: The kiwi, a flightless bird found only in New Zealand, is unusual in many ways. Its nostrils are at the tip of its long beak, which it uses to smell insects and worms underground. The female lays an extremely large egg, which can weigh about a fifth of her body weight.",
        [('Found only in New ____', ['Zealand']), ('Its nostrils are at the tip of its ____', ['beak']), ("An egg can weigh about a ____ of the female's weight", ['fifth'])], W1),
    'Lecture: Migration of birds': ('Lecture: The Venus flytrap',
        "Lecturer: The Venus flytrap grows naturally only in a small area of North and South Carolina in the United States. Its soil is poor in nutrients, so the plant catches insects to obtain nitrogen. The trap closes only when tiny trigger hairs are touched twice within about twenty seconds, which stops it closing on raindrops.",
        [('It grows naturally only in North and South ____', ['Carolina']), ('Insects provide the plant with ____', ['nitrogen']), ('The trigger hairs must be touched ____', ['twice'])], W1),
    'Lecture: Chocolate': ('Lecture: Morse code',
        "Lecturer: Morse code was developed in the 1830s and 1840s for use with the electric telegraph. Letters are represented by combinations of short and long signals, known as dots and dashes. The most common letters were given the shortest codes; the letter E, for example, is a single dot. In 1844 the first long-distance message was sent from Washington to Baltimore.",
        [('It was developed for the electric ____', ['telegraph']), ('The letter E is a single ____', ['dot']), ('The first long-distance message was sent to ____', ['Baltimore'])], W1),
    'Lecture: Glass-making': ('Lecture: The first photographs',
        "Lecturer: The oldest surviving photograph taken from nature was made in France in the 1820s by Nicephore Niepce. It shows the view from a window, and the exposure took at least several hours. In 1839 Louis Daguerre announced a process that produced sharp images in minutes rather than hours, and photography quickly became popular.",
        [('The oldest surviving photograph shows the view from a ____', ['window']), ('Its exposure took several ____', ['hours']), ('Daguerre announced his process in ____', ['1839'])], W1),
}
