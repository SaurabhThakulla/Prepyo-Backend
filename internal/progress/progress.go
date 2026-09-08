// Package progress provides score estimates and per-skill breakdown calculations based on attempt history.
package progress

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// recentWindow limits the estimate to work done in the last 60 days.
const recentWindow = 60 * 24 * time.Hour

type Service struct {
	exams *exams.Repository
}

func NewService(examRepo *exams.Repository) *Service {
	return &Service{exams: examRepo}
}

// SkillBreakdown is one row of the progress view.
type SkillBreakdown struct {
	Skill    models.SkillType `json:"skill"`
	Attempts int              `json:"attempts"`
	Accuracy int              `json:"accuracy"`
	Estimate *float64         `json:"estimate"`
	Status   string           `json:"status"` // strong, steady, needs_work, no_data

	Available int `json:"available"`
	Completed int `json:"completed"`
}

// Estimate computes the learner's current standing for their target exam.
func (s *Service) Estimate(ctx context.Context, db database.DB, user models.User) (models.ScoreEstimate, error) {
	version, err := s.exams.Current(ctx, user.TargetExam)
	if err != nil {
		return models.ScoreEstimate{}, err
	}
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}

	var earned, max float64
	var attempts int
	err = db.QueryRow(ctx, `
		SELECT COALESCE(sum(a.score), 0), COALESCE(sum(a.max_score), 0), count(*)
		FROM practice_attempts a
		WHERE a.user_id = $1 AND a.exam = $2 AND a.created_at > $3`,
		user.ID, user.TargetExam, time.Now().Add(-recentWindow)).Scan(&earned, &max, &attempts)
	if err != nil {
		return models.ScoreEstimate{}, fmt.Errorf("read practice totals: %w", err)
	}

	estimate := models.ScoreEstimate{
		Confidence:  scoring.Confidence(attempts),
		BasedOn:     attempts,
		TargetScore: user.TargetScore,
		UpdatedAt:   time.Now(),
	}

	// If a mock attempt exists, prioritize its score.
	var mockScore *float64
	err = db.QueryRow(ctx, `
		SELECT user_score FROM mock_attempts
		WHERE user_id = $1 AND exam = $2
		ORDER BY completed_at DESC
		LIMIT 1`,
		user.ID, user.TargetExam).Scan(&mockScore)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return models.ScoreEstimate{}, fmt.Errorf("read latest mock: %w", err)
	}

	switch {
	case mockScore != nil:
		estimate.Value = mockScore
		if attempts < 8 {
			estimate.Confidence = "medium"
		}
	case max > 0:
		value := scale.EstimateFromAccuracy(earned / max)
		estimate.Value = &value
	default:
		// No evidence yet; Value stays nil.
		estimate.Confidence = "low"
	}

	if estimate.Value != nil && user.TargetScore != nil {
		gap := math.Max(0, *user.TargetScore-*estimate.Value)
		estimate.TargetGap = &gap

		// Calculate readiness towards target score.
		span := *user.TargetScore - scale.Min
		readiness := 100
		if span > 0 {
			readiness = int(math.Round(math.Min(1, (*estimate.Value-scale.Min)/span) * 100))
		}
		if readiness < 0 {
			readiness = 0
		}
		estimate.Readiness = &readiness
	}

	return estimate, nil
}

// Skills reports accuracy per skill over the same window.
func (s *Service) Skills(ctx context.Context, db database.DB, user models.User) ([]SkillBreakdown, error) {
	version, err := s.exams.Current(ctx, user.TargetExam)
	if err != nil {
		return nil, err
	}
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}

	rows, err := db.Query(ctx, `
		SELECT q.skill, count(*), COALESCE(sum(a.score), 0), COALESCE(sum(a.max_score), 0)
		FROM practice_attempts a
		JOIN questions q ON q.id = a.question_id
		WHERE a.user_id = $1 AND a.exam = $2 AND a.created_at > $3
		GROUP BY q.skill`,
		user.ID, user.TargetExam, time.Now().Add(-recentWindow))
	if err != nil {
		return nil, fmt.Errorf("read skill totals: %w", err)
	}
	defer rows.Close()

	type totals struct {
		attempts    int
		earned, max float64
	}
	bySkill := map[models.SkillType]totals{}
	for rows.Next() {
		var skill models.SkillType
		var t totals
		if err := rows.Scan(&skill, &t.attempts, &t.earned, &t.max); err != nil {
			return nil, fmt.Errorf("scan skill totals: %w", err)
		}
		bySkill[skill] = t
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	available, completed, err := s.bankCoverage(ctx, db, user)
	if err != nil {
		return nil, err
	}

	// Return all skills in stable order.
	breakdown := make([]SkillBreakdown, 0, len(models.AllSkills))
	for _, skill := range models.AllSkills {
		t := bySkill[skill]
		row := SkillBreakdown{
			Skill:     skill,
			Attempts:  t.attempts,
			Status:    "no_data",
			Available: available[skill],
			Completed: completed[skill],
		}

		if t.max > 0 {
			accuracy := t.earned / t.max
			value := scale.EstimateFromAccuracy(accuracy)
			row.Accuracy = int(math.Round(accuracy * 100))
			row.Estimate = &value
			row.Status = statusFor(accuracy)
		}
		breakdown = append(breakdown, row)
	}
	return breakdown, nil
}

