package evaluations

import (
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// shortAnswerScoringVersion is stamped on Answer Short Question evaluations marked by matching, not a model.
const shortAnswerScoringVersion = "short-answer-match.v1"

// matchesShortAnswer reports whether an item is marked by matching: Answer Short Question with an answer key.
func matchesShortAnswer(q models.Question) bool {
	return expectsShortAnswer(q.TypeID) && len(shortAnswerKey(q)) > 0
}

// shortAnswerKey is the item's accepted answers, or its model answer when it has none.
func shortAnswerKey(q models.Question) []string {
	if len(q.CorrectAnswers) > 0 {
		return q.CorrectAnswers
	}
	if answer := strings.TrimSpace(q.ModelAnswer); answer != "" {
		return []string{answer}
	}
	return nil
}

// shortAnswerEvaluation marks Answer Short Question on Pearson's one trait, Vocabulary (0-1), shown on the 10-90 scale.
func shortAnswerEvaluation(question models.Question, version models.ExamVersion, transcript string) models.Evaluation {
	transcript = strings.Join(strings.Fields(strings.ReplaceAll(transcript, scoring.PauseMark, " ")), " ")
	key := shortAnswerKey(question)
	// The key may list many variants; the first is the one shown.
	accepted := ""
	if len(key) > 0 {
		accepted = key[0]
	}

	score := version.MinScore
	summary, feedback := "Your answer did not match the expected answer: "+accepted+".", "Expected: "+accepted+"."
	strengths, weaknesses := []string{}, []string{"Listen for the key word in the question and answer with the word it asks for."}
	switch {
	case scoring.MatchSpokenAnswer(key, transcript):
		score = version.MaxScore
		summary, feedback = "Correct: your answer matches "+accepted+".", "Right answer."
		strengths, weaknesses = []string{"You gave the right answer"}, []string{}
	case transcript == "":
		summary = "No answer could be made out in your recording, so it scores the minimum. The expected answer was " + accepted + "."
	}

	return models.Evaluation{
		Exam:              question.Exam,
		Skill:             models.SkillSpeaking,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &score,
		ScoreConfidence:   "medium",
		Summary:           summary,
		Criteria:          []models.EvaluationCriterion{{Name: "Vocabulary", Score: score, MaxScore: version.MaxScore, Feedback: feedback}},
		Strengths:         strengths,
		Weaknesses:        weaknesses,
		SentenceFeedback:  []models.SentenceFeedback{},
		Transcript:        transcript,
	}
}
