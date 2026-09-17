package admin

import (
	"reflect"
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

func TestPTEListeningRequiresTranscriptForTTS(t *testing.T) {
	for _, spec := range authorableTypes {
		if spec.Exam != "PTE" || spec.Skill != "listening" {
			continue
		}
		t.Run(spec.TypeID, func(t *testing.T) {
			for _, script := range []string{"", " \n\t "} {
				_, problems := newAuthoredQuestion{
					Exam: "PTE", TypeID: spec.TypeID, Title: "TTS validation",
					AudioURL: "https://example.com/recording.mp3", AudioTranscript: script,
				}.normalise()
				if problems["audioTranscript"] == "" {
					t.Fatalf("a URL without a script must be rejected: %v", problems)
				}
			}
		})
	}
}

func TestPTEListeningBlanksKeepDisplaySeparateFromTTS(t *testing.T) {
	const passage = "The [[b1]] closes at nine."
	const script = "The library closes at nine."
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-fib", Title: "Library hours",
		ContextPassage: "  " + passage + "  ", AudioTranscript: "  " + script + "  ",
		Blanks: []newBlank{{CorrectAnswer: "library"}},
	}.normalise()
	if len(problems) != 0 {
		t.Fatalf("transcript-only authoring failed: %v", problems)
	}
	if q.contextPassage != passage || q.audioTranscript != script || q.audioURL != "" {
		t.Fatalf("display and spoken text were not preserved separately: %+v", q)
	}
	if q.spec.Passage != fieldRequired {
		t.Fatal("the authoring catalogue must request a displayed passage")
	}
}

func TestPTEListeningBlanksRequireDisplayedPassage(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-listening-fib", Title: "Library hours",
		ContextPassage: " \n ", AudioTranscript: "The library closes at nine.",
		Blanks: []newBlank{{CorrectAnswer: "library"}},
	}.normalise()
	if problems["contextPassage"] == "" {
		t.Fatalf("expected a displayed passage problem: %v", problems)
	}
}

func TestSummaryAuthoringKeepsGradingKey(t *testing.T) {
	req := newAuthoredQuestion{
		Exam: "PTE", TypeID: "summarize-spoken-text", Title: "Energy",
		AudioTranscript: "Battery storage helps the grid.", ModelAnswer: "Battery storage helps the grid.",
		CorrectAnswers: []string{" Battery ", "battery", "", "GRID"},
	}
	q, problems := req.normalise()
	if len(problems) != 0 || !reflect.DeepEqual(q.correctAnswers, []string{"battery", "grid"}) || q.modelAnswer != req.ModelAnswer {
		t.Fatalf("summary normalisation: %+v, %v", q, problems)
	}
	result, ok := scoring.Grade(models.Question{
		TypeID: q.spec.TypeID, Points: q.points, CorrectAnswers: q.correctAnswers,
	}, models.AnswerSubmission{TextResponse: "battery grid " + strings.Repeat("energy ", 48)})
	if !ok || !result.IsCorrect {
		t.Fatalf("summary key did not reach the existing grader: %+v", result)
	}
	req.ModelAnswer, req.CorrectAnswers = "", nil
	_, problems = req.normalise()
	if problems["modelAnswer"] == "" || problems["correctAnswers"] == "" {
		t.Fatalf("missing summary fields accepted: %v", problems)
	}
}

func TestHighlightedWordsUseOccurrenceIDs(t *testing.T) {
	const text = "The library and the library are open to all the students today."
	q, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-highlight-incorrect-word", Title: "Library",
		ContextPassage: text, AudioTranscript: "The library and the museum are open to all the students today.",
		CorrectAnswers: []string{"w5", "w5"},
	}.normalise()
	if len(problems) != 0 || len(q.options) != 12 || !reflect.DeepEqual(q.correctAnswers, []string{"w5"}) {
		t.Fatalf("word normalisation: %+v, %v", q, problems)
	}
	if q.options[1].Text != q.options[4].Text || q.options[1].ID == q.options[4].ID || q.contextPassage != text {
		t.Fatal("repeated word occurrences or displayed transcript lost")
	}
	options := make([]models.QuestionOption, len(q.options))
	for i, option := range q.options {
		options[i] = models.QuestionOption{ID: option.ID, Text: option.Text}
	}
	question := models.Question{TypeID: q.spec.TypeID, Points: 10, Options: options, CorrectAnswers: q.correctAnswers}
	for _, tc := range []struct {
		selected []string
		score    float64
	}{
		{[]string{"w5"}, 10}, {[]string{"w2"}, 0}, {[]string{"w5", "w2"}, 0},
	} {
		result, ok := scoring.Grade(question, models.AnswerSubmission{SelectedOptions: tc.selected})
		if !ok || result.Score != tc.score {
			t.Fatalf("word scoring %v = %+v", tc.selected, result)
		}
	}
	for _, answers := range [][]string{nil, {"library"}, {"w0"}, {"w14"}} {
		_, _, problems := highlightedWordKey(text, answers)
		if problems["correctAnswers"] == "" {
			t.Fatalf("invalid word IDs accepted: %v", answers)
		}
	}
	_, _, problems = highlightedWordKey("one two", []string{"w1", "w2"})
	if problems["correctAnswers"] == "" {
		t.Fatal("all words marked accepted")
	}
	_, _, problems = highlightedWordKey(strings.Repeat("word ", 501), []string{"w1"})
	if problems["contextPassage"] == "" {
		t.Fatal("unbounded transcript accepted")
	}
}

func TestLegacyPTEListeningAuthoringUsesCanonicalType(t *testing.T) {
	for _, id := range []string{"multiple-choice-multiple", "fill-in-the-blanks", "highlight-correct-summary", "select-missing-word"} {
		q, problems := newAuthoredQuestion{
			Exam: "PTE", TypeID: id, Title: "Legacy authoring", AudioTranscript: "The library closes.",
			ContextPassage: "The [[b1]] closes.", Blanks: []newBlank{{CorrectAnswer: "library"}},
			Options: []string{"library", "museum"}, CorrectAnswers: []string{"library"},
		}.normalise()
		if len(problems) != 0 || q.spec.TypeID != models.CanonicalPTEListeningType(id) {
			t.Fatalf("legacy %s: %+v, %v", id, q, problems)
		}
	}
}

func TestSpeakingStillAcceptsRecordingWithoutTranscript(t *testing.T) {
	_, problems := newAuthoredQuestion{
		Exam: "PTE", TypeID: "pte-repeat-sentence", Title: "Repeat the sentence",
		AudioURL: "https://example.com/recording.mp3",
	}.normalise()
	if len(problems) != 0 {
		t.Fatalf("listening-only validation changed speaking authoring: %v", problems)
	}
}
