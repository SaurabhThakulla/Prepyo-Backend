package referrals

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
)

const (
	RewardReferrerXP     = 200
	RewardRefereeXP      = 100
	DefaultMonthlyLimit  = 50
	MilestoneThreeReward = 1 // 1 free mock test
	MilestoneFiveDays    = 3 // 3 pro days
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
		Message:      "Valid referral code. Your friend earns 3 membership days after your eligible subscription payment is confirmed.",
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

// Overview fetches the authenticated user's referral summary.
func (s *Service) Overview(ctx context.Context, user models.User) (models.ReferralOverview, error) {
	return s.repo.Overview(ctx, s.pool, user, s.webAppURL)
}
