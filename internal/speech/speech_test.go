package speech

import (
	"strings"
	"testing"
)

func TestSplitTurnsDropsLabelsAndKeepsPunctuation(t *testing.T) {
	turns := SplitTurns("Receptionist: The course is full. Caller: Which day is it? Receptionist: It is on Thursday.")
	want := []Turn{
		{"Receptionist", "The course is full."},
		{"Caller", "Which day is it?"},
		{"Receptionist", "It is on Thursday."},
	}
	if len(turns) != len(want) {
		t.Fatalf("turns = %+v", turns)
	}
	for i := range want {
		if turns[i] != want[i] {
			t.Errorf("turn %d = %+v, want %+v", i, turns[i], want[i])
		}
	}
}

func TestSplitTurnsLeavesUnlabelledScriptsWhole(t *testing.T) {
	turns := SplitTurns("Go north from the entrance. Note: the café is closed.")
	if len(turns) != 1 || turns[0].Speaker != "" {
		t.Fatalf("turns = %+v, want one unlabelled turn", turns)
	}
}

func TestPiecesFitTheProviderLimit(t *testing.T) {
	long := strings.Repeat("The museum reopened after the storm, with new galleries, a larger café and longer opening hours; ", 6)
	pieces := SplitPieces([]Turn{{Speaker: "Guide", Text: long + "Tickets are free."}}, 200)
	if len(pieces) < 3 {
		t.Fatalf("pieces = %d, want the long turn split", len(pieces))
	}
	var rebuilt []string
	for _, p := range pieces {
		if n := len([]rune(p.Text)); n == 0 || n > 200 {
			t.Fatalf("piece of %d characters: %q", n, p.Text)
		}
		if p.Speaker != "Guide" {
			t.Fatalf("speaker = %q", p.Speaker)
		}
		rebuilt = append(rebuilt, p.Text)
	}
	if got := strings.Join(strings.Fields(strings.Join(rebuilt, " ")), " "); got != strings.Join(strings.Fields(long+"Tickets are free."), " ") {
		t.Fatal("splitting lost or changed words")
	}
}

func TestEachSpeakerGetsTheirOwnVoice(t *testing.T) {
	pieces := []Piece{{"Receptionist", "a"}, {"Caller", "b"}, {"Receptionist", "c"}, {"Examiner", "d"}}
	voices := VoiceFor(pieces, []string{"diana", "daniel", "hannah"}, map[string]string{"Examiner": "troy"})
	if voices["Receptionist"] != "diana" || voices["Caller"] != "daniel" || voices["Examiner"] != "troy" {
		t.Fatalf("voices = %v", voices)
	}
}
