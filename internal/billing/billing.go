// Package billing owns plans, entitlements and purchase rewards.
package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
)

var (
	ErrPlanNotFound     = errors.New("plan not found")
	ErrLimitReached     = errors.New("daily sub-test limit reached")
	ErrMockLimitReached = errors.New("mock test allowance reached for your plan")
	ErrInvalidPayment   = errors.New("invalid payment confirmation")
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

const planFields = `id, name, price_npr, duration_months, duration_days, bonus_days, features, sub_tests_per_day, mock_tests_included, is_popular`

func (r *Repository) Plans(ctx context.Context) ([]models.Plan, error) {
	rows, err := r.db.Query(ctx, `SELECT `+planFields+` FROM plans ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	plans := []models.Plan{}
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceNPR, &p.DurationMonths, &p.DurationDays, &p.BonusDays,
			&p.Features, &p.SubTestsPerDay, &p.MockTestsIncluded, &p.IsPopular); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		p.AIEvaluationsPerDay = p.SubTestsPerDay // deprecated duplicate, see models.Plan
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *Repository) Plan(ctx context.Context, id string) (models.Plan, error) {
	var p models.Plan
	err := r.db.QueryRow(ctx, `SELECT `+planFields+` FROM plans WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.PriceNPR, &p.DurationMonths, &p.DurationDays, &p.BonusDays,
			&p.Features, &p.SubTestsPerDay, &p.MockTestsIncluded, &p.IsPopular)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Plan{}, ErrPlanNotFound
	}
	if err != nil {
		return models.Plan{}, fmt.Errorf("get plan: %w", err)
	}
	p.AIEvaluationsPerDay = p.SubTestsPerDay // deprecated duplicate, see models.Plan
	return p, nil
}

type Service struct {
	repo          *Repository
	notifications *notifications.Repository
}

func NewService(repo *Repository, notifs *notifications.Repository) *Service {
	return &Service{repo: repo, notifications: notifs}
}

// State returns the learner's plan together with today's usage and mock test entitlement.
func (s *Service) State(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	plan, err := s.effectivePlan(ctx, user)
	if err != nil {
		return models.SubscriptionState{}, err
	}

	dayStart, dayEnd := gamification.LocalDayStart(user), gamification.LocalDayEnd(user)

	const subTestsToday = `
		(SELECT count(DISTINCT COALESCE(q.group_id, q.id))
		   FROM practice_attempts pa
		   JOIN questions q ON q.id = pa.question_id
		  WHERE pa.user_id = $1 AND pa.created_at >= $2 AND pa.created_at < $3)
		+
		(SELECT count(*) FROM ai_evaluations
		  WHERE user_id = $1 AND created_at >= $2 AND created_at < $3)`

	// For free plans, full mocks are counted lifetime (1 included on signup +
	// bonuses). For paid active plans, they are counted per calendar month.
	var subTestsUsed, mocksUsed int
	if plan.ID == "free" {
		err = db.QueryRow(ctx, `
			SELECT `+subTestsToday+`,
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic)`,
			user.ID, dayStart, dayEnd).Scan(&subTestsUsed, &mocksUsed)
	} else {
		err = db.QueryRow(ctx, `
			SELECT `+subTestsToday+`,
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic AND ma.completed_at >= date_trunc('month', now()))`,
			user.ID, dayStart, dayEnd).Scan(&subTestsUsed, &mocksUsed)
	}

	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("read usage: %w", err)
	}

	totalMockTestsAllowed := plan.MockTestsIncluded + user.BonusMockTests

	state := models.SubscriptionState{
		PlanID:             plan.ID,
		PlanName:           plan.Name,
		IsActive:           planIsActive(user),
		DailySubTestsUsed:  subTestsUsed,
		DailySubTestsLimit: plan.SubTestsPerDay,

		// Deprecated duplicates, see models.SubscriptionState.
		DailyEvaluationsUsed:  subTestsUsed,
		DailyEvaluationsLimit: plan.SubTestsPerDay,

		MockTestsIncluded:     plan.MockTestsIncluded,
		BonusMockTests:        user.BonusMockTests,
		TotalMockTestsAllowed: totalMockTestsAllowed,
		MockTestsUsed:         mocksUsed,
		BonusDays:             plan.BonusDays,
	}
	if user.PlanValidUntil != nil {
		state.ValidUntil = user.PlanValidUntil.Format(time.DateOnly)
	}
	return state, nil
}

