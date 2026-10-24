package institution

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/referrals"
)

var (
	errSeatsFull = errors.New("institution has no free seats")
	errNoPlan    = errors.New("institution has no plan")
)

// Student is a learner an institution added.
type Student struct {
	ID        string
	Name      string
	Username  string
	CreatedAt time.Time
	Disabled  bool
	Usage     Usage
}

// students lists an institution's students, oldest first, with what each did since `since` (nil = all time).
func (s *store) students(ctx context.Context, institutionID string, since *time.Time) ([]Student, error) {
	rows, err := s.pool.Query(ctx, `WITH`+StudentUsageSQL+`
		SELECT u.id, u.name, st.username, st.created_at, st.disabled_at IS NOT NULL,
			su.practice, su.mocks, su.evaluations, su.tutor, su.last_active,
			(SELECT COALESCE(SUM(g.units), 0) FROM ai_grading_usage g
			  WHERE g.user_id = st.user_id AND g.created_at >= date_trunc('month', now() AT TIME ZONE 'UTC') AT TIME ZONE 'UTC')
		FROM institution_students st
		JOIN users u ON u.id = st.user_id
		JOIN student_usage su ON su.user_id = st.user_id
		WHERE st.institution_id = $2
		ORDER BY st.created_at`, since, institutionID)
	if err != nil {
		return nil, fmt.Errorf("list institution students: %w", err)
	}
	defer rows.Close()

	list := []Student{}
	for rows.Next() {
		var student Student
		var units int
		u := &student.Usage
		if err := rows.Scan(&student.ID, &student.Name, &student.Username, &student.CreatedAt, &student.Disabled,
			&u.PracticeAttempts, &u.MocksCompleted, &u.AIEvaluations, &u.TutorMessages, &u.LastActiveAt, &units); err != nil {
			return nil, fmt.Errorf("scan institution student: %w", err)
		}
		u.AIGradingsThisMonth = float64(units) / unitsPerGrading
		list = append(list, student)
	}
	return list, rows.Err()
}

// addStudent creates a learner account for the institution, or returns errSeatsFull.
func (s *store) addStudent(ctx context.Context, institutionID, name, passwordHash string) (Student, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Student{}, fmt.Errorf("begin add student: %w", err)
	}
	defer tx.Rollback(ctx)

	// Locking the institution makes two adds at once take turns, so neither slips past the seat limit.
	var plan, exam string
	if err := tx.QueryRow(ctx, `SELECT COALESCE(plan, ''), COALESCE(exams[1], 'PTE') FROM institutions WHERE id = $1 FOR UPDATE`, institutionID).
		Scan(&plan, &exam); err != nil {
		return Student{}, fmt.Errorf("lock institution: %w", err)
	}
	seats := Institution{Plan: plan}.Seats()
	if seats == 0 {
		return Student{}, errNoPlan
	}

	taken, err := takenUsernames(ctx, tx, institutionID)
	if err != nil {
		return Student{}, err
	}
	if len(taken) >= seats {
		return Student{}, errSeatsFull
	}

	student := Student{Name: name, Username: freeUsername(name, taken)}
	if student.ID, err = insertLearner(ctx, tx, name, exam); err != nil {
		return Student{}, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO institution_students (user_id, institution_id, username, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		student.ID, institutionID, student.Username, passwordHash).Scan(&student.CreatedAt)
	if err != nil {
		return Student{}, fmt.Errorf("create institution student: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Student{}, fmt.Errorf("commit add student: %w", err)
	}
	return student, nil
}

func takenUsernames(ctx context.Context, tx pgx.Tx, institutionID string) (map[string]bool, error) {
	rows, err := tx.Query(ctx, `SELECT lower(username) FROM institution_students WHERE institution_id = $1`, institutionID)
	if err != nil {
		return nil, fmt.Errorf("read usernames: %w", err)
	}
	defer rows.Close()

	taken := map[string]bool{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("scan username: %w", err)
		}
		taken[username] = true
	}
	return taken, rows.Err()
}

// insertLearner creates the row in users that makes the student an ordinary learner.
// Students skip the learner onboarding: they start on the institution's first exam.
func insertLearner(ctx context.Context, tx pgx.Tx, name, exam string) (string, error) {
	// Every user needs a unique referral code; a clash is rare, so a few tries are plenty.
	for attempt := 0; attempt < 10; attempt++ {
		code, err := referrals.GenerateCode()
		if err != nil {
			return "", err
		}

		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO users (name, nepal_region, timezone, referral_code, target_exam, onboarding_completed_at)
			VALUES ($1, 'Kathmandu', 'Asia/Kathmandu', $2, $3, now())
			ON CONFLICT (upper(btrim(referral_code))) DO NOTHING
			RETURNING id`, name, referrals.NormalizeCode(code), exam).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("create learner: %w", err)
		}
		return id, nil
	}
	return "", errors.New("create learner: no free referral code")
}

// studentForLogin finds a student by institution code and username, with the password hash.
func (s *store) studentForLogin(ctx context.Context, code, username string) (userID, passwordHash string, disabled bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT st.user_id, st.password_hash, st.disabled_at IS NOT NULL
		FROM institution_students st
		JOIN institutions i ON i.id = st.institution_id
		WHERE i.code = $1 AND lower(st.username) = $2`,
		code, strings.ToLower(username)).Scan(&userID, &passwordHash, &disabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, errNotFound
	}
	if err != nil {
		return "", "", false, fmt.Errorf("find institution student: %w", err)
	}
	return userID, passwordHash, disabled, nil
}

// setStudentDisabled turns a student's account off (signing them out everywhere) or back on.
func (s *store) setStudentDisabled(ctx context.Context, institutionID, userID string, disabled bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin disable student: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE institution_students SET disabled_at = CASE WHEN $3 THEN COALESCE(disabled_at, now()) END
		WHERE institution_id = $1 AND user_id::text = $2`, institutionID, userID, disabled)
	if err != nil {
		return fmt.Errorf("disable student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	if disabled {
		if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id::text = $1`, userID); err != nil {
			return fmt.Errorf("end student sessions: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit disable student: %w", err)
	}
	return nil
}

// setStudentPassword changes a student's password and signs them out everywhere.
func (s *store) setStudentPassword(ctx context.Context, institutionID, userID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE institution_students SET password_hash = $3
		WHERE institution_id = $1 AND user_id = $2`, institutionID, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("set student password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("end student sessions: %w", err)
	}
	return nil
}

// removeStudent deletes the student's whole learner account, which frees the seat.
func (s *store) removeStudent(ctx context.Context, institutionID, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM users
		WHERE id = $2 AND id IN (SELECT user_id FROM institution_students WHERE institution_id = $1)`,
		institutionID, userID)
	if err != nil {
		return fmt.Errorf("remove institution student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}
