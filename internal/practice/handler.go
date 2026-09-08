package practice

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mistakes"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/pkg/httpx"
)

type ReferralsService interface {
	QualifyReferral(ctx context.Context, refereeID string) error
}

type Handler struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	mistakes  *mistakes.Repository
	exams     *exams.Repository
	xp        *gamification.Service
	billing   *billing.Service
	referrals ReferralsService
	log       *slog.Logger
}

func NewHandler(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	mistakeRepo *mistakes.Repository,
	examRepo *exams.Repository,
	xp *gamification.Service,
	billingService *billing.Service,
	referrals ReferralsService,
	log *slog.Logger,
) *Handler {
	return &Handler{
		db:        db,
		repo:      repo,
		questions: questionRepo,
		mistakes:  mistakeRepo,
		exams:     examRepo,
		xp:        xp,
		billing:   billingService,
		referrals: referrals,
		log:       log,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/attempts", h.listAttempts)
	r.Post("/attempts", h.submit)
	return r
}

func (h *Handler) listAttempts(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	list, total, err := h.repo.List(r.Context(), ListParams{UserID: user.ID, Limit: page.Limit, Offset: page.Offset})
	if err != nil {
		httpx.Internal(w, h.log, "practice.listAttempts", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"attempts":   list,
		"pagination": page.Meta(total),
	})
}

// submit grades one answer and records the result.
func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var sub models.AnswerSubmission
	if !httpx.Decode(w, r, &sub, h.log, "practice.submit") {
		return
	}
	if strings.TrimSpace(sub.QuestionID) == "" {
		httpx.ValidationError(w, map[string]string{"questionId": "Required."})
		return
	}

	user := reqctx.MustUser(r.Context())
	ctx := r.Context()

	question, err := h.questions.ByID(ctx, sub.QuestionID)
	if err != nil {
		if errors.Is(err, questions.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That question does not exist.")
			return
		}
		httpx.Internal(w, h.log, "practice.submit.question", err)
		return
	}

	if !scoring.Deterministic(question.Skill) {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"Speaking and writing tasks are scored by evaluation. Submit them to /api/v1/evaluations.")
		return
	}

	exam, ok := examFor(sub.Exam, user)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}

	if !question.SupportsExam(exam) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeBadRequest,
			"That question is not part of the "+string(exam)+" syllabus.")
		return
	}

	version, err := h.exams.Current(ctx, exam)
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.examVersion", err)
		return
	}

	result, ok := scoring.Grade(question, sub)
	if !ok {
		h.log.Error("no grader for question type", "questionId", question.ID, "typeId", question.TypeID)
		httpx.Error(w, http.StatusNotImplemented, httpx.CodeInternal,
			"This task type cannot be scored yet. Your answer was not recorded.")
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.begin", err)
		return
	}
	defer tx.Rollback(ctx)

	// Quota check under user lock.
	if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
		httpx.Internal(w, h.log, "practice.submit.lock", err)
		return
	}
	if _, err := h.billing.CheckSubTestAllowance(ctx, tx, user, billing.SubTestKeyForQuestion(question)); err != nil {
		if errors.Is(err, billing.ErrLimitReached) {
			httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached,
				"You have used all of today's practice sub-tests. They reset at midnight.")
			return
		}
		httpx.Internal(w, h.log, "practice.submit.allowance", err)
		return
	}

	attempt, err := h.repo.Save(ctx, tx, SaveParams{
		UserID:             user.ID,
		QuestionID:         question.ID,
		Exam:               exam,
		ExamVersionID:      version.ID,
		IsCorrect:          result.IsCorrect,
		Score:              result.Score,
		MaxScore:           result.MaxScore,
		AccuracyPercentage: result.AccuracyPercentage,
		UserResponse:       result.UserDisplay,
		Feedback:           result.Feedback,
		TimeSpentSeconds:   sub.TimeSpentSeconds,
	})
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.save", err)
		return
	}

	if !result.IsCorrect {
		if err := h.mistakes.Record(ctx, tx, mistakes.RecordParams{
			UserID:          user.ID,
			QuestionID:      question.ID,
			Exam:            exam,
			ErrorTag:        result.ErrorTag,
			UserResponse:    result.UserDisplay,
			CorrectResponse: result.CorrectDisplay,
			Explanation:     question.Explanation,
		}); err != nil {
			httpx.Internal(w, h.log, "practice.submit.mistake", err)
			return
		}
	}

	amount := gamification.XPPracticeAttempted
	if result.IsCorrect {
		amount = gamification.XPPracticeCorrect
	}
	awarded, err := h.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    user.ID,
		Amount:    amount,
		Reason:    "Practice: " + question.TypeName,
		SourceKey: "practice:" + question.ID + ":" + gamification.LocalDay(user),
	})
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.xp", err)
		return
	}

	streak, err := h.xp.TouchStreak(ctx, tx, user)
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.streak", err)
		return
	}

	missions, err := h.xp.RecordActivity(ctx, tx, user, question.Skill)
	if err != nil {
		httpx.Internal(w, h.log, "practice.submit.missions", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		httpx.Internal(w, h.log, "practice.submit.commit", err)
		return
	}

	// If referee has completed at least 10 practice questions, qualify referral
	if h.referrals != nil {
		var totalPracticeAttempts int
		if err := h.db.QueryRow(ctx, `SELECT count(*) FROM practice_attempts WHERE user_id = $1`, user.ID).Scan(&totalPracticeAttempts); err == nil && totalPracticeAttempts >= 10 {
			if err := h.referrals.QualifyReferral(ctx, user.ID); err != nil {
				h.log.Error("referral qualification failed for practice", "error", err, "userId", user.ID)
			}
		}
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":   attempt,
		"review":    question.ForReview(),
		"xpAwarded": awarded,
		"streak":    streak,
		"missions":  missions,
	})
}

// examFor resolves the exam from a submission, defaulting to the user's target exam.
func examFor(raw models.ExamType, user models.User) (models.ExamType, bool) {
	if strings.TrimSpace(string(raw)) == "" {
		return user.TargetExam, true
	}
	return raw, raw.Valid()
}
