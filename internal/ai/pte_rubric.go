package ai

import (
	"fmt"
	"strings"
)

// PTE Academic assessment guidance handed to the evaluator.
//
// This is Prepyo's own summary of Pearson's published PTE Academic scoring
// criteria, written in our words rather than copied. Every rule here is
// traceable to Pearson's published scoring guides. It deliberately contains no
// population priors ("most learners score…"), no demands for native-like
// language and no artificial score caps or words-per-minute thresholds that
// Pearson does not state.

// PTE writing task kinds.
const (
	pteTaskWriteEssay           = "write-essay"
	pteTaskSummarizeWrittenText = "summarize-written-text"
	pteTaskSummarizeSpokenText  = "summarize-spoken-text"
)

// pteWritingTask classifies a writing question by its type id, falling back
// to the displayed task name for older rows.
func pteWritingTask(typeID, taskName string) string {
	id := strings.ToLower(typeID)
	switch {
	case strings.Contains(id, "summarize-written-text") || strings.Contains(id, "swt"):
		return pteTaskSummarizeWrittenText
	case strings.Contains(id, "summarize-spoken-text") || strings.Contains(id, "sst"):
		return pteTaskSummarizeSpokenText
	case strings.Contains(id, "write-essay") || strings.Contains(id, "essay"):
		return pteTaskWriteEssay
	}
	name := strings.ToLower(taskName)
	switch {
	case strings.Contains(name, "summarize written text"):
		return pteTaskSummarizeWrittenText
	case strings.Contains(name, "summarize spoken text"):
		return pteTaskSummarizeSpokenText
	case strings.Contains(name, "write essay") || strings.Contains(name, "essay"):
		return pteTaskWriteEssay
	}
	return pteTaskWriteEssay
}

// PTE speaking task kinds.
const (
	pteTaskReadAloud                = "read-aloud"
	pteTaskRepeatSentence           = "repeat-sentence"
	pteTaskDescribeImage            = "describe-image"
	pteTaskRetellLecture            = "retell-lecture"
	pteTaskAnswerShortQuestion      = "answer-short-question"
	pteTaskSummarizeGroupDiscussion = "summarize-group-discussion"
	pteTaskRespondToSituation       = "respond-to-situation"
)

func pteSpeakingTask(typeID, taskName string) string {
	id := strings.ToLower(typeID)
	switch {
	case strings.Contains(id, "read-aloud"):
		return pteTaskReadAloud
	case strings.Contains(id, "repeat-sentence"):
		return pteTaskRepeatSentence
	case strings.Contains(id, "describe-image"):
		return pteTaskDescribeImage
	case strings.Contains(id, "retell-lecture") || strings.Contains(id, "re-tell-lecture"):
		return pteTaskRetellLecture
	case strings.Contains(id, "answer-short-question"):
		return pteTaskAnswerShortQuestion
	case strings.Contains(id, "summarise-group-discussion") || strings.Contains(id, "summarize-group-discussion") || strings.Contains(id, "group-discussion"):
		return pteTaskSummarizeGroupDiscussion
	case strings.Contains(id, "respond-to-situation") || strings.Contains(id, "respond-to-a-situation"):
		return pteTaskRespondToSituation
	}
	name := strings.ToLower(taskName)
	switch {
	case strings.Contains(name, "read aloud"):
		return pteTaskReadAloud
	case strings.Contains(name, "repeat sentence"):
		return pteTaskRepeatSentence
	case strings.Contains(name, "describe image"):
		return pteTaskDescribeImage
	case strings.Contains(name, "retell") || strings.Contains(name, "re-tell"):
		return pteTaskRetellLecture
	case strings.Contains(name, "short question") || strings.Contains(name, "answer short question"):
		return pteTaskAnswerShortQuestion
	case strings.Contains(name, "group discussion") || strings.Contains(name, "summarise group discussion") || strings.Contains(name, "summarize group discussion"):
		return pteTaskSummarizeGroupDiscussion
	case strings.Contains(name, "respond to a situation") || strings.Contains(name, "respond to situation"):
		return pteTaskRespondToSituation
	}
	return ""
}

