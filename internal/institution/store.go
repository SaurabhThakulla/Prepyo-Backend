package institution

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/notifications"
)

var (
	errNotFound   = errors.New("institution account not found")
	errEmailTaken = errors.New("email already has an institution account")
)

// store holds every SQL statement the package runs.
type store struct {
	pool *pgxpool.Pool
}

// accountColumns is read by scanAccount; a = institution_admins, i = institutions, l = institution_logos.
const accountColumns = `a.id, a.name, a.email, i.id, i.name, i.code, i.phone, i.valid_until, i.seat_limit,
	i.exams, i.city, i.branches, i.monthly_students, i.brand_color, i.onboarded_at, l.updated_at`

func scanAccount(row pgx.Row, extra ...any) (Account, error) {
	var account Account
	fields := []any{
		&account.Admin.ID, &account.Admin.Name, &account.Admin.Email,
		&account.Institution.ID, &account.Institution.Name, &account.Institution.Code,
		&account.Institution.Phone, &account.Institution.ValidUntil, &account.Institution.SeatLimit,
		&account.Institution.Profile.Exams, &account.Institution.Profile.City, &account.Institution.Profile.Branches,
		&account.Institution.Profile.MonthlyStudents, &account.Institution.Profile.BrandColor,
		&account.Institution.OnboardedAt, &account.Institution.LogoUpdatedAt,
	}
	err := row.Scan(append(fields, extra...)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, errNotFound
	}
	return account, err
}

// createAccount stores an institution with its admin and notifies the Prepyo admins, all or nothing.
func (s *store) createAccount(ctx context.Context, form SignupForm, passwordHash string) (Account, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Account{}, fmt.Errorf("begin institution signup: %w", err)
	}
	defer tx.Rollback(ctx)

	institution, err := insertInstitution(ctx, tx, form)
	if err != nil {
		return Account{}, err
	}

	admin := Admin{Name: form.Name, Email: form.Email}
	err = tx.QueryRow(ctx, `
		INSERT INTO institution_admins (institution_id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		institution.ID, admin.Name, admin.Email, passwordHash).Scan(&admin.ID)
	if isUniqueViolation(err) {
		return Account{}, errEmailTaken
	}
	if err != nil {
		return Account{}, fmt.Errorf("create institution admin: %w", err)
	}

	err = notifications.SendToAdmins(ctx, tx, notifications.CreateParams{
		Type:      notifications.TypeSystem,
		Title:     "New institution signed up",
		Message:   fmt.Sprintf("%s (%s, %s) signed up on the free plan.", institution.Name, admin.Name, institution.Phone),
		ActionURL: "/admin",
	})
	if err != nil {
		return Account{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Account{}, fmt.Errorf("commit institution signup: %w", err)
	}
	return Account{Admin: admin, Institution: institution}, nil
}

// insertInstitution stores the institution; when its code is taken it tries name-2, name-3...
func insertInstitution(ctx context.Context, tx pgx.Tx, form SignupForm) (Institution, error) {
	institution := Institution{Name: form.InstitutionName, Phone: form.Phone, Profile: Profile{Exams: []string{}, Branches: 1}}
	base := CodeFromName(form.InstitutionName)

	for attempt := 1; attempt <= 50; attempt++ {
		code := base
		if attempt > 1 {
			code = fmt.Sprintf("%s-%d", base, attempt)
		}

		err := tx.QueryRow(ctx, `
			INSERT INTO institutions (name, code, phone) VALUES ($1, $2, $3)
			ON CONFLICT (code) DO NOTHING
			RETURNING id, seat_limit`,
			institution.Name, code, institution.Phone).Scan(&institution.ID, &institution.SeatLimit)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // code taken, try the next one
		}
		if err != nil {
			return Institution{}, fmt.Errorf("create institution: %w", err)
		}
		institution.Code = code
		return institution, nil
	}
	return Institution{}, fmt.Errorf("create institution: no free address for %q", base)
}

// accountForLogin finds the account for an email, with its password hash.
func (s *store) accountForLogin(ctx context.Context, email string) (Account, string, error) {
	var passwordHash string
	account, err := scanAccount(s.pool.QueryRow(ctx, `
		SELECT `+accountColumns+`, a.password_hash
		FROM institution_admins a
		JOIN institutions i ON i.id = a.institution_id
		LEFT JOIN institution_logos l ON l.institution_id = i.id
		WHERE a.email = $1`, email), &passwordHash)
	if err != nil && !errors.Is(err, errNotFound) {
		err = fmt.Errorf("find institution admin: %w", err)
	}
	return account, passwordHash, err
}

// accountBySession finds who a session belongs to. An expired session counts as none.
func (s *store) accountBySession(ctx context.Context, tokenHash []byte) (Account, error) {
	account, err := scanAccount(s.pool.QueryRow(ctx, `
		SELECT `+accountColumns+`
		FROM institution_sessions s
		JOIN institution_admins a ON a.id = s.admin_id
		JOIN institutions i ON i.id = a.institution_id
		LEFT JOIN institution_logos l ON l.institution_id = i.id
		WHERE s.token_hash = $1 AND s.expires_at > now()`, tokenHash))
	if err != nil && !errors.Is(err, errNotFound) {
		err = fmt.Errorf("find institution session: %w", err)
	}
	return account, err
}

// saveSession stores a new session and clears the admin's expired ones.
func (s *store) saveSession(ctx context.Context, adminID string, tokenHash []byte, expiresAt time.Time) error {
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM institution_sessions WHERE admin_id = $1 AND expires_at <= now()`, adminID); err != nil {
		return fmt.Errorf("clear expired institution sessions: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO institution_sessions (token_hash, admin_id, expires_at) VALUES ($1, $2, $3)`,
		tokenHash, adminID, expiresAt); err != nil {
		return fmt.Errorf("save institution session: %w", err)
	}
	return nil
}

func (s *store) deleteSession(ctx context.Context, tokenHash []byte) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM institution_sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete institution session: %w", err)
	}
	return nil
}

// isUniqueViolation reports whether a write failed because the value already exists.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
