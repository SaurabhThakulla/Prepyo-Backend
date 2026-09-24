package reading

import (
	"context"
	"errors"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// The practice list shows the passages practice deals from, and a passage
// chosen from it is the one dealt.
func TestPractisesThePassageChosenFromTheList(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()
	user := newLearner(t, pool, models.ExamIELTS)
	family := []string{TypeMCQSingle, TypeMCQMultiple}

	list, err := svc.PracticePassages(ctx, user, models.ExamIELTS, family)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) < 2 {
		t.Skipf("the bank has %d multiple-choice passages; the test needs two", len(list))
	}
	for _, p := range list {
		if p.QuestionCount == 0 || p.Title == "" || p.PractisedAt != nil {
			t.Fatalf("list entry %+v", p)
		}
	}

	chosen := list[len(list)-1]
	set, err := svc.PracticeSet(ctx, user, PracticeParams{
		Exam: models.ExamIELTS, TypeID: TypeMCQSingle, TypeIDs: family, PassageID: chosen.ID,
	})
	if err != nil {
		t.Fatalf("practise %s: %v", chosen.ID, err)
	}
	if set.Passage == nil || set.Passage.ID != chosen.ID || len(set.Groups) == 0 {
		t.Fatalf("dealt %+v, want passage %s", set.Passage, chosen.ID)
	}
	for _, g := range set.Groups {
		if g.TypeID != TypeMCQSingle && g.TypeID != TypeMCQMultiple {
			t.Fatalf("dealt a %s set for multiple choice", g.TypeID)
		}
	}

	// Practising it marks it on the list.
	list, err = svc.PracticePassages(ctx, user, models.ExamIELTS, family)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range list {
		if (p.ID == chosen.ID) != (p.PractisedAt != nil) {
			t.Fatalf("%s practised at %v", p.ID, p.PractisedAt)
		}
	}

	// A passage with no set of the type, or none at all, is not dealt.
	if _, err := svc.PracticeSet(ctx, user, PracticeParams{
		Exam: models.ExamIELTS, TypeIDs: []string{TypeMCQSingle}, PassageID: "no-such-passage",
	}); !errors.Is(err, ErrNoPassage) {
		t.Fatalf("unknown passage: %v", err)
	}
}
