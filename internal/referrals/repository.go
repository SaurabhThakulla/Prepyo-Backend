package referrals

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var (
	ErrNotFound       = errors.New("referral not found")
	ErrSelfReferral   = errors.New("cannot refer yourself")
	ErrAlreadyReferred = errors.New("user has already been referred")
	ErrCodeNotFound   = errors.New("referral code not found")
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

// UserByCode finds the referrer associated with a given referral code.
func (r *Repository) UserByCode(ctx context.Context, db database.DB, code string) (models.User, error) {
	norm := NormalizeCode(code)
	var u models.User
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, target_exam, target_score, exam_date,
		       nepal_region, xp, streak_days, streak_last_active_date, timezone,
		       plan_id, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at
		FROM users
		WHERE referral_code = $1`, norm).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.TargetExam,
			&u.TargetScore, &u.ExamDate, &u.NepalRegion, &u.XP, &u.StreakDays,
			&u.StreakLastActiveDate, &u.Timezone, &u.PlanID, &u.PlanValidUntil,
			&u.ReferralCode, &u.BonusMockTests, &u.BonusProDays, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrCodeNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("find user by referral code: %w", err)
	}
	return u, nil
}

// Create records a new pending referral link between referrer and referee.
func (r *Repository) Create(ctx context.Context, db database.DB, ref models.Referral) (models.Referral, error) {
	if ref.ReferrerID == ref.RefereeID {
		return models.Referral{}, ErrSelfReferral
	}

	var created models.Referral
	err := db.QueryRow(ctx, `
		INSERT INTO referrals (
			referrer_id, referee_id, referral_code, status,
			reward_referrer_xp, reward_referee_xp
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, referrer_id, referee_id, referral_code, status,
		          reward_referrer_xp, reward_referee_xp, created_at, completed_at`,
		ref.ReferrerID, ref.RefereeID, NormalizeCode(ref.ReferralCode),
		ref.Status, ref.RewardReferrerXP, ref.RewardRefereeXP,
	).Scan(&created.ID, &created.ReferrerID, &created.RefereeID, &created.ReferralCode,
		&created.Status, &created.RewardReferrerXP, &created.RewardRefereeXP,
		&created.CreatedAt, &created.CompletedAt)

	if err != nil {
		return models.Referral{}, fmt.Errorf("create referral: %w", err)
	}
	return created, nil
}

// ByRefereeID fetches any referral where the specified user is the referee (with lock for qualification).
func (r *Repository) ByRefereeIDForUpdate(ctx context.Context, db database.DB, refereeID string) (models.Referral, error) {
	var ref models.Referral
	err := db.QueryRow(ctx, `
		SELECT id, referrer_id, referee_id, referral_code, status,
		       reward_referrer_xp, reward_referee_xp, created_at, completed_at
		FROM referrals
		WHERE referee_id = $1
		FOR UPDATE`, refereeID).
		Scan(&ref.ID, &ref.ReferrerID, &ref.RefereeID, &ref.ReferralCode,
			&ref.Status, &ref.RewardReferrerXP, &ref.RewardRefereeXP,
			&ref.CreatedAt, &ref.CompletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Referral{}, ErrNotFound
	}
	if err != nil {
		return models.Referral{}, fmt.Errorf("find referral by referee: %w", err)
	}
	return ref, nil
}

// MarkCompleted transitions a referral from pending to completed.
func (r *Repository) MarkCompleted(ctx context.Context, db database.DB, referralID string) error {
	tag, err := db.Exec(ctx, `
		UPDATE referrals
		SET status = 'completed', completed_at = now()
		WHERE id = $1 AND status = 'pending'`, referralID)
	if err != nil {
		return fmt.Errorf("mark referral completed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetCompletedCount returns total number of successfully qualified referrals for a referrer.
func (r *Repository) GetCompletedCount(ctx context.Context, db database.DB, referrerID string) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT count(*)
		FROM referrals
		WHERE referrer_id = $1 AND status = 'completed'`, referrerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count completed referrals: %w", err)
	}
	return count, nil
}

// MonthlyReferralCount checks how many referrals occurred in the current month for fraud throttling.
func (r *Repository) MonthlyReferralCount(ctx context.Context, db database.DB, referrerID string) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT count(*)
		FROM referrals
		WHERE referrer_id = $1
		  AND created_at >= date_trunc('month', now())`, referrerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count monthly referrals: %w", err)
	}
	return count, nil
}

// IncrementBonusMockTests grants an extra free mock test entitlement.
func (r *Repository) IncrementBonusMockTests(ctx context.Context, db database.DB, userID string, amount int) error {
	_, err := db.Exec(ctx, `
		UPDATE users
		SET bonus_mock_tests = bonus_mock_tests + $2, updated_at = now()
		WHERE id = $1`, userID, amount)
	if err != nil {
		return fmt.Errorf("increment bonus mock tests: %w", err)
	}
	return nil
}

// AddBonusProDays extends or grants Pro subscription days.
func (r *Repository) AddBonusProDays(ctx context.Context, db database.DB, userID string, days int) error {
	_, err := db.Exec(ctx, `
		UPDATE users
		SET bonus_pro_days = bonus_pro_days + $2,
		    plan_valid_until = COALESCE(GREATEST(plan_valid_until, CURRENT_DATE), CURRENT_DATE) + ($2 || ' days')::INTERVAL,
		    updated_at = now()
		WHERE id = $1`, userID, days)
	if err != nil {
		return fmt.Errorf("add bonus pro days: %w", err)
	}
	return nil
}

// Overview aggregates the complete referral profile for the authenticated learner.
func (r *Repository) Overview(ctx context.Context, db database.DB, user models.User, webAppURL string) (models.ReferralOverview, error) {
	var totalInvited, pendingCount, completedCount, totalXpEarned int
	err := db.QueryRow(ctx, `
		SELECT
			count(*),
			count(*) FILTER (WHERE status = 'pending'),
			count(*) FILTER (WHERE status = 'completed'),
			COALESCE(sum(reward_referrer_xp) FILTER (WHERE status = 'completed'), 0)
		FROM referrals
		WHERE referrer_id = $1`, user.ID).
		Scan(&totalInvited, &pendingCount, &completedCount, &totalXpEarned)
	if err != nil {
		return models.ReferralOverview{}, fmt.Errorf("aggregate referral stats: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT r.id, u.name, r.status, r.created_at, r.completed_at
		FROM referrals r
		JOIN users u ON r.referee_id = u.id
		WHERE r.referrer_id = $1
		ORDER BY r.created_at DESC
		LIMIT 15`, user.ID)
	if err != nil {
		return models.ReferralOverview{}, fmt.Errorf("list recent referrals: %w", err)
	}
	defer rows.Close()

	recent := []models.RecentReferralItem{}
	for rows.Next() {
		var item models.RecentReferralItem
		var rawName string
		if err := rows.Scan(&item.ID, &rawName, &item.Status, &item.CreatedAt, &item.CompletedAt); err != nil {
			return models.ReferralOverview{}, fmt.Errorf("scan recent referral: %w", err)
		}
		item.FriendName = maskName(rawName)
		recent = append(recent, item)
	}

	if webAppURL == "" {
		webAppURL = "https://prepyo.com"
	}
	shareLink := fmt.Sprintf("%s/signup?ref=%s", strings.TrimRight(webAppURL, "/"), user.ReferralCode)

	return models.ReferralOverview{
		ReferralCode: user.ReferralCode,
		ShareLink:    shareLink,
		Stats: models.ReferralStats{
			TotalInvited:  totalInvited,
			Pending:       pendingCount,
			Completed:     completedCount,
			TotalXPEarned: totalXpEarned,
		},
		Milestones: models.ReferralMilestones{
			ThreeReferrals: models.ReferralMilestone{
				Required:  3,
				Current:   completedCount,
				Completed: completedCount >= 3,
			},
			FiveReferrals: models.ReferralMilestone{
				Required:  5,
				Current:   completedCount,
				Completed: completedCount >= 5,
			},
		},
		RecentReferrals: recent,
	}, nil
}

// maskName hides sensitive personal details in the referee list (e.g., "Aprapya Rana" -> "A*** R***").
func maskName(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "Learner"
	}
	var masked []string
	for _, p := range parts {
		runes := []rune(p)
		if len(runes) <= 1 {
			masked = append(masked, string(runes)+"*")
		} else {
			masked = append(masked, string(runes[0])+"***")
		}
	}
	return strings.Join(masked, " ")
}