// SubTestKeyForQuestion returns the group ID if present, otherwise the question ID.
func SubTestKeyForQuestion(question models.Question) string {
	if question.GroupID != "" {
		return question.GroupID
	}
	return question.ID
}

// LockUserForQuota acquires an exclusive row lock on the user for atomic quota checks.
func LockUserForQuota(ctx context.Context, db database.DB, userID string) error {
	if _, err := db.Exec(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, userID); err != nil {
		return fmt.Errorf("lock user for quota: %w", err)
	}
	return nil
}

// CheckSubTestAllowance checks if the user has remaining daily sub-test allowance.
func (s *Service) CheckSubTestAllowance(ctx context.Context, db database.DB, user models.User, taskSetKey string) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.DailySubTestsUsed < state.DailySubTestsLimit {
		return state, nil
	}

	if taskSetKey != "" {
		counted, err := s.taskSetCountedToday(ctx, db, user, taskSetKey)
		if err != nil {
			return state, err
		}
		if counted {
			return state, nil
		}
	}
	return state, ErrLimitReached
}

// taskSetCountedToday reports whether this task set was already attempted today.
func (s *Service) taskSetCountedToday(ctx context.Context, db database.DB, user models.User, taskSetKey string) (bool, error) {
	var counted bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM practice_attempts pa
			  JOIN questions q ON q.id = pa.question_id
			 WHERE pa.user_id = $1
			   AND pa.created_at >= $2 AND pa.created_at < $3
			   AND COALESCE(q.group_id, q.id) = $4)`,
		user.ID, gamification.LocalDayStart(user), gamification.LocalDayEnd(user), taskSetKey,
	).Scan(&counted)
	if err != nil {
		return false, fmt.Errorf("read task set usage: %w", err)
	}
	return counted, nil
}

// CheckMockAllowance checks if the learner has available full mock tests in their quota.
func (s *Service) CheckMockAllowance(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.MockTestsUsed >= state.TotalMockTestsAllowed {
		return state, ErrMockLimitReached
	}
	return state, nil
}

type ConfirmPaymentParams struct {
	UserID         string
	PlanID         string
	PaymentGateway string
	TransactionID  string
	AmountNPR      int
}

// ConfirmPayment processes a successful subscription purchase, granting base duration plus purchase bonus days.
// It is strictly idempotent against duplicate webhooks or retries.
func (s *Service) ConfirmPayment(ctx context.Context, pool *pgxpool.Pool, p ConfirmPaymentParams) (models.SubscriptionState, error) {
	if p.UserID == "" || p.PlanID == "" || p.TransactionID == "" {
		return models.SubscriptionState{}, ErrInvalidPayment
	}

	plan, err := s.repo.Plan(ctx, p.PlanID)
	if err != nil {
		return models.SubscriptionState{}, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("begin payment confirmation tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if this transaction ID was already confirmed (idempotency)
	var existingStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM subscription_payments
		WHERE transaction_id = $1`, p.TransactionID).Scan(&existingStatus)
	if err == nil && existingStatus == "success" {
		// Already processed successfully! Return current state without duplicating entitlement.
		var u models.User
		_ = tx.QueryRow(ctx, `SELECT id, email, name, role, target_exam, target_score, exam_date, nepal_region, xp, streak_days, streak_last_active_date, timezone, plan_id, plan_started_at, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at FROM users WHERE id = $1`, p.UserID).
			Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.TargetExam, &u.TargetScore, &u.ExamDate, &u.NepalRegion, &u.XP, &u.StreakDays, &u.StreakLastActiveDate, &u.Timezone, &u.PlanID, &u.PlanStartedAt, &u.PlanValidUntil, &u.ReferralCode, &u.BonusMockTests, &u.BonusProDays, &u.CreatedAt)
		return s.State(ctx, tx, u)
	}

	// Calculate base + purchase bonus days:
	// Weekly: 7 days (+0 bonus) = 7 days
	// Normal (pro): 30 days (+3 bonus) = 33 days
	// Max (elite): 90 days (+7 bonus) = 97 days
	baseDays := plan.DurationDays
	if baseDays <= 0 && plan.DurationMonths > 0 {
		baseDays = plan.DurationMonths * 30
	}
	bonusDays := plan.BonusDays
	effectiveDays := baseDays + bonusDays

	var updatedUser models.User
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET plan_id = $2,
		    plan_started_at = CURRENT_DATE,
		    plan_valid_until = COALESCE(GREATEST(plan_valid_until, CURRENT_DATE), CURRENT_DATE) + ($3 || ' days')::INTERVAL,
		    role = CASE WHEN role = 'admin' THEN 'admin' ELSE $4 END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, name, role, target_exam, target_score, exam_date, nepal_region, xp, streak_days, streak_last_active_date, timezone, plan_id, plan_started_at, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at`,
		p.UserID, plan.ID, effectiveDays, models.RoleForPlan(plan.ID)).
		Scan(&updatedUser.ID, &updatedUser.Email, &updatedUser.Name,
			&updatedUser.Role, &updatedUser.TargetExam, &updatedUser.TargetScore, &updatedUser.ExamDate,
			&updatedUser.NepalRegion, &updatedUser.XP, &updatedUser.StreakDays, &updatedUser.StreakLastActiveDate,
			&updatedUser.Timezone, &updatedUser.PlanID, &updatedUser.PlanStartedAt, &updatedUser.PlanValidUntil,
			&updatedUser.ReferralCode, &updatedUser.BonusMockTests, &updatedUser.BonusProDays, &updatedUser.CreatedAt)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("update user plan: %w", err)
	}

	// Record subscription payment in ledger
	_, err = tx.Exec(ctx, `
		INSERT INTO subscription_payments (
			user_id, plan_id, payment_gateway, transaction_id,
			amount_npr, status, base_days, bonus_days, effective_days, processed_at
		)
		VALUES ($1, $2, $3, $4, $5, 'success', $6, $7, $8, now())
		ON CONFLICT (transaction_id) DO UPDATE SET
			status = 'success',
			processed_at = now()`,
		p.UserID, plan.ID, p.PaymentGateway, p.TransactionID,
		p.AmountNPR, baseDays, bonusDays, effectiveDays)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("record subscription payment: %w", err)
	}

	// Send confirmation notification
	if s.notifications != nil {
		msg := fmt.Sprintf("Your %s plan is activated for %d days (%d base + %d purchase bonus days)!", plan.Name, effectiveDays, baseDays, bonusDays)
		if bonusDays == 0 {
			msg = fmt.Sprintf("Your %s plan is activated for %d days!", plan.Name, effectiveDays)
		}
		_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
			UserID:    p.UserID,
			Title:     "Subscription Activated! 🎉",
			Message:   msg,
			Type:      "system",
			ActionURL: "/subscription",
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return models.SubscriptionState{}, fmt.Errorf("commit payment confirmation: %w", err)
	}

	return s.State(ctx, pool, updatedUser)
}

// effectivePlan returns the user's active plan, falling back to free if lapsed.
func (s *Service) effectivePlan(ctx context.Context, user models.User) (models.Plan, error) {
	if !planIsActive(user) {
		return s.repo.Plan(ctx, "free")
	}
	plan, err := s.repo.Plan(ctx, user.PlanID)
	if errors.Is(err, ErrPlanNotFound) {
		return s.repo.Plan(ctx, "free")
	}
	return plan, err
}

func planIsActive(user models.User) bool {
	if user.PlanID == "free" {
		return true
	}
	if user.PlanValidUntil == nil {
		return false
	}
	return user.PlanValidUntil.After(time.Now())
}
