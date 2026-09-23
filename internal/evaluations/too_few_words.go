package evaluations

import (
	"fmt"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
)

const tooFewWordsVersion = "too-few-words.v1"

// tooFewWordsEvaluation marks a spoken answer the device barely heard.
//
// It used to be refused, which left a learner with a finished recording and no
// way forward. It is scored now, at the bottom of the exam's scale, because in
// the exam an answer with almost no words scores almost nothing too. The
// summary says plainly that this is about what was heard, since the usual
// cause is the microphone or the browser's speech recognition, not the learner.
func tooFewWordsEvaluation(question models.Question, version models.ExamVersion, transcript string, words int) models.Evaluation {
	score := version.MinScore

	heard := "No words could be made out in your recording"
	if words > 0 {
		heard = fmt.Sprintf("Only %d %s could be made out in your recording", words, pluralWord(words))
	}
	summary := heard + ", so this answer gets the lowest score. " +
		"If you did speak, the microphone or your browser's speech recognition may not have picked you up: " +
		"try Chrome or Edge, speak a little closer to the microphone, and record the task again."

	return models.Evaluation{
		Exam:              question.Exam,
		Skill:             models.SkillSpeaking,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &score,
		ScoreConfidence:   "low",
		Summary:           summary,
		Criteria: []models.EvaluationCriterion{{
			Name:     "Content",
			Score:    score,
			MaxScore: version.MaxScore,
			Feedback: "Too little was heard to judge what you said.",
		}},
		Strengths: []string{},
		Weaknesses: []string{
			"Speak for most of the recording time: an answer that is too short loses content and fluency marks.",
		},
		SentenceFeedback: []models.SentenceFeedback{},
		Transcript:       transcript,
	}
}

func pluralWord(n int) string {
	if n == 1 {
		return "word"
	}
	return "words"
}
