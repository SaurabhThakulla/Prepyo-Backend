// Package practice records single-question practice attempts.
//
// Reading and listening are graded here by internal/scoring. Speaking and
// writing need judgement, so they go to /api/v1/evaluations instead; this
// package refuses them rather than inventing a score.
package practice

import (
	"context"
	"fmt"

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
	ExamVersionID      string
	IsCorrect          bool
	Score              float64
	MaxScore           float64
	AccuracyPercentage int
	UserResponse       string
	Feedback           string
	TimeSpentSeconds   int
}

func (r *Repository) Save(ctx context.Context, db database.DB, p SaveParams) (models.PracticeAttempt, error) {
	var a models.PracticeAttempt
	err := db.QueryRow(ctx, `
		INSERT INTO practice_attempts (
			user_id, question_id, exam_version_id, is_correct, score, max_score,
			accuracy_percentage, user_response, feedback, time_spent_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, question_id, exam_version_id, is_correct, score, max_score,
		          accuracy_percentage, COALESCE(feedback, ''), COALESCE(user_response, ''),
		          time_spent_seconds, created_at`,
		p.UserID, p.QuestionID, p.ExamVersionID, p.IsCorrect, p.Score, p.MaxScore,
		p.AccuracyPercentage, p.UserResponse, p.Feedback, p.TimeSpentSeconds,
	).Scan(&a.ID, &a.QuestionID, &a.ExamVersionID, &a.IsCorrect, &a.Score, &a.MaxScore,
		&a.AccuracyPercentage, &a.Feedback, &a.UserResponse, &a.TimeSpentSeconds, &a.CreatedAt)
	if err != nil {
		return models.PracticeAttempt{}, fmt.Errorf("save practice attempt: %w", err)
	}
	return a, nil
}

type ListParams struct {
	UserID string
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.PracticeAttempt, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM practice_attempts WHERE user_id = $1`, p.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count practice attempts: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, question_id, exam_version_id, is_correct, score, max_score,
		       accuracy_percentage, COALESCE(feedback, ''), COALESCE(user_response, ''),
		       time_spent_seconds, created_at
		FROM practice_attempts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		p.UserID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list practice attempts: %w", err)
	}
	defer rows.Close()

	list := []models.PracticeAttempt{}
	for rows.Next() {
		var a models.PracticeAttempt
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.ExamVersionID, &a.IsCorrect, &a.Score,
			&a.MaxScore, &a.AccuracyPercentage, &a.Feedback, &a.UserResponse,
			&a.TimeSpentSeconds, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan practice attempt: %w", err)
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}
