package admin

import (
	"context"
	"testing"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/scoring"
)

func TestAuthoredQuestionsAreStoredSoTheyGrade(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()
	bank := questions.NewRepository(h.db)

	cases := []struct {
		name   string
		req    newAuthoredQuestion
		answer models.AnswerSubmission
	}{
		{
			name: "multiple answers",
			req: newAuthoredQuestion{
				Exam: "PTE", TypeID: "pte-listening-mcma", Title: "Authoring check: choices",
				AudioTranscript: "Urban bees forage further than rural ones.",
				Options:         []string{"Cities", "Farmland", "Forests"},
				CorrectAnswers:  []string{"Cities", "Forests"},
			},
			answer: models.AnswerSubmission{SelectedOptions: []string{"A", "C"}},
		},
		{
			name: "fill in the blanks",
			req: newAuthoredQuestion{
				Exam: "IELTS", TypeID: "ielts-listening-fill-blanks", Title: "Authoring check: blanks",
				AudioTranscript: "The tour leaves from the harbour at nine.",
				Blanks:          []newBlank{{CorrectAnswer: "harbour"}, {CorrectAnswer: "nine"}},
			},
			answer: models.AnswerSubmission{BlankResponses: map[string]string{"b1": "Harbour", "b2": "nine"}},
		},
		{
			name: "dictation",
			req: newAuthoredQuestion{
				Exam: "PTE", TypeID: "write-from-dictation", Title: "Authoring check: dictation",
				AudioTranscript: "Late submissions lose ten percent per day.",
			},
			answer: models.AnswerSubmission{TextResponse: "late submissions lose ten percent per day"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, problems := tc.req.normalise()
			if len(problems) > 0 {
				t.Fatalf("normalise: %v", problems)
			}
			id, err := h.insertAuthoredQuestion(ctx, q)
			if err != nil {
				t.Fatalf("insert: %v", err)
			}
			t.Cleanup(func() {
				_, _ = h.db.Exec(context.Background(), `DELETE FROM questions WHERE id = $1`, id)
			})

			found, err := bank.ByIDs(ctx, []string{id})
			if err != nil {
				t.Fatalf("read back: %v", err)
			}
			stored, ok := found[id]
			if !ok {
				t.Fatalf("question %s was not stored", id)
			}
			if stored.TypeName != q.spec.TypeName || string(stored.Skill) != "listening" {
				t.Fatalf("stored as %s / %s", stored.Skill, stored.TypeName)
			}

			result, graded := scoring.Grade(stored, tc.answer)
			if !graded || !result.IsCorrect {
				t.Fatalf("graded=%v result=%+v", graded, result)
			}
		})
	}
}

func TestAuthoredFigureTaskKeepsItsFigure(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	q, problems := newAuthoredQuestion{
		Exam: "IELTS", TypeID: "ielts-writing-task1-figure", Title: "Authoring check: figure",
		ImageURL: "https://example.com/chart.png", FigureData: "2000: 10%, 2020: 90%",
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("normalise: %v", problems)
	}
	id, err := h.insertAuthoredQuestion(ctx, q)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() {
		_, _ = h.db.Exec(context.Background(), `DELETE FROM questions WHERE id = $1`, id)
	})

	found, err := questions.NewRepository(h.db).ByIDs(ctx, []string{id})
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	stored := found[id]
	if stored.ImageURL != q.imageURL || stored.FigureData != q.figureData || stored.Prompt == "" {
		t.Fatalf("stored %+v", stored)
	}
}
