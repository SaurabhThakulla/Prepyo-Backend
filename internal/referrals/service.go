package referrals

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
)

const (
	RewardReferrerXP      = 200
	RewardRefereeXP       = 100
	DefaultMonthlyLimit   = 50
	MilestoneThreeReward  = 1 // 1 free mock test
	MilestoneFiveDays     = 3 // 3 pro days
)

type Service struct {
	pool          *pgxpool.Pool
	repo          *Repository
	xp            *gamification.Service
	notifications *notifications.Repository
	monthlyLimit  int
	webAppURL     string
	log           *slog.Logger
}

func NewService(
	pool *pgxpool.Pool,
	repo *Repository,
	xp *gamification.Service,
	notifs *notifications.Repository,
	webAppURL string,
	log *slog.Logger,
) *Service {
	return &Service{
		pool:          pool,
		repo:          repo,
		xp:            xp,
		notifications: notifs,
		monthlyLimit:  DefaultMonthlyLimit,
		webAppURL:     webAppURL,
		log:           log,
	}
}

// ValidateCode verifies if a referral code is valid and active.
func (s *Service) ValidateCode(ctx context.Context, code string) (models.ReferralValidation, error) {
	norm := NormalizeCode(code)
	if norm == "" {
		return models.ReferralValidation{
			Valid:   false,
			Message: "Please provide a referral code.",
		}, nil
	}

	user, err := s.repo.UserByCode(ctx, s.pool, norm)
	if err != nil {
		if errors.Is(err, ErrCodeNotFound) {
			return models.ReferralValidation{
				Valid:   false,
				Message: "Referral code not found. Please check and try again.",
			}, nil
		}
		return models.ReferralValidation{}, err
	}

	return models.ReferralValidation{
		Valid:        true,
		ReferrerName: maskName(user.Name),
		Message:      "Valid referral code from a Prepyo learner. You'll receive +100 XP upon completing your first activity!",
	}, nil
}

