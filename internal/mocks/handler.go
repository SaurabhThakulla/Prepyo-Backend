package mocks

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
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
	examRepo *exams.Repository,
	xp *gamification.Service,
	billing *billing.Service,
	referrals ReferralsService,
	log *slog.Logger,
) *Handler {
	return &Handler{
		db:        db,
		repo:      repo,
		questions: questionRepo,
		exams:     examRepo,
		xp:        xp,
		billing:   billing,
		referrals: referrals,
		log:       log,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/attempts", h.attempts)
	r.Get("/{mockID}", h.get)
	r.Post("/{mockID}/submit", h.submit)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	exam := models.ExamType(r.URL.Query().Get("exam"))
	if exam == "" {
		exam = user.TargetExam
	}

	list, err := h.repo.List(r.Context(), exam)
	if err != nil {
		httpx.Internal(w, h.log, "mocks.list", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"mocks": list})
}

// get returns the blueprint plus its questions, with the answer key stripped.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	mock, err := h.repo.ByID(r.Context(), chi.URLParam(r, "mockID"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That mock exam does not exist.")
			return
		}
		httpx.Internal(w, h.log, "mocks.get", err)
		return
	}
	if mock.IsGenerated {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, generatedMockMessage)
		return
	}

	found, err := h.questions.ByIDs(r.Context(), questionIDsOf(mock))
	if err != nil {
		httpx.Internal(w, h.log, "mocks.get.questions", err)
		return
	}

	safe := make([]models.Question, 0, len(found))
	for _, q := range found {
		safe = append(safe, q.PublicQuestion())
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"mock": mock, "questions": safe})
}

func (h *Handler) attempts(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	list, total, err := h.repo.Attempts(r.Context(), user.ID, page.Limit, page.Offset)
	if err != nil {
		httpx.Internal(w, h.log, "mocks.attempts", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"attempts":   list,
		"pagination": page.Meta(total),
	})
}

type submitRequest struct {
	Answers         []models.AnswerSubmission `json:"answers"`
	DurationSeconds int                       `json:"durationSeconds"`
}

// submit grades every answer the learner gave and stores the result.
//
// The score comes from their answers. Speaking and writing sections are left
// out of the score rather than guessed at, and the response says how many
// questions the score covers.
func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if !httpx.Decode(w, r, &req, h.log, "mocks.submit") {
		return
	}

	user := reqctx.MustUser(r.Context())
	ctx := r.Context()

	mock, err := h.repo.ByID(ctx, chi.URLParam(r, "mockID"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That mock exam does not exist.")
			return
		}
		httpx.Internal(w, h.log, "mocks.submit.mock", err)
		return
	}
	if mock.IsGenerated {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, generatedMockMessage)
		return
	}

	version, err := h.exams.ByID(ctx, mock.ExamVersionID)
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.version", err)
		return
	}

	bank, err := h.questions.ByIDs(ctx, questionIDsOf(mock))
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.questions", err)
		return
	}

	graded := gradeAll(bank, req.Answers)
	if graded.total == 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"None of the submitted answers belong to this mock exam.")
		return
	}

	// Full mock exams consume entitlement; check plan quota and bonus mock tests
	if !mock.IsDiagnostic && h.billing != nil {
		if _, err := h.billing.CheckMockAllowance(ctx, h.db, user); err != nil {
			if errors.Is(err, billing.ErrMockLimitReached) {
				httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
					"You have used all mock tests included in your plan. Upgrade or refer friends to unlock more.")
				return
			}
			httpx.Internal(w, h.log, "mocks.submit.checkAllowance", err)
			return
		}
	}

	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	overall := scale.EstimateFromAccuracy(graded.accuracy())

	skillScores := make(map[models.SkillType]float64, len(graded.bySkill))
	for skill, tally := range graded.bySkill {
		skillScores[skill] = scale.EstimateFromAccuracy(tally.accuracy())
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.begin", err)
		return
	}
	defer tx.Rollback(ctx)

	attempt, err := h.repo.SaveAttempt(ctx, tx, SaveAttemptParams{
		UserID:          user.ID,
		MockID:          mock.ID,
		ExamVersionID:   mock.ExamVersionID,
		Exam:            mock.Exam,
		UserScore:       overall,
		SkillScores:     skillScores,
		TotalCorrect:    graded.correct,
		TotalQuestions:  graded.total,
		DurationSeconds: req.DurationSeconds,
	})
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.save", err)
		return
	}

	// Keyed by mock and day: retaking a mock is useful practice and the result
	// is still stored, but the 300 XP is paid once per mock per day rather than
	// on every retake.
	awarded, err := h.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    user.ID,
		Amount:    gamification.XPMockCompleted,
		Reason:    "Completed mock: " + mock.Title,
		SourceKey: "mock:" + mock.ID + ":" + gamification.LocalDay(user),
	})
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.xp", err)
		return
	}

	streak, err := h.xp.TouchStreak(ctx, tx, user)
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.streak", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		httpx.Internal(w, h.log, "mocks.submit.commit", err)
		return
	}

	// If this was a diagnostic test, qualify pending referral
	if mock.IsDiagnostic && h.referrals != nil {
		if err := h.referrals.QualifyReferral(ctx, user.ID); err != nil {
			h.log.Error("referral qualification failed for diagnostic", "error", err, "userId", user.ID)
		}
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":   attempt,
		"xpAwarded": awarded,
		"streak":    streak,
		// Says plainly what the score is based on, so a mock with speaking
		// sections does not look like it graded everything.
		"scoredQuestions":   graded.total,
		"ungradedQuestions": graded.ungraded,
		"scoreConfidence":   scoring.Confidence(graded.total),
	})
}

