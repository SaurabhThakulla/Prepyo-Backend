// Package users owns the users table: lookups, profile updates and XP.
//
// Credentials and sessions live in internal/auth, which uses this package for
// the user row itself.
package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailTaken    = errors.New("email already registered")
	uniqueViolation  = "23505"
	selectUserFields = `
		id, email, password_hash, name, role, target_exam, target_score, exam_date,
		nepal_region, xp, streak_days, streak_last_active_date, timezone,
		plan_id, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at`
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type CreateParams struct {
	Email        string
	PasswordHash string
	Name         string
	NepalRegion  string
	Timezone     string
	ReferralCode string
}

func (r *Repository) Create(ctx context.Context, p CreateParams) (models.User, error) {
	return r.CreateTx(ctx, r.db, p)
}

func (r *Repository) CreateTx(ctx context.Context, db database.DB, p CreateParams) (models.User, error) {
	row := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, nepal_region, timezone, referral_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+selectUserFields,
		p.Email, p.PasswordHash, p.Name, p.NepalRegion, p.Timezone, p.ReferralCode)

	user, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *Repository) ByID(ctx context.Context, id string) (models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectUserFields+` FROM users WHERE id = $1`, id)
	return r.scanOne(row, "by id")
}

func (r *Repository) ByEmail(ctx context.Context, email string) (models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectUserFields+` FROM users WHERE email = lower($1)`, email)
	return r.scanOne(row, "by email")
}

// UpdateProfile writes the onboarding and goal fields. Every argument is
// optional; a nil pointer leaves that column alone.
type UpdateProfileParams struct {
	Name        *string
	TargetExam  *models.ExamType
	TargetScore *float64
	ExamDate    *time.Time
	NepalRegion *string
	Timezone    *string
}

func (r *Repository) UpdateProfile(ctx context.Context, userID string, p UpdateProfileParams) (models.User, error) {
	// COALESCE keeps the existing value when the parameter is NULL, which means
	// one statement covers every combination of supplied fields.
	row := r.db.QueryRow(ctx, `
		UPDATE users SET
			name         = COALESCE($2, name),
			target_exam  = COALESCE($3, target_exam),
			target_score = COALESCE($4, target_score),
			exam_date    = COALESCE($5, exam_date),
			nepal_region = COALESCE($6, nepal_region),
			timezone     = COALESCE($7, timezone),
			updated_at   = now()
		WHERE id = $1
		RETURNING `+selectUserFields,
		userID, p.Name, p.TargetExam, p.TargetScore, p.ExamDate, p.NepalRegion, p.Timezone)

	return r.scanOne(row, "update profile")
}

func (r *Repository) SetPlan(ctx context.Context, userID, planID string, validUntil *time.Time) (models.User, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE users SET plan_id = $2, plan_valid_until = $3, updated_at = now()
		WHERE id = $1
		RETURNING `+selectUserFields,
		userID, planID, validUntil)

	return r.scanOne(row, "set plan")
}

// Delete removes the account. Every learner table cascades from users, so this
// is a full account deletion.
func (r *Repository) Delete(ctx context.Context, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) scanOne(row pgx.Row, op string) (models.User, error) {
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func scanUser(row pgx.Row) (models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.TargetExam,
		&u.TargetScore, &u.ExamDate, &u.NepalRegion, &u.XP, &u.StreakDays,
		&u.StreakLastActiveDate, &u.Timezone, &u.PlanID, &u.PlanValidUntil,
		&u.ReferralCode, &u.BonusMockTests, &u.BonusProDays,
		&u.CreatedAt,
	)
	return u, err
}
