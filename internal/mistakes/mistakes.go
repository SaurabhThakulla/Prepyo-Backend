// Package mistakes manages recorded user mistakes and resolution status.
package mistakes

import (
	"context"
	"errors"
	"fmt"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("mistake not found")

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type RecordParams struct {
	UserID          string
	QuestionID      string
	Exam            models.ExamType
	ErrorTag        string
	UserResponse    string
	CorrectResponse string
	Explanation     string
}

// Record adds or updates a mistake, incrementing failure count on duplicate.
func (r *Repository) Record(ctx context.Context, db database.DB, p RecordParams) error {
	_, err := db.Exec(ctx, `
		INSERT INTO mistakes (user_id, question_id, exam, error_tag, user_response, correct_response, explanation)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, question_id) DO UPDATE SET
			failed_count      = mistakes.failed_count + 1,
			exam              = EXCLUDED.exam,
			error_tag         = EXCLUDED.error_tag,
			user_response     = EXCLUDED.user_response,
			resolved          = FALSE,
			last_attempted_at = now()`,
		p.UserID, p.QuestionID, p.Exam, p.ErrorTag, p.UserResponse, p.CorrectResponse, p.Explanation)
	if err != nil {
		return fmt.Errorf("record mistake: %w", err)
	}
	return nil
}

type ListParams struct {
	UserID         string
	Exam           models.ExamType
	Skill          models.SkillType
	UnresolvedOnly bool
	Period         string
	Limit          int
	Offset         int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.Mistake, int, error) {
	filter := `
		WHERE m.user_id = $1
		  AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
		  AND ($3 = '' OR LOWER(q.skill) = LOWER($3))
		  AND (NOT $4 OR NOT m.resolved)
		  AND ` + periodWindow("$5")

	var total int
	err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM mistakes m
		JOIN questions q ON q.id = m.question_id`+filter,
		p.UserID, p.Exam, p.Skill, p.UnresolvedOnly, normalizePeriod(p.Period)).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count mistakes: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.question_id, q.title, m.exam, q.skill, q.type_name, q.prompt,
		       m.user_response, m.correct_response, m.explanation, m.error_tag,
		       m.failed_count, m.resolved, m.last_attempted_at
		FROM mistakes m
		JOIN questions q ON q.id = m.question_id`+filter+`
		ORDER BY m.resolved, m.failed_count DESC, m.last_attempted_at DESC
		LIMIT $6 OFFSET $7`,
		p.UserID, p.Exam, p.Skill, p.UnresolvedOnly, normalizePeriod(p.Period), p.Limit, p.Offset)

	if err != nil {
		return nil, 0, fmt.Errorf("list mistakes: %w", err)
	}
	defer rows.Close()

	list := []models.Mistake{}
	for rows.Next() {
		var m models.Mistake
		if err := rows.Scan(&m.ID, &m.QuestionID, &m.QuestionTitle, &m.Exam, &m.Skill,
			&m.TypeName, &m.Prompt, &m.UserResponse, &m.CorrectResponse, &m.Explanation,
			&m.ErrorTag, &m.FailedCount, &m.Resolved, &m.LastAttemptedAt); err != nil {
			return nil, 0, fmt.Errorf("scan mistake: %w", err)
		}
		list = append(list, m)
	}
	return list, total, rows.Err()
}

