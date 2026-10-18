package practice

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/prepyo/backend/internal/evaluations"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// historyLimit caps each list; the screen shows recent attempts, not an archive.
const historyLimit = 50

// maxHistoryQuestions covers the largest reading set, whose questions are asked for together.
const maxHistoryQuestions = 40

// history is the learner's own earlier attempts at these questions, newest
// first: right/wrong attempts and AI evaluations. Only what was already marked
// for them comes back, never an answer key, and nothing for a question in a
// mock they are sitting.
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	ids := []string{}
	for _, id := range r.URL.Query()["questionId"] {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 || len(ids) > maxHistoryQuestions {
		httpx.ValidationError(w, map[string]string{"questionId": fmt.Sprintf("Give 1 to %d question ids.", maxHistoryQuestions)})
		return
	}

	for _, id := range ids {
		open, err := h.questions.InOpenMock(r.Context(), user.ID, id)
		if err != nil {
			httpx.Internal(w, h.log, "practice.history.mock", err)
			return
		}
		if open {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "This isn't available while the question is part of a mock test you are taking.")
			return
		}
	}

	attempts, err := h.repo.ForQuestions(r.Context(), user.ID, ids, historyLimit)
	if err != nil {
		httpx.Internal(w, h.log, "practice.history.attempts", err)
		return
	}
	evals, err := evaluations.ForQuestions(r.Context(), h.db, user.ID, ids, historyLimit)
	if err != nil {
		httpx.Internal(w, h.log, "practice.history.evaluations", err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, http.StatusOK, map[string]any{"attempts": attempts, "evaluations": evals})
}

// ForQuestions is one learner's own attempts at these questions, newest first.
func (r *Repository) ForQuestions(ctx context.Context, userID string, questionIDs []string, limit int) ([]models.PracticeAttempt, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, question_id, exam, exam_version_id, is_correct, score, max_score,
		       accuracy_percentage, COALESCE(feedback, ''), COALESCE(user_response, ''),
		       time_spent_seconds, created_at
		FROM practice_attempts
		WHERE user_id = $1 AND question_id = ANY($2)
		ORDER BY created_at DESC
		LIMIT $3`,
		userID, questionIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("list practice attempts for questions: %w", err)
	}
	defer rows.Close()

	list := []models.PracticeAttempt{}
	for rows.Next() {
		var a models.PracticeAttempt
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.Exam, &a.ExamVersionID, &a.IsCorrect, &a.Score,
			&a.MaxScore, &a.AccuracyPercentage, &a.Feedback, &a.UserResponse,
			&a.TimeSpentSeconds, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan practice attempt: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