// tally accumulates marks for a group of questions.
type tally struct {
	earned float64
	max    float64
}

func (t tally) accuracy() float64 {
	if t.max <= 0 {
		return 0
	}
	return t.earned / t.max
}

type gradedMock struct {
	tally
	correct  int
	total    int
	ungraded int
	bySkill  map[models.SkillType]tally
}

// gradeAll scores each answer that belongs to this mock.
//
// Answers for questions outside the blueprint are ignored, so a client cannot
// pad its score by submitting extra answers.
func gradeAll(bank map[string]models.Question, answers []models.AnswerSubmission) gradedMock {
	result := gradedMock{bySkill: map[models.SkillType]tally{}}
	seen := map[string]bool{}

	for _, answer := range answers {
		question, ok := bank[answer.QuestionID]
		if !ok || seen[answer.QuestionID] {
			continue
		}
		seen[answer.QuestionID] = true

		if !scoring.Deterministic(question.Skill) {
			// Speaking and writing need evaluation; they are reported
			// separately instead of being counted as right or wrong.
			result.ungraded++
			continue
		}

		graded, ok := scoring.Grade(question, answer)
		if !ok {
			result.ungraded++
			continue
		}

		result.earned += graded.Score
		result.max += graded.MaxScore
		result.total++
		if graded.IsCorrect {
			result.correct++
		}

		skill := result.bySkill[question.Skill]
		skill.earned += graded.Score
		skill.max += graded.MaxScore
		result.bySkill[question.Skill] = skill
	}
	return result
}

// GradedSet is what GradeAnswers reports back: how much of a paper was
// gradable, and how accurately it was answered overall and per skill.
type GradedSet struct {
	Correct  int
	Total    int
	Ungraded int
	Accuracy float64
	// BySkill holds accuracy in 0..1 for each skill that had a graded
	// question. A skill with nothing gradable is absent rather than zero,
	// because "not measured" and "scored nothing" are different results.
	BySkill map[models.SkillType]float64
}

// GradeAnswers grades a set of answers against a bank of questions.
//
// It exists so a mock that composes its own paper — see internal/reading — is
// scored by exactly the same code as a fixed blueprint, rather than by a second
// implementation that can drift away from this one.
func GradeAnswers(bank map[string]models.Question, answers []models.AnswerSubmission) GradedSet {
	graded := gradeAll(bank, answers)

	set := GradedSet{
		Correct:  graded.correct,
		Total:    graded.total,
		Ungraded: graded.ungraded,
		Accuracy: graded.accuracy(),
		BySkill:  make(map[models.SkillType]float64, len(graded.bySkill)),
	}
	for skill, t := range graded.bySkill {
		set.BySkill[skill] = t.accuracy()
	}
	return set
}

// generatedMockMessage is what a caller gets for trying to read or submit a
// generated blueprint through the fixed-mock endpoints.
const generatedMockMessage = "This mock is composed for you when you start it. " +
	"Start it at POST /api/v1/reading/mocks and submit it to that session."

func questionIDsOf(mock models.Mock) []string {
	var ids []string
	for _, section := range mock.Sections {
		ids = append(ids, section.QuestionIDs...)
	}
	return ids
}
