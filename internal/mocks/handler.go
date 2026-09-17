package mocks

import (
	"errors"
	"log/slog"
	"net/http"

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

type Handler struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	exams     *exams.Repository
	xp        *gamification.Service
	billing   *billing.Service
	mistakes  *mistakes.Repository
	log       *slog.Logger
}

func NewHandler(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	examRepo *exams.Repository,
	xp *gamification.Service,
	billing *billing.Service,
	mistakeRepo *mistakes.Repository,
	log *slog.Logger,
) *Handler {
	return &Handler{
		db:        db,
		repo:      repo,
		questions: questionRepo,
		exams:     examRepo,
		xp:        xp,
		billing:   billing,
		mistakes:  mistakeRepo,
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

// get returns the mock blueprint and its public questions.
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

// submit grades the submitted answers and saves the mock attempt.
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

	graded, err := gradeCanonical(bank, questionIDsOf(mock), req.Answers)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Answers must match this mock's question list without duplicates.")
		return
	}
	if graded.total == 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"None of the submitted answers belong to this mock exam.")
		return
	}

	// Check plan quota and bonus mock tests.
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

	skillScores := make(map[models.SkillType]float64, len(graded.bySkill))
	for skill, t := range graded.bySkill {
		skillScores[skill] = scale.EstimateFromRawMarks(string(mock.Exam), string(skill), t.correct, t.total)
	}

	overall := scale.EstimateOverall(string(mock.Exam), skillScores, graded.correct, graded.total)

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
		Answers:         req.Answers,
	})
	if err != nil {
		httpx.Internal(w, h.log, "mocks.submit.save", err)
		return
	}

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

	if h.mistakes != nil {
		for _, answer := range req.Answers {
			question, ok := bank[answer.QuestionID]
			if !ok {
				continue
			}
			if !scoring.Deterministic(question.Skill) {
				continue
			}
			gradedRes, ok := scoring.Grade(question, answer)
			if !ok || gradedRes.IsCorrect {
				continue
			}
			_ = h.mistakes.Record(ctx, tx, mistakes.RecordParams{
				UserID:          user.ID,
				QuestionID:      question.ID,
				Exam:            mock.Exam,
				ErrorTag:        gradedRes.ErrorTag,
				UserResponse:    gradedRes.UserDisplay,
				CorrectResponse: gradedRes.CorrectDisplay,
				Explanation:     question.Explanation,
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {

		httpx.Internal(w, h.log, "mocks.submit.commit", err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":           attempt,
		"xpAwarded":         awarded,
		"streak":            streak,
		"scoredQuestions":   graded.total,
		"ungradedQuestions": graded.ungraded,
		"scoreConfidence":   scoring.Confidence(graded.total),
	})
}

// tally accumulates marks and question counts for a group of questions.
type tally struct {
	earned  float64
	max     float64
	correct int
	total   int
}

func (t tally) accuracy() float64 {
	if t.max <= 0 {
		return 0
	}
	return t.earned / t.max
}

var ErrInvalidAnswers = errors.New("answers do not match the canonical mock question list")

func gradeCanonical(bank map[string]models.Question, canonicalIDs []string, answers []models.AnswerSubmission) (gradedMock, error) {
	result := gradedMock{bySkill: map[models.SkillType]tally{}}
	canonical := make(map[string]bool, len(canonicalIDs))
	submissions := make(map[string]models.AnswerSubmission, len(answers))
	for _, id := range canonicalIDs {
		canonical[id] = true
	}
	for _, answer := range answers {
		if !canonical[answer.QuestionID] {
			return gradedMock{}, ErrInvalidAnswers
		}
		if _, exists := submissions[answer.QuestionID]; exists {
			return gradedMock{}, ErrInvalidAnswers
		}
		submissions[answer.QuestionID] = answer
	}

	for _, id := range canonicalIDs {
		question, ok := bank[id]
		if !ok {
			return gradedMock{}, ErrInvalidAnswers
		}
		answer := submissions[id]
		if !scoring.Deterministic(question.Skill) {
			result.ungraded++
			continue
		}
		graded, ok := scoring.Grade(question, answer)
		if !ok {
			return gradedMock{}, ErrInvalidAnswers
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
		skill.total++
		if graded.IsCorrect {
			skill.correct++
		}
		result.bySkill[question.Skill] = skill
	}
	return result, nil
}

type gradedMock struct {
	tally
	correct  int
	total    int
	ungraded int
	bySkill  map[models.SkillType]tally
}

// gradeAll scores each answer that belongs to this mock.
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
		skill.total++
		if graded.IsCorrect {
			skill.correct++
		}
		result.bySkill[question.Skill] = skill
	}
	return result
}

// GradedSet holds aggregate scoring metrics for a set of answers.
type GradedSet struct {
	Correct  int
	Total    int
	Ungraded int
	Accuracy float64
	BySkill  map[models.SkillType]float64
}

// GradeAnswers grades a set of answers against a bank of questions.
func GradeAnswers(bank map[string]models.Question, canonicalIDs []string, answers []models.AnswerSubmission) (GradedSet, error) {
	graded, err := gradeCanonical(bank, canonicalIDs, answers)
	if err != nil {
		return GradedSet{}, err
	}

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
	return set, nil
}

const generatedMockMessage = "This mock is composed for you when you start it. " +
	"Start it at POST /api/v1/reading/mocks and submit it to that session."

func questionIDsOf(mock models.Mock) []string {
	var ids []string
	for _, section := range mock.Sections {
		ids = append(ids, section.QuestionIDs...)
	}
	return ids
}
