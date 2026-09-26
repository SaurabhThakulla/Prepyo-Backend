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
	b.WriteString("- Rate each criterion on its own, in whole bands from 0 to 9. Award a band only when the response fully fits the positive features of that band; a negative feature of a lower band limits the rating.\n")
	b.WriteString("- Judge only the evidence in this response. Do not compare it with typical candidates or aim for an expected average.\n")
	b.WriteString("- High bands do not require native-like writing. Band 7 grammar has frequent error-free sentences with a few persistent errors; at band 8 the majority of sentences are error-free; at band 9 errors are extremely rare and minor.\n\n")

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

	b.WriteString("\nCoherence and Cohesion:\n")
	b.WriteString("- Logical organisation and clear progression, paragraphing, and cohesive devices including reference and substitution.\n")
	b.WriteString("- Linking words are not a fault in themselves. Cohesion that is faulty or mechanical because of misuse, overuse or omission fits band 6; flexible use with some inaccuracies or some over- or under-use still fits band 7.\n")
	if task == ieltsTask2 {
		b.WriteString("- Missing or inadequate paragraphing limits the rating to band 5 in Task 2.\n")
	}

	b.WriteString("\nLexical Resource:\n")
	b.WriteString("- Range and precision of vocabulary, awareness of style and collocation, and less common or idiomatic items used appropriately (some ability at band 7, skilful use at band 8).\n")
	b.WriteString("- Spelling and word-formation errors count by how often they occur and how much they affect communication.\n")
	b.WriteString("- Do not penalise ordinary words that are used precisely. Memorised phrases or formulaic language used inappropriately is a band 4 feature.\n")

	b.WriteString("\nGrammatical Range and Accuracy:\n")
	b.WriteString("- Range of simple and complex structures, how often sentences are error-free, the effect of errors on communication, and punctuation.\n")

	b.WriteString("\nLength and copying:\n")
	if req.WordCount > 0 {
		fmt.Fprintf(&b, "- Prepyo counted %d words (hyphenated words count as one; wording copied from the task is not counted). The task asks for at least %d.\n", req.WordCount, req.MinimumWords)
	}
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
	b.WriteString("- Rate each criterion on its own, in whole bands from 0 to 9. Award a band only when the performance fully fits the positive features of that band.\n")
	b.WriteString("- Judge only this response. Do not compare it with typical candidates or aim for an expected average. High bands do not require a native speaker's delivery.\n")
	b.WriteString("- A real Speaking test rates average performance across all three parts; one recorded answer supports a practice estimate only.\n\n")

	b.WriteString("Fluency and Coherence:\n")
	b.WriteString("- Ability to keep going and produce extended answers, how coherent the answer is, discourse markers and connectives, and topic development.\n")
	b.WriteString("- What matters is why the speaker hesitates. Hesitation to plan content is normal even at bands 8-9. Some hesitation, repetition or self-correction to find language still fits band 7 when it does not affect coherence; relying on it to keep going fits band 5.\n")
	b.WriteString("- Judge the overall pattern. Never cap the rating because of one pause or a few filler words.\n\n")

	b.WriteString("Lexical Resource:\n")
	b.WriteString("- Range and flexibility of vocabulary, precision, paraphrase, and less common or idiomatic items used appropriately (some ability at band 7, skilful use at band 8).\n\n")

	b.WriteString("Grammatical Range and Accuracy:\n")
	b.WriteString("- Range of structures, and how often utterances are error-free: frequent at band 7, the majority at band 8. Spoken grammar is judged as speech, not as written text.\n\n")

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