func pteWritingGuidance(req WritingRequest) string {
	task := pteWritingTask(req.TypeID, req.TaskName)

	var b strings.Builder
	b.WriteString("You are an experienced PTE Academic assessor giving practice feedback. Your estimate is not an official score.\n\n")
	b.WriteString("Assess against Pearson's published PTE Academic scoring criteria:\n")
	b.WriteString("- Rate each trait on the evidence in this response only. Weigh strengths and errors together.\n")
	b.WriteString("- Judge only the text the learner wrote. Never quote or invent a sentence they did not write.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from the learner's text into `original` and give an improved version with a short explanation.\n")
	b.WriteString("- If the response is too short or unclear to judge, set estimatedScore.value to null and state so in the summary.\n")
	b.WriteString("- Linking words and cohesive devices are normal in structured writing and are not a fault in themselves. Only misuse or heavy mechanical overuse counts against coherence.\n")
	b.WriteString("- Grammar and vocabulary errors count by how often they occur and how much they affect meaning.\n\n")

	traits := pteWritingTraits(task)
	switch task {
	case pteTaskSummarizeWrittenText:
		b.WriteString("Task: Summarize Written Text.\n")
		b.WriteString(pteSummaryContentLadder)
		b.WriteString("Form (0-1): 1 = one single complete sentence of 5 to 75 words; 0 = more than one sentence, not a complete sentence, fewer than 5 or more than 75 words, or written in capitals.\n")
		b.WriteString(pteGrammarVocabularyLadder)
	case pteTaskSummarizeSpokenText:
		b.WriteString("Task: Summarize Spoken Text.\n")
		b.WriteString(pteSummaryContentLadder)
		b.WriteString("Form (0-2): 2 = 50 to 70 words; 1 = 40-49 or 71-100 words; 0 = fewer than 40 or more than 100 words, written in capitals, no punctuation, or only bullet points or very short sentences.\n")
		b.WriteString(pteGrammarVocabularyLadder)
		b.WriteString(pteSpellingLadder)
	default:
		b.WriteString("Task: Write Essay.\n")
		b.WriteString(pteEssayLadder)
		b.WriteString(pteSpellingLadder)
	}

	fmt.Fprintf(&b, "\nReturn exactly these traits, each scored in whole points on Pearson's own scale, with maxScore set to that trait's maximum: %s.\n", pteTraitList(traits))
	b.WriteString("- Content and Form are gates: a response scored 0 on either earns no points at all. Score Content 0 for an off-topic answer or one built mainly from memorised material.\n")
	b.WriteString("- Choose the level whose description fits best. When a response sits between two levels, give the lower one unless it fully fits the higher one.\n")
	b.WriteString("- Prepyo converts the points to the 10-90 scale itself. Set estimatedScore.value to the total of your trait points, or null if the response cannot be judged.\n")
	b.WriteString("- Set estimatedScore.confidence to low, medium or high based on how much evidence the response gives you.\n")

	return b.String()
}

func pteSpeakingGuidance(req SpeakingRequest) string {
	task := pteSpeakingTask("", req.TaskName)

	var b strings.Builder
	b.WriteString("You are an experienced PTE Academic assessor giving practice feedback on one spoken response. Your estimate is not an official score.\n\n")
	b.WriteString("You are given one audio recording of the learner's spoken response.\n")
	b.WriteString("- First transcribe what you actually hear into `transcript`, verbatim. Include the learner's own errors, repetitions and false starts. Do not tidy them up.\n")
	b.WriteString("- If the recording is silent, unintelligible, or contains no speech, set `transcript` to \"\" and estimatedScore.value to null, and say so plainly in the summary.\n")
	b.WriteString("- Judge oral fluency and pronunciation from the audio itself, not from the transcript alone.\n")
	b.WriteString("- Judge only this recording. Never quote or invent words the learner did not say.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from `transcript` into `original` and give an improved version with a short explanation.\n\n")

	b.WriteString("Assess against Pearson's published PTE Academic speaking criteria:\n")
	b.WriteString("Oral Fluency (0-5):\n")
	b.WriteString("- Judge rhythm, phrasing and smoothness across the whole response. Do not apply words-per-minute cut-offs.\n")
	b.WriteString(pteFluencyLadder + "\n")

	b.WriteString("Pronunciation (0-5):\n")
	b.WriteString("- Judge intelligibility and how easily a listener understands the speech, including stress, rhythm, and intonation.\n")
	b.WriteString("- An accent is not a fault: regional or non-native accents that do not impair intelligibility must not lower the score. Lower the score only for sounds, stress or intonation that make words hard to understand.\n")
	b.WriteString(ptePronunciationLadder + "\n")

	if task == pteTaskRespondToSituation {
		b.WriteString("Appropriacy:\n")
		b.WriteString("- Judge whether the response suits the situation, the person being spoken to and the purpose (for example, a polite request to a tutor or a clear explanation to a colleague), covering what the situation asks for, in a suitable register.\n")
	} else {
		b.WriteString("Content:\n")
		b.WriteString("- Judge how accurately and fully the response addresses the prompt.\n")
		if strings.TrimSpace(req.ExpectedText) != "" {
			b.WriteString("- The learner was given a fixed text to say. Compare what you heard against it and treat omissions, substitutions and additions as content errors.\n")
		}
	}
	b.WriteString(pteSpeakingContentLadder(task) + "\n")
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))

	fmt.Fprintf(&b, "\nReturn exactly these traits, each in whole points on Pearson's own scale with maxScore set to that trait's maximum, and evidence-based feedback: %s.\n", pteTraitList(pteSpeakingTraits(task, true)))
	b.WriteString(ptePointsRules)
	b.WriteString("- Set estimatedScore.confidence to low, medium or high based on how much the recording gives you. A very short recording is low confidence.\n")

	return b.String()
}

