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

func TestAuthoredQuestionSupportsCurrentSituationTask(t *testing.T) {
	const typeID = "pte-respond-to-situation"
	spec, ok := authorableByID[typeID]
	if !ok || spec.PrepSeconds != 10 || spec.TimeLimitSeconds != 40 {
		t.Fatalf("missing current PTE task or incorrect timing: %+v", spec)
	}
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: typeID, Title: "Delayed appointment",
		ContextPassage:  "Your train is delayed. Call your tutor and request a later appointment.",
		AudioTranscript: "Your train is delayed. Call your tutor and request a later appointment.",
	}.normalise()
	if len(problems) != 0 || q.prepSeconds != 10 || q.timeLimitSeconds != 40 {
		t.Fatalf("situation task rejected or wrong defaults: %v %+v", problems, q)
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
		ContextPassage: "The [[b1]] flows [[b2]].",
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

func TestAssetLinksAllowed(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-describe-image", Title: "x", ImageURL: "/api/v1/questions/assets/img_123",
	}.normalise()
	if problems["imageUrl"] != "" {
		t.Fatalf("expected asset url to be accepted, got error: %v", problems["imageUrl"])
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
		if spec.Exam != "PTE" && spec.Exam != "IELTS" && spec.Exam != "EPS_TOPIK" {
			t.Errorf("type %s has exam %q", spec.TypeID, spec.Exam)
		}
		if spec.Skill == "reading" && spec.Exam != "EPS_TOPIK" {
			t.Errorf("reading type %s is authored on passages, not here", spec.TypeID)
		}
		if spec.OptionImages != fieldUnused && spec.Answer != answerSingle && spec.Answer != answerMultiple {
			t.Errorf("type %s has picture options but no choices", spec.TypeID)
		}
		if spec.TimeLimitSeconds <= 0 || spec.Points <= 0 {
			t.Errorf("type %s has no default time or points", spec.TypeID)
		}
		if spec.Skill == "listening" && spec.Answer == answerRubric {
			t.Errorf("listening type %s is graded without a model, so it needs an answer key", spec.TypeID)
		}
	}
}

func TestEPSTopikHasTheOfficialSubtasks(t *testing.T) {
	want := map[string][]string{
		"reading": {"eps-r-picture", "eps-r-grammar", "eps-r-practical", "eps-r-word-relation", "eps-r-blank",
			"eps-r-definition", "eps-r-topic", "eps-r-detail", "eps-r-text-to-picture"},
		"listening": {"eps-l-sound", "eps-l-picture", "eps-l-response", "eps-l-next", "eps-l-number",
			"eps-l-picture-question", "eps-l-dialogue"},
	}
	for skill, ids := range want {
		for _, id := range ids {
			spec, ok := authorableByID[id]
			if !ok {
				t.Errorf("missing EPS-TOPIK subtask %s", id)
				continue
			}
			if spec.Exam != "EPS_TOPIK" || spec.Skill != skill || spec.Answer != answerSingle || spec.Points != 1 {
				t.Errorf("subtask %s is not a one-point %s choice: %+v", id, skill, spec)
			}
		}
	}
}

func TestPictureOptionsKeepTheirPictures(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "EPS_TOPIK", TypeID: "eps-l-picture", Title: "망치", AudioTranscript: "망치입니다.",
		Options:        []string{"망치", "", "가위", "드라이버"},
		OptionImages:   []string{"/api/v1/questions/assets/img_1", "", "/api/v1/questions/assets/img_2", "https://cdn.example.com/c.png"},
		CorrectAnswers: []string{"망치"},
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	want := []questionOption{
		{ID: "A", Text: "망치", ImageURL: "/api/v1/questions/assets/img_1"},
		{ID: "B", Text: "가위", ImageURL: "/api/v1/questions/assets/img_2"},
		{ID: "C", Text: "드라이버", ImageURL: "https://cdn.example.com/c.png"},
	}
	if !reflect.DeepEqual(q.options, want) || !reflect.DeepEqual(q.correctAnswers, []string{"A"}) {
		t.Fatalf("options = %+v, correct = %v", q.options, q.correctAnswers)
	}
}

func TestPictureTaskNeedsAPictureForEveryOption(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "EPS_TOPIK", TypeID: "eps-l-picture", Title: "x", AudioTranscript: "망치입니다.",
		Options: []string{"망치", "가위"}, OptionImages: []string{"/api/v1/questions/assets/img_1"},
		CorrectAnswers: []string{"망치"},
	}.normalise()
	if problems["optionImages"] == "" {
		t.Fatalf("expected an optionImages problem, got %v", problems)
	}
}

func TestTextTaskDropsOptionPictures(t *testing.T) {
	q, problems := newAuthoredQuestion{
		Exam: "EPS_TOPIK", TypeID: "eps-l-sound", Title: "x", AudioTranscript: "가구",
		Options: []string{"가구", "기구"}, OptionImages: []string{"/api/v1/questions/assets/img_1", ""},
		CorrectAnswers: []string{"가구"},
	}.normalise()
	if len(problems) > 0 || q.options[0].ImageURL != "" {
		t.Fatalf("problems = %v, options = %+v", problems, q.options)
	}
}

func TestNoticeNeedsTextOrAPicture(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "EPS_TOPIK", TypeID: "eps-r-detail", Title: "x",
		Options: []string{"a", "b"}, CorrectAnswers: []string{"a"},
	}.normalise()
	if problems["contextPassage"] == "" {
		t.Fatalf("expected a contextPassage problem, got %v", problems)
	}
	_, problems = newAuthoredQuestion{
		Exam: "EPS_TOPIK", TypeID: "eps-r-detail", Title: "x", ImageURL: "/api/v1/questions/assets/img_1",
		Options: []string{"a", "b"}, CorrectAnswers: []string{"a"},
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("a picture of the notice should be enough, got %v", problems)
	}
}
