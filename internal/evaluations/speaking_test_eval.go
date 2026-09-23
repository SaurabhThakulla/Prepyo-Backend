package evaluations

import (
	"context"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
)

// EvaluateSpeakingTest rates a whole Speaking mock and stores the result, so it
// counts towards the learner's speaking estimate like any other evaluation. It
// is tied to no single question. The mock paid for itself when it started.
func (s *Service) EvaluateSpeakingTest(ctx context.Context, user models.User, sessionID string, turns []ai.SpeakingTestTurn) (models.Evaluation, error) {
	var answers strings.Builder
	for _, turn := range turns {
		answers.WriteString(turn.Question + "\x00" + turn.Answer + "\x00")
	}
	fingerprint := fingerprintOf(user.ID, "speaking-mock:"+sessionID+":"+ai.SpeakingTestPromptVersion, answers.String())
	if existing, err := s.repo.ByFingerprint(ctx, user.ID, fingerprint); err == nil {
		return existing, nil
	}

	version, err := s.exams.Current(ctx, models.ExamIELTS)
	if err != nil {
		return models.Evaluation{}, err
	}
	evaluation, usage, err := s.gateway.EvaluateSpeakingTest(ctx, ai.SpeakingTestRequest{
		Turns:    turns,
		MinScore: version.MinScore,
		MaxScore: version.MaxScore,
	})
	if err != nil {
		return models.Evaluation{}, err
	}

	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), persistTimeout)
	defer cancel()
	saved, err := s.repo.Save(saveCtx, s.db, SaveParams{
		UserID:      user.ID,
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
		return models.Evaluation{}, fmt.Errorf("save speaking test evaluation: %w", err)
	}
	return saved, nil
}
