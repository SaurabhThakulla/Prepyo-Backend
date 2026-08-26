package evaluations

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

var (
	// ErrEmptyResponse means there is nothing worth evaluating.
	ErrEmptyResponse = errors.New("response is empty")
	ErrLimitReached  = billing.ErrLimitReached
)

// minWordsToEvaluate stops a one-word submission from consuming an evaluation
// from the learner's daily allowance.
const minWordsToEvaluate = 20

type Service struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	exams     *exams.Repository
	billing   *billing.Service
	gateway   *ai.Gateway
	xp        *gamification.Service
}

func NewService(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	examRepo *exams.Repository,
	billingService *billing.Service,
	gateway *ai.Gateway,
	xp *gamification.Service,
) *Service {
	return &Service{
		db: db, repo: repo, questions: questionRepo, exams: examRepo,
		billing: billingService, gateway: gateway, xp: xp,
	}
}

type Request struct {
	User       models.User
	QuestionID string
	Text       string
}

type Outcome struct {
	Evaluation models.Evaluation
	// Reused is true when an identical submission had already been evaluated,
	// so no provider call was made and no allowance was spent.
	Reused       bool
	XPAwarded    int
	Streak       int
	Missions     []models.DailyMission
	Subscription models.SubscriptionState
}

// EvaluateWriting runs the full flow: allowance, deduplication, provider call,
// validation, persistence and rewards.
func (s *Service) EvaluateWriting(ctx context.Context, req Request) (Outcome, error) {
	text := strings.TrimSpace(req.Text)
	if len(strings.Fields(text)) < minWordsToEvaluate {
		return Outcome{}, ErrEmptyResponse
	}

	question, err := s.questions.ByID(ctx, req.QuestionID)
	if err != nil {
		return Outcome{}, err
	}

	// Same learner, same question, same words means the same feedback. Return
	// the stored result rather than paying for an identical call.
	fingerprint := fingerprintOf(req.User.ID, question.ID, text)
	if existing, err := s.repo.ByFingerprint(ctx, req.User.ID, fingerprint); err == nil {
		state, err := s.billing.State(ctx, s.db, req.User)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{Evaluation: existing, Reused: true, Subscription: state}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Outcome{}, err
	}

	state, err := s.billing.CheckEvaluationAllowance(ctx, s.db, req.User)
	if err != nil {
		return Outcome{Subscription: state}, err
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return Outcome{}, err
	}

	evaluation, usage, err := s.gateway.EvaluateWriting(ctx, ai.WritingRequest{
		Exam:        question.Exam,
		TaskName:    question.TypeName,
		Prompt:      question.Prompt,
		LearnerText: text,
		MinScore:    version.MinScore,
		MaxScore:    version.MaxScore,
	})
	if err != nil {
		return Outcome{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Outcome{}, fmt.Errorf("begin evaluation: %w", err)
	}
	defer tx.Rollback(ctx)

	saved, err := s.repo.Save(ctx, tx, SaveParams{
		UserID:      req.User.ID,
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
		return Outcome{}, err
	}

	awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    req.User.ID,
		Amount:    gamification.XPEvaluation,
		Reason:    "Writing evaluated: " + question.TypeName,
		SourceKey: "evaluation:" + saved.ID,
	})
	if err != nil {
		return Outcome{}, err
	}

	streak, err := s.xp.TouchStreak(ctx, tx, req.User)
	if err != nil {
		return Outcome{}, err
	}

	missions, err := s.xp.RecordActivity(ctx, tx, req.User, question.Skill)
	if err != nil {
		return Outcome{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Outcome{}, fmt.Errorf("commit evaluation: %w", err)
	}

	// Read the allowance again so the response shows usage including this call.
	state, err = s.billing.State(ctx, s.db, req.User)
	if err != nil {
		return Outcome{}, err
	}

	return Outcome{
		Evaluation:   saved,
		XPAwarded:    awarded,
		Streak:       streak,
		Missions:     missions,
		Subscription: state,
	}, nil
}

// fingerprintOf identifies a submission. Whitespace is normalised so reformatting
// the same essay counts as the same submission.
func fingerprintOf(userID, questionID, text string) []byte {
	normalised := strings.ToLower(strings.Join(strings.Fields(text), " "))
	sum := sha256.Sum256([]byte(userID + "\x00" + questionID + "\x00" + normalised))
	return sum[:]
}
