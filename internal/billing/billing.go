// Package billing owns plans, entitlements and purchase rewards.
package billing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
	"github.com/prepyo/backend/internal/referrals"
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

const planFields = `id, name, price_npr, duration_months, duration_days, bonus_days, features, sub_tests_per_day, mock_tests_included, is_popular, unlimited, ai_gradings_per_period, ai_gradings_period`

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
			&p.Features, &p.SubTestsPerDay, &p.MockTestsIncluded, &p.IsPopular, &p.Unlimited, &p.AIGradingsPerPeriod,
			&p.AIGradingsPeriod); err != nil {
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
			&p.Features, &p.SubTestsPerDay, &p.MockTestsIncluded, &p.IsPopular, &p.Unlimited, &p.AIGradingsPerPeriod,
			&p.AIGradingsPeriod)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Plan{}, ErrPlanNotFound
	}
	if err != nil {
		return models.Plan{}, fmt.Errorf("get plan: %w", err)
	}
	p.AIEvaluationsPerDay = p.SubTestsPerDay // deprecated duplicate, see models.Plan
	return p, nil
}

// FullMockID is the attempt row of an IELTS full mock, which is counted
// against the allowance by its session rather than its attempt.
const FullMockID = "mock-ielts-full"

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

	// Session starts are the credit ledger. Answers and AI feedback never add
	// usage, and stopping or completing a session never refunds its start.
	const subTestsToday = `
		(SELECT COALESCE(SUM(credits), 0) FROM practice_sessions
		  WHERE user_id = $1 AND created_at >= $2 AND created_at < $3)`

	// For free plans, full mocks are counted lifetime (1 included on signup +
	// bonuses). For paid active plans, they are counted per calendar month.
	// Generated mocks are single-section papers; they are paid for in
	// sub-tests at start (SectionMockSubTests) and never touch this allowance.
	// An IELTS full mock spends it when it starts (full_mock_sessions), so its
	// attempt row is not counted a second time. A PTE full mock does the same
	// through pte_mock_sessions; its attempt rows are marked generated.
	var subTestsUsed, mocksUsed int
	if plan.ID == "free" {
		err = db.QueryRow(ctx, `
			SELECT `+subTestsToday+`,
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic AND NOT m.is_generated AND m.id <> '`+FullMockID+`')
				+ (SELECT count(*) FROM full_mock_sessions f WHERE f.user_id = $1)
				+ (SELECT count(*) FROM pte_mock_sessions p WHERE p.user_id = $1 AND p.kind = 'full')`,
			user.ID, dayStart, dayEnd).Scan(&subTestsUsed, &mocksUsed)
	} else {
		err = db.QueryRow(ctx, `
			SELECT `+subTestsToday+`,
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic AND NOT m.is_generated AND m.id <> '`+FullMockID+`'
				    AND ma.completed_at >= date_trunc('month', now()))
				+ (SELECT count(*) FROM full_mock_sessions f
				    WHERE f.user_id = $1 AND f.created_at >= date_trunc('month', now()))
				+ (SELECT count(*) FROM pte_mock_sessions p
				    WHERE p.user_id = $1 AND p.kind = 'full' AND p.created_at >= date_trunc('month', now()))`,
			user.ID, dayStart, dayEnd).Scan(&subTestsUsed, &mocksUsed)
	}

	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("read usage: %w", err)
	}

	totalMockTestsAllowed := plan.MockTestsIncluded + user.BonusMockTests

	gradingStart, gradingRenews := gradingPeriod(user, plan)
	gradingUnits, err := gradingUnitsUsed(ctx, db, user.ID, gradingStart)
	if err != nil {
		return models.SubscriptionState{}, err
	}

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
		Unlimited:             plan.Unlimited,

		AIGradingsUsed:     float64(gradingUnits) / UnitsPerGrading,
		AIGradingsLimit:    plan.AIGradingsPerPeriod,
		AIGradingsRenewOn:  gradingRenews.Format(time.DateOnly),
		AIGradingsPeriod:   plan.AIGradingsPeriod,
		AIGradingUnitsUsed: gradingUnits,
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

