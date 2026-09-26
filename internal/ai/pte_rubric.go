package ai

import (
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

	switch task {
	case pteTaskSummarizeWrittenText:
		b.WriteString("Task: Summarize Written Text.\n")
		b.WriteString("Return the published Pearson traits: Content, Form, Grammar, and Vocabulary.\n")
		b.WriteString("- Content: provides a good summary of the text by including the main points and key supporting aspects.\n")
		b.WriteString("- Form: exactly ONE single complete sentence between 5 and 75 words.\n")
		b.WriteString("- Grammar: correct grammatical structure for a single comprehensive sentence.\n")
		b.WriteString("- Vocabulary: appropriate word choice relevant to the passage.\n")
	case pteTaskSummarizeSpokenText:
		b.WriteString("Task: Summarize Spoken Text.\n")
		b.WriteString("Return the published Pearson traits: Content, Form, Grammar, Vocabulary, and Spelling.\n")
		b.WriteString("- Content: summarizes the main points and key supporting ideas.\n")
		b.WriteString("- Form: 50 to 70 words.\n")
		b.WriteString("- Grammar, Vocabulary, and Spelling: accurate sentence structures, appropriate academic vocabulary, and correct spelling.\n")
	default:
		b.WriteString("Task: Write Essay.\n")
		b.WriteString("Return the published Pearson traits: Content, Form, Development, Structure and Coherence, Grammar, General Linguistic Range, Vocabulary Range, and Spelling.\n")
		b.WriteString("- Content: addresses the prompt and topic adequately with relevant supporting details.\n")
		b.WriteString("- Form: standard essay length is 200 to 300 words. Responses between 120-199 words or 301-380 words receive partial credit; responses under 120 or over 380 words receive 0 for Form.\n")
		b.WriteString("- Development, Structure and Coherence: logical progression of ideas, clear paragraph structure, and appropriate transitional devices.\n")
		b.WriteString("- Grammar: control of complex syntax and grammatical accuracy.\n")
		b.WriteString("- General Linguistic Range: clarity and precision in expressing ideas.\n")
		b.WriteString("- Vocabulary Range: appropriate academic vocabulary and collocations.\n")
		b.WriteString("- Spelling: standard English spelling conventions.\n")
	}

	b.WriteString("\n- CRITICAL FOR PTE: Every criterion is scored on the 10-90 scale with maxScore 90.\n")
	b.WriteString("- estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 70.0, 79, 85) and should be consistent with the trait scores (roughly their overall level, not a separate guess).\n")
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
	b.WriteString("Oral Fluency:\n")
	b.WriteString("- Judge rhythm, phrasing and smoothness across the whole response.\n")
	b.WriteString("- Occasional hesitations, a self-correction or a natural pause are not failures and do not cap the score; only a consistent pattern of long or frequent hesitations, false starts or unnatural stops lowers the score.\n")
	b.WriteString("- Do not apply words-per-minute cut-offs.\n\n")

	b.WriteString("Pronunciation:\n")
	b.WriteString("- Judge intelligibility and how easily a listener understands the speech, including stress, rhythm, and intonation.\n")
	b.WriteString("- An accent is not a fault: regional or non-native accents that do not impair intelligibility must not lower the score. Lower the score only for sounds, stress or intonation that make words hard to understand.\n\n")

	if task == pteTaskRespondToSituation {
		b.WriteString("Appropriacy:\n")
		b.WriteString("- Judge whether the response suits the situation, the person being spoken to and the purpose (for example, a polite request to a tutor or a clear explanation to a colleague), covering what the situation asks for, in a suitable register.\n\n")
	} else {
		b.WriteString("Content:\n")
		b.WriteString("- Judge how accurately and fully the response addresses the prompt.\n")
		if strings.TrimSpace(req.ExpectedText) != "" {
			b.WriteString("- The learner was given a fixed text to say. Compare what you heard against it and treat omissions, substitutions and additions as content errors.\n")
		}
	}
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))

	switch task {
	case pteTaskAnswerShortQuestion:
		b.WriteString("\nReturn the published PTE criterion: Content. It has maxScore 90 and evidence-based feedback.\n")
	case pteTaskRespondToSituation:
		b.WriteString("\nReturn the published PTE speaking criteria: Appropriacy, Oral Fluency, and Pronunciation. Each has maxScore 90 and evidence-based feedback citing specific strengths and weaknesses.\n")
	default:
		b.WriteString("\nReturn the published PTE speaking criteria: Content, Oral Fluency, and Pronunciation. Each has maxScore 90 and evidence-based feedback citing specific strengths and weaknesses.\n")
	}

	b.WriteString("- CRITICAL FOR PTE: Every criterion is scored on the 10-90 scale with maxScore 90.\n")
	b.WriteString("- estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 70.0, 79, 85) and should be consistent with the trait scores (roughly their overall level, not a separate guess).\n")
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
	b.WriteString("Oral Fluency:\n")
	b.WriteString("- Judge rhythm, phrasing and smoothness from the transcript and timing.\n")
	b.WriteString("- Occasional hesitations, a self-correction or a natural pause are not failures and do not cap the score; only a consistent pattern of long or frequent hesitations, false starts or unnatural stops lowers the score.\n")
	b.WriteString("- Do not apply words-per-minute cut-offs.\n\n")

	if task == pteTaskRespondToSituation {
		b.WriteString("Appropriacy:\n")
		b.WriteString("- Judge whether the response suits the situation, the person being spoken to and the purpose (for example, a polite request to a tutor or a clear explanation to a colleague), covering what the situation asks for, in a suitable register.\n\n")
	} else {
		b.WriteString("Content:\n")
		b.WriteString("- Judge how accurately and fully the response addresses the prompt.\n")
		if strings.TrimSpace(req.ExpectedText) != "" {
			b.WriteString("- The learner was given a fixed text to say. Compare the transcript against it and treat omissions, substitutions and additions as content errors, allowing reasonable margin for speech recognition anomalies.\n")
		}
	}
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))

	switch task {
	case pteTaskAnswerShortQuestion:
		b.WriteString("\nReturn the criterion: Content. It has maxScore 90 and evidence-based feedback. Do not return a Pronunciation criterion.\n")
	case pteTaskRespondToSituation:
		b.WriteString("\nReturn exactly two criteria: Appropriacy and Oral Fluency. Each has maxScore 90 and evidence-based feedback citing specific weaknesses. Do not return a Pronunciation criterion.\n")
	default:
		b.WriteString("\nReturn exactly two criteria: Content and Oral Fluency. Each has maxScore 90 and evidence-based feedback citing specific weaknesses. Do not return a Pronunciation criterion.\n")
	}

	b.WriteString("- CRITICAL FOR PTE: Every criterion is scored on the 10-90 scale with maxScore 90.\n")
	b.WriteString("- estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 70.0, 79, 85) and should be consistent with the trait scores (roughly their overall level, not a separate guess).\n")
	b.WriteString("- Set estimatedScore.confidence to low or medium. A transcript-only judgement is never high confidence.\n")

	return b.String()
}
