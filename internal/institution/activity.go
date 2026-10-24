package institution

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/pkg/httpx"
)

// unitsPerGrading matches billing.UnitsPerGrading (billing imports this package, so it cannot be imported here).
const unitsPerGrading = 5

// Usage is what a student did in a period; LastActiveAt is all time, AIGradingsThisMonth is this calendar month.
type Usage struct {
	PracticeAttempts    int        `json:"practiceAttempts"`
	MocksCompleted      int        `json:"mocksCompleted"`
	AIEvaluations       int        `json:"aiEvaluations"`
	TutorMessages       int        `json:"tutorMessages"`
	LastActiveAt        *time.Time `json:"lastActiveAt"`
	AIGradingsThisMonth float64    `json:"aiGradingsThisMonth"`
}

// StudentUsageSQL is a CTE, student_usage: one row per institution student with counts since $1 (NULL = all time).
const StudentUsageSQL = `
	student_usage AS (
		SELECT s.user_id, s.institution_id,
			p.n AS practice, m.n AS mocks, e.n AS evaluations, t.n AS tutor,
			GREATEST(p.last, m.last, e.last, t.last) AS last_active
		FROM institution_students s
		CROSS JOIN LATERAL (SELECT count(*) FILTER (WHERE $1::timestamptz IS NULL OR created_at >= $1) AS n, max(created_at) AS last
			FROM practice_attempts WHERE user_id = s.user_id) p
		CROSS JOIN LATERAL (SELECT count(*) FILTER (WHERE $1::timestamptz IS NULL OR completed_at >= $1) AS n, max(completed_at) AS last
			FROM mock_attempts WHERE user_id = s.user_id) m
		CROSS JOIN LATERAL (SELECT count(*) FILTER (WHERE $1::timestamptz IS NULL OR created_at >= $1) AS n, max(created_at) AS last
			FROM ai_evaluations WHERE user_id = s.user_id) e
		CROSS JOIN LATERAL (SELECT count(*) FILTER (WHERE $1::timestamptz IS NULL OR created_at >= $1) AS n, max(created_at) AS last
			FROM tutor_messages WHERE user_id = s.user_id) t
	)`

// readPeriod reads ?days (7, 30, 90 or 0 for all time; default 30) and returns the start, nil for all time.
func readPeriod(r *http.Request) (int, *time.Time, bool) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || (n != 0 && n != 7 && n != 30 && n != 90) {
			return 0, nil, false
		}
		days = n
	}
	if days == 0 {
		return 0, nil, true
	}
	since := time.Now().AddDate(0, 0, -days)
	return days, &since, true
}

// ActivityItem is one practice answer or finished mock.
type ActivityItem struct {
	Kind     string    `json:"kind"` // practice or mock
	Exam     string    `json:"exam"`
	Skill    string    `json:"skill"`
	Task     string    `json:"task"`
	Title    string    `json:"title"`
	Score    *float64  `json:"score"`
	MaxScore *float64  `json:"maxScore"`
	Accuracy *int      `json:"accuracy"` // percent
	At       time.Time `json:"at"`
}

// SkillSummary is a student's practice in one skill over the period.
type SkillSummary struct {
	Skill           string   `json:"skill"`
	Attempts        int      `json:"attempts"`
	AverageAccuracy *float64 `json:"averageAccuracy"`
}

// studentActivity shows one student's practice by skill over ?days and their latest 40 answers and mocks.
func (h *Handler) studentActivity(w http.ResponseWriter, r *http.Request) {
	days, since, ok := readPeriod(r)
	if !ok {
		httpx.ValidationError(w, map[string]string{"days": "Choose 7, 30, 90 or 0 (all time)."})
		return
	}
	userID := chi.URLParam(r, "id")
	ctx := r.Context()

	var name, username string
	err := h.store.pool.QueryRow(ctx, `
		SELECT u.name, st.username FROM institution_students st JOIN users u ON u.id = st.user_id
		WHERE st.institution_id = $1 AND st.user_id::text = $2`,
		signedInAccount(r).Institution.ID, userID).Scan(&name, &username)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That student is not in your institution.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.studentActivity", err)
		return
	}

	skills, err := h.store.skillSummary(ctx, userID, since)
	if err != nil {
		httpx.Internal(w, h.log, "institution.studentActivity", err)
		return
	}
	recent, err := h.store.recentActivity(ctx, userID)
	if err != nil {
		httpx.Internal(w, h.log, "institution.studentActivity", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"days": days, "name": name, "username": username, "skills": skills, "recent": recent,
	})
}

func (s *store) skillSummary(ctx context.Context, userID string, since *time.Time) ([]SkillSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT COALESCE(q.skill, 'other'), count(*), avg(pa.accuracy_percentage)::float8
		FROM practice_attempts pa
		LEFT JOIN questions q ON q.id = pa.question_id
		WHERE pa.user_id::text = $1 AND ($2::timestamptz IS NULL OR pa.created_at >= $2)
		GROUP BY 1 ORDER BY 2 DESC`, userID, since)
	if err != nil {
		return nil, fmt.Errorf("student skill summary: %w", err)
	}
	defer rows.Close()

	list := []SkillSummary{}
	for rows.Next() {
		var s SkillSummary
		if err := rows.Scan(&s.Skill, &s.Attempts, &s.AverageAccuracy); err != nil {
			return nil, fmt.Errorf("scan skill summary: %w", err)
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (s *store) recentActivity(ctx context.Context, userID string) ([]ActivityItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM (
			SELECT 'practice' AS kind, COALESCE(pa.exam, ''), COALESCE(q.skill, ''), COALESCE(q.type_name, ''),
				COALESCE(q.title, ''), pa.score::float8, pa.max_score::float8, pa.accuracy_percentage, pa.created_at AS at
			FROM practice_attempts pa
			LEFT JOIN questions q ON q.id = pa.question_id
			WHERE pa.user_id::text = $1
			UNION ALL
			SELECT 'mock', COALESCE(ma.exam, ''), '', 'Mock test', COALESCE(m.title, ''),
				ma.user_score::float8, NULL, NULL, ma.completed_at
			FROM mock_attempts ma
			LEFT JOIN mocks m ON m.id = ma.mock_id
			WHERE ma.user_id::text = $1 AND ma.completed_at IS NOT NULL
		) a
		ORDER BY at DESC
		LIMIT 40`, userID)
	if err != nil {
		return nil, fmt.Errorf("student recent activity: %w", err)
	}
	defer rows.Close()

	list := []ActivityItem{}
	for rows.Next() {
		var a ActivityItem
		if err := rows.Scan(&a.Kind, &a.Exam, &a.Skill, &a.Task, &a.Title, &a.Score, &a.MaxScore, &a.Accuracy, &a.At); err != nil {
			return nil, fmt.Errorf("scan recent activity: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