func pteTranscriptGuidance(req SpokenTranscriptRequest) string {
	task := pteSpeakingTask("", req.TaskName)

	var b strings.Builder
	b.WriteString("You are an experienced PTE Academic assessor giving practice feedback. Your estimate is not an official score.\n\n")
	b.WriteString("You are given a transcript of the learner's spoken answer, produced by speech recognition on their device. You cannot hear the recording.\n")
	b.WriteString("- Speech recognition makes its own mistakes and often removes hesitations, fillers and false starts. A single odd word is more likely a recognition error than a learner error: ignore it unless the pattern repeats.\n")
	b.WriteString("- Measured timing, if given below, is evidence about pace and pauses. Use it to describe delivery in a natural way; do not apply words-per-minute cut-offs.\n")
	b.WriteString("- NEVER score or comment on pronunciation, accent, intonation or stress. You did not hear them. Do not include a pronunciation criterion.\n")
	if task == pteTaskRespondToSituation {
		b.WriteString("- Judge only what the transcript shows: appropriacy, structure, vocabulary and grammar.\n")
	} else {
		b.WriteString("- Judge only what the transcript shows: content, structure, vocabulary and grammar.\n")
	}
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from the transcript into `original` and give an improved version with a short explanation.\n")
	b.WriteString("- Say in the summary that this is a provisional estimate from a transcript, and that pronunciation was not assessed.\n\n")

	b.WriteString("Assess against Pearson's published PTE Academic criteria:\n")
	b.WriteString("Oral Fluency (0-5):\n")
	b.WriteString("- Judge rhythm, phrasing and smoothness from the transcript and timing. Do not apply words-per-minute cut-offs.\n")
	b.WriteString("- A transcript does not show rhythm. Fillers, repetitions and false starts it does show count; give 4 or 5 only when measured timing below shows continuous speech.\n")
	b.WriteString(pteFluencyLadder + "\n")

	if task == pteTaskRespondToSituation {
		b.WriteString("Appropriacy:\n")
		b.WriteString("- Judge whether the response suits the situation, the person being spoken to and the purpose (for example, a polite request to a tutor or a clear explanation to a colleague), covering what the situation asks for, in a suitable register.\n")
	} else {
		b.WriteString("Content:\n")
		b.WriteString("- Judge how accurately and fully the response addresses the prompt.\n")
		if strings.TrimSpace(req.ExpectedText) != "" {
			b.WriteString("- The learner was given a fixed text to say. Compare the transcript against it and treat omissions, substitutions and additions as content errors, allowing reasonable margin for speech recognition anomalies.\n")
		}
	}
	b.WriteString(pteSpeakingContentLadder(task) + "\n")
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))

	traits := pteSpeakingTraits(task, false)
	names := make([]string, len(traits))
	for i, t := range traits {
		names[i] = t.Name
	}
	fmt.Fprintf(&b, "\nReturn exactly %s: %s. Each has evidence-based feedback citing specific weaknesses. Do not return a Pronunciation criterion.\n", pluralCriteria(len(traits)), strings.Join(names, " and "))
	fmt.Fprintf(&b, "- Score each in whole points on Pearson's own scale, with maxScore set to that trait's maximum: %s.\n", pteTraitList(traits))
	b.WriteString(ptePointsRules)
	b.WriteString("- Set estimatedScore.confidence to low or medium. A transcript-only judgement is never high confidence.\n")

	return b.String()
}

