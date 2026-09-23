package scoring

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func chooseTwo(exam models.ExamType) models.Question {
	return models.Question{
		Exam: exam, TypeID: "reading-mcq-multiple", Points: 1,
		Options:        []models.QuestionOption{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}, {ID: "E"}},
		CorrectAnswers: []string{"D", "E"},
	}
}

// IELTS gives one mark per correct answer and takes nothing away for a wrong
// one; a "Choose TWO" item is two marks.
func TestIELTSChooseTwoMarksEachCorrectLetter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		selected []string
		marks    int
	}{
		{"both right", []string{"D", "E"}, 2},
		{"one right one wrong", []string{"D", "A"}, 1},
		{"one right only", []string{"E"}, 1},
		{"both wrong", []string{"A", "B"}, 0},
		{"an extra letter cancels a correct one", []string{"D", "E", "A"}, 1},
		{"every letter", []string{"A", "B", "C", "D", "E"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Grade(chooseTwo(models.ExamIELTS), models.AnswerSubmission{SelectedOptions: tc.selected})
			if !ok {
				t.Fatal("not graded")
			}
			if got.Marks != tc.marks || got.MarksAvailable != 2 {
				t.Fatalf("marks = %d/%d, want %d/2", got.Marks, got.MarksAvailable, tc.marks)
			}
		})
	}
}

// PTE keeps its own negative marking; the IELTS mode must not leak into it.
func TestPTEChoiceStillNetsWrongAnswers(t *testing.T) {
	got, _ := Grade(chooseTwo(models.ExamPTE), models.AnswerSubmission{SelectedOptions: []string{"D", "A"}})
	if got.Score != 0 {
		t.Fatalf("PTE score = %v, want 0: one right minus one wrong", got.Score)
	}
}

// The exam the learner was working under decides the marking, so a shared
// question on an IELTS paper is marked the IELTS way.
func TestSubmissionExamDecidesMarking(t *testing.T) {
	q := chooseTwo(models.ExamPTE)
	got, _ := Grade(q, models.AnswerSubmission{Exam: models.ExamIELTS, SelectedOptions: []string{"D", "A"}})
	if got.Marks != 1 {
		t.Fatalf("marks = %d, want 1 under IELTS marking", got.Marks)
	}
}

func TestBlanksIgnorePunctuationAndAcceptVariants(t *testing.T) {
	q := models.Question{Exam: models.ExamIELTS, Points: 2, Blanks: []models.Blank{
		{ID: "1", CorrectAnswer: "Thursday"},
		{ID: "2", CorrectAnswer: "ten", AcceptedAnswers: []string{"10"}},
	}}
	for _, tc := range []struct {
		one, two string
		marks    int
	}{
		{"Thursday.", "10", 2},
		{" thursday ", "ten", 2},
		{"on Thursday", "10", 1}, // more words than the key: IELTS marks it wrong
		{"Thurs", "eleven", 0},
	} {
		got, _ := Grade(q, models.AnswerSubmission{BlankResponses: map[string]string{"1": tc.one, "2": tc.two}})
		if got.Marks != tc.marks || got.MarksAvailable != 2 {
			t.Errorf("%q/%q: marks = %d/%d, want %d/2", tc.one, tc.two, got.Marks, got.MarksAvailable, tc.marks)
		}
	}
}

func TestSingleAnswerItemsAreOneMark(t *testing.T) {
	q := models.Question{Exam: models.ExamIELTS, TypeID: "reading-true-false", Points: 1,
		Options: []models.QuestionOption{{ID: "TRUE"}, {ID: "FALSE"}}, CorrectAnswers: []string{"TRUE"}}
	right, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"TRUE"}})
	wrong, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"FALSE"}})
	if right.Marks != 1 || right.MarksAvailable != 1 || wrong.Marks != 0 || wrong.MarksAvailable != 1 {
		t.Fatalf("right %d/%d wrong %d/%d, want 1/1 and 0/1",
			right.Marks, right.MarksAvailable, wrong.Marks, wrong.MarksAvailable)
	}
}

// A paper that is not exactly 40 marks is scaled to 40 before the table, so
// one extra question cannot move the band by a whole point.
func TestIELTSRawMarksScaleToForty(t *testing.T) {
	s := Scale{Min: 0, Max: 9, Step: 0.5}
	if a, b := s.EstimateFromRawMarks("IELTS", "reading", 20, 40), s.EstimateFromRawMarks("IELTS", "reading", 20, 41); a != b {
		t.Fatalf("20/40 = %.1f but 20/41 = %.1f; a one-question difference must not change the band", a, b)
	}
	if got := s.EstimateFromRawMarks("IELTS", "listening", 15, 20); got != IELTSListeningBand(30) {
		t.Fatalf("15/20 listening = %.1f, want the band for 30/40 (%.1f)", got, IELTSListeningBand(30))
	}
}

// General Training Reading has its own table: the published anchors are
// 15->4, 23->5, 30->6 and 35->7.
func TestGeneralTrainingReadingAnchors(t *testing.T) {
	for raw, band := range map[int]float64{15: 4, 23: 5, 30: 6, 35: 7} {
		if got := IELTSGeneralTrainingReadingBand(raw); got != band {
			t.Errorf("GT %d/40 = %.1f, want %.1f", raw, got, band)
		}
		if got := IELTSBandFromRawMarks(ModuleGeneralTraining, "reading", raw, 40); got != band {
			t.Errorf("module lookup GT %d/40 = %.1f, want %.1f", raw, got, band)
		}
	}
	// The published Academic anchors are one band higher for the same marks.
	for raw, band := range map[int]float64{15: 5, 23: 6, 30: 7, 35: 8} {
		if got := IELTSBandFromRawMarks(ModuleAcademic, "reading", raw, 40); got != band {
			t.Errorf("Academic %d/40 = %.1f, want %.1f", raw, got, band)
		}
	}
	for raw, band := range map[int]float64{16: 5, 23: 6, 30: 7, 35: 8} {
		if got := IELTSListeningBand(raw); got != band {
			t.Errorf("Listening %d/40 = %.1f, want %.1f", raw, got, band)
		}
	}
}
