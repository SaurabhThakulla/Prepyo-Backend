// Package progress turns a learner's attempt history into an estimate and a
// per-skill breakdown.
//
// Nothing here is stored on the user row. The estimate is recomputed from
// attempts each time it is asked for, so it cannot drift away from the evidence
// behind it.
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

// recentWindow limits the estimate to work done in the last two months. Older
// attempts say more about who the learner used to be than who they are now.
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

	// Available is how many published questions exist for this skill and exam,
	// and Completed how many distinct ones this learner has answered.
	//
	// Both are all-time, unlike Attempts and Accuracy above, which look at a
	// recent window. "You have done 12 of 133 reading questions" is a statement
	// about the bank, and a question answered two months ago is still done.
	Available int `json:"available"`
	Completed int `json:"completed"`
}

// Estimate computes the learner's current standing for their target exam.
//
// With no attempts the value is nil rather than a default number, so the UI can
// say "take a diagnostic" instead of showing a score nobody earned.
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
		JOIN questions q ON q.id = a.question_id
		WHERE a.user_id = $1 AND q.exam = $2 AND a.created_at > $3`,
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

	// A mock is a better signal than scattered practice, so when one exists it
	// takes precedence.
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

		// Readiness is progress along the scale towards the target, not a raw
		// score ratio, so IELTS band 6 of 7 does not read as 86% ready.
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
		WHERE a.user_id = $1 AND q.exam = $2 AND a.created_at > $3
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

	// Always return all four skills in a stable order so the UI does not have
	// rows appearing and disappearing between refreshes.
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

// bankCoverage counts what exists and what this learner has done, per skill.
//
// The available count is every published question for the exam, deliberately
// including passage-backed reading questions and re-order items. Those are
// excluded from the generic /questions listing because they cannot be dealt
// standalone, but they are absolutely part of the bank a learner works through,
// and leaving them out would report a reading bank of 3 against a real 133.
//
// Completed unions the two places an answer can land: reading and listening
// write practice_attempts, writing and speaking write ai_evaluations. Counting
// only the first would leave those two skills permanently at zero.
func (s *Service) bankCoverage(ctx context.Context, db database.DB, user models.User) (map[models.SkillType]int, map[models.SkillType]int, error) {
	available := map[models.SkillType]int{}
	completed := map[models.SkillType]int{}

	rows, err := db.Query(ctx, `
		SELECT skill, count(*) FROM questions
		 WHERE is_published AND exam = $1
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
		 WHERE q.exam = $2
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

// ActivityDay is one day of practice, bucketed in the learner's own timezone.
//
// Only days with attempts are returned. A year of empty squares is the client's
// job to draw, not ours to send.
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

// maxActivityDays caps the window at roughly two years. The heatmap asks for
// one; anything past that is someone probing the query.
const maxActivityDays = 750

// Activity returns per-day practice counts for the last `days` days.
//
// Days are bucketed in the learner's timezone, not UTC: a 9pm session in
// Kathmandu belongs to that evening's square, not to the next morning's.
func (s *Service) Activity(ctx context.Context, db database.DB, user models.User, days int) (ActivitySummary, error) {
	if days <= 0 {
		days = 365
	}
	days = min(days, maxActivityDays)

	// An unknown zone must not take the whole page down, and Postgres would
	// reject it too, so fall back to UTC and carry on.
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	today := time.Now().In(loc)
	end := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	start := end.AddDate(0, 0, -(days - 1))

	// The extra created_at bound is what lets the query use
	// idx_practice_attempts_user; the date cast alone would not. A day of slack
	// covers every UTC offset either side of the local midnight.
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

// currentStreak counts back from today. Today not being practised yet does not
// break a streak — it is still early — so the walk starts at yesterday then.
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
