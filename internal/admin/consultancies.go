package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/institution"
	"github.com/prepyo/backend/pkg/httpx"
)

// usage is what a set of students did in the chosen period; LastActiveAt is all-time.
type usage struct {
	PracticeAttempts int     `json:"practiceAttempts"`
	MocksCompleted   int     `json:"mocksCompleted"`
	AIEvaluations    int     `json:"aiEvaluations"`
	TutorMessages    int     `json:"tutorMessages"`
	LastActiveAt     *string `json:"lastActiveAt"`
}

type consultancy struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Phone      string  `json:"phone"`
	Status     string  `json:"status"`
	ValidUntil *string `json:"validUntil"`
	SeatLimit  int     `json:"seatLimit"`
	CreatedAt  string  `json:"createdAt"`
	AdminName  string  `json:"adminName"`
	AdminEmail string  `json:"adminEmail"`
	// Profile is what the institution answered in its setup; Onboarded says it finished.
	Profile   institution.Profile `json:"profile"`
	Onboarded bool                `json:"onboarded"`
	Students  int                 `json:"students"`
	// ActiveStudents did anything in the period.
	ActiveStudents int   `json:"activeStudents"`
	Usage          usage `json:"usage"`
}

type consultancyStudent struct {
	UserID         string  `json:"userId"`
	Name           string  `json:"name"`
	Username       string  `json:"username"`
	TargetExam     string  `json:"targetExam"`
	PlanID         string  `json:"planId"`
	PlanValidUntil *string `json:"planValidUntil"`
	AddedAt        string  `json:"addedAt"`
	Usage          usage   `json:"usage"`
}

// periodDays are the windows the board offers; 0 means all time.
var periodDays = map[int]bool{0: true, 7: true, 30: true, 90: true}