// SectionMockSubTests is what one section mock (a full reading paper, and
// later writing, listening or speaking) costs from the daily allowance.
const SectionMockSubTests = 5

// CheckSubTestAllowance checks whether a new session can spend one credit.
// Submission uses RequireStartedSubTest instead; it must not need a second credit.
func (s *Service) CheckSubTestAllowance(ctx context.Context, db database.DB, user models.User, _ string) (models.SubscriptionState, error) {
	return s.CheckSubTestCredits(ctx, db, user, 1)
}

// CheckSubTestCredits checks whether a new session can spend `credits` at once.
// It is all or nothing: a section mock needing 5 is refused with 4 left, rather
// than started on credit the learner does not have.
func (s *Service) CheckSubTestCredits(ctx context.Context, db database.DB, user models.User, credits int) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.Unlimited || state.DailySubTestsUsed+credits <= state.DailySubTestsLimit {
		return state, nil
	}

	// Only once the day's allowance is really gone: a five-credit section mock
	// refused with four left is not "you have used everything".
	if state.DailySubTestsUsed >= state.DailySubTestsLimit {
		message := fmt.Sprintf("You have used all %d practice tasks for today. They reset at midnight.", state.DailySubTestsLimit)
		if !user.HasActivePaidPlan() {
			message += " Upgrade for a bigger daily allowance."
		}
		s.notifyOnce(ctx, notifications.CreateParams{
			UserID: user.ID, Type: notifications.TypeLimit,
			Title: "Daily practice limit reached", Message: message,
			ActionURL: "/subscription",
			DedupeKey: "limit:practice:" + gamification.LocalDay(user),
		})
	}
	return state, ErrLimitReached
}

// notifyOnce sends a notice on its own connection, outside the caller's
// transaction: it goes out exactly when the request is being refused, and that
// refusal rolls the transaction back. A failure is logged, never returned; a
// missed notice must not turn a clear "limit reached" into a server error.
func (s *Service) notifyOnce(ctx context.Context, p notifications.CreateParams) {
	if s.notifications == nil {
		return
	}
	if err := s.notifications.Notify(context.WithoutCancel(ctx), p); err != nil {
		slog.Default().Warn("limit notification failed", "type", p.Type, "error", err)
	}
}

var ErrSessionRequired = errors.New("start this task before submitting answers")

