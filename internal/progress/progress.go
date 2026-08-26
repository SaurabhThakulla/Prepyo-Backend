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

	// Always return all four skills in a stable order so the UI does not have
	// rows appearing and disappearing between refreshes.
	breakdown := make([]SkillBreakdown, 0, len(models.AllSkills))
	for _, skill := range models.AllSkills {
		t := bySkill[skill]
		row := SkillBreakdown{Skill: skill, Attempts: t.attempts, Status: "no_data"}

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