// Resolve requires a newer correct attempt for this learner, question and exam.
func (r *Repository) Resolve(ctx context.Context, db database.DB, userID, mistakeID string) error {
	tag, err := db.Exec(ctx, `
		UPDATE mistakes m SET resolved = TRUE
		WHERE m.id = $1 AND m.user_id = $2 AND NOT m.resolved
		  AND EXISTS (
		    SELECT 1 FROM practice_attempts a
		    WHERE a.user_id=m.user_id AND a.question_id=m.question_id
		      AND a.exam=m.exam AND a.is_correct AND a.created_at > m.last_attempted_at
		      AND NOT EXISTS (
		        SELECT 1 FROM practice_attempts newer
		        WHERE newer.user_id=a.user_id AND newer.question_id=a.question_id
		          AND newer.exam=a.exam AND newer.created_at>a.created_at
		      )
		  )`,
		mistakeID, userID)
	if err != nil {
		return fmt.Errorf("resolve mistake: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type AnalyticsParams struct {
	UserID string
	Exam   models.ExamType
	Period string // "weekly" or "monthly"; anything else is weekly
}

type TimelinePoint struct {
	Label    string `json:"label"`
	Date     string `json:"date"`
	Count    int    `json:"count"`
	Resolved int    `json:"resolved"`
}

type TagStat struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type AnalyticsSummary struct {
	Period      string                   `json:"period"`
	Total       int                      `json:"total"`
	Unresolved  int                      `json:"unresolved"`
	Resolved    int                      `json:"resolved"`
	RepeatCount int                      `json:"repeatCount"`
	BySkill     map[models.SkillType]int `json:"bySkill"`
	ByTag       []TagStat                `json:"byTag"`
	Timeline    []TimelinePoint          `json:"timeline"`
}

func (r *Repository) Analytics(ctx context.Context, p AnalyticsParams) (AnalyticsSummary, error) {
	period := normalizePeriod(p.Period)

	summary := AnalyticsSummary{
		Period:   period,
		BySkill:  map[models.SkillType]int{},
		ByTag:    []TagStat{},
		Timeline: []TimelinePoint{},
	}

	// 1. Overall stats
	statsQuery := `
		SELECT
			count(*) as total,
			count(*) FILTER (WHERE NOT m.resolved) as unresolved,
			count(*) FILTER (WHERE m.resolved) as resolved,
			count(*) FILTER (WHERE m.failed_count > 1) as repeat_count
		FROM mistakes m
		WHERE m.user_id = $1
		  AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
		  AND ` + periodWindow("$3")

	if err := r.db.QueryRow(ctx, statsQuery, p.UserID, p.Exam, period).Scan(
		&summary.Total, &summary.Unresolved, &summary.Resolved, &summary.RepeatCount,
	); err != nil {
		return summary, fmt.Errorf("read mistake stats: %w", err)
	}

	// 2. Breakdown by skill
	skillQuery := `
		SELECT q.skill, count(m.id)
		FROM mistakes m
		JOIN questions q ON q.id = m.question_id
		WHERE m.user_id = $1
		  AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
		  AND ` + periodWindow("$3") + `
		GROUP BY q.skill`

	skillRows, err := r.db.Query(ctx, skillQuery, p.UserID, p.Exam, period)
	if err == nil {
		for skillRows.Next() {
			var skill models.SkillType
			var count int
			if err := skillRows.Scan(&skill, &count); err == nil {
				summary.BySkill[skill] = count
			}
		}
		skillRows.Close()
	}

	// 3. Top tags
	tagQuery := `
		SELECT m.error_tag, count(m.id)
		FROM mistakes m
		WHERE m.user_id = $1
		  AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
		  AND m.error_tag <> ''
		  AND ` + periodWindow("$3") + `
		GROUP BY m.error_tag
		ORDER BY count(m.id) DESC
		LIMIT 5`

	tagRows, err := r.db.Query(ctx, tagQuery, p.UserID, p.Exam, period)
	if err == nil {
		for tagRows.Next() {
			var tag string
			var count int
			if err := tagRows.Scan(&tag, &count); err == nil {
				summary.ByTag = append(summary.ByTag, TagStat{Tag: tag, Count: count})
			}
		}
		tagRows.Close()
	}

	// 4. Timeline points
	var timelineQuery string
	switch period {
	case "monthly":
		// Four whole weeks, oldest first: the 1st week is 21-27 days ago, the 4th is the last 7 days.
		timelineQuery = `
			SELECT (ARRAY['1st week', '2nd week', '3rd week', '4th week'])[w + 1] AS label, to_char(ws.s, 'YYYY-MM-DD') AS date_str,
			       count(m.id) AS count,
			       count(m.id) FILTER (WHERE m.resolved) AS resolved
			FROM generate_series(0, 3) w
			CROSS JOIN LATERAL (SELECT CURRENT_DATE - 27 + w * 7 AS s) ws
			LEFT JOIN (
			    SELECT m.id, m.resolved, m.last_attempted_at FROM mistakes m
			    WHERE m.user_id = $1 AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
			) m ON m.last_attempted_at::date BETWEEN ws.s AND ws.s + 6
			GROUP BY w, ws.s
			ORDER BY w`
	default: // weekly
		timelineQuery = `
			SELECT to_char(d, 'Dy') as label, to_char(d, 'YYYY-MM-DD') as date_str,
			       count(m.id) as count,
			       count(m.id) FILTER (WHERE m.resolved) as resolved
			FROM generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day'::interval) d
			LEFT JOIN (
			    SELECT m.id, m.resolved, m.last_attempted_at FROM mistakes m
			    WHERE m.user_id = $1 AND ($2 = '' OR UPPER(m.exam) = UPPER($2))
			) m ON m.last_attempted_at::date = d::date
			GROUP BY d
			ORDER BY d`
	}

	timeRows, err := r.db.Query(ctx, timelineQuery, p.UserID, p.Exam)
	if err != nil {
		return summary, fmt.Errorf("read mistake timeline: %w", err)
	}
	defer timeRows.Close()

	for timeRows.Next() {
		var pt TimelinePoint
		if err := timeRows.Scan(&pt.Label, &pt.Date, &pt.Count, &pt.Resolved); err == nil {
			summary.Timeline = append(summary.Timeline, pt)
		}
	}

	return summary, nil
}

// The mistake bank has two views: the last 7 days, and the last 4 weeks.
func normalizePeriod(period string) string {
	if period == "monthly" {
		return "monthly"
	}
	return "weekly"
}

// periodWindow is the date filter for a period, against the SQL parameter that
// carries it. One definition keeps the list, the totals and the graph on exactly
// the same window. Monthly is four whole weeks so it splits into the 1st to 4th
// week.
func periodWindow(param string) string {
	return fmt.Sprintf(`(
		      (%[1]s = 'weekly' AND m.last_attempted_at >= CURRENT_DATE - INTERVAL '6 days') OR
		      (%[1]s = 'monthly' AND m.last_attempted_at >= CURRENT_DATE - INTERVAL '27 days')
		  )`, param)
}
