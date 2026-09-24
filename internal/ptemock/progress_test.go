package ptemock

import (
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
)

// advance is the whole of the test's timing, so it is tested on its own,
// without a database: a paper is a session row and its items.

func paper(tasks ...string) (sessionRow, []itemRow) {
	items := make([]itemRow, len(tasks))
	for i, code := range tasks {
		items[i] = itemRow{Position: i + 1, Task: code, Part: Tasks[code].Part, Status: ItemPending}
	}
	return sessionRow{Status: StatusInProgress, Current: 1, Total: len(items), Deadlines: map[string]time.Time{}}, items
}

func clockAt(at time.Time) *Service {
	return &Service{now: func() time.Time { return at }}
}

func TestAdvanceStartsTheClockOfTheItemOnScreen(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("RA", "RWFIB", "RO")
	svc := clockAt(start)

	svc.advance(&session, items)
	if session.Current != 1 || items[0].ShownAt == nil {
		t.Fatal("the first item is on screen")
	}
	if len(session.Deadlines) != 0 {
		t.Fatal("a speaking item starts no server clock")
	}

	items[0].Status = ItemAnswered
	svc.advance(&session, items)
	if session.Current != 2 {
		t.Fatalf("current %d", session.Current)
	}
	deadline, ok := session.Deadlines["part:reading"]
	if !ok || !deadline.Equal(start.Add(300*time.Second)) {
		t.Fatalf("reading clock %v %v", deadline, ok)
	}
}

func TestAdvanceClosesAPartWhenItsTimeIsUp(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("RWFIB", "RO", "RMCSA", "SST")
	clockAt(start).advance(&session, items)

	// A draft saved on the item on screen is what is marked.
	items[0].Response = &models.AnswerSubmission{BlankResponses: map[string]string{"b1": "word"}}

	late := start.Add(time.Duration(clockSeconds("part:reading", items))*time.Second + SubmitGrace + time.Second)
	clockAt(late).advance(&session, items)

	if items[0].Status != ItemAnswered {
		t.Fatalf("draft kept as the answer: %s", items[0].Status)
	}
	if items[1].Status != ItemTimedOut || items[2].Status != ItemTimedOut {
		t.Fatalf("the rest of the part timed out: %s %s", items[1].Status, items[2].Status)
	}
	if session.Current != 4 || items[3].ShownAt == nil {
		t.Fatalf("the test moved on to the next part: current %d", session.Current)
	}
	if _, ok := session.Deadlines["item:4"]; !ok {
		t.Fatal("Summarize Spoken Text starts its own clock")
	}
}

func TestAdvanceWithinGraceWaits(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("WE")
	clockAt(start).advance(&session, items)

	clockAt(start.Add(1200*time.Second+SubmitGrace/2)).advance(&session, items)
	if items[0].Status != ItemPending || session.Status != StatusInProgress {
		t.Fatal("an answer sent at zero is still on its way; the item stays open for the grace period")
	}
}

func TestAnAnswerInGraceDoesNotOpenAnItemWithNoTime(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("RWFIB", "RO", "RMCSA")
	clockAt(start).advance(&session, items)

	// The first item's answer arrives just after zero, within the grace period.
	inGrace := start.Add(time.Duration(clockSeconds("part:reading", items))*time.Second + SubmitGrace/2)
	svc := clockAt(inGrace)
	svc.advance(&session, items)
	if items[0].Status != ItemPending {
		t.Fatal("the item on screen still takes its answer")
	}
	items[0].Status = ItemAnswered
	svc.advance(&session, items)
	if items[1].Status != ItemTimedOut || items[2].Status != ItemTimedOut || session.Status != StatusScoring {
		t.Fatalf("the rest of the part closes rather than opening with no time: %s %s %s",
			items[1].Status, items[2].Status, session.Status)
	}
}

func TestAnItemDeletedWithItsQuestionIsPassedOver(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("RA", "DI", "RS")
	// Deleting a question deletes its item, leaving a gap at position 2.
	items = append(items[:1], items[2:]...)
	svc := clockAt(start)
	svc.advance(&session, items)
	items[0].Status = ItemAnswered
	svc.advance(&session, items)
	if session.Current != 3 || items[1].ShownAt == nil {
		t.Fatalf("the test moved past the gap: current %d", session.Current)
	}
	items[1].Status = ItemAnswered
	svc.advance(&session, items)
	if session.Status != StatusScoring {
		t.Fatalf("status %s", session.Status)
	}
}

func TestAdvanceSendsAFinishedPaperToScoring(t *testing.T) {
	start := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	session, items := paper("RA", "DI")
	svc := clockAt(start)
	svc.advance(&session, items)
	items[0].Status, items[1].Status = ItemAnswered, ItemSkipped
	svc.advance(&session, items)
	if session.Status != StatusScoring || session.ScoringStartedAt == nil {
		t.Fatalf("status %s", session.Status)
	}
	// Scoring is not undone by a second look.
	svc.advance(&session, items)
	if session.Status != StatusScoring {
		t.Fatal("advance only moves papers in progress")
	}
}

func TestCleanResponseBindsTheAnswerToItsItem(t *testing.T) {
	sub := cleanResponse(&models.AnswerSubmission{QuestionID: "someone-else", TextResponse: "hello"}, "q1")
	if sub.QuestionID != "q1" || sub.Exam != models.ExamPTE {
		t.Fatalf("%+v", sub)
	}
	if answered(cleanResponse(&models.AnswerSubmission{TextResponse: "  "}, "q1")) {
		t.Fatal("whitespace is not an answer")
	}
	if cleanResponse(nil, "q1") != nil {
		t.Fatal("no response stays none")
	}
}

func TestSaneDelivery(t *testing.T) {
	if saneDelivery(&Delivery{DurationSeconds: 10, PauseCount: 2, LongestPauseSeconds: 1, SpeakingRatio: 0.8}) == nil {
		t.Fatal("a plausible measurement is kept")
	}
	if saneDelivery(&Delivery{DurationSeconds: 10, SpeakingRatio: 1.5}) != nil {
		t.Fatal("a ratio above one is not a measurement")
	}
	if saneDelivery(&Delivery{DurationSeconds: 5, LongestPauseSeconds: 9}) != nil {
		t.Fatal("a pause longer than the recording is not a measurement")
	}
}
