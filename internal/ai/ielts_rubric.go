package ai

import (
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// IELTS assessment guidance handed to the evaluator.
//
// This is Prepyo's own summary of the public IELTS band descriptors (Writing,
// updated May 2023; Speaking), written in our words rather than copied. Every
// rule here is traceable to a descriptor feature. It deliberately contains no
// population priors ("most learners score…"), no demands for native-like
// language and no rules the descriptors do not make, because each of those
// pulls a practice estimate away from what an IELTS rating describes.

// IELTS writing task kinds. Task 1 differs by module; Task 2 does not.
const (
	ieltsTaskAcademicTask1 = "academic-task1"
	ieltsTaskGeneralTask1  = "general-task1"
	ieltsTask2             = "task2"
)

// ieltsWritingTask classifies a writing question by its type id, falling back
// to the displayed task name for older rows.
func ieltsWritingTask(typeID, taskName string) string {
	id := strings.ToLower(typeID)
	switch {
	case strings.Contains(id, "task1-letter") || strings.Contains(id, "letter"):
		return ieltsTaskGeneralTask1
	case strings.Contains(id, "task1") || strings.Contains(id, "figure"):
		return ieltsTaskAcademicTask1
	case strings.Contains(id, "task2"):
		return ieltsTask2
	}
	name := strings.ToLower(taskName)
	switch {
	case strings.Contains(name, "letter"):
		return ieltsTaskGeneralTask1
	case strings.Contains(name, "figure") || strings.Contains(name, "task 1"):
		return ieltsTaskAcademicTask1
	}
	return ieltsTask2
}

// ieltsTask2Parts names the parts of the prompt a Task 2 essay type must
// address, which is what "all parts of the prompt" means for that question.
// Each rule follows from the descriptor: a part left out means the main parts
// are addressed incompletely. Unknown types get no extra line.
func ieltsTask2Parts(typeID string) string {
	id := strings.ToLower(typeID)
	switch {
	case strings.HasSuffix(id, "-opinion"):
		return "- This prompt asks how far the writer agrees or disagrees: the position may be full or partial agreement, but it must be clear and sustained to the conclusion.\n"
	case strings.HasSuffix(id, "-discussion"):
		return "- This prompt asks the writer to discuss both views and give their own opinion: both views must be discussed and an opinion given. Discussing only one view, or giving no opinion, addresses the main parts incompletely.\n"
	case strings.HasSuffix(id, "-advantages"):
		return "- This prompt concerns advantages and disadvantages: both must be discussed. If it asks whether one outweighs the other, the response must give a clear judgement; listing both sides without weighing them addresses the main parts incompletely.\n"
	case strings.HasSuffix(id, "-problem-solution"):
		return "- This prompt asks for problems and solutions: both must be addressed, and the solutions should respond to the problems identified. Covering only one part addresses the main parts incompletely.\n"
	case strings.HasSuffix(id, "-cause-effect"):
		return "- This prompt asks for causes (or reasons) and effects: both must be addressed. Covering only one part addresses the main parts incompletely.\n"
	case strings.HasSuffix(id, "-two-part"):
		return "- This prompt asks two direct questions: each must be answered. Where a question asks whether something is positive or negative, or asks for a view, the response must give a clear answer. Leaving a question unanswered addresses the main parts incompletely.\n"
	}
	return ""
}

// IELTSMinimumWords is the length each writing task asks for.
func IELTSMinimumWords(typeID, taskName string) int {
	if ieltsWritingTask(typeID, taskName) == ieltsTask2 {
		return 250
	}
	return 150
}

func ieltsWritingGuidance(req WritingRequest) string {
	task := ieltsWritingTask(req.TypeID, req.TaskName)
	first := "Task Response"
	if task != ieltsTask2 {
		first = "Task Achievement"
	}

	var b strings.Builder
	b.WriteString("You are an experienced IELTS Writing assessor giving practice feedback. You are not an official IELTS examiner, and your estimate is not an official score.\n\n")
	b.WriteString("Assess against the public IELTS Writing band descriptors:\n")
	b.WriteString(ieltsRatingRules)

	fmt.Fprintf(&b, "%s:\n", first)
	switch task {
	case ieltsTaskAcademicTask1:
		b.WriteString("- Academic Task 1: the response summarises the visual by selecting and reporting key features and making relevant comparisons.\n")
		b.WriteString("- A clear overview of the main trends, differences or stages is expected at band 7 and above; an attempted, relevant overview fits band 6.\n")
		b.WriteString("- Recounting details mechanically, or describing the figure with no data to support it, limits the rating to band 5 or below. Inaccurate or irrelevant data lowers it.\n")
		b.WriteString("- Use the figure data supplied to check accuracy. Do not reward opinions or explanations the figure does not support.\n")
	case ieltsTaskGeneralTask1:
		b.WriteString("- General Training Task 1 is a letter. All three bullet points must be presented and adequately covered; a missing bullet point limits the rating to band 4, and one covered inadequately limits it to band 5.\n")
		b.WriteString("- The purpose must be clear, and the tone must suit the recipient and stay consistent (formal, semi-formal or informal as the task implies). Inconsistent tone lowers the rating; inappropriate tone limits it.\n")
		b.WriteString("- Band 7 and above: bullet points clearly highlighted and extended, a clear purpose, and consistent appropriate tone.\n")
	default:
		b.WriteString("- Task 2: all parts of the prompt are addressed; a clear position is presented and developed through the response; main ideas are relevant, extended and supported.\n")
		b.WriteString("- Band 7 allows some over-generalisation or lack of focus in supporting ideas; addressing the main parts incompletely limits the rating to band 5.\n")
		b.WriteString("- A response that is barely related to the prompt or off-topic is limited to band 2 on this criterion.\n")
		if parts := ieltsTask2Parts(req.TypeID); parts != "" {
			b.WriteString(parts)
		}
	}
	b.WriteString(ieltsTaskLadder)

	b.WriteString("\nCoherence and Cohesion:\n")
	b.WriteString("- Logical organisation and clear progression, paragraphing, and cohesive devices including reference and substitution. Linking words are not a fault in themselves.\n")
	if task == ieltsTask2 {
		b.WriteString("- Missing or inadequate paragraphing limits the rating to band 5 in Task 2.\n")
	}
	b.WriteString(ieltsCoherenceLadder)

	b.WriteString("\nLexical Resource:\n")
	b.WriteString("- Range and precision of vocabulary, style and collocation, and spelling and word formation. Do not penalise ordinary words used precisely. Memorised phrases used inappropriately are a band 4 feature.\n")
	b.WriteString(ieltsWritingLexisLadder)

	b.WriteString("\nGrammatical Range and Accuracy:\n")
	b.WriteString("- Range of simple and complex structures, how often sentences are error-free, the effect of errors on communication, and punctuation.\n")
	b.WriteString(ieltsWritingGrammarLadder)

	b.WriteString("\nLength and copying:\n")
	// The count itself is in the learner's message, so this prompt stays the same for every answer and can be cached.
	b.WriteString("- The word count given with the response is Prepyo's: hyphenated words count as one, and wording copied from the task is not counted.\n")
	b.WriteString("- An answer that is too short may not give enough evidence for higher bands, and usually leaves part of the task unaddressed. Judge those consequences; do not apply a fixed deduction.\n")
	b.WriteString("- Wording copied from the task prompt must be discounted: do not credit it as the learner's own language.\n")
	b.WriteString("- Responses of 20 words or fewer are rated band 1 on every criterion; Prepyo applies that rule before you see a response.\n")

	fmt.Fprintf(&b, "\nReturn exactly four criteria: %s, Coherence and Cohesion, Lexical Resource, and Grammatical Range and Accuracy. Each has maxScore 9, a whole-band score, and feedback that cites evidence from the response. ", first)
	b.WriteString("estimatedScore.value is the mean of the four criteria rounded to the nearest half band: a mean ending in .25 rounds up to the next half band, and one ending in .75 rounds up to the next whole band. This is a task-level practice estimate, not a Writing band, which weights Task 2 twice as heavily as Task 1.\n")
	b.WriteString("If the response is too brief or too unclear to judge, set estimatedScore.value to null and say so in the summary.\n")
	b.WriteString("Every entry in sentenceFeedback must copy an exact sentence from the learner's text into `original` and give an improved version with a short explanation. Never quote or invent a sentence the learner did not write.\n")
	b.WriteString("Set estimatedScore.confidence to low, medium or high according to how much evidence the response gives you.\n")
	return b.String()
}

// ieltsSpeakingGuidance is the shared descriptor summary for both speaking
// paths. withPronunciation is false when only a transcript is available.
func ieltsSpeakingGuidance(taskName string, withPronunciation bool) string {
	var b strings.Builder
	b.WriteString("Assess against the public IELTS Speaking band descriptors:\n")
	b.WriteString(ieltsRatingRules)
	b.WriteString("- A real Speaking test rates average performance across all three parts; one recorded answer supports a practice estimate only.\n\n")

	b.WriteString("Fluency and Coherence:\n")
	b.WriteString("- Ability to keep going and produce extended answers, how coherent the answer is, discourse markers and connectives, and topic development.\n")
	b.WriteString("- What matters is why the speaker hesitates. Hesitation to plan content is normal even at bands 8-9. Some hesitation, repetition or self-correction to find language still fits band 7 when it does not affect coherence; relying on it to keep going fits band 5.\n")
	b.WriteString("- Judge the overall pattern. Never cap the rating because of one pause or a few filler words.\n")
	b.WriteString(ieltsFluencyLadder + "\n")

	b.WriteString("Lexical Resource:\n")
	b.WriteString("- Range and flexibility of vocabulary, precision, paraphrase, and less common or idiomatic items used appropriately.\n")
	b.WriteString(ieltsSpeakingLexisLadder + "\n")

	b.WriteString("Grammatical Range and Accuracy:\n")
	b.WriteString("- Range of structures and how often utterances are error-free. Spoken grammar is judged as speech, not as written text.\n")
	b.WriteString(ieltsSpeakingGrammarLadder + "\n")

	if withPronunciation {
		b.WriteString("Pronunciation:\n")
		b.WriteString("- How easily the speaker can be understood, how much effort the listener needs, and the range and control of stress, rhythm, intonation, chunking and connected speech.\n")
		b.WriteString("- An accent is not a fault. A first-language accent that does not reduce intelligibility must not lower the rating: the descriptors expect accent to have minimal effect at band 8 and no effect at band 9. Penalise only features that cause a lack of clarity or make the listener work harder.\n\n")
	}

	switch name := strings.ToLower(taskName); {
	case strings.Contains(name, "full test"):
		b.WriteString("Task: a complete Speaking test. Rate the candidate's average performance across all three parts. Part 1 answers on familiar topics are naturally short; the Part 2 long turn lasts up to two minutes after a minute's preparation; Part 3 asks the candidate to explain, justify, compare and speculate about more abstract issues linked to the Part 2 topic.\n")
	case strings.Contains(name, "part 2") || strings.Contains(name, "cue card"):
		b.WriteString("Task: Part 2 long turn. The candidate speaks for up to two minutes on the cue card after one minute of preparation. Covering the cue card points is expected but they need not be covered in order; a turn well under a minute gives little evidence of extended speech.\n")
	case strings.Contains(name, "part 3") || strings.Contains(name, "discussion"):
		b.WriteString("Task: Part 3 discussion. Questions are more abstract; the candidate is expected to explain, justify, compare and speculate.\n")
	case strings.Contains(name, "introduction") || strings.Contains(name, "part 1"):
		b.WriteString("Task: Part 1 questions on familiar topics. Answers are naturally shorter; relevant extension beyond a one-word reply is expected.\n")
	}
	b.WriteString("There are no right or wrong opinions in IELTS Speaking. Do not mark an answer down for the ideas or opinion it expresses.\n")
	return b.String()
}

// ieltsSpeakingCriteria names the criteria the reply must contain.
func ieltsSpeakingCriteria(withPronunciation bool) string {
	if withPronunciation {
		return "Fluency and Coherence, Lexical Resource, Grammatical Range and Accuracy, and Pronunciation"
	}
	return "Fluency and Coherence, Lexical Resource, and Grammatical Range and Accuracy"
}

// isIELTS reports whether an exam is IELTS, for prompt branches.
func isIELTS(exam models.ExamType) bool { return exam == models.ExamIELTS }

// The ladders below are Prepyo's own wording of the public IELTS band
// descriptors (Writing, May 2023; Speaking), bands 4-9. Without them a model
// knows the top of the scale well and places most clear, competent answers
// there; the lower rungs are what keep a band 6 a band 6.

const ieltsRatingRules = `- Rate each criterion on its own, in whole bands from 0 to 9, using the ladder under it. Criteria often differ by a band; do not let one strong criterion lift the others.
- Choose the band whose description fits the response as a whole. When it shows features of two adjacent bands, give the lower one unless it fully fits every positive feature of the higher one.
- Judge only the evidence in this response. Do not compare it with typical candidates or aim for an expected average. High bands do not require native-like language, but bands 8 and 9 need their own features to be clearly present, not merely an absence of serious errors.

`

const ieltsTaskLadder = `- 9: every part fully addressed; position fully developed; ideas relevant, fully extended and well supported.
- 8: every part sufficiently addressed; ideas relevant, well extended and supported; only occasional omissions or lapses.
- 7: every part addressed; a clear position or purpose throughout; main ideas extended and supported, though some over-generalise or lose focus.
- 6: every part addressed but some more fully than others; a relevant position or purpose, though conclusions may be unclear or repetitive; some main ideas insufficiently developed or unclear.
- 5: main parts addressed only partly; development is not always clear; main ideas limited and not sufficiently developed; some irrelevant detail or repetition.
- 4: the task is answered only minimally or tangentially; position or purpose unclear; main ideas hard to identify, repetitive or poorly supported.
`

const ieltsCoherenceLadder = `- 9: the message is followed effortlessly; cohesion rarely draws attention; paragraphing is skilful.
- 8: ideas are sequenced logically and followed with ease; cohesion is well managed with occasional lapses; paragraphing is sufficient and appropriate.
- 7: clear progression throughout with few lapses; a range of cohesive devices used flexibly, with some inaccuracy or over- or under-use; each paragraph has a clear central topic.
- 6: generally coherent with clear overall progression; cohesive devices used to some good effect, but linking within or between sentences is sometimes faulty or mechanical; referencing not always clear; paragraphs not always logical.
- 5: some organisation but not wholly logical, and overall progression may be missing; sentences not fluently linked; devices limited or overused, with some inaccuracy; paragraphing may be inadequate.
- 4: ideas not arranged coherently; no clear progression; basic devices that are inaccurate or repetitive.
`

const ieltsWritingLexisLadder = `- 9: full flexibility and precise use throughout; only rare minor slips.
- 8: a wide resource used fluently and flexibly for precise meaning; skilful use of less common and idiomatic items, with occasional inaccuracies in word choice or collocation; spelling errors rare.
- 7: enough range for some flexibility and precision; some less common and idiomatic items; awareness of style and collocation, though some choices are inappropriate; only a few spelling or word-formation errors.
- 6: vocabulary adequate for the task; meaning generally clear despite a rather restricted range or imprecision; less common words attempted with some inaccuracy; some spelling or word-formation errors that do not impede meaning.
- 5: limited but minimally adequate range; simple vocabulary may be accurate but allows little variation; frequent lapses in appropriacy; noticeable spelling or word-formation errors that may cause some difficulty.
- 4: basic vocabulary used repetitively or inappropriately; errors that cause strain.
`

const ieltsWritingGrammarLadder = `- 9: a wide range of structures with full flexibility and control; errors extremely rare.
- 8: a wide range used flexibly and accurately; the majority of sentences are error-free; only occasional, non-systematic errors.
- 7: a variety of complex structures with some flexibility and accuracy; error-free sentences are frequent; a few errors persist without impeding communication.
- 6: a mix of simple and complex sentences but limited flexibility; complex structures are less accurate than simple ones; errors occur but rarely impede communication.
- 5: a limited, repetitive range; complex sentences attempted but often faulty; errors may be frequent and cause some difficulty; punctuation may be faulty.
- 4: a very limited range; simple sentences predominate and subordinate clauses are rare; errors are frequent and may impede meaning.
`

const ieltsFluencyLadder = `- 9: fluent; hesitation is only to think about content; fully coherent with fully developed topics.
- 8: fluent with only occasional repetition or self-correction; hesitation is usually about content, rarely about finding words; topics developed coherently.
- 7: keeps going and gives long turns without noticeable effort; some hesitation, repetition or self-correction to find language, without loss of coherence; discourse markers used flexibly.
- 6: willing and able to give long turns, but coherence is sometimes lost through hesitation, repetition or self-correction; a range of discourse markers, not always appropriate.
- 5: usually keeps going but relies on repetition, self-correction or slow speech; hesitates to find basic words or grammar; overuses some markers; more complex speech breaks down.
- 4: cannot keep going without noticeable pauses; speaks slowly with frequent repetition; links simple sentences only.
`

const ieltsSpeakingLexisLadder = `- 9: total flexibility and precision on every topic; idiomatic language used naturally.
- 8: a wide resource used readily and flexibly; skilful less common and idiomatic items with occasional inaccuracy; paraphrases effectively.
- 7: discusses a variety of topics flexibly; some less common and idiomatic items, with some inappropriate choices; paraphrases effectively.
- 6: enough vocabulary to discuss topics at length; meaning clear despite some inappropriate choices; generally paraphrases successfully.
- 5: enough for familiar and some unfamiliar topics, but little flexibility; paraphrase attempted with mixed success.
- 4: enough for familiar topics only; frequent wrong word choices; rarely paraphrases.
`

const ieltsSpeakingGrammarLadder = `- 9: precise and accurate structures throughout, apart from slips any speaker makes.
- 8: a wide range used flexibly; the majority of sentences are error-free, with occasional non-systematic errors.
- 7: a range of structures used flexibly; error-free sentences are frequent; a few basic errors persist.
- 6: a mix of short and complex forms with limited flexibility; complex structures often contain errors, which rarely impede meaning.
- 5: basic sentence forms fairly well controlled; complex structures attempted but nearly always contain errors.
- 4: mostly short basic sentences; subordinate structures rare; errors frequent and may cause misunderstanding.
`
