package ptemock

import "github.com/prepyo/backend/internal/models"

// Part is one of the three parts of PTE Academic, in the order of the test.
type Part string

const (
	PartSpeakingWriting Part = "speaking_writing"
	PartReading         Part = "reading"
	PartListening       Part = "listening"
)

// PartOrder is the order the parts are taken in.
var PartOrder = []Part{PartSpeakingWriting, PartReading, PartListening}

// Clock says how an item is timed.
type Clock string

const (
	// ClockResponse items run their own preparation and recording windows in
	// the test player, as the speaking items do. The server keeps no deadline
	// for them: the windows are seconds long and the recording stops itself.
	ClockResponse Clock = "response"
	// ClockItem items have a countdown of their own: Summarize Written Text,
	// Write Essay and Summarize Spoken Text.
	ClockItem Clock = "item"
	// ClockSection items draw on their part's shared countdown, which starts
	// when the first of them is shown.
	ClockSection Clock = "section"
)

// Marking is how an item is marked.
type Marking string

const (
	// MarkKey items are marked against an answer key (scoring.Grade).
	MarkKey Marking = "key"
	// MarkSpoken items are rated from what the device heard.
	MarkSpoken Marking = "spoken"
	// MarkWritten items are rated from the text written.
	MarkWritten Marking = "written"
)

// Task is one PTE Academic item type as the test runs it.
//
// Timings are the published ones for the current (2025) test. Weights are how
// much a task counts towards each communicative skill score. Pearson publishes
// which skills each task contributes to but not the exact weights; these follow
// the widely used approximations the practice pages already show, and they are
// relative - a skill's score is the weighted mean of its tasks, so only their
// proportions matter.
type Task struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// TypeIDs are the bank's ids for this task, canonical first. Older seeds
	// used other ids for several tasks and still carry them.
	TypeIDs []string         `json:"-"`
	Skill   models.SkillType `json:"skill"`
	Part    Part             `json:"part"`
	Clock   Clock            `json:"clock"`
	Marking Marking          `json:"-"`
	// PrepSeconds and ResponseSeconds are a speaking item's windows.
	PrepSeconds     int `json:"prepSeconds,omitempty"`
	ResponseSeconds int `json:"responseSeconds,omitempty"`
	// AudioDelaySeconds is the "Beginning in N seconds" before a recording plays.
	AudioDelaySeconds int `json:"audioDelaySeconds,omitempty"`
	// Seconds is a ClockItem item's own time, or the share a ClockSection item
	// adds to its part's clock.
	Seconds int `json:"seconds,omitempty"`
	// EstimateSeconds is roughly how long a speaking item takes end to end,
	// recording included, for the durations shown before a test.
	EstimateSeconds int `json:"-"`
	// Instructions are the on-screen instructions, in the test's own words.
	Instructions string `json:"instructions"`
	// Weights is how much the task counts towards each skill score.
	Weights map[models.SkillType]float64 `json:"weights"`
	// MinWords and MaxWords are a written response's form limits. Outside
	// them PTE scores the response zero on every trait, so it is not rated.
	MinWords int `json:"minWords,omitempty"`
	MaxWords int `json:"maxWords,omitempty"`
}

const (
	skSpeaking  = models.SkillSpeaking
	skWriting   = models.SkillWriting
	skReading   = models.SkillReading
	skListening = models.SkillListening
)

