// Package practice records single-question practice attempts.
//
// Reading and listening are graded here by internal/scoring. Speaking and
// writing need judgement, so they go to /api/v1/evaluations instead; this
// package refuses them rather than inventing a score.
package practice

import (
	"context"
	"fmt"
	"time"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type SaveParams struct {
	UserID             string
	QuestionID         string
	Exam               models.ExamType
	ExamVersionID      string
	IsCorrect          bool
	Score              float64
	MaxScore           float64
	AccuracyPercentage int
	UserResponse       string
	Feedback           string
	TimeSpentSeconds   int
}

// AttemptedBetween reports whether the learner already submitted this question
// in [from, to), which is how a resubmission is told apart from new work.
func (r *Repository) AttemptedBetween(ctx context.Context, db database.DB, userID, questionID string, from, to time.Time) (bool, error) {
	var seen bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM practice_attempts
			WHERE user_id = $1 AND question_id = $2 AND created_at >= $3 AND created_at < $4)`,
		userID, questionID, from, to).Scan(&seen)
	if err != nil {
		return false, fmt.Errorf("read earlier attempt: %w", err)
	}
	return seen, nil
}

func (r *Repository) Save(ctx context.Context, db database.DB, p SaveParams) (models.PracticeAttempt, error) {
	var a models.PracticeAttempt
	err := db.QueryRow(ctx, `
		INSERT INTO practice_attempts (
			user_id, question_id, exam, exam_version_id, is_correct, score, max_score,
			accuracy_percentage, user_response, feedback, time_spent_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, question_id, exam, exam_version_id, is_correct, score, max_score,
		          accuracy_percentage, COALESCE(feedback, ''), COALESCE(user_response, ''),
		          time_spent_seconds, created_at`,
		p.UserID, p.QuestionID, p.Exam, p.ExamVersionID, p.IsCorrect, p.Score, p.MaxScore,
		p.AccuracyPercentage, p.UserResponse, p.Feedback, p.TimeSpentSeconds,
	).Scan(&a.ID, &a.QuestionID, &a.Exam, &a.ExamVersionID, &a.IsCorrect, &a.Score, &a.MaxScore,
		&a.AccuracyPercentage, &a.Feedback, &a.UserResponse, &a.TimeSpentSeconds, &a.CreatedAt)
	if err != nil {
		return models.PracticeAttempt{}, fmt.Errorf("save practice attempt: %w", err)
	}
	return a, nil
}

type ListParams struct {
	UserID string
	Exam   models.ExamType
	Skill  models.SkillType
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.PracticeAttempt, int, error) {
	filter := `
		WHERE a.user_id = $1
		  AND ($2 = '' OR UPPER(a.exam) = UPPER($2))
		  AND ($3 = '' OR LOWER(q.skill) = LOWER($3))`

	var total int
	err := r.db.QueryRow(ctx, `
		SELECT count(*)
		FROM practice_attempts a
		LEFT JOIN questions q ON q.id = a.question_id`+filter,
		p.UserID, p.Exam, p.Skill).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count practice attempts: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.question_id, COALESCE(q.title, ''), COALESCE(q.skill, ''),
		       COALESCE(q.type_name, ''), COALESCE(q.prompt, ''),
		       a.exam, a.exam_version_id, a.is_correct, a.score, a.max_score,
		       a.accuracy_percentage, COALESCE(a.feedback, ''), COALESCE(a.user_response, ''),
		       a.time_spent_seconds, a.created_at
		FROM practice_attempts a
		LEFT JOIN questions q ON q.id = a.question_id`+filter+`
		ORDER BY a.created_at DESC
		LIMIT $4 OFFSET $5`,
		p.UserID, p.Exam, p.Skill, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list practice attempts: %w", err)
	}
	defer rows.Close()

	list := []models.PracticeAttempt{}
	for rows.Next() {
		var a models.PracticeAttempt
		var skill string
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.QuestionTitle, &skill,
			&a.TypeName, &a.Prompt,
			&a.Exam, &a.ExamVersionID, &a.IsCorrect, &a.Score,
			&a.MaxScore, &a.AccuracyPercentage, &a.Feedback, &a.UserResponse,
			&a.TimeSpentSeconds, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan practice attempt: %w", err)
		}
		a.Skill = models.SkillType(skill)
		list = append(list, a)
	}
	return list, total, rows.Err()
}
