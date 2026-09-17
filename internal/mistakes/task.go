package mistakes

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

// RetryTask includes the exact question and its supporting material, never its key.
type RetryTask struct {
	Question models.Question        `json:"question"`
	Passage  *models.ReadingPassage `json:"passage,omitempty"`
	Group    *models.ReadingGroup   `json:"group,omitempty"`
}

func (r *Repository) Task(ctx context.Context, userID, mistakeID string) (RetryTask, error) {
	var questionID string
	var exam models.ExamType
	err := r.db.QueryRow(ctx, `SELECT question_id, exam FROM mistakes WHERE id=$1 AND user_id=$2`, mistakeID, userID).Scan(&questionID, &exam)
	if errors.Is(err, pgx.ErrNoRows) {
		return RetryTask{}, ErrNotFound
	}
	if err != nil {
		return RetryTask{}, fmt.Errorf("get mistake task: %w", err)
	}
	q, err := questions.NewRepository(r.db).ByID(ctx, questionID)
	if errors.Is(err, questions.ErrNotFound) {
		return RetryTask{}, ErrNotFound
	}
	if err != nil {
		return RetryTask{}, err
	}
	if !q.SupportsExam(exam) {
		return RetryTask{}, ErrNotFound
	}
	task := RetryTask{Question: q.PublicQuestion()}
	if q.TypeID == "reorder-paragraphs" {
		rand.Shuffle(len(task.Question.Options), func(i, j int) {
			task.Question.Options[i], task.Question.Options[j] = task.Question.Options[j], task.Question.Options[i]
		})
	}
	if q.GroupID == "" {
		return task, nil
	}
	var g models.ReadingGroup
	err = r.db.QueryRow(ctx, `SELECT id, passage_id, type_id, type_name, instructions,
 resources, passage_display, time_limit_seconds, box_title FROM reading_question_groups WHERE id=$1`, q.GroupID).
		Scan(&g.ID, &g.PassageID, &g.TypeID, &g.TypeName, &g.Instructions, &g.Resources, &g.PassageDisplay, &g.TimeLimitSeconds, &g.BoxTitle)
	if err != nil {
		return RetryTask{}, fmt.Errorf("get retry group: %w", err)
	}
	task.Group = &g
	// Gap-fill tasks must not receive the clean passage: it reveals the answers.
	if g.PassageDisplay == "hidden" {
		return task, nil
	}
	var p models.ReadingPassage
	err = r.db.QueryRow(ctx, `SELECT id, title, subtitle, paragraphs, sources, word_count
 FROM reading_passages WHERE id=$1 AND is_published`, g.PassageID).
		Scan(&p.ID, &p.Title, &p.Subtitle, &p.Paragraphs, &p.Sources, &p.WordCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return RetryTask{}, ErrNotFound
	}
	if err != nil {
		return RetryTask{}, fmt.Errorf("get retry passage: %w", err)
	}
	task.Passage = &p
	return task, nil
}
