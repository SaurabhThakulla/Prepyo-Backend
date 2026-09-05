package evaluations

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	// ErrWrongSkill means the recording was submitted against a question that
	// is not a speaking task.
	ErrWrongSkill   = errors.New("question is not a speaking task")
	ErrLimitReached = billing.ErrLimitReached
)

// minWordsToEvaluate stops a one-word submission from consuming an evaluation
// from the learner's daily allowance.
const minWordsToEvaluate = 20

// minRecordingSeconds is the speaking equivalent: below this there is not
// enough speech to say anything useful about, and a learner should not spend an
// evaluation to be told so.
const minRecordingSeconds = 2

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

	// Cheap pre-check, deliberately unlocked and deliberately before the model
	// call: an over-quota learner should never cost a provider call. The check
	// that actually enforces the quota runs under a lock in persist().
	state, err := s.billing.CheckSubTestAllowance(ctx, s.db, req.User, "")
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

	return s.persist(ctx, persistParams{
		User:        req.User,
		Question:    question,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage:       usage,
		XPReason:    "Writing evaluated: " + question.TypeName,
	})
}

// SpeakingRequest is one recording to evaluate. Audio arrives already encoded
// in a format the provider accepts; see ai.AudioFormats.
type SpeakingRequest struct {
	User            models.User
	QuestionID      string
	Audio           []byte
	AudioFormat     string
	DurationSeconds int
}

// EvaluateSpeaking runs the same flow as EvaluateWriting for a recording.
//
// The recording itself is not stored. It is sent to the provider, transcribed,
// scored, and dropped; what persists is the transcript and the feedback. That
// keeps a learner's voice out of the database entirely, and the transcript is
// the part the feedback actually quotes.
func (s *Service) EvaluateSpeaking(ctx context.Context, req SpeakingRequest) (Outcome, error) {
	if len(req.Audio) == 0 || req.DurationSeconds < minRecordingSeconds {
		return Outcome{}, ErrEmptyResponse
	}

	question, err := s.questions.ByID(ctx, req.QuestionID)
	if err != nil {
		return Outcome{}, err
	}
	// A recording only means something against a speaking task. Without this a
	// reading question could be sent here and scored on criteria it has none of.
	if question.Skill != models.SkillSpeaking {
		return Outcome{}, ErrWrongSkill
	}

	// Two identical uploads are a double-click or a retry, not two attempts.
	// Distinct recordings of the same words hash differently, which is right:
	// they are different performances and deserve their own feedback.
	fingerprint := fingerprintOfAudio(req.User.ID, question.ID, req.Audio)
	if existing, err := s.repo.ByFingerprint(ctx, req.User.ID, fingerprint); err == nil {
		state, err := s.billing.State(ctx, s.db, req.User)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{Evaluation: existing, Reused: true, Subscription: state}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Outcome{}, err
	}

	// Cheap pre-check, deliberately unlocked and deliberately before the model
	// call: an over-quota learner should never cost a provider call. The check
	// that actually enforces the quota runs under a lock in persist().
	state, err := s.billing.CheckSubTestAllowance(ctx, s.db, req.User, "")
	if err != nil {
		return Outcome{Subscription: state}, err
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return Outcome{}, err
	}

	evaluation, usage, err := s.gateway.EvaluateSpeaking(ctx, ai.SpeakingRequest{
		Exam:     question.Exam,
		TaskName: question.TypeName,
		Prompt:   question.Prompt,
		// Read Aloud puts the words to say in the passage; Repeat Sentence puts
		// them in the transcript. Either way the model needs them to judge
		// content, and an open task such as a cue card has neither.
		ExpectedText:    firstNonEmpty(question.ContextPassage, question.AudioTranscript),
		AudioBase64:     base64.StdEncoding.EncodeToString(req.Audio),
		AudioFormat:     req.AudioFormat,
		DurationSeconds: req.DurationSeconds,
		MinScore:        version.MinScore,
		MaxScore:        version.MaxScore,
	})
	if err != nil {
		return Outcome{}, err
	}

	return s.persist(ctx, persistParams{
		User:        req.User,
		Question:    question,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage:       usage,
		XPReason:    "Speaking evaluated: " + question.TypeName,
	})
}

type persistParams struct {
	User        models.User
	Question    models.Question
	Fingerprint []byte
	Evaluation  models.Evaluation
	Usage       ai.Usage
	XPReason    string
}

// persist stores a validated evaluation and pays out for it, in one transaction
// so a learner cannot end up with XP for feedback that was never saved.
func (s *Service) persist(ctx context.Context, p persistParams) (Outcome, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Outcome{}, fmt.Errorf("begin evaluation: %w", err)
	}
	defer tx.Rollback(ctx)

	// The authoritative quota check. It runs here rather than beside the
	// pre-check because this is the first point at which the model call is
	// already done — the lock must never be held across provider inference, or
	// one submission would block every other submission by the same learner for
	// the several seconds the model takes.
	//
	// Two requests can therefore both clear the unlocked pre-check and both call
	// the model. Only one gets past this check, so the quota is never exceeded;
	// the cost of that race is one wasted provider call.
	if err := billing.LockUserForQuota(ctx, tx, p.User.ID); err != nil {
		return Outcome{}, err
	}
	if _, err := s.billing.CheckSubTestAllowance(ctx, tx, p.User, ""); err != nil {
		return Outcome{}, err
	}

	saved, err := s.repo.Save(ctx, tx, SaveParams{
		UserID:      p.User.ID,
		QuestionID:  p.Question.ID,
		Fingerprint: p.Fingerprint,
		Evaluation:  p.Evaluation,
		Usage: models.EvaluationUsage{
			Provider:         p.Usage.Provider,
			Model:            p.Usage.Model,
			PromptVersion:    p.Usage.PromptVersion,
			PromptTokens:     p.Usage.PromptTokens,
			CompletionTokens: p.Usage.CompletionTokens,
			LatencyMS:        p.Usage.LatencyMS,
		},
	})
	if err != nil {
		return Outcome{}, err
	}

	awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    p.User.ID,
		Amount:    gamification.XPEvaluation,
		Reason:    p.XPReason,
		SourceKey: "evaluation:" + saved.ID,
	})
	if err != nil {
		return Outcome{}, err
	}

	streak, err := s.xp.TouchStreak(ctx, tx, p.User)
	if err != nil {
		return Outcome{}, err
	}

	missions, err := s.xp.RecordActivity(ctx, tx, p.User, p.Question.Skill)
	if err != nil {
		return Outcome{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Outcome{}, fmt.Errorf("commit evaluation: %w", err)
	}

	// Read the allowance again so the response shows usage including this call.
	state, err := s.billing.State(ctx, s.db, p.User)
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// fingerprintOf identifies a submission. Whitespace is normalised so reformatting
// the same essay counts as the same submission.
func fingerprintOf(userID, questionID, text string) []byte {
	normalised := strings.ToLower(strings.Join(strings.Fields(text), " "))
	sum := sha256.Sum256([]byte(userID + "\x00" + questionID + "\x00" + normalised))
	return sum[:]
}

// fingerprintOfAudio is the same idea for a recording. There is nothing to
// normalise: only the identical bytes count as the identical submission.
func fingerprintOfAudio(userID, questionID string, audio []byte) []byte {
	h := sha256.New()
	h.Write([]byte(userID + "\x00" + questionID + "\x00"))
	h.Write(audio)
	return h.Sum(nil)
}
