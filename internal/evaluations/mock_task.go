package evaluations

import (
	"context"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
)

const ieltsNotAttemptedVersion = "ielts-not-attempted.v1"

// EvaluateMockTask rates one IELTS writing task sat inside a writing mock.
//
// The mock paid for both tasks when it started, so there is no per-task credit
// check, and it awards XP once for the sitting rather than per task. The
// evaluation is stored like any other, so it counts towards the learner's
// writing estimate and appears in their feedback history.
//
// The descriptors decide the two short cases without a model: a task left
// blank was not attempted (Band 0), and 20 words or fewer of the learner's own
// is Band 1 on every criterion.
func (s *Service) EvaluateMockTask(ctx context.Context, user models.User, question models.Question, text, sessionID string) (models.Evaluation, error) {
	if question.Skill != models.SkillWriting || question.Exam != models.ExamIELTS {
		return models.Evaluation{}, ErrWrongWritingSkill
	}
	text = strings.TrimSpace(text)
	fingerprint := fingerprintOf(user.ID, question.ID+":mock:"+sessionID+":"+ai.WritingPromptVersion, text)
	if existing, err := s.repo.ByFingerprint(ctx, user.ID, fingerprint); err == nil {
		return existing, nil
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return models.Evaluation{}, err
	}

	var evaluation models.Evaluation
	var usage ai.Usage
	words := IELTSWordCount(text, question.Prompt)
	switch {
	case text == "":
		evaluation = ieltsNotAttemptedEvaluation(question)
		usage = ai.Usage{Provider: "prepyo", Model: "ielts-length-rule", PromptVersion: ieltsNotAttemptedVersion}
	case words <= ieltsBand1WordLimit:
		evaluation = ieltsShortWritingEvaluation(question, words)
		usage = ai.Usage{Provider: "prepyo", Model: "ielts-length-rule", PromptVersion: ieltsShortWritingVersion}
	default:
		evaluation, usage, err = s.gateway.EvaluateWriting(ctx, ai.WritingRequest{
			Exam:           question.Exam,
			TypeID:         question.TypeID,
			TaskName:       question.TypeName,
			Prompt:         question.Prompt,
			FigureData:     question.FigureData,
			ContextPassage: question.ContextPassage,
			LearnerText:    text,
			WordCount:      words,
			MinimumWords:   ai.IELTSMinimumWords(question.TypeID, question.TypeName),
			MinScore:       version.MinScore,
			MaxScore:       version.MaxScore,
		})
		if err != nil {
			return models.Evaluation{}, err
		}
	}

	// Saving must not be cut short by the request deadline once the provider
	// has answered; see persist.
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), persistTimeout)
	defer cancel()
	saved, err := s.repo.Save(saveCtx, s.db, SaveParams{
		UserID:      user.ID,
		QuestionID:  question.ID,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage: models.EvaluationUsage{
			Provider:         usage.Provider,
			Model:            usage.Model,
			PromptVersion:    usage.PromptVersion,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			LatencyMS:        usage.LatencyMS,
		},
	})
	if err != nil {
		return models.Evaluation{}, fmt.Errorf("save mock task evaluation: %w", err)
	}
	return saved, nil
}

// ieltsNotAttemptedEvaluation is a task left blank. The public descriptors keep
// Band 0 for a candidate who did not attempt the question in any way.
func ieltsNotAttemptedEvaluation(question models.Question) models.Evaluation {
	band := 0.0
	first := "Task Response"
	if ai.IELTSMinimumWords(question.TypeID, question.TypeName) < 250 {
		first = "Task Achievement"
	}
	criteria := make([]models.EvaluationCriterion, 0, 4)
	for _, name := range []string{first, "Coherence and Cohesion", "Lexical Resource", "Grammatical Range and Accuracy"} {
		criteria = append(criteria, models.EvaluationCriterion{
			Name: name, Score: 0, MaxScore: 9,
			Feedback: "Nothing was written for this task, so it was not attempted.",
		})
	}
	return models.Evaluation{
		Exam:              models.ExamIELTS,
		Skill:             models.SkillWriting,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &band,
		ScoreConfidence:   "high",
		Summary:           "This task was left blank. IELTS keeps Band 0 for a question that was not attempted, and it counts towards the Writing band like any other task.",
		Criteria:          criteria,
		Strengths:         []string{},
		Weaknesses:        []string{"Leave time to write both tasks: Task 2 counts twice as much as Task 1."},
		SentenceFeedback:  []models.SentenceFeedback{},
	}
}
