package reading

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func sampleQuestions() []models.Question {
	list := make([]models.Question, 8)
	for i := range list {
		list[i] = models.Question{
			ID:             string(rune('a' + i)),
			TypeID:         TypeSentenceCompletion,
			CorrectAnswers: []string{"secret"},
			ModelAnswer:    "secret",
			Explanation:    "because",
			Blanks:         []models.Blank{{ID: "b1", CorrectAnswer: "secret"}},
		}
	}
	return list
}

func idsOf(list []models.Question) []string {
	ids := make([]string, len(list))
	for i, q := range list {
		ids[i] = q.ID
	}
	return ids
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Nothing a learner is given while answering may carry the answer key, and a
// reading question travels with more of one than most: the accepted spellings,
// the model answer, the explanation, and the answer inside each blank.
func TestBuildGroupStripsTheAnswerKey(t *testing.T) {
	group := buildGroup(Group{ShuffleQuestions: false}, sampleQuestions(), "")

	for _, q := range group.Questions {
		if len(q.CorrectAnswers) != 0 || q.ModelAnswer != "" || q.Explanation != "" {
			t.Fatalf("question %s still carries its answer key", q.ID)
		}
		for _, b := range q.Blanks {
			if b.CorrectAnswer != "" {
				t.Fatalf("question %s leaks the answer through blank %s", q.ID, b.ID)
			}
		}
	}
}

func TestBuildGroupShufflesWhenTheTaskAllowsIt(t *testing.T) {
	original := idsOf(sampleQuestions())

	// Any single shuffle can legitimately come back in order, so this asks
	// whether shuffling happens at all rather than whether one draw moved.
	moved := false
	for range 20 {
		if !equal(idsOf(buildGroup(Group{ShuffleQuestions: true}, sampleQuestions(), "").Questions), original) {
			moved = true
			break
		}
	}
	if !moved {
		t.Error("20 shuffles all returned the authored order; questions are not being randomised")
	}
}

// An ordering task is the position of each question in the set, so shuffling it
// would destroy the question.
func TestBuildGroupKeepsOrderWhenTheTaskDependsOnIt(t *testing.T) {
	original := idsOf(sampleQuestions())

	for range 20 {
		got := idsOf(buildGroup(Group{ShuffleQuestions: false}, sampleQuestions(), "").Questions)
		if !equal(got, original) {
			t.Fatalf("order = %v, want the authored order %v", got, original)
		}
	}
}

func TestBuildGroupCarriesTheGroupThrough(t *testing.T) {
	group := buildGroup(Group{
		ReadingGroup: models.ReadingGroup{
			ID:           "g-1",
			PassageID:    "p-1",
			TypeID:       TypeMatchingInformation,
			TypeName:     "Matching Information",
			Instructions: "Which paragraph contains each of the following?",
			Resources:    []models.ReadingParagraph{{Label: "A", Text: "box"}},
		},
	}, sampleQuestions(), "")

	if group.ID != "g-1" || group.PassageID != "p-1" || group.TypeID != TypeMatchingInformation {
		t.Errorf("group identity lost: %+v", group)
	}
	if group.Instructions == "" || len(group.Resources) != 1 {
		t.Error("instructions or resources were dropped; the client cannot render the task without them")
	}
	if len(group.Questions) != 8 {
		t.Errorf("questions = %d, want 8", len(group.Questions))
	}
}

func TestReviewOfFollowsTheDealtOrderAndRestoresAnswers(t *testing.T) {
	bank := map[string]models.Question{
		"a": {ID: "a", CorrectAnswers: []string{"gods"}, Explanation: "paragraph A"},
		"b": {ID: "b", CorrectAnswers: []string{"bitter"}},
	}

	review := reviewOf([]string{"b", "a", "missing"}, bank)

	if len(review) != 2 {
		t.Fatalf("review length = %d, want 2 (the unknown id is skipped)", len(review))
	}
	if review[0].ID != "b" || review[1].ID != "a" {
		t.Errorf("review order = %s,%s; want the dealt order b,a", review[0].ID, review[1].ID)
	}
	if len(review[1].CorrectAnswers) != 1 || review[1].Explanation != "paragraph A" {
		t.Error("review must carry the answer key: it is what the results screen shows")
	}
}

// A question the exam does not set is not a question the learner can be asked,
// whichever route reached it. buildGroup is the last place every path passes
// through, so the filter lives there and is asserted here.
func TestBuildGroupDropsQuestionsTheExamDoesNotSet(t *testing.T) {
	list := []models.Question{
		{ID: "ielts-only", SupportedExams: []models.ExamType{models.ExamIELTS}},
		{ID: "pte-only", SupportedExams: []models.ExamType{models.ExamPTE}},
		{ID: "both", SupportedExams: []models.ExamType{models.ExamIELTS, models.ExamPTE}},
	}

	got := idsOf(buildGroup(Group{ShuffleQuestions: false}, list, models.ExamPTE).Questions)
	if !equal(got, []string{"pte-only", "both"}) {
		t.Errorf("PTE set = %v, want the PTE-only and shared questions", got)
	}

	got = idsOf(buildGroup(Group{ShuffleQuestions: false}, list, models.ExamIELTS).Questions)
	if !equal(got, []string{"ielts-only", "both"}) {
		t.Errorf("IELTS set = %v, want the IELTS-only and shared questions", got)
	}

	// An empty exam is "no filter", which is what a direct passage read wants.
	got = idsOf(buildGroup(Group{ShuffleQuestions: false}, list, "").Questions)
	if len(got) != 3 {
		t.Errorf("unfiltered set = %v, want all three", got)
	}
}

// A question written before eligibility existed carries no supported exams, and
// has to stay answerable under the exam it was authored for.
func TestSupportsExamFallsBackToTheAuthoredExam(t *testing.T) {
	legacy := models.Question{ID: "old", Exam: models.ExamIELTS}

	if !legacy.SupportsExam(models.ExamIELTS) {
		t.Error("a legacy IELTS question stopped being answerable under IELTS")
	}
	if legacy.SupportsExam(models.ExamPTE) {
		t.Error("a legacy IELTS question became answerable under PTE")
	}
}
