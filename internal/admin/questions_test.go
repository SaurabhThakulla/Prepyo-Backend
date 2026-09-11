package admin

import (
	"reflect"
	"testing"
)

func TestAuthoredQuestionRejectsUnknownType(t *testing.T) {
	_, problems := newAuthoredQuestion{Exam: "PTE", TypeID: "no-such-task", Title: "x"}.normalise()
	if problems["typeId"] == "" {
		t.Fatalf("expected a typeId problem, got %v", problems)
	}
}

func TestAuthoredQuestionRejectsTheWrongExam(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "IELTS", TypeID: "read-aloud", Title: "x", ContextPassage: "Some text to read.",
	}.normalise()
	if problems["exam"] == "" {
		t.Fatalf("expected an exam problem, got %v", problems)
	}
}

func TestAuthoredQuestionFillsTaskDefaults(t *testing.T) {
	q, problems := newAuthoredQuestion{Exam: "PTE", TypeID: "pte-write-essay", Title: "Tuition fees"}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	spec := authorableByID["pte-write-essay"]
	if q.prompt != spec.Prompt || q.timeLimitSeconds != spec.TimeLimitSeconds || q.points != spec.Points {
		t.Fatalf("defaults not applied: prompt=%q time=%d points=%d", q.prompt, q.timeLimitSeconds, q.points)
	}
	if q.difficulty != "medium" || len(q.tags) == 0 {
		t.Fatalf("difficulty=%q tags=%v", q.difficulty, q.tags)
	}
}

func TestAuthoredQuestionDropsFieldsTheTaskDoesNotUse(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-write-essay", Title: "Tuition fees",
		ContextPassage: "stray", AudioURL: "https://example.com/a.mp3",
		Options: []string{"a", "b"}, CorrectAnswers: []string{"a"}, PrepTimeSeconds: 30,
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if q.contextPassage != "" || q.audioURL != "" || q.options != nil || q.correctAnswers != nil || q.prepSeconds != 0 {
		t.Fatalf("unused fields kept: %+v", q)
	}
}

func TestListeningNeedsSomethingToListenTo(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "IELTS", TypeID: "ielts-listening-mcq", Title: "x",
		Options: []string{"a", "b"}, CorrectAnswers: []string{"a"},
	}.normalise()
	if problems["audioTranscript"] == "" {
		t.Fatalf("expected an audio problem, got %v", problems)
	}
}

func TestSingleChoiceTakesExactlyOneAnswer(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "IELTS", TypeID: "ielts-listening-mcq", Title: "x", AudioTranscript: "script",
		Options: []string{"a", "b", "c"}, CorrectAnswers: []string{"a", "b"},
	}.normalise()
	if problems["correctAnswers"] == "" {
		t.Fatalf("expected a correctAnswers problem, got %v", problems)
	}
}

func TestChoiceAnswersBecomeOptionLetters(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-mcma", Title: "x", AudioTranscript: "script",
		Options: []string{" red ", "green", "", "blue"}, CorrectAnswers: []string{"Blue", "red"},
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	wantOptions := []questionOption{{ID: "A", Text: "red"}, {ID: "B", Text: "green"}, {ID: "C", Text: "blue"}}
	if !reflect.DeepEqual(q.options, wantOptions) {
		t.Fatalf("options = %v", q.options)
	}
	if !reflect.DeepEqual(q.correctAnswers, []string{"C", "A"}) {
		t.Fatalf("correct = %v", q.correctAnswers)
	}
}

func TestMultipleChoiceNeedsAWrongOption(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-mcma", Title: "x", AudioTranscript: "script",
		Options: []string{"a", "b"}, CorrectAnswers: []string{"a", "b"},
	}.normalise()
	if problems["correctAnswers"] == "" {
		t.Fatalf("expected a correctAnswers problem, got %v", problems)
	}
}

func TestDictationIsAnsweredByItsScript(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "write-from-dictation", Title: "x",
		AudioTranscript: " The library closes at nine. ",
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if !reflect.DeepEqual(q.correctAnswers, []string{"The library closes at nine."}) {
		t.Fatalf("correct = %v", q.correctAnswers)
	}

	_, problems = newAuthoredQuestion{
		Exam: "PTE", TypeID: "write-from-dictation", Title: "x", AudioURL: "https://example.com/a.mp3",
	}.normalise()
	if problems["audioTranscript"] == "" {
		t.Fatalf("a recording alone gives no answer key, got %v", problems)
	}
}

func TestBlanksAreNumberedAndChecked(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-fib", Title: "x", AudioTranscript: "script",
		Blanks: []newBlank{
			{CorrectAnswer: " river "},
			{},
			{Options: []string{"north", "South"}, CorrectAnswer: "south"},
		},
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	want := []storedBlank{
		{ID: "b1", Options: []string{}, CorrectAnswer: "river"},
		{ID: "b2", Options: []string{"north", "South"}, CorrectAnswer: "South"},
	}
	if !reflect.DeepEqual(q.blanks, want) {
		t.Fatalf("blanks = %+v", q.blanks)
	}

	_, problems = newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-fib", Title: "x", AudioTranscript: "script",
		Blanks: []newBlank{{Options: []string{"north", "south"}, CorrectAnswer: "east"}},
	}.normalise()
	if problems["blanks"] == "" {
		t.Fatalf("expected a blanks problem, got %v", problems)
	}
}

func TestLinksMustBeWebLinks(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-describe-image", Title: "x", ImageURL: "javascript:alert(1)",
	}.normalise()
	if problems["imageUrl"] == "" {
		t.Fatalf("expected an imageUrl problem, got %v", problems)
	}
}

func TestEveryAuthorableTypeIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, spec := range authorableTypes {
		if seen[spec.TypeID] {
			t.Errorf("type %s is listed twice", spec.TypeID)
		}
		seen[spec.TypeID] = true
		if !authorableSkill(spec.Skill) {
			t.Errorf("type %s has skill %q", spec.TypeID, spec.Skill)
		}
		if spec.Exam != "PTE" && spec.Exam != "IELTS" {
			t.Errorf("type %s has exam %q", spec.TypeID, spec.Exam)
		}
		if spec.TimeLimitSeconds <= 0 || spec.Points <= 0 {
			t.Errorf("type %s has no default time or points", spec.TypeID)
		}
		if spec.Skill == "listening" && spec.Answer == answerRubric {
			t.Errorf("listening type %s is graded without a model, so it needs an answer key", spec.TypeID)
		}
	}
}