// pluralCriteria words the count of transcript traits, which is one or two.
func pluralCriteria(n int) string {
	if n == 1 {
		return "one criterion"
	}
	return "two criteria"
}

// ptePointsRules closes every PTE speaking prompt: how the points are used.
const ptePointsRules = `- Content (or Appropriacy) is a gate: a response scored 0 on it earns no points at all. Score it 0 for an irrelevant answer or one built mainly from memorised material.
- Choose the level whose description fits best. When a response sits between two levels, give the lower one unless it fully fits the higher one.
- Prepyo converts the points to the 10-90 scale itself. Set estimatedScore.value to the total of your trait points, or null if the response cannot be judged.
`

// The ladders below are Prepyo's own wording of the trait descriptions in
// Pearson's PTE Academic Test Taker Score Guide (July 2025), section 7.

const pteEssayLadder = `Content (0-6):
- 6: addresses the prompt in depth, reframing the issue in the writer's own words; the argument is supported convincingly with sub-points and relevant examples throughout.
- 5: addresses the prompt adequately with a persuasive argument; main points are supported, with minor exceptions.
- 4: addresses the main point; the argument is generally convincing but lacks depth; support is present for some points and weak or missing for others.
- 3: relevant, but the main points are not addressed adequately; support is often missing or unsuitable.
- 2: a superficial attempt with generic statements or heavy reuse of the prompt's wording; few relevant details.
- 1: an incomplete understanding of the prompt; generic or repetitive phrasing; details disjointed.
- 0: does not deal with the prompt.
Form (0-2): 2 = 200 to 300 words; 1 = 120-199 or 301-380 words; 0 = under 120 or over 380 words, written in capitals, no punctuation, or only bullet points or very short sentences.
Development, Structure and Coherence (0-6):
- 6: an effective logical structure that flows smoothly; a clear argument developed systematically; well-developed introduction and conclusion; logical paragraphs; varied connectives used effectively.
- 5: a conventional, appropriate structure that is logical if not always smooth; a clear argument with some points developed; introduction, conclusion and paragraphs present; some gaps or abrupt transitions.
- 4: conventional structure mostly present but some elements missing; parts of the argument undeveloped or hard to follow; paragraphs not always effective or well linked.
- 3: traces of structure; simple points or disconnected ideas; a position that is not developed into an argument; paragraphs not used effectively.
- 2: little recognisable structure; disorganised ideas; only simple connectives such as "and", "but", "because".
- 1: disconnected ideas with no clear position.
- 0: no recognisable structure.
Grammar (0-2): 2 = consistent control of complex language, errors rare and hard to spot; 1 = a relatively high degree of control with no errors that cause misunderstanding; 0 = mainly simple structures and/or several basic mistakes.
General Linguistic Range (0-6):
- 6: varied expression used with ease and precision; no sign of limitation; errors rare and minor.
- 5: varied expression used appropriately; ideas clear with little sign of limits; occasional errors, meaning clear.
- 4: enough range for basic ideas; limits show on complex or abstract ideas (repetition, roundabout phrasing); errors cause occasional lapses in clarity.
- 3: a narrow range with simple expressions repeated; only simple ideas expressed; errors cause some disruption.
- 2: limited vocabulary and simple expressions dominate; some ideas unclear; frequent basic errors.
- 1: very limited language; ideas generally unclear.
- 0: meaning not accessible.
Vocabulary Range (0-2): 2 = good command of a broad range including idiomatic expressions; 1 = a good range for general academic topics, with gaps that cause roundabout phrasing or imprecision; 0 = mainly basic vocabulary, not enough for the topic.
`

