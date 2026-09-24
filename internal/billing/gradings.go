package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
)

// AI gradings are the speaking and writing answers an AI model marks. Each
// plan has a pool of them per plan period (plans.ai_gradings_per_period),
// separate from the daily practice sub-tests: reading and listening cost
// nothing to mark, so only what the AI marks is counted here.
//
// Usage is kept in fifths of a grading so that cheap marking can cost less
// than a whole one.
const (
	// UnitsPerGrading is one answer marked by the AI.
	UnitsPerGrading = 5
	// WordMatchUnits is one Read Aloud or Repeat Sentence answer, marked by
	// word-matching against its text: only its transcription is paid for.
	WordMatchUnits = 1
)

// What a sectional mock marked by the AI takes from the pool when it starts,
// in units: its real AI cost, so a mock is never a cheaper route to the same
// marking than practice. Reading and listening mocks take nothing, and full
// mocks are paid for from the full-mock allowance instead.
const (
	PTESpeakingMockUnits   = 14 * UnitsPerGrading
	PTEWritingMockUnits    = 3 * UnitsPerGrading
	IELTSSpeakingMockUnits = 5 * UnitsPerGrading
	IELTSWritingMockUnits  = 2 * UnitsPerGrading
)

// ErrGradingLimitReached means the plan period's AI gradings are used up.
var ErrGradingLimitReached = errors.New("AI grading allowance reached for this plan period")

// GradingLimitMessage is what a learner is told when the pool is used up.
const GradingLimitMessage = "You have used all your AI gradings for speaking and writing for now. " +
	"Reading and listening practice is still available, and your gradings renew on the date shown on your plan. " +
	"Upgrade for more."

// Plan grading periods (plans.ai_gradings_period).
const (
	GradingsPerDay   = "day"
	GradingsPerPlan  = "plan"
	GradingsPerMonth = "month"
)

// gradingPeriod is when the learner's current pool started and when it
// renews: their own day, the paid plan's term, or the calendar month.
func gradingPeriod(user models.User, plan models.Plan) (start, renews time.Time) {
	if plan.AIGradingsPeriod == GradingsPerDay {
		return gamification.LocalDayStart(user), gamification.LocalDayEnd(user)
	}
	if plan.AIGradingsPeriod != GradingsPerMonth && plan.ID != "free" && user.PlanValidUntil != nil {
		start = user.PlanStartedAt
		if start.IsZero() {
			start = time.Now().AddDate(0, 0, -(plan.DurationDays + plan.BonusDays))
		}
		return start, *user.PlanValidUntil
	}
	now := time.Now().UTC()
	start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

// gradingUnitsUsed is what the learner has spent from the pool since start.
func gradingUnitsUsed(ctx context.Context, db database.DB, userID string, start time.Time) (int, error) {
	var units int
	if err := db.QueryRow(ctx, `SELECT COALESCE(SUM(units), 0) FROM ai_grading_usage
		WHERE user_id = $1 AND created_at >= $2`, userID, start).Scan(&units); err != nil {
		return 0, fmt.Errorf("read ai gradings: %w", err)
	}
	return units, nil
}

// CheckAIGradings checks that the pool has `units` left for one more charge.
// It is all or nothing, as sub-test credits are.
func (s *Service) CheckAIGradings(ctx context.Context, db database.DB, user models.User, units int) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.AIGradingUnitsUsed+units <= state.AIGradingsLimit*UnitsPerGrading {
		return state, nil
	}
	period := "in this plan period"
	switch state.AIGradingsPeriod {
	case GradingsPerDay:
		period = "for today. They reset at midnight"
	case GradingsPerMonth:
		period = "this month"
	}
	message := fmt.Sprintf("You have used all %d AI gradings for speaking and writing %s. "+
		"Reading and listening practice is still available.", state.AIGradingsLimit, period)
	if !user.HasActivePaidPlan() {
		message += " Upgrade for more AI gradings."
	}
	s.notifyOnce(ctx, notifications.CreateParams{
		UserID: user.ID, Type: notifications.TypeLimit,
		Title: "AI gradings used up", Message: message,
		ActionURL: "/subscription",
		DedupeKey: "limit:gradings:" + user.PlanID + ":" + state.AIGradingsRenewOn,
	})
	return state, ErrGradingLimitReached
}

// RecordAIGrading spends units from the pool. Callers record a charge only
// once the marking it pays for has happened (or, for a mock, in the
// transaction that starts it), so a failed call costs the learner nothing.
func (s *Service) RecordAIGrading(ctx context.Context, db database.DB, userID string, units int, source, ref string) error {
	if units <= 0 {
		return nil
	}
	if _, err := db.Exec(ctx, `INSERT INTO ai_grading_usage (user_id, units, source, ref) VALUES ($1, $2, $3, $4)`,
		userID, units, source, ref); err != nil {
		return fmt.Errorf("record ai grading: %w", err)
	}
	return nil
}
