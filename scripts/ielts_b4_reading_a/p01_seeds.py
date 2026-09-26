"""Passage 1: seed banks."""

from scripts.ielts_b4_reading_a._build import opts, paras

PASSAGE = {
    "id_slug": "seedbanks",
    "title": "Storing Seeds for the Future",
    "subtitle": "Why scientists freeze millions of seeds, and what a freezer cannot do",
    "topic": "Botany and agricultural conservation",
    "paragraphs": paras(
        "Deep inside a sandstone mountain on the Norwegian island of Spitsbergen, about 1,300 kilometres from "
        "the North Pole, there is a storeroom that holds more than a million packets of seeds. The Svalbard "
        "Global Seed Vault opened in 2008, and it is by far the best known of the roughly 1,700 seed banks in "
        "the world. Most of the others are much less dramatic: a row of freezers in a university basement, or "
        "a cold room at an agricultural research station. What they share is a simple aim. Over some ten "
        "thousand years, farmers have developed hundreds of thousands of local varieties of wheat, rice, beans "
        "and other crops, and many of these are disappearing as growers switch to a small number of "
        "high-yielding modern types. A seed bank tries to keep samples of this variety alive in case it is "
        "needed again.",

        "The reason for keeping old varieties is practical rather than sentimental. A traditional variety that "
        "gives a poor harvest may still carry a gene that protects it against a particular disease, pest or "
        "drought, and plant breeders can cross it with a modern variety in order to transfer that quality. "
        "There are well-documented cases. In the 1970s, a virus that stunted rice plants spread across large "
        "parts of Asia, and breeders screening thousands of samples found resistance in a single wild rice "
        "population from India. Crosses made with it were being grown on millions of hectares within a few "
        "years. As the climate changes, breeders expect to return to the collections more often, looking for "
        "plants that flower earlier, tolerate heat or survive with less water.",

        "The storage method itself relies on a fortunate property of most seeds. Many species produce what are "
        "called orthodox seeds, which survive being dried until they contain only a few per cent water. Once "
        "dried, they can be sealed in foil packets and kept at around minus 18 or minus 20 degrees Celsius, "
        "where their metabolism almost stops. Under these conditions the seeds of some cereals are expected to "
        "remain alive for decades, and in some cases for centuries. They cannot be kept indefinitely, however. "
        "Every few years, staff take a small sample from each collection and test how many of the seeds "
        "germinate. When the figure falls below a set level, often 85 per cent of the original rate, the "
        "collection has to be sown, grown to maturity and harvested again to produce fresh seed, a slow and "
        "expensive process known as regeneration.",

        "Not every plant cooperates. A significant minority of species, including cocoa, avocado, mango and "
        "many tropical trees, produce recalcitrant seeds, which die if they are dried. Freezing them while "
        "they are still moist is no solution, because ice crystals form inside the cells and destroy them. For "
        "these plants, other approaches have been needed. Some are kept as living collections in fields or "
        "orchards, which take up a great deal of space and are exposed to disease and storms. Others are stored "
        "as small pieces of tissue, such as the tip of a shoot or the embryo removed from a seed, which are "
        "treated with protective chemicals and plunged into liquid nitrogen at about minus 196 degrees. This "
        "method, called cryopreservation, is promising, but each species requires its own procedure, and "
        "finding a reliable one for a single plant can take years of trial and error.",

        "Not everyone is convinced that such collections are enough on their own. A seed held in a freezer is a "
        "snapshot of a population at the moment it was collected. In the field, by contrast, farmers' varieties "
        "keep changing, as growers save seed from the plants that did best in a hot year or survived a new "
        "pest. Seed stored in 1985 has not been exposed to the conditions of the decades since. This is a fair "
        "point, and it has led a number of agricultural scientists to argue that banks should work alongside "
        "what is called on-farm conservation, in which farmers are supported in continuing to grow and exchange "
        "local varieties. I think they are right to see the two as partners rather than rivals. They do, "
        "however, differ greatly in cost: a freezer is cheap to run, while supporting farming communities over "
        "the long term requires steady funding and the cooperation of local people.",

        "The Svalbard vault has attracted more public attention than any other seed bank, but it was designed "
        "as a back-up rather than a working collection. National and regional banks send duplicate samples "
        "there, which remain the property of the depositors and are not opened by the vault's staff. The site "
        "was chosen partly because the surrounding permafrost would keep the seeds cold even if the electricity "
        "failed. The value of the arrangement was shown in 2015, when a research centre that had lost access to "
        "its own collection became the first depositor to withdraw samples in order to rebuild it. Yet it would "
        "be a mistake to think that the vault alone can guarantee the future of our crops. Its seeds are only "
        "as good as the collections that sent them, and many of the world's smaller banks, particularly in "
        "poorer countries, struggle to pay for the regular testing and regeneration on which everything else "
        "depends.",
    ),
    "mcq_single": [
        {
            "prompt": "What does the writer say about most of the world's seed banks?",
            "options": opts(
                "They are found in cold, remote regions.",
                "They are far less striking than the Svalbard vault.",
                "They were set up after the Svalbard vault opened.",
                "They concentrate on wheat and rice.",
            ),
            "key": "B",
            "explanation": "Paragraph A says that most of the others are much less dramatic: a row of freezers in a university basement, or a cold room at a research station.",
        },
        {
            "prompt": "The example of the rice virus is used to show that",
            "options": opts(
                "wild plants are generally more productive than cultivated ones.",
                "viral diseases spread faster in Asia than elsewhere.",
                "breeders prefer to work with modern varieties.",
                "a single stored sample can prove extremely valuable.",
            ),
            "key": "D",
            "explanation": "Paragraph B describes how resistance was found in a single wild rice population, and crosses made with it were soon grown on millions of hectares.",
        },
        {
            "prompt": "What happens when the germination rate of a stored collection falls below a set level?",
            "options": opts(
                "The seeds are moved to a colder store.",
                "The seeds are sent to Svalbard for safe keeping.",
                "The collection is grown to produce new seed.",
                "The collection is thrown away and replaced.",
            ),
            "key": "C",
            "explanation": "Paragraph C states that the collection has to be sown, grown to maturity and harvested again to produce fresh seed, a process known as regeneration.",
        },
        {
            "prompt": "Why can recalcitrant seeds not simply be frozen while they are still moist?",
            "options": opts(
                "Ice forms inside their cells and destroys them.",
                "They lose their useful qualities when cold.",
                "They start to germinate too early.",
                "They take up too much space in a freezer.",
            ),
            "key": "A",
            "explanation": "Paragraph D says freezing them while moist is no solution, because ice crystals form inside the cells and destroy them.",
        },
        {
            "prompt": "According to the writer, what limits the value of the Svalbard vault?",
            "options": opts(
                "Its remote position makes it hard for depositors to reach.",
                "Its staff do not have permission to test the seeds.",
                "It depends on collections elsewhere that may lack funds.",
                "Its electricity supply is unreliable.",
            ),
            "key": "C",
            "explanation": "Paragraph F says its seeds are only as good as the collections that sent them, and many smaller banks struggle to pay for testing and regeneration.",
        },
    ],
    "mcq_multiple": [
        {
            "prompt": "Which TWO kinds of plant does the writer say breeders will look for as the climate changes?",
            "options": opts(
                "plants that flower earlier",
                "plants with larger seeds",
                "plants that resist frost",
                "plants that need more fertiliser",
                "plants that tolerate heat",
            ),
            "keys": ["A", "E"],
            "explanation": "Paragraph B mentions plants that flower earlier (A) and plants that tolerate heat (E). Larger seeds, frost and fertiliser are not mentioned.",
        },
        {
            "prompt": "Which TWO statements about orthodox seeds are made in the passage?",
            "options": opts(
                "They must be stored in liquid nitrogen.",
                "They can be dried until they hold very little water.",
                "They are found only in cereals.",
                "Their metabolism almost stops when they are frozen.",
                "They germinate more readily after long storage.",
            ),
            "keys": ["B", "D"],
            "explanation": "Paragraph C says orthodox seeds survive being dried until they contain only a few per cent water (B) and that when frozen their metabolism almost stops (D).",
        },
        {
            "prompt": "Which TWO disadvantages of keeping plants in fields or orchards are mentioned?",
            "options": opts(
                "They need to be watered constantly.",
                "They can only be used for tropical species.",
                "They occupy a lot of space.",
                "They produce seed of poor quality.",
                "They are exposed to disease and storms.",
            ),
            "keys": ["C", "E"],
            "explanation": "Paragraph D says living collections take up a great deal of space (C) and are exposed to disease and storms (E).",
        },
        {
            "prompt": "Which TWO points are made about on-farm conservation?",
            "options": opts(
                "It is cheaper to run than a freezer.",
                "It allows varieties to go on changing in response to conditions.",
                "It has replaced seed banks in many countries.",
                "It was first proposed in 1985.",
                "It needs long-term funding and local cooperation.",
            ),
            "keys": ["B", "E"],
            "explanation": "Paragraph E explains that varieties in the field keep changing (B), and that supporting farming communities requires steady funding and the cooperation of local people (E). A freezer, not on-farm conservation, is described as cheap to run.",
        },
        {
            "prompt": "Which TWO facts about the Svalbard vault are given?",
            "options": opts(
                "Depositors keep ownership of the samples they send.",
                "Its staff test the seeds at regular intervals.",
                "It holds the largest working collection in the world.",
                "Its location was chosen partly because of the permafrost.",
                "It was paid for by a group of poorer countries.",
            ),
            "keys": ["A", "D"],
            "explanation": "Paragraph F says the samples remain the property of the depositors (A) and that the site was chosen partly because the permafrost would keep the seeds cold (D). The samples are not opened by the staff, and the vault is a back-up rather than a working collection.",
        },
    ],
    "tfng": [
        {
            "statement": "Most of the world's seed banks are housed inside mountains.",
            "key": "FALSE",
            "explanation": "Paragraph A says most seed banks are much less dramatic, such as freezers in a university basement or a cold room at a research station.",
        },
        {
            "statement": "The rice that resisted the virus came from a wild population.",
            "key": "TRUE",
            "explanation": "Paragraph B says breeders found resistance in a single wild rice population from India.",
        },
        {
            "statement": "Frozen cereal seeds can in some cases stay alive for hundreds of years.",
            "key": "TRUE",
            "explanation": "Paragraph C says the seeds of some cereals are expected to remain alive for decades, and in some cases for centuries.",
        },
        {
            "statement": "Cryopreservation was first developed for cocoa plants.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph D names cocoa as a species with recalcitrant seeds but says nothing about which plant cryopreservation was first used for.",
        },
        {
            "statement": "Samples sent to Svalbard become the property of the vault.",
            "key": "FALSE",
            "explanation": "Paragraph F says the duplicate samples remain the property of the depositors.",
        },
    ],
    "ynng": [
        {
            "statement": "Old crop varieties are kept mainly because of their cultural importance.",
            "key": "NO",
            "explanation": "Paragraph B says the reason for keeping old varieties is practical rather than sentimental.",
        },
        {
            "statement": "Cryopreservation is likely to replace field collections within the next ten years.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph D calls cryopreservation promising but gives no view on whether or when it will replace field collections.",
        },
        {
            "statement": "The criticism that stored seed cannot adapt to new conditions is reasonable.",
            "key": "YES",
            "explanation": "Paragraph E says of this criticism: 'This is a fair point.'",
        },
        {
            "statement": "On-farm conservation should take the place of seed banks.",
            "key": "NO",
            "explanation": "Paragraph E says the writer thinks the scientists are right to see the two as partners rather than rivals.",
        },
        {
            "statement": "The Svalbard vault cannot by itself make the future of crop diversity secure.",
            "key": "YES",
            "explanation": "Paragraph F says it would be a mistake to think that the vault alone can guarantee the future of our crops.",
        },
    ],
    "matching": [
        {
            "prompt": "a reason why many local crop varieties are being lost",
            "key": "A",
            "evidence": "many of these are disappearing as growers switch to a small number of high-yielding modern types",
            "explanation": "Paragraph A explains that local varieties are disappearing as growers switch to a few modern types.",
        },
        {
            "prompt": "an example of a stored sample that solved a problem for farmers",
            "key": "B",
            "evidence": "breeders screening thousands of samples found resistance in a single wild rice population from India",
            "explanation": "Paragraph B describes the wild rice population that provided resistance to a virus.",
        },
        {
            "prompt": "the very low temperature used to preserve plant tissue",
            "key": "D",
            "evidence": "plunged into liquid nitrogen at about minus 196 degrees",
            "explanation": "Paragraph D gives the temperature of liquid nitrogen used in cryopreservation.",
        },
        {
            "prompt": "a comparison of the costs of two ways of conserving crops",
            "key": "E",
            "evidence": "a freezer is cheap to run, while supporting farming communities over the long term requires steady funding",
            "explanation": "Paragraph E contrasts the low cost of a freezer with the funding needed for on-farm conservation.",
        },
        {
            "prompt": "the first occasion on which samples were taken out of a back-up store",
            "key": "F",
            "evidence": "became the first depositor to withdraw samples in order to rebuild it",
            "explanation": "Paragraph F describes the research centre that withdrew samples from Svalbard in 2015.",
        },
    ],
    "completion": [
        {
            "prompt": "Farmers are abandoning local varieties in favour of a few modern types that are ________.",
            "answer": "high-yielding",
            "explanation": "Paragraph A says growers are switching to a small number of high-yielding modern types.",
        },
        {
            "prompt": "Seeds that survive being dried are known as ________ seeds.",
            "answer": "orthodox",
            "explanation": "Paragraph C says many species produce what are called orthodox seeds, which survive being dried.",
        },
        {
            "prompt": "Growing a stored collection again to obtain fresh seed is called ________.",
            "answer": "regeneration",
            "explanation": "Paragraph C calls this slow and expensive process regeneration.",
        },
        {
            "prompt": "Seeds that die if they are dried are described as ________.",
            "answer": "recalcitrant",
            "explanation": "Paragraph D says cocoa, avocado, mango and many tropical trees produce recalcitrant seeds, which die if they are dried.",
        },
        {
            "prompt": "Helping farmers to continue growing local varieties is known as on-farm ________.",
            "answer": "conservation",
            "explanation": "Paragraph E refers to what is called on-farm conservation.",
        },
    ],
}
