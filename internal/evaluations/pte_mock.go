package evaluations

import (
	"context"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// EvaluatePTEMockWriting rates one PTE writing item sat inside a PTE mock.
//
// Like EvaluateMockTask, the mock paid for its items when it started, so there
// is no per-item credit check, and XP is awarded once for the sitting rather
// than per item. The evaluation is stored like any other, so it appears in the
// learner's feedback history. The PTE form rules (word limits) are applied by
// the mock before it asks; this rates what is left.
func (s *Service) EvaluatePTEMockWriting(ctx context.Context, userID string, question models.Question, text, sessionID string) (models.Evaluation, error) {
	if question.Skill != models.SkillWriting {
		return models.Evaluation{}, ErrWrongWritingSkill
	}
	text = strings.TrimSpace(text)
	fingerprint := fingerprintOf(userID, question.ID+":pte-mock:"+sessionID+":"+ai.WritingPromptVersion, text)
	if existing, err := s.repo.ByFingerprint(ctx, userID, fingerprint); err == nil {
		return existing, nil
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return models.Evaluation{}, err
	}
	evaluation, usage, err := s.gateway.EvaluateWriting(ctx, ai.WritingRequest{
		Exam:           models.ExamPTE,
		TypeID:         question.TypeID,
		TaskName:       question.TypeName,
		Prompt:         question.Prompt,
		FigureData:     question.FigureData,
		ContextPassage: question.ContextPassage,
		LearnerText:    text,
		WordCount:      len(strings.Fields(text)),
		MinScore:       version.MinScore,
		MaxScore:       version.MaxScore,
	})
	if err != nil {
		return models.Evaluation{}, err
	}
	return s.saveMockEvaluation(ctx, userID, question.ID, fingerprint, evaluation, usage)
}

// EvaluatePTEMockSpeaking rates one PTE speaking item sat inside a PTE mock,
// from the transcript the learner's device produced, by the same rules as
// practice (EvaluateSpeakingTranscript): Read Aloud and Repeat Sentence are
// aligned word for word, an answer the device barely heard is marked at the
// bottom of the scale, and the rest are rated by the model.
func (s *Service) EvaluatePTEMockSpeaking(ctx context.Context, userID string, question models.Question,
	transcript string, durationSeconds int, delivery scoring.Delivery, sessionID string) (models.Evaluation, error) {
	if question.Skill != models.SkillSpeaking {
		return models.Evaluation{}, ErrWrongSkill
	}
	transcript = strings.TrimSpace(transcript)
	words := len(strings.Fields(transcript))
	fingerprint := fingerprintOf(userID, question.ID+":pte-mock:"+sessionID+":"+ai.SpeakingTranscriptPromptVersion, transcript)
	if existing, err := s.repo.ByFingerprint(ctx, userID, fingerprint); err == nil {
		return existing, nil
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return models.Evaluation{}, err
	}

	material := speakingMaterialOf(question)
	var evaluation models.Evaluation
	var usage ai.Usage
	switch {
	case strings.TrimSpace(material.expected) != "":
		evaluation = verbatimEvaluation(question, version, material.expected, transcript, delivery)
		usage = ai.Usage{Provider: "prepyo", Model: "verbatim-alignment", PromptVersion: verbatimScoringVersion}
	case words < minTranscriptWords && !(expectsShortAnswer(question.TypeID) && words > 0):
		evaluation = tooFewWordsEvaluation(question, version, transcript, words)
		usage = ai.Usage{Provider: "prepyo", Model: "too-few-words", PromptVersion: tooFewWordsVersion}
	default:
		evaluation, usage, err = s.gateway.EvaluateSpokenTranscript(ctx, ai.SpokenTranscriptRequest{
			Exam:            models.ExamPTE,
			TaskName:        question.TypeName,
			Prompt:          question.Prompt,
			ExpectedText:    material.expected,
			SourceText:      material.source,
			ReferenceAnswer: material.reference,
			Transcript:      transcript,
			DurationSeconds: durationSeconds,
			WordCount:       words,
			Delivery:        delivery.Summary(),
			MinScore:        version.MinScore,
			MaxScore:        version.MaxScore,
		})
		if err != nil {
			return models.Evaluation{}, err
		}
	}
	return s.saveMockEvaluation(ctx, userID, question.ID, fingerprint, evaluation, usage)
}

// saveMockEvaluation stores a mock item's rating. Saving must not be cut short
// by the caller's deadline once the provider has answered; see persist.
func (s *Service) saveMockEvaluation(ctx context.Context, userID, questionID string, fingerprint []byte,
	evaluation models.Evaluation, usage ai.Usage) (models.Evaluation, error) {
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), persistTimeout)
	defer cancel()
	saved, err := s.repo.Save(saveCtx, s.db, SaveParams{
		UserID:      userID,
		QuestionID:  questionID,
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
		return models.Evaluation{}, fmt.Errorf("save pte mock evaluation: %w", err)
	}
	return saved, nil
}

// expectsShortAnswer reports whether a spoken task is answered in a word or
// two. Answer Short Question is: "often just one or a few words is enough", so
// the rule that marks an answer of under minTranscriptWords at the bottom of
// the scale would mark every right answer wrong. Its answers are rated.
func expectsShortAnswer(typeID string) bool {
	switch typeID {
	case "pte-answer-short-question", "answer-short-question":
		return true
	}
	return false
}
