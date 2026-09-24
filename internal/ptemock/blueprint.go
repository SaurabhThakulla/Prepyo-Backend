package ptemock

import "github.com/prepyo/backend/internal/models"

// Kind is a kind of PTE mock: the whole test, or one skill's sectional test.
type Kind string

const (
	KindFull      Kind = "full"
	KindSpeaking  Kind = "speaking"
	KindWriting   Kind = "writing"
	KindReading   Kind = "reading"
	KindListening Kind = "listening"
)

// Slot is how many items of one task a paper deals.
type Slot struct {
	Task  string `json:"task"`
	Count int    `json:"count"`
}

// Blueprint is what one kind of mock is made of, in the order of the test.
type Blueprint struct {
	Kind        Kind   `json:"kind"`
	MockID      string `json:"mockId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Skills are the scores the report leads with: all four for the full test,
	// the section's own for a sectional test.
	Skills []models.SkillType `json:"skills"`
	Slots  []Slot             `json:"slots"`
	// FullAllowance marks the kind paid for from the plan's full-mock
	// allowance; every other kind costs SectionMockSubTests sub-tests.
	FullAllowance bool `json:"fullAllowance"`
}

// Item counts follow the published ranges for each task in the current test,
// at the level most test forms use. A paper is dealt from the bank as it is:
// where a task has fewer items than the count, the paper carries fewer, and a
// task with none is left out and named (see Paper.Missing). Skill scores are
// weighted means over tasks, so a short task does not tilt the result.
var speakingSlots = []Slot{
	{"RA", 6}, {"RS", 10}, {"DI", 5}, {"RL", 2}, {"ASQ", 5}, {"SGD", 2}, {"RTS", 2},
}

var readingSlots = []Slot{
	{"RWFIB", 5}, {"RMCMA", 1}, {"RO", 2}, {"RFIB", 4}, {"RMCSA", 1},
}

var listeningSlots = []Slot{
	{"SST", 1}, {"LMCMA", 1}, {"LFIB", 2}, {"HCS", 1}, {"LMCSA", 1}, {"SMW", 1}, {"HIW", 2}, {"WFD", 3},
}

func join(parts ...[]Slot) []Slot {
	var out []Slot
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// Blueprints is every kind of PTE mock, in the order they are offered.
var Blueprints = []Blueprint{
	{
		Kind:          KindFull,
		MockID:        "mock-pte-full",
		Title:         "PTE Academic Full Mock Test",
		Description:   "The whole test in one sitting: Speaking & Writing, Reading, then Listening, with an overall score and all four communicative skill scores.",
		Skills:        models.AllSkills,
		Slots:         join(speakingSlots, []Slot{{"SWT", 1}, {"WE", 1}}, readingSlots, listeningSlots),
		FullAllowance: true,
	},
	{
		Kind:        KindSpeaking,
		MockID:      "mock-pte-speaking",
		Title:       "Speaking Sectional Test",
		Description: "All seven speaking tasks in the order of the test, each with its own preparation and recording time.",
		Skills:      []models.SkillType{skSpeaking},
		Slots:       speakingSlots,
	},
	{
		Kind:        KindWriting,
		MockID:      "mock-pte-writing",
		Title:       "Writing Sectional Test",
		Description: "Summarize Written Text and Write Essay, each on its own clock.",
		Skills:      []models.SkillType{skWriting},
		Slots:       []Slot{{"SWT", 2}, {"WE", 1}},
	},
	{
		Kind:        KindReading,
		MockID:      "mock-pte-reading",
		Title:       "Reading Sectional Test",
		Description: "The reading part: five task types, one item per screen, on one shared clock.",
		Skills:      []models.SkillType{skReading},
		Slots:       readingSlots,
	},
	{
		Kind:        KindListening,
		MockID:      "mock-pte-listening",
		Title:       "Listening Sectional Test",
		Description: "The listening part: every recording plays once, and Summarize Spoken Text has its own clock.",
		Skills:      []models.SkillType{skListening},
		Slots:       listeningSlots,
	},
}

// BlueprintFor is a kind's blueprint.
func BlueprintFor(kind Kind) (Blueprint, bool) {
	for _, b := range Blueprints {
		if b.Kind == kind {
			return b, true
		}
	}
	return Blueprint{}, false
}

// MockIDs are the attempt rows every PTE mock records against.
func MockIDs() []string {
	ids := make([]string, 0, len(Blueprints))
	for _, b := range Blueprints {
		ids = append(ids, b.MockID)
	}
	return ids
}

// ItemCount is how many items the blueprint deals when the bank is full.
func (b Blueprint) ItemCount() int {
	n := 0
	for _, s := range b.Slots {
		n += s.Count
	}
	return n
}

// EstimatedMinutes is how long the blueprint's paper takes, from the tasks'
// own timings: each item clock in full, each section clock in full, and each
// speaking item end to end.
func (b Blueprint) EstimatedMinutes() int {
	seconds := 0
	for _, s := range b.Slots {
		task := Tasks[s.Task]
		switch task.Clock {
		case ClockResponse:
			seconds += s.Count * task.EstimateSeconds
		default:
			seconds += s.Count * task.Seconds
		}
	}
	return (seconds + 59) / 60
}

// Parts are the parts the blueprint covers, in the order of the test.
func (b Blueprint) Parts() []Part {
	has := map[Part]bool{}
	for _, s := range b.Slots {
		has[Tasks[s.Task].Part] = true
	}
	var parts []Part
	for _, p := range PartOrder {
		if has[p] {
			parts = append(parts, p)
		}
	}
	return parts
}
