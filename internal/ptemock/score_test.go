package ptemock

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func TestBlueprintsNameRealTasks(t *testing.T) {
	seen := map[Kind]bool{}
	for _, bp := range Blueprints {
		if seen[bp.Kind] {
			t.Fatalf("%s: two blueprints", bp.Kind)
		}
		seen[bp.Kind] = true
		if len(bp.Skills) == 0 || len(bp.Slots) == 0 {
			t.Fatalf("%s: empty blueprint", bp.Kind)
		}
		lastPart := -1
		for _, slot := range bp.Slots {
			task, ok := Tasks[slot.Task]
			if !ok {
				t.Fatalf("%s: unknown task %q", bp.Kind, slot.Task)
			}
			if slot.Count <= 0 {
				t.Fatalf("%s/%s: count %d", bp.Kind, slot.Task, slot.Count)
			}
			// Parts come in the order of the test and are never revisited.
			part := indexOf(task.Part)
			if part < lastPart {
				t.Fatalf("%s: %s (%s) comes after a later part", bp.Kind, slot.Task, task.Part)
			}
			lastPart = part
		}
	}
	for _, kind := range []Kind{KindFull, KindSpeaking, KindWriting, KindReading, KindListening} {
		if !seen[kind] {
			t.Fatalf("no blueprint for %s", kind)
		}
	}
}

func indexOf(p Part) int {
	for i, q := range PartOrder {
		if q == p {
			return i
		}
	}
	return -1
}

func TestTasksAreConsistent(t *testing.T) {
	for code, task := range Tasks {
		if task.Code != code {
			t.Errorf("%s: code %q", code, task.Code)
		}
		if len(task.TypeIDs) == 0 || task.Name == "" || task.Instructions == "" {
			t.Errorf("%s: incomplete", code)
		}
		if task.Weights[task.Skill] <= 0 {
			t.Errorf("%s: does not count towards its own skill %s", code, task.Skill)
		}
		switch task.Clock {
		case ClockResponse:
			if task.Marking != MarkSpoken || task.ResponseSeconds <= 0 || task.EstimateSeconds <= 0 {
				t.Errorf("%s: a speaking item needs a response window and an estimate", code)
			}
		case ClockItem, ClockSection:
			if task.Seconds <= 0 {
				t.Errorf("%s: a timed item needs seconds", code)
			}
		default:
			t.Errorf("%s: unknown clock %q", code, task.Clock)
		}
		if task.Marking == MarkWritten && (task.MinWords <= 0 || task.MaxWords <= task.MinWords) {
			t.Errorf("%s: written form limits", code)
		}
	}
}

func TestFullTestTakesAboutTwoHours(t *testing.T) {
	full, _ := BlueprintFor(KindFull)
	if m := full.EstimatedMinutes(); m < 105 || m > 140 {
		t.Fatalf("full test estimated at %d minutes; PTE Academic is about two hours", m)
	}
	reading, _ := BlueprintFor(KindReading)
	if m := reading.EstimatedMinutes(); m < 25 || m > 32 {
		t.Fatalf("reading part estimated at %d minutes; the test gives 29-30", m)
	}
	if got := full.Parts(); len(got) != 3 {
		t.Fatalf("full test parts %v", got)
	}
}

func TestToScale(t *testing.T) {
	for _, tc := range []struct {
		fraction float64
		want     int
	}{{0, 10}, {1, 90}, {0.5, 50}, {-1, 10}, {2, 90}, {0.625, 60}} {
		if got := ToScale(tc.fraction); got != tc.want {
			t.Errorf("ToScale(%v) = %d, want %d", tc.fraction, got, tc.want)
		}
	}
}

func TestScoreSectional(t *testing.T) {
	reading, _ := BlueprintFor(KindReading)
	marks := []Mark{
		{Task: "RWFIB", Scored: true, Score: 4, MaxScore: 4},
		{Task: "RWFIB", Scored: true, Score: 0, MaxScore: 4},
		{Task: "RO", Scored: true, Score: 3, MaxScore: 3},
	}
	result := Score(reading, marks, 3)
	// RWFIB mean 0.5 at weight 13, RO 1.0 at weight 7: (6.5 + 7) / 20 = 0.675.
	if got, want := result.Skills[models.SkillReading], ToScale(0.675); got != want {
		t.Fatalf("reading %d, want %d", got, want)
	}
	if result.Overall != nil {
		t.Fatal("a sectional test has no overall score")
	}
	if len(result.Skills) != 1 {
		t.Fatalf("a reading test reports reading only, got %v", result.Skills)
	}
	if headline, ok := result.Headline(reading); !ok || int(headline) != result.Skills[models.SkillReading] {
		t.Fatalf("headline %v %v", headline, ok)
	}
}