// bankCoverage counts total available and completed questions per skill.
func (s *Service) bankCoverage(ctx context.Context, db database.DB, user models.User) (map[models.SkillType]int, map[models.SkillType]int, error) {
	available := map[models.SkillType]int{}
	completed := map[models.SkillType]int{}

	rows, err := db.Query(ctx, `
		SELECT skill, count(*) FROM questions
		 WHERE is_published AND $1 = ANY(supported_exams)
		 GROUP BY skill`, user.TargetExam)
	if err != nil {
		return nil, nil, fmt.Errorf("read bank size: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var skill models.SkillType
		var n int
		if err := rows.Scan(&skill, &n); err != nil {
			return nil, nil, fmt.Errorf("scan bank size: %w", err)
		}
		available[skill] = n
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	done, err := db.Query(ctx, `
		SELECT q.skill, count(DISTINCT q.id)
		  FROM questions q
		 WHERE $2 = ANY(q.supported_exams)
		   AND (EXISTS (SELECT 1 FROM practice_attempts a
		                 WHERE a.question_id = q.id AND a.user_id = $1)
		     OR EXISTS (SELECT 1 FROM ai_evaluations e
		                 WHERE e.question_id = q.id AND e.user_id = $1))
		 GROUP BY q.skill`, user.ID, user.TargetExam)
	if err != nil {
		return nil, nil, fmt.Errorf("read completed questions: %w", err)
	}
	defer done.Close()
	for done.Next() {
		var skill models.SkillType
		var n int
		if err := done.Scan(&skill, &n); err != nil {
			return nil, nil, fmt.Errorf("scan completed questions: %w", err)
		}
		completed[skill] = n
	}
	return available, completed, done.Err()
}

func statusFor(accuracy float64) string {
	switch {
	case accuracy >= 0.8:
		return "strong"
	case accuracy >= 0.6:
		return "steady"
	default:
		return "needs_work"
	}
}

// ActivityDay represents one day of practice activity.
type ActivityDay struct {
	Date    string                   `json:"date"` // YYYY-MM-DD
	Count   int                      `json:"count"`
	Minutes int                      `json:"minutes"`
	Skills  map[models.SkillType]int `json:"skills"`
}

// ActivitySummary backs the practice heatmap and the counters above it.
type ActivitySummary struct {
	From          string        `json:"from"`
	To            string        `json:"to"`
	Days          []ActivityDay `json:"days"`
	TotalSessions int           `json:"totalSessions"`
	TotalMinutes  int           `json:"totalMinutes"`
	CurrentStreak int           `json:"currentStreak"`
	LongestStreak int           `json:"longestStreak"`
}

const maxActivityDays = 750

// Activity returns per-day practice counts for the last `days` days in the user's timezone.
func (s *Service) Activity(ctx context.Context, db database.DB, user models.User, days int) (ActivitySummary, error) {
	if days <= 0 {
		days = 365
	}
	days = min(days, maxActivityDays)

	// Fall back to UTC if timezone is invalid.
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	today := time.Now().In(loc)
	end := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	start := end.AddDate(0, 0, -(days - 1))

	rows, err := db.Query(ctx, `
		SELECT (a.created_at AT TIME ZONE $2)::date AS day,
		       q.skill,
		       count(*),
		       COALESCE(sum(a.time_spent_seconds), 0)
		FROM practice_attempts a
		JOIN questions q ON q.id = a.question_id
		WHERE a.user_id = $1
		  AND a.created_at >= $3
		  AND (a.created_at AT TIME ZONE $2)::date >= $4::date
		GROUP BY day, q.skill
		ORDER BY day`,
		user.ID, user.Timezone, start.AddDate(0, 0, -1), start.Format(time.DateOnly))
	if err != nil {
		return ActivitySummary{}, fmt.Errorf("read practice activity: %w", err)
	}
	defer rows.Close()

	byDate := map[string]*ActivityDay{}
	order := []string{}
	totalSeconds := 0

	for rows.Next() {
		var day time.Time
		var skill models.SkillType
		var count, seconds int
		if err := rows.Scan(&day, &skill, &count, &seconds); err != nil {
			return ActivitySummary{}, fmt.Errorf("scan practice activity: %w", err)
		}

		key := day.Format(time.DateOnly)
		entry, ok := byDate[key]
		if !ok {
			entry = &ActivityDay{Date: key, Skills: map[models.SkillType]int{}}
			byDate[key] = entry
			order = append(order, key)
		}
		entry.Count += count
		entry.Skills[skill] += count
		entry.Minutes += seconds / 60
		totalSeconds += seconds
	}
	if err := rows.Err(); err != nil {
		return ActivitySummary{}, err
	}

	summary := ActivitySummary{
		From:         start.Format(time.DateOnly),
		To:           end.Format(time.DateOnly),
		Days:         make([]ActivityDay, 0, len(order)),
		TotalMinutes: totalSeconds / 60,
	}
	for _, key := range order {
		summary.Days = append(summary.Days, *byDate[key])
		summary.TotalSessions += byDate[key].Count
	}

	summary.CurrentStreak = currentStreak(byDate, end)
	summary.LongestStreak = longestStreak(byDate, start, end)
	return summary, nil
}

func currentStreak(byDate map[string]*ActivityDay, end time.Time) int {
	cursor := end
	if byDate[cursor.Format(time.DateOnly)] == nil {
		cursor = cursor.AddDate(0, 0, -1)
	}

	streak := 0
	for byDate[cursor.Format(time.DateOnly)] != nil {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

func longestStreak(byDate map[string]*ActivityDay, start, end time.Time) int {
	best, run := 0, 0
	for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 0, 1) {
		if byDate[cursor.Format(time.DateOnly)] != nil {
			run++
			best = max(best, run)
			continue
		}
		run = 0
	}
	return best
}
