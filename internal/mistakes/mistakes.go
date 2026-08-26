// Package mistakes keeps the mistake bank: questions a learner got wrong,
// grouped so repeat failures show up as one entry with a rising count.
package mistakes

import (
	"context"
	"errors"
	"fmt"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("mistake not found")

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type RecordParams struct {
	UserID          string
	QuestionID      string
	ErrorTag        string
	UserResponse    string
	CorrectResponse string
	Explanation     string
}

// Record adds or updates a mistake.
//
// A second failure on the same question bumps failed_count and reopens the
// entry rather than creating a duplicate, which is what keeps the bank a list
// of recurring weaknesses instead of a log of every wrong click.
func (r *Repository) Record(ctx context.Context, db database.DB, p RecordParams) error {
	_, err := db.Exec(ctx, `
		INSERT INTO mistakes (user_id, question_id, error_tag, user_response, correct_response, explanation)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, question_id) DO UPDATE SET
			failed_count      = mistakes.failed_count + 1,
			error_tag         = EXCLUDED.error_tag,
			user_response     = EXCLUDED.user_response,
			resolved          = FALSE,
			last_attempted_at = now()`,
		p.UserID, p.QuestionID, p.ErrorTag, p.UserResponse, p.CorrectResponse, p.Explanation)
	if err != nil {
		return fmt.Errorf("record mistake: %w", err)
	}
	return nil
}

type ListParams struct {
	UserID         string
	Exam           models.ExamType
	Skill          models.SkillType
	UnresolvedOnly bool
	Limit          int
	Offset         int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.Mistake, int, error) {
	const filter = `
		WHERE m.user_id = $1
		  AND ($2 = '' OR q.exam = $2)
		  AND ($3 = '' OR q.skill = $3)
		  AND (NOT $4 OR NOT m.resolved)`

	var total int
	err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM mistakes m
		JOIN questions q ON q.id = m.question_id`+filter,
		p.UserID, p.Exam, p.Skill, p.UnresolvedOnly).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count mistakes: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.question_id, q.title, q.exam, q.skill, q.type_name, q.prompt,
		       m.user_response, m.correct_response, m.explanation, m.error_tag,
		       m.failed_count, m.resolved, m.last_attempted_at
		FROM mistakes m
		JOIN questions q ON q.id = m.question_id`+filter+`
		ORDER BY m.resolved, m.failed_count DESC, m.last_attempted_at DESC
		LIMIT $5 OFFSET $6`,
		p.UserID, p.Exam, p.Skill, p.UnresolvedOnly, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list mistakes: %w", err)
	}
	defer rows.Close()

	list := []models.Mistake{}
	for rows.Next() {
		var m models.Mistake
		if err := rows.Scan(&m.ID, &m.QuestionID, &m.QuestionTitle, &m.Exam, &m.Skill,
			&m.TypeName, &m.Prompt, &m.UserResponse, &m.CorrectResponse, &m.Explanation,
			&m.ErrorTag, &m.FailedCount, &m.Resolved, &m.LastAttemptedAt); err != nil {
			return nil, 0, fmt.Errorf("scan mistake: %w", err)
		}
		list = append(list, m)
	}
	return list, total, rows.Err()
}

// Resolve marks a mistake as handled.
//
// The user_id in the WHERE clause is the ownership check: another learner's id
// simply matches no rows, so there is no way to resolve someone else's entry.
func (r *Repository) Resolve(ctx context.Context, db database.DB, userID, mistakeID string) error {
	tag, err := db.Exec(ctx, `
		UPDATE mistakes SET resolved = TRUE
		WHERE id = $1 AND user_id = $2 AND NOT resolved`,
		mistakeID, userID)
	if err != nil {
		return fmt.Errorf("resolve mistake: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
