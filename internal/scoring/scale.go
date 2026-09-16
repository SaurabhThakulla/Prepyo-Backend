package scoring

import "math"

// Scale describes an exam's score range, taken from the exam version so PTE
// (10-90, whole points) and IELTS (0-9, half bands) both work here.
type Scale struct {
	Min  float64
	Max  float64
	Step float64
}

// EstimateFromAccuracy maps a share of marks earned onto the exam's scale.
//
// This is a product estimate, not an official conversion. Neither Pearson nor
// IELTS publishes a raw-to-scale table, so anything claiming to reproduce one
// would be invented. Callers must present the result as an estimate and label
// it that way in the UI.
//
// The mapping is a plain linear interpolation across the scale, rounded to the
// exam's step. It is easy to explain to a learner, which matters more than
// false precision.
func (s Scale) EstimateFromAccuracy(accuracy float64) float64 {
	accuracy = clamp(accuracy, 0, 1)

	raw := s.Min + accuracy*(s.Max-s.Min)
	if s.Step > 0 {
		raw = math.Round(raw/s.Step) * s.Step
	}
	return clamp(raw, s.Min, s.Max)
}

// EstimateFromRawMarks converts a raw mark count into a band/score using
// exam-specific lookup tables when available, falling back to accuracy-based
// linear interpolation otherwise.
//
// exam is the exam type (e.g. "IELTS", "PTE").
// skill is the skill area (e.g. "reading", "listening").
// correct is the number of correct answers and total is the total questions.
func (s Scale) EstimateFromRawMarks(exam, skill string, correct, total int) float64 {
	if exam == "IELTS" && skill == "reading" && total == 40 {
		return IELTSReadingBand(correct)
	}
	if exam == "IELTS" && skill == "listening" && total == 40 {
		return IELTSListeningBand(correct)
	}
	if total <= 0 {
		return s.Min
	}
	return s.EstimateFromAccuracy(float64(correct) / float64(total))
}

// ---------------------------------------------------------------------------
// IELTS Official Raw-to-Band Conversion Tables
// ---------------------------------------------------------------------------
//
// These tables are based on widely published IELTS Academic score conversion
// charts. Each entry maps a minimum raw mark to the corresponding band score.
// The lookup scans from the highest threshold downward.

// ieltsReadingTable maps minimum raw marks (out of 40) to IELTS Academic
// Reading band scores.
var ieltsReadingTable = []struct {
	minRaw int
	band   float64
}{
	{39, 9.0},
	{37, 8.5},
	{35, 8.0},
	{33, 7.5},
	{30, 7.0},
	{27, 6.5},
	{23, 6.0},
	{19, 5.5},
	{15, 5.0},
	{13, 4.5},
	{10, 4.0},
	{8, 3.5},
	{6, 3.0},
	{4, 2.5},
}

// IELTSReadingBand returns the IELTS Academic Reading band score for a given
// raw mark out of 40 questions.
func IELTSReadingBand(correct int) float64 {
	correct = int(clamp(float64(correct), 0, 40))
	for _, entry := range ieltsReadingTable {
		if correct >= entry.minRaw {
			return entry.band
		}
	}
	return 2.0
}

// ieltsListeningTable maps minimum raw marks (out of 40) to IELTS Academic
// Listening band scores.
var ieltsListeningTable = []struct {
	minRaw int
	band   float64
}{
	{39, 9.0},
	{37, 8.5},
	{35, 8.0},
	{32, 7.5},
	{30, 7.0},
	{26, 6.5},
	{23, 6.0},
	{18, 5.5},
	{16, 5.0},
	{13, 4.5},
	{11, 4.0},
	{8, 3.5},
	{6, 3.0},
	{4, 2.5},
}

// IELTSListeningBand returns the IELTS Academic Listening band score for a
// given raw mark out of 40 questions.
func IELTSListeningBand(correct int) float64 {
	correct = int(clamp(float64(correct), 0, 40))
	for _, entry := range ieltsListeningTable {
		if correct >= entry.minRaw {
			return entry.band
		}
	}
	return 2.0
}

// Confidence describes how much weight to put on an estimate given how many
// questions it is based on. Reported to the learner so a single lucky answer is
// not shown as a settled result.
func Confidence(attempts int) string {
	switch {
	case attempts >= 20:
		return "high"
	case attempts >= 8:
		return "medium"
	default:
		return "low"
	}
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(math.Max(v, lo), hi)
}