func TestScoreFullCountsIntegratedSkills(t *testing.T) {
	full, _ := BlueprintFor(KindFull)
	marks := []Mark{
		{Task: "RA", Scored: true, Score: 1, MaxScore: 1},  // speaking and reading
		{Task: "WFD", Scored: true, Score: 0, MaxScore: 1}, // listening and writing
		{Task: "WE", Scored: true, Score: 1, MaxScore: 1},  // writing
		{Task: "RO", Scored: true, Score: 0, MaxScore: 1},  // reading
	}
	result := Score(full, marks, 4)
	if result.Skills[models.SkillSpeaking] != 90 {
		t.Fatalf("speaking %d", result.Skills[models.SkillSpeaking])
	}
	if result.Skills[models.SkillListening] != 10 {
		t.Fatalf("listening %d", result.Skills[models.SkillListening])
	}
	// Writing: WFD 0 at weight 10, WE 1 at weight 12 -> 12/22.
	if got, want := result.Skills[models.SkillWriting], ToScale(12.0/22.0); got != want {
		t.Fatalf("writing %d, want %d", got, want)
	}
	// Reading: RA 1 at weight 6, RO 0 at weight 7 -> 6/13.
	if got, want := result.Skills[models.SkillReading], ToScale(6.0/13.0); got != want {
		t.Fatalf("reading %d, want %d", got, want)
	}
	if result.Overall == nil {
		t.Fatal("a full test has an overall score")
	}
}

func TestUnscoredItemsAreLeftOutNotZero(t *testing.T) {
	speaking, _ := BlueprintFor(KindSpeaking)
	marks := []Mark{
		{Task: "RA", Scored: true, Score: 1, MaxScore: 1},
		{Task: "DI", Scored: false},
	}
	result := Score(speaking, marks, 2)
	if result.Skills[models.SkillSpeaking] != 90 {
		t.Fatalf("an unscored item pulled the score down: %d", result.Skills[models.SkillSpeaking])
	}
	if result.Unscored != 1 {
		t.Fatalf("unscored %d", result.Unscored)
	}
}

func TestNoScoredItemsMeansNoScore(t *testing.T) {
	writing, _ := BlueprintFor(KindWriting)
	result := Score(writing, []Mark{{Task: "WE", Scored: false}}, 1)
	if _, ok := result.Headline(writing); ok {
		t.Fatal("a paper with nothing scored has no headline score")
	}
}

func TestFractionOfEvaluation(t *testing.T) {
	e := models.Evaluation{Criteria: []models.EvaluationCriterion{
		{Score: 2, MaxScore: 2}, {Score: 1, MaxScore: 2}, {Score: 5, MaxScore: 0},
	}}
	if f, ok := FractionOfEvaluation(e); !ok || f != 0.75 {
		t.Fatalf("criteria fraction %v %v", f, ok)
	}
	estimate := 50.0
	if f, ok := FractionOfEvaluation(models.Evaluation{EstimatedScore: &estimate}); !ok || f != 0.5 {
		t.Fatalf("estimate fraction %v %v", f, ok)
	}
	if _, ok := FractionOfEvaluation(models.Evaluation{}); ok {
		t.Fatal("an empty evaluation has no fraction")
	}
}

func TestReorderKeepsDealtOrder(t *testing.T) {
	options := []models.QuestionOption{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}}
	got := reorder(options, []string{"C", "A", "D", "B"})
	want := "CADB"
	for i, o := range got {
		if o.ID != string(want[i]) {
			t.Fatalf("order %v", got)
		}
	}
	// An option the order does not name still reaches the screen.
	got = reorder(options, []string{"B"})
	if len(got) != 4 || got[0].ID != "B" {
		t.Fatalf("partial order %v", got)
	}
}

func TestClockKeys(t *testing.T) {
	items := []itemRow{
		{Position: 1, Task: "RA", Part: PartSpeakingWriting},
		{Position: 2, Task: "WE", Part: PartSpeakingWriting},
		{Position: 3, Task: "RWFIB", Part: PartReading},
		{Position: 4, Task: "RO", Part: PartReading},
	}
	if clockKey(items[0]) != "" {
		t.Fatal("speaking items keep their own windows")
	}
	if clockKey(items[1]) != "item:2" || clockSeconds("item:2", items) != 1200 {
		t.Fatal("the essay has its own 20 minutes")
	}
	if clockKey(items[2]) != "part:reading" || clockSeconds("part:reading", items) != 300 {
		t.Fatalf("reading shares one clock: %d", clockSeconds("part:reading", items))
	}
}