// LinkReferralOnRegister links a new user (referee) to an existing referrer during signup.
func (s *Service) LinkReferralOnRegister(ctx context.Context, db database.DB, refereeID, code string) (*models.Referral, error) {
	norm := NormalizeCode(code)
	if norm == "" {
		return nil, nil
	}

	referrer, err := s.repo.UserByCode(ctx, db, norm)
	if err != nil {
		if errors.Is(err, ErrCodeNotFound) {
			return nil, ErrCodeNotFound
		}
		return nil, err
	}

	if referrer.ID == refereeID {
		return nil, ErrSelfReferral
	}

	// Check monthly referral limits for abuse prevention
	count, err := s.repo.MonthlyReferralCount(ctx, db, referrer.ID)
	if err != nil {
		return nil, err
	}
	if count >= s.monthlyLimit {
		s.log.Warn("referrer hit monthly limit", "referrerId", referrer.ID, "limit", s.monthlyLimit)
		// We still record the relationship or cap it gracefully
	}

	created, err := s.repo.Create(ctx, db, models.Referral{
		ReferrerID:       referrer.ID,
		RefereeID:        refereeID,
		ReferralCode:     norm,
		Status:           models.ReferralPending,
		RewardReferrerXP: RewardReferrerXP,
		RewardRefereeXP:  RewardRefereeXP,
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// QualifyReferral checks and converts a pending referral into a completed one,
// awarding XP, milestone perks, and notifications. Idempotent against repeat triggers.
func (s *Service) QualifyReferral(ctx context.Context, refereeID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin qualification tx: %w", err)
	}
	defer tx.Rollback(ctx)

	ref, err := s.repo.ByRefereeIDForUpdate(ctx, tx, refereeID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// User was not referred; no referral to qualify
			return nil
		}
		return err
	}

	// If already completed or cancelled, do nothing (idempotent)
	if ref.Status != models.ReferralPending {
		return nil
	}

	// 1. Mark referral as completed
	if err := s.repo.MarkCompleted(ctx, tx, ref.ID); err != nil {
		return err
	}

	// 2. Award Referrer +200 XP with deterministic ledger key
	referrerSourceKey := fmt.Sprintf("referral:%s:referrer", ref.ID)
	_, err = s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    ref.ReferrerID,
		Amount:    ref.RewardReferrerXP,
		Reason:    "Friend qualified via practice/diagnostic",
		SourceKey: referrerSourceKey,
	})
	if err != nil {
		return fmt.Errorf("award referrer xp: %w", err)
	}

	// 3. Award Referee +100 XP with deterministic ledger key
	refereeSourceKey := fmt.Sprintf("referral:%s:referee", ref.ID)
	_, err = s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    ref.RefereeID,
		Amount:    ref.RewardRefereeXP,
		Reason:    "Welcome referral bonus",
		SourceKey: refereeSourceKey,
	})
	if err != nil {
		return fmt.Errorf("award referee xp: %w", err)
	}

	// 4. Send Notifications
	_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
		UserID:    ref.ReferrerID,
		Title:     "Friend Completed Activity!",
		Message:   "Your friend completed their first Prepyo activity. You earned +200 XP!",
		Type:      "referral",
		ActionURL: "/referrals",
	})

	_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
		UserID:    ref.RefereeID,
		Title:     "Welcome to Prepyo!",
		Message:   "Welcome to Prepyo! You received +100 XP from your referral.",
		Type:      "referral",
		ActionURL: "/dashboard",
	})

	// 5. Evaluate Milestones for Referrer
	completedCount, err := s.repo.GetCompletedCount(ctx, tx, ref.ReferrerID)
	if err != nil {
		return fmt.Errorf("get completed count: %w", err)
	}

	// Milestone 3: 1 Free AI Mock Test
	if completedCount >= 3 {
		milestone3Key := fmt.Sprintf("referral:milestone:%s:3", ref.ReferrerID)
		awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
			UserID:    ref.ReferrerID,
			Amount:    50, // bonus milestone XP
			Reason:    "Milestone: 3 qualified referrals",
			SourceKey: milestone3Key,
		})
		if err != nil {
			return fmt.Errorf("award milestone 3: %w", err)
		}
		if awarded > 0 {
			// First time hitting milestone 3 -> increment bonus mock tests
			if err := s.repo.IncrementBonusMockTests(ctx, tx, ref.ReferrerID, MilestoneThreeReward); err != nil {
				return fmt.Errorf("grant milestone mock test: %w", err)
			}
			_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
				UserID:    ref.ReferrerID,
				Title:     "Milestone Unlocked!",
				Message:   "You reached 3 qualified referrals! Your free AI Mock Test has been unlocked.",
				Type:      "referral",
				ActionURL: "/mocks",
			})
		}
	}

	// Milestone 5: 3 Pro Days
	if completedCount >= 5 {
		milestone5Key := fmt.Sprintf("referral:milestone:%s:5", ref.ReferrerID)
		awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
			UserID:    ref.ReferrerID,
			Amount:    100, // bonus milestone XP
			Reason:    "Milestone: 5 qualified referrals",
			SourceKey: milestone5Key,
		})
		if err != nil {
			return fmt.Errorf("award milestone 5: %w", err)
		}
		if awarded > 0 {
			// First time hitting milestone 5 -> add 3 Pro days
			if err := s.repo.AddBonusProDays(ctx, tx, ref.ReferrerID, MilestoneFiveDays); err != nil {
				return fmt.Errorf("grant milestone pro days: %w", err)
			}
			_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
				UserID:    ref.ReferrerID,
				Title:     "Milestone Unlocked!",
				Message:   "You reached 5 qualified referrals! 3 Pro days have been added to your account.",
				Type:      "referral",
				ActionURL: "/subscription",
			})
		}
	}

	return tx.Commit(ctx)
}

// Overview fetches the authenticated user's referral summary.
func (s *Service) Overview(ctx context.Context, user models.User) (models.ReferralOverview, error) {
	return s.repo.Overview(ctx, s.pool, user, s.webAppURL)
}
