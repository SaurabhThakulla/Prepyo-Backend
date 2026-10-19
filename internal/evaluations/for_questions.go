package evaluations

import (
	"context"
	"fmt"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

// ForQuestions is one learner's own evaluations of these questions, newest first.
func ForQuestions(ctx context.Context, db database.DB, userID string, questionIDs []string, limit int) ([]models.Evaluation, error) {
	rows, err := db.Query(ctx, `
		SELECT `+selectFields+` FROM ai_evaluations
		WHERE user_id = $1 AND question_id = ANY($2)
		ORDER BY created_at DESC
		LIMIT $3`,
		userID, questionIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("list evaluations for questions: %w", err)
	}
	defer rows.Close()

	list := []models.Evaluation{}
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan evaluation: %w", err)
		}
		list = append(list, e)
	}
	return list, rows.Err()
}
