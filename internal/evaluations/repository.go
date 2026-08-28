// Package evaluations turns a learner's writing or speaking into stored AI
// feedback.
//
// It sits between the product and internal/ai and owns the three things the
// gateway should not: checking the learner's plan allowance before spending a
// call, refusing to pay twice for the same submission, and persisting the
// result with its usage.
package evaluations

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("evaluation not found")

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

const selectFields = `
	id, COALESCE(question_id, ''), exam, skill, evaluation_version, estimated_score,
	score_confidence, summary, criteria, strengths, weaknesses, sentence_feedback,
	COALESCE(model_rewrite, ''), transcript, created_at`

type SaveParams struct {
	UserID      string
	QuestionID  string
	Fingerprint []byte
	Evaluation  models.Evaluation
	Usage       models.EvaluationUsage
}

func (r *Repository) Save(ctx context.Context, db database.DB, p SaveParams) (models.Evaluation, error) {
	e := p.Evaluation

	// questionID is nullable in the schema; an empty string would break the
	// foreign key, so send NULL instead.
	var questionID *string
	if p.QuestionID != "" {
		questionID = &p.QuestionID
	}

	err := db.QueryRow(ctx, `
		INSERT INTO ai_evaluations (
			user_id, question_id, exam, skill, evaluation_version, request_fingerprint,
			estimated_score, score_confidence, summary, criteria, strengths, weaknesses,
			sentence_feedback, model_rewrite, transcript, provider, model, prompt_version,
			prompt_tokens, completion_tokens, latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING `+selectFields,
		p.UserID, questionID, e.Exam, e.Skill, e.EvaluationVersion, p.Fingerprint,
		e.EstimatedScore, e.ScoreConfidence, e.Summary, e.Criteria, e.Strengths, e.Weaknesses,
		e.SentenceFeedback, e.ModelRewrite, e.Transcript, p.Usage.Provider, p.Usage.Model, p.Usage.PromptVersion,
		p.Usage.PromptTokens, p.Usage.CompletionTokens, p.Usage.LatencyMS,
	).Scan(&e.ID, &e.QuestionID, &e.Exam, &e.Skill, &e.EvaluationVersion, &e.EstimatedScore,
		&e.ScoreConfidence, &e.Summary, &e.Criteria, &e.Strengths, &e.Weaknesses,
		&e.SentenceFeedback, &e.ModelRewrite, &e.Transcript, &e.CreatedAt)
	if err != nil {
		return models.Evaluation{}, fmt.Errorf("save evaluation: %w", err)
	}
	return e, nil
}

// ByFingerprint finds an earlier evaluation of the identical submission.
func (r *Repository) ByFingerprint(ctx context.Context, userID string, fingerprint []byte) (models.Evaluation, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+selectFields+` FROM ai_evaluations
		WHERE user_id = $1 AND request_fingerprint = $2`,
		userID, fingerprint)

	e, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Evaluation{}, ErrNotFound
	}
	if err != nil {
		return models.Evaluation{}, fmt.Errorf("find evaluation: %w", err)
	}
	return e, nil
}

type ListParams struct {
	UserID string
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.Evaluation, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM ai_evaluations WHERE user_id = $1`, p.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count evaluations: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+selectFields+` FROM ai_evaluations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		p.UserID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list evaluations: %w", err)
	}
	defer rows.Close()

	list := []models.Evaluation{}
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

func scan(row pgx.Row) (models.Evaluation, error) {
	var e models.Evaluation
	err := row.Scan(&e.ID, &e.QuestionID, &e.Exam, &e.Skill, &e.EvaluationVersion,
		&e.EstimatedScore, &e.ScoreConfidence, &e.Summary, &e.Criteria, &e.Strengths,
		&e.Weaknesses, &e.SentenceFeedback, &e.ModelRewrite, &e.Transcript, &e.CreatedAt)
	return e, err
}