const pteSummaryContentLadder = `Content (0-4):
- 4: a comprehensive summary; all main ideas identified and combined concisely in the writer's own words; extra detail left out; flows smoothly with varied connectives.
- 3: an adequate summary; main ideas present with minor omissions; paraphrase used but not consistently; ideas connected but not combined efficiently.
- 2: a partial summary; no separation of main points from detail; relies heavily on copied excerpts.
- 1: relevant but not a real summary; disconnected excerpts; main ideas missing or misrepresented.
- 0: too limited to show any understanding of the source.
`

const pteGrammarVocabularyLadder = `Grammar (0-2): 2 = correct grammatical structure; 1 = errors that do not hinder communication; 0 = defective structure that could hinder communication.
Vocabulary (0-2): 2 = appropriate word choice; 1 = word errors that do not hinder communication; 0 = defective word choice that could hinder communication.
`

const pteSpellingLadder = "Spelling (0-2): 2 = no spelling errors; 1 = one spelling error; 0 = more than one.\n"

const pteFluencyLadder = `- 5: smooth rhythm and phrasing; no hesitations, repetitions or false starts.
- 4: acceptable rhythm and phrasing; at most one hesitation, repetition or false start.
- 3: acceptable speed but may be uneven; more than one hesitation, but most words in continuous phrases; few repetitions or false starts; no long pauses.
- 2: uneven or staccato; at least one smooth three-word run; two or three hesitations, repetitions or false starts; at most one long pause.
- 1: irregular phrasing; many hesitations, repetitions or false starts; notably uneven; one or two long pauses.
- 0: slow and laboured; most words isolated; more than one long pause.`

const ptePronunciationLadder = `- 5: every sound easily understood; stress correct in words and sentences.
- 4: clear, with a few minor distortions that do not affect intelligibility.
- 3: most sounds correct; some consistent errors make a few words unclear.
- 2: some sounds consistently mispronounced; at least two thirds intelligible, the listener has to adjust.
- 1: many sounds mispronounced; about a third of the words hard to understand.
- 0: more than half of the speech unintelligible.`

// pteSpeakingContentLadder is the Content (or Appropriacy) scale for a task.
func pteSpeakingContentLadder(task string) string {
	switch task {
	case pteTaskAnswerShortQuestion:
		return "Content (0-1): 1 = an answer with the right meaning; 0 = a wrong answer or none."
	case pteTaskRepeatSentence:
		return "Content (0-3): 3 = every word in the right order; 2 = at least half the words in order; 1 = fewer than half; 0 = almost nothing from the sentence."
	case pteTaskDescribeImage:
		return `Content (0-6):
- 6: describes the image fully and accurately and expands on relationships between its features; a listener could picture all of it.
- 5: the main features described accurately, with some relationships not expanded; minor details missing.
- 4: some accurate simple descriptions and basic relationships, but not all main features covered.
- 3: mainly superficial description with minor inaccuracies; narrow, repeated expressions.
- 2: minimal, superficial description with some inaccuracies.
- 1: disconnected elements or a list of items without description.
- 0: relevant but too limited to score higher.`
	case pteTaskRespondToSituation:
		return `Appropriacy (0-6):
- 6: deals with the situation effectively; fully achieves the purpose with full regard to the context; appropriate, fully developed and persuasive.
- 5: deals with the situation adequately; achieves the purpose with some regard to the context; minor omissions.
- 4: partly achieves the purpose; some omissions or misunderstandings; expression adequate with some limits.
- 3: achieves only the most basic part of the purpose; limited range; some inappropriate elements or reuse of the prompt's wording.
- 2: some relevant content but the purpose is not achieved.
- 1: relevant words but shows little understanding of the situation, or an appropriate reply with no elaboration at all.
- 0: relevant but too limited to score higher.`
	}
	return `Content (0-6):
- 6: full understanding; main ideas paraphrased smoothly in own words, important points expanded with specifics (for a discussion: each speaker's view and how the views relate); well connected.
- 5: main ideas and some important details captured accurately in own words, with minor slips; generally smooth.
- 4: some main ideas and details, possibly a few inaccuracies or a focus on minor points; ideas not well connected.
- 3: some ideas, not fully accurate or not separating main points from detail; may repeat the source's wording; mostly unconnected.
- 2: mostly inaccurate or incomplete, missing the main ideas; relies on repeating the source's wording.
- 1: isolated words and phrases from the source without meaning.
- 0: relevant but too limited to score higher.`
}