// RequireStartedSubTest verifies payment at start without checking remaining
// credits. A group session covers every question in the group. Active sessions
// may finish across midnight; stopped sessions remain answerable on their day
// because clients can stop the timer before sending their answers.
func (s *Service) RequireStartedSubTest(ctx context.Context, db database.DB, user models.User, question models.Question, exam models.ExamType) error {
	var started bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM practice_sessions
			WHERE user_id = $1 AND item_id = $2 AND exam = $3 AND skill = $4
			  AND (status = 'active' OR (created_at >= $5 AND created_at < $6)))`,
		user.ID, SubTestKeyForQuestion(question), string(exam), string(question.Skill),
		gamification.LocalDayStart(user), gamification.LocalDayEnd(user),
	).Scan(&started)
	if err != nil {
		return fmt.Errorf("read started sub-test: %w", err)
	}
	if !started {
		return ErrSessionRequired
	}
	return nil
}

// CheckMockAllowance checks if the learner has available full mock tests in their quota.
func (s *Service) CheckMockAllowance(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	// Full mocks are counted on every plan, unlimited ones included.
	if state.MockTestsUsed >= state.TotalMockTestsAllowed {
		validUntil := "none"
		if user.PlanValidUntil != nil {
			validUntil = user.PlanValidUntil.Format(time.DateOnly)
		}
		s.notifyOnce(ctx, notifications.CreateParams{
			UserID: user.ID, Type: notifications.TypeLimit,
			Title:     "All mock tests used",
			Message:   fmt.Sprintf("You have taken all %d full mock tests in your plan. Upgrade or renew to take more.", state.TotalMockTestsAllowed),
			ActionURL: "/subscription",
			// Once per plan period, not once a day: the allowance does not reset daily.
			DedupeKey: "limit:mock:" + user.PlanID + ":" + validUntil,
		})
		return state, ErrMockLimitReached
	}
	return state, nil
}

// RecordSessionStart records the start of a practice task or sub-test session, consuming allowance.
func (s *Service) RecordSessionStart(ctx context.Context, db database.DB, user models.User, exam, skill, itemID string) (string, error) {
	return s.RecordSessionStartCredits(ctx, db, user, exam, skill, itemID, 1)
}

// RecordSessionStartCredits records a session that spends `credits` sub-tests.
func (s *Service) RecordSessionStartCredits(ctx context.Context, db database.DB, user models.User, exam, skill, itemID string, credits int) (string, error) {
	var sessionID string
	err := db.QueryRow(ctx, `
		INSERT INTO practice_sessions (user_id, exam, skill, item_id, status, credits)
		VALUES ($1, $2, $3, $4, 'active', $5)
		RETURNING id::text`,
		user.ID, exam, skill, itemID, credits,
	).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("record session start: %w", err)
	}
	return sessionID, nil
}

// RecordSessionStop marks an active practice session as stopped.
func (s *Service) RecordSessionStop(ctx context.Context, db database.DB, user models.User, sessionID string) error {
	_, err := db.Exec(ctx, `
		UPDATE practice_sessions
		   SET status = 'stopped', stopped_at = now()
		 WHERE id = $1 AND user_id = $2 AND status = 'active'`,
		sessionID, user.ID,
	)
	if err != nil {
		return fmt.Errorf("record session stop: %w", err)
	}
	return nil
}

type ConfirmPaymentParams struct {
	UserID         string
	PlanID         string
	PaymentGateway string
	TransactionID  string
	AmountNPR      int
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

// ErrDuplicateTransaction means the learner has already submitted that
// transaction id. The unique index is what catches it, so two taps of the same
// button cannot become two requests for the same money.
var ErrDuplicateTransaction = errors.New("transaction already submitted")

type RequestPaymentParams struct {
	UserID         string
	Plan           models.Plan
	PaymentGateway string
	TransactionID  string
	PhoneNumber    string
	ProofImage     []byte
	ProofImageType string
}

// RequestPayment records a purchase waiting for review.
//
// It writes the entitlement it is asking for — the plan and the days — but
// grants nothing. Storing the days now means an approval months later gives
// what was advertised at the time of payment, not whatever the plan has since
// become.
func (s *Service) RequestPayment(ctx context.Context, pool *pgxpool.Pool, p RequestPaymentParams) error {
	baseDays := p.Plan.DurationDays
	if baseDays <= 0 && p.Plan.DurationMonths > 0 {
		baseDays = p.Plan.DurationMonths * 30
	}
	effectiveDays := baseDays + p.Plan.BonusDays

	var proof []byte
	var proofType *string
	if len(p.ProofImage) > 0 {
		proof = p.ProofImage
		proofType = &p.ProofImageType
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = referrals.LockLifecycle(ctx, tx); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO subscription_payments
			(user_id, plan_id, payment_gateway, transaction_id, phone_number, amount_npr, status,
			 base_days, bonus_days, effective_days, proof_image, proof_image_type)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8, $9, $10, $11)`,
		p.UserID, p.Plan.ID, p.PaymentGateway, p.TransactionID, p.PhoneNumber, p.Plan.PriceNPR,
		baseDays, p.Plan.BonusDays, effectiveDays, proof, proofType)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateTransaction
		}
		return fmt.Errorf("record payment request: %w", err)
	}

	// The learner hears that it arrived; every admin hears there is one to check.
	if _, err := notifications.Send(ctx, tx, notifications.CreateParams{
		UserID:    p.UserID,
		Type:      notifications.TypePayment,
		Title:     "Payment received for " + p.Plan.Name,
		Message:   "We are checking it now and will activate your plan within a few hours.",
		ActionURL: "/subscription",
	}); err != nil {
		return err
	}
	var learner string
	if err := tx.QueryRow(ctx, `SELECT name FROM users WHERE id = $1`, p.UserID).Scan(&learner); err != nil {
		return fmt.Errorf("read payer name: %w", err)
	}
	if err := notifications.SendToAdmins(ctx, tx, notifications.CreateParams{
		Type:      notifications.TypePayment,
		Title:     "New payment to review",
		Message:   fmt.Sprintf("%s submitted NPR %d for %s via %s.", learner, p.Plan.PriceNPR, p.Plan.Name, p.PaymentGateway),
		ActionURL: "/admin/payments",
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// QueuedPlan is a paid plan waiting for the current one to end.
type QueuedPlan struct {
	ID        string `json:"id"`
	PlanID    string `json:"planId"`
	PlanName  string `json:"planName"`
	Days      int    `json:"days"`
	CreatedAt string `json:"createdAt"`
}

// QueuedPlans lists what a learner has bought and not yet started.
func (s *Service) QueuedPlans(ctx context.Context, db database.DB, userID string) ([]QueuedPlan, error) {
	rows, err := db.Query(ctx, `
		SELECT q.id, q.plan_id, p.name, q.days, q.created_at
		FROM queued_plans q
		JOIN plans p ON p.id = q.plan_id
		WHERE q.user_id = $1 AND q.status = 'queued'
		ORDER BY q.created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list queued plans: %w", err)
	}
	defer rows.Close()

	list := make([]QueuedPlan, 0, 2)
	for rows.Next() {
		var item QueuedPlan
		var created time.Time
		if err := rows.Scan(&item.ID, &item.PlanID, &item.PlanName, &item.Days, &created); err != nil {
			return nil, fmt.Errorf("scan queued plan: %w", err)
		}
		item.CreatedAt = created.Format(time.RFC3339)
		list = append(list, item)
	}
	return list, rows.Err()
}

// ErrQueuedPlanNotFound means the row is gone, already started, or belongs to
// somebody else. One error for all three: the caller may only act on their own
// queue, and saying which it was would confirm another learner's row exists.
var ErrQueuedPlanNotFound = errors.New("queued plan not found")

// ActivateQueuedPlan starts a waiting plan now, ending the current one.
//
// Whatever was left of the running plan is forfeited — the new plan's expiry is
// counted from today, not added to what remained. That is a real loss, so it
// only ever happens because the learner asked: nothing calls this on their
// behalf.
func (s *Service) ActivateQueuedPlan(ctx context.Context, pool *pgxpool.Pool, userID, queuedID string) (models.User, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("begin activation tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var planID string
	var days int
	err = tx.QueryRow(ctx, `
		SELECT plan_id, days FROM queued_plans
		WHERE id = $1 AND user_id = $2 AND status = 'queued'
		FOR UPDATE`, queuedID, userID).Scan(&planID, &days)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrQueuedPlanNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("read queued plan: %w", err)
	}

	// The updated row comes back from the write. Reporting the caller's copy
	// instead would answer with the plan they had a moment ago, which is the
	// one this call just replaced.
	var updated models.User
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET plan_id = $2,
		    plan_started_at = CURRENT_DATE,
		    plan_valid_until = CURRENT_DATE + make_interval(days => $3),
		    role = CASE WHEN role = 'admin' THEN 'admin' ELSE $4 END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, name, role, target_exam, target_score, exam_date, nepal_region,
		          xp, streak_days, streak_last_active_date, timezone, plan_id, plan_started_at,
		          plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at,
		          target_module`,
		userID, planID, days, models.RoleForPlan(planID)).
		Scan(&updated.ID, &updated.Email, &updated.Name, &updated.Role, &updated.TargetExam,
			&updated.TargetScore, &updated.ExamDate, &updated.NepalRegion, &updated.XP,
			&updated.StreakDays, &updated.StreakLastActiveDate, &updated.Timezone,
			&updated.PlanID, &updated.PlanStartedAt, &updated.PlanValidUntil,
			&updated.ReferralCode, &updated.BonusMockTests, &updated.BonusProDays, &updated.CreatedAt,
			&updated.TargetModule)
	if err != nil {
		return models.User{}, fmt.Errorf("activate plan: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE queued_plans SET status = 'activated', activated_at = now()
		WHERE id = $1`, queuedID); err != nil {
		return models.User{}, fmt.Errorf("mark queued plan activated: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, fmt.Errorf("commit activation: %w", err)
	}
	return updated, nil
}