// readPeriod reads ?days (default 30) and returns the start of the window, nil for all time.
func readPeriod(r *http.Request) (int, *time.Time, bool) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || !periodDays[n] {
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

// studentUsage is one row per student: counts since $1 (NULL = all time) and last activity.
const studentUsage = `
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

// consultancyColumns is read by scanConsultancy; i = institutions, a = institution_admins.
const consultancyColumns = `
	i.id, i.name, i.code, i.phone, i.valid_until, i.seat_limit, i.created_at,
	i.exams, i.city, i.branches, i.monthly_students, i.brand_color, i.onboarded_at IS NOT NULL,
	COALESCE(a.name, ''), COALESCE(a.email, ''),
	count(u.user_id),
	count(u.user_id) FILTER (WHERE u.practice + u.mocks + u.evaluations + u.tutor > 0),
	COALESCE(sum(u.practice), 0)::int, COALESCE(sum(u.mocks), 0)::int,
	COALESCE(sum(u.evaluations), 0)::int, COALESCE(sum(u.tutor), 0)::int,
	max(u.last_active)`

const consultancyFrom = `
	FROM institutions i
	LEFT JOIN institution_admins a ON a.institution_id = i.id
	LEFT JOIN student_usage u ON u.institution_id = i.id`

func scanConsultancy(r row) (consultancy, error) {
	var c consultancy
	var validUntil, lastActive *time.Time
	var createdAt time.Time
	err := r.Scan(&c.ID, &c.Name, &c.Code, &c.Phone, &validUntil, &c.SeatLimit, &createdAt,
		&c.Profile.Exams, &c.Profile.City, &c.Profile.Branches, &c.Profile.MonthlyStudents, &c.Profile.BrandColor, &c.Onboarded,
		&c.AdminName, &c.AdminEmail, &c.Students, &c.ActiveStudents,
		&c.Usage.PracticeAttempts, &c.Usage.MocksCompleted, &c.Usage.AIEvaluations, &c.Usage.TutorMessages, &lastActive)
	if err != nil {
		return consultancy{}, err
	}
	c.Status = institution.Institution{ValidUntil: validUntil}.Status(time.Now())
	c.ValidUntil = formatDate(validUntil, time.DateOnly)
	c.CreatedAt = createdAt.Format(time.RFC3339)
	c.Usage.LastActiveAt = formatDate(lastActive, time.RFC3339)
	return c, nil
}

func formatDate(t *time.Time, layout string) *string {
	if t == nil {
		return nil
	}
	s := t.Format(layout)
	return &s
}

// consultancies lists every institution with its students and their usage over ?days.
func (h *Handler) consultancies(w http.ResponseWriter, r *http.Request) {
	days, since, ok := readPeriod(r)
	if !ok {
		httpx.ValidationError(w, map[string]string{"days": "Choose 7, 30, 90 or 0 (all time)."})
		return
	}
	list, err := h.listConsultancies(r.Context(), since)
	if err != nil {
		httpx.Internal(w, h.log, "admin.consultancies", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"days": days, "consultancies": list})
}

func (h *Handler) listConsultancies(ctx context.Context, since *time.Time) ([]consultancy, error) {
	rows, err := h.db.Query(ctx, `WITH`+studentUsage+`
		SELECT`+consultancyColumns+consultancyFrom+`
		GROUP BY i.id, a.id
		ORDER BY i.created_at DESC`, since)
	if err != nil {
		return nil, fmt.Errorf("list consultancies: %w", err)
	}
	defer rows.Close()

	list := []consultancy{}
	for rows.Next() {
		c, err := scanConsultancy(rows)
		if err != nil {
			return nil, fmt.Errorf("scan consultancy: %w", err)
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// consultancy is one institution with each of its students' usage over ?days.
func (h *Handler) consultancy(w http.ResponseWriter, r *http.Request) {
	days, since, ok := readPeriod(r)
	if !ok {
		httpx.ValidationError(w, map[string]string{"days": "Choose 7, 30, 90 or 0 (all time)."})
		return
	}
	id := chi.URLParam(r, "id")

	c, err := scanConsultancy(h.db.QueryRow(r.Context(), `WITH`+studentUsage+`
		SELECT`+consultancyColumns+consultancyFrom+`
		WHERE i.id::text = $2
		GROUP BY i.id, a.id`, since, id))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That consultancy no longer exists.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.consultancy", err)
		return
	}

	students, err := h.consultancyStudents(r.Context(), since, c.ID)
	if err != nil {
		httpx.Internal(w, h.log, "admin.consultancy", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"days": days, "consultancy": c, "students": students})
}

func (h *Handler) consultancyStudents(ctx context.Context, since *time.Time, institutionID string) ([]consultancyStudent, error) {
	rows, err := h.db.Query(ctx, `WITH`+studentUsage+`
		SELECT s.user_id, us.name, s.username, us.target_exam, us.plan_id, us.plan_valid_until, s.created_at,
			u.practice, u.mocks, u.evaluations, u.tutor, u.last_active
		FROM institution_students s
		JOIN users us ON us.id = s.user_id
		JOIN student_usage u ON u.user_id = s.user_id
		WHERE s.institution_id = $2
		ORDER BY u.last_active DESC NULLS LAST, s.created_at`, since, institutionID)
	if err != nil {
		return nil, fmt.Errorf("list consultancy students: %w", err)
	}
	defer rows.Close()

	list := []consultancyStudent{}
	for rows.Next() {
		var s consultancyStudent
		var planValidUntil, lastActive *time.Time
		var addedAt time.Time
		if err := rows.Scan(&s.UserID, &s.Name, &s.Username, &s.TargetExam, &s.PlanID, &planValidUntil, &addedAt,
			&s.Usage.PracticeAttempts, &s.Usage.MocksCompleted, &s.Usage.AIEvaluations, &s.Usage.TutorMessages, &lastActive); err != nil {
			return nil, fmt.Errorf("scan consultancy student: %w", err)
		}
		s.PlanValidUntil = formatDate(planValidUntil, time.DateOnly)
		s.AddedAt = addedAt.Format(time.RFC3339)
		s.Usage.LastActiveAt = formatDate(lastActive, time.RFC3339)
		list = append(list, s)
	}
	return list, rows.Err()
}