// Tasks is every task a PTE mock can deal, by code.
var Tasks = map[string]Task{
	// ── Part 1: Speaking ───────────────────────────────────────────────────
	"RA": {
		Code: "RA", Name: "Read Aloud", TypeIDs: []string{"read-aloud", "pte-read-aloud"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		PrepSeconds: 35, ResponseSeconds: 40, EstimateSeconds: 80,
		Instructions: "Look at the text below. In 35 seconds, you must read this text aloud as naturally and clearly as possible. You have 40 seconds to read aloud.",
		Weights:      map[models.SkillType]float64{skSpeaking: 9, skReading: 6},
	},
	"RS": {
		Code: "RS", Name: "Repeat Sentence", TypeIDs: []string{"pte-repeat-sentence", "repeat-sentence"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		AudioDelaySeconds: 3, ResponseSeconds: 15, EstimateSeconds: 28,
		Instructions: "You will hear a sentence. Please repeat the sentence exactly as you hear it. You will hear the sentence only once.",
		Weights:      map[models.SkillType]float64{skSpeaking: 15, skListening: 10},
	},
	"DI": {
		Code: "DI", Name: "Describe Image", TypeIDs: []string{"pte-describe-image", "describe-image"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		PrepSeconds: 25, ResponseSeconds: 40, EstimateSeconds: 70,
		Instructions: "Look at the image below. In 25 seconds, please speak into the microphone and describe in detail what the image is showing. You will have 40 seconds to give your response.",
		Weights:      map[models.SkillType]float64{skSpeaking: 8},
	},
	"RL": {
		Code: "RL", Name: "Re-tell Lecture", TypeIDs: []string{"pte-retell-lecture", "retell-lecture", "re-tell-lecture"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		AudioDelaySeconds: 3, PrepSeconds: 10, ResponseSeconds: 40, EstimateSeconds: 140,
		Instructions: "You will hear a lecture. After listening to the lecture, in 10 seconds, please speak into the microphone and retell what you have just heard from the lecture in your own words. You will have 40 seconds to give your response.",
		Weights:      map[models.SkillType]float64{skSpeaking: 8, skListening: 5},
	},
	"ASQ": {
		Code: "ASQ", Name: "Answer Short Question", TypeIDs: []string{"pte-answer-short-question", "answer-short-question"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		AudioDelaySeconds: 3, ResponseSeconds: 10, EstimateSeconds: 22,
		Instructions: "You will hear a question. Please give a simple and short answer. Often just one or a few words is enough.",
		Weights:      map[models.SkillType]float64{skSpeaking: 3, skListening: 2},
	},
	"SGD": {
		Code: "SGD", Name: "Summarize Group Discussion", TypeIDs: []string{"pte-summarise-group-discussion", "pte-summarize-group-discussion"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		AudioDelaySeconds: 3, PrepSeconds: 10, ResponseSeconds: 120, EstimateSeconds: 250,
		Instructions: "You will hear three people having a discussion. When you hear the beep, summarize the whole discussion. You will have 10 seconds to prepare and 2 minutes to give your response.",
		Weights:      map[models.SkillType]float64{skSpeaking: 5, skListening: 4},
	},
	"RTS": {
		Code: "RTS", Name: "Respond to a Situation", TypeIDs: []string{"pte-respond-to-situation", "respond-to-situation"},
		Skill: skSpeaking, Part: PartSpeakingWriting, Clock: ClockResponse, Marking: MarkSpoken,
		AudioDelaySeconds: 3, PrepSeconds: 20, ResponseSeconds: 40, EstimateSeconds: 85,
		Instructions: "Listen to and read a description of a situation. You will have 20 seconds to think about your answer. Then you will hear a beep. You will have 40 seconds to answer the question. Please answer as completely as you can.",
		Weights:      map[models.SkillType]float64{skSpeaking: 5},
	},

	// ── Part 1: Writing ────────────────────────────────────────────────────
	"SWT": {
		Code: "SWT", Name: "Summarize Written Text", TypeIDs: []string{"summarize-written-text", "pte-summarize-written-text"},
		Skill: skWriting, Part: PartSpeakingWriting, Clock: ClockItem, Marking: MarkWritten,
		Seconds: 600, MinWords: 5, MaxWords: 75,
		Instructions: "Read the passage below and summarize it using one sentence. Type your response in the box at the bottom of the screen. You have 10 minutes to finish this task. Your response will be judged on the quality of your writing and on how well your response presents the key points in the passage.",
		Weights:      map[models.SkillType]float64{skWriting: 8, skReading: 5},
	},
	"WE": {
		Code: "WE", Name: "Write Essay", TypeIDs: []string{"pte-write-essay", "write-essay"},
		Skill: skWriting, Part: PartSpeakingWriting, Clock: ClockItem, Marking: MarkWritten,
		Seconds: 1200, MinWords: 120, MaxWords: 380,
		Instructions: "You will have 20 minutes to plan, write and revise an essay about the topic below. Your response will be judged on how well you develop a position, organize your ideas, present supporting details, and control the elements of standard written English. You should write 200-300 words.",
		Weights:      map[models.SkillType]float64{skWriting: 12},
	},

	// ── Part 2: Reading ────────────────────────────────────────────────────
	"RWFIB": {
		Code: "RWFIB", Name: "Reading & Writing: Fill in the Blanks", TypeIDs: []string{"fill-in-blanks-rw"},
		Skill: skReading, Part: PartReading, Clock: ClockSection, Marking: MarkKey, Seconds: 150,
		Instructions: "Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.",
		Weights:      map[models.SkillType]float64{skReading: 13, skWriting: 8},
	},
	"RMCMA": {
		Code: "RMCMA", Name: "Multiple Choice, Multiple Answers", TypeIDs: []string{"reading-mcq-multiple", "pte-reading-mcma"},
		Skill: skReading, Part: PartReading, Clock: ClockSection, Marking: MarkKey, Seconds: 150,
		Instructions: "Read the text and answer the question by selecting all the correct responses. More than one response is correct.",
		Weights:      map[models.SkillType]float64{skReading: 5},
	},
	"RO": {
		Code: "RO", Name: "Re-order Paragraphs", TypeIDs: []string{"reorder-paragraphs"},
		Skill: skReading, Part: PartReading, Clock: ClockSection, Marking: MarkKey, Seconds: 150,
		Instructions: "The text boxes below have been placed in a random order. Restore the original order.",
		Weights:      map[models.SkillType]float64{skReading: 7},
	},
	"RFIB": {
		Code: "RFIB", Name: "Reading: Fill in the Blanks", TypeIDs: []string{"fill-in-blanks-r"},
		Skill: skReading, Part: PartReading, Clock: ClockSection, Marking: MarkKey, Seconds: 120,
		Instructions: "In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.",
		Weights:      map[models.SkillType]float64{skReading: 11},
	},
	"RMCSA": {
		Code: "RMCSA", Name: "Multiple Choice, Single Answer", TypeIDs: []string{"reading-mcq-single", "pte-reading-mcsa"},
		Skill: skReading, Part: PartReading, Clock: ClockSection, Marking: MarkKey, Seconds: 90,
		Instructions: "Read the text and answer the multiple-choice question by selecting the correct response. Only one response is correct.",
		Weights:      map[models.SkillType]float64{skReading: 4},
	},

	// ── Part 3: Listening ──────────────────────────────────────────────────
	"SST": {
		Code: "SST", Name: "Summarize Spoken Text", TypeIDs: []string{"summarize-spoken-text"},
		Skill: skListening, Part: PartListening, Clock: ClockItem, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 600,
		Instructions: "You will hear a short lecture. Write a summary for a fellow student who was not present at the lecture. The summary should be 50-70 words. You have 10 minutes to finish this task. Your response will be judged on the quality of your writing and on how well your response presents the key points presented in the lecture.",
		Weights:      map[models.SkillType]float64{skListening: 12, skWriting: 6},
	},
	"LMCMA": {
		Code: "LMCMA", Name: "Multiple Choice, Multiple Answers", TypeIDs: []string{"pte-listening-mcma", "multiple-choice-multiple"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 120,
		Instructions: "Listen to the recording and answer the question by selecting all the correct responses. You will need to select more than one response.",
		Weights:      map[models.SkillType]float64{skListening: 5},
	},
	"LFIB": {
		Code: "LFIB", Name: "Fill in the Blanks", TypeIDs: []string{"pte-listening-fib", "fill-in-the-blanks"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 120,
		Instructions: "You will hear a recording. Type the missing words in each blank.",
		Weights:      map[models.SkillType]float64{skListening: 10, skWriting: 5},
	},
	"HCS": {
		Code: "HCS", Name: "Highlight Correct Summary", TypeIDs: []string{"pte-highlight-correct-summary", "highlight-correct-summary"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 120,
		Instructions: "You will hear a recording. Click on the paragraph that best relates to the recording.",
		Weights:      map[models.SkillType]float64{skListening: 4, skReading: 3},
	},
	"LMCSA": {
		Code: "LMCSA", Name: "Multiple Choice, Single Answer", TypeIDs: []string{"pte-listening-mcsa"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 90,
		Instructions: "Listen to the recording and answer the multiple-choice question by selecting the correct response. Only one response is correct.",
		Weights:      map[models.SkillType]float64{skListening: 4},
	},
	"SMW": {
		Code: "SMW", Name: "Select Missing Word", TypeIDs: []string{"pte-select-missing-word", "select-missing-word"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 90,
		Instructions: "You will hear a recording. At the end of the recording the last word or group of words has been replaced by a beep. Select the correct option to complete the recording.",
		Weights:      map[models.SkillType]float64{skListening: 3},
	},
	"HIW": {
		Code: "HIW", Name: "Highlight Incorrect Words", TypeIDs: []string{"pte-highlight-incorrect-word"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 10, Seconds: 120,
		Instructions: "You will hear a recording. Below is a transcription of the recording. Some words in the transcription differ from what the speaker said. Please click on the words that are different.",
		Weights:      map[models.SkillType]float64{skListening: 8, skReading: 5},
	},
	"WFD": {
		Code: "WFD", Name: "Write from Dictation", TypeIDs: []string{"write-from-dictation"},
		Skill: skListening, Part: PartListening, Clock: ClockSection, Marking: MarkKey,
		AudioDelaySeconds: 7, Seconds: 90,
		Instructions: "You will hear a sentence. Type the sentence in the box below exactly as you hear it. Write as much of the sentence as you can. You will hear the sentence only once.",
		Weights:      map[models.SkillType]float64{skListening: 14, skWriting: 10},
	},
}
