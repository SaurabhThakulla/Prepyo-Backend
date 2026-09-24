package ptemock

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/prepyo/backend/internal/mockpapers"
)

// publishReading publishes a numbered PTE Reading test of the given question
// IDs, and retires it when the test ends so no other test is dealt it.
func (f fixture) publishReading(t *testing.T, title string, ids ...string) mockpapers.MockPaper {
	t.Helper()
	type item struct {
		QuestionID string `json:"questionId"`
		Task       string `json:"task"`
		Part       string `json:"part"`
	}
	items := make([]item, 0, len(ids))
	for _, id := range ids {
		task := "RWFIB"
		if id == f.prefix+"ro" {
			task = "RO"
		}
		items = append(items, item{QuestionID: id, Task: task, Part: string(PartReading)})
	}
	content, err := json.Marshal(map[string]any{"items": items, "missing": []string{}, "title": title + f.prefix})
	if err != nil {
		t.Fatal(err)
	}
	repo := mockpapers.NewRepository(f.pool)
	paper, err := repo.Publish(context.Background(), mockpapers.DraftPaper{
		Exam: mockpapers.ExamPTE, Section: string(KindReading), Module: mockpapers.ModuleAny,
		Title: title, ContentSchema: 1, Content: content,
	})
	if err != nil {
		t.Fatalf("publish %s: %v", title, err)
	}
	t.Cleanup(func() { _ = repo.Retire(context.Background(), paper.ID) })
	return paper
}

func TestStartsTheNumberedTestAskedFor(t *testing.T) {
	ctx := context.Background()
	f := setup(t)
	a := f.publishReading(t, "Test A", f.prefix+"rwfib", f.prefix+"ro")
	b := f.publishReading(t, "Test B", f.prefix+"ro")

	view, err := f.svc.Start(ctx, f.user, KindReading, a.ID)
	if err != nil {
		t.Fatalf("start test A: %v", err)
	}
	if view.PaperID == nil || *view.PaperID != a.ID || view.TotalItems != 2 {
		t.Fatalf("test A dealt paper %v with %d items", view.PaperID, view.TotalItems)
	}

	// Another test of the section is refused while A is open, not swapped for A.
	if _, err := f.svc.Start(ctx, f.user, KindReading, b.ID); !errors.Is(err, mockpapers.ErrOtherPaperOpen) {
		t.Fatalf("starting test B while A is open: %v", err)
	}
	for _, asked := range []string{a.ID, ""} {
		again, err := f.svc.Start(ctx, f.user, KindReading, asked)
		if err != nil || again.ID != view.ID {
			t.Fatalf("asking for %q resumes test A: %s %v", asked, again.ID, err)
		}
	}
}

func TestANumberedTestSkipsADeletedQuestion(t *testing.T) {
	ctx := context.Background()
	f := setup(t)
	paper := f.publishReading(t, "Test with a gap", f.prefix+"rwfib", f.prefix+"deleted", f.prefix+"ro")

	view, err := f.svc.Start(ctx, f.user, KindReading, paper.ID)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if view.TotalItems != 2 {
		t.Fatalf("%d items, want the 2 still in the bank", view.TotalItems)
	}
}

func TestAPaperWithAPartLeftEmptyCannotStart(t *testing.T) {
	f := setup(t)
	paper := f.publishReading(t, "Test with nothing left", f.prefix+"gone")
	if _, err := f.svc.Start(context.Background(), f.user, KindReading, paper.ID); !errors.Is(err, ErrBankTooSmall) {
		t.Fatalf("start: %v", err)
	}
}
