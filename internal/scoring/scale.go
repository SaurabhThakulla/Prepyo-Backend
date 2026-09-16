package scoring

import (
	"math"

	"github.com/prepyo/backend/internal/models"
)

// Scale describes an exam's score range, taken from the exam version so PTE
// (10-90, whole points) and IELTS (0-9, half bands) both work here.
type Scale struct {
	Min  float64
	Max  float64
	Step float64
}

// EstimateFromAccuracy maps a share of marks earned onto the exam's scale.
func (s Scale) EstimateFromAccuracy(accuracy float64) float64 {
	accuracy = clamp(accuracy, 0, 1)

	raw := s.Min + accuracy*(s.Max-s.Min)
	if s.Step > 0 {
		raw = math.Round(raw/s.Step) * s.Step
	}
	return clamp(raw, s.Min, s.Max)
}

// EstimateFromRawMarks converts a raw mark count into an official exam score
// following industry standards:
// - IELTS Academic Reading (40 questions): Official 40-mark band conversion table
// - IELTS Academic Listening (40 questions): Official 40-mark band conversion table
// - IELTS partial / generic tasks: Linear accuracy rounded with the official IELTS rounding rule
// - PTE Academic: Official 10-90 scaled score with integer rounding
func (s Scale) EstimateFromRawMarks(exam, skill string, correct, total int) float64 {
	if total <= 0 {
		return s.Min
	}

	if exam == "IELTS" {
		if skill == "reading" && total == 40 {
			return IELTSReadingBand(correct)
		}
		if skill == "listening" && total == 40 {
			return IELTSListeningBand(correct)
		}
		accuracy := float64(correct) / float64(total)
		return RoundIELTSBand(s.Min + accuracy*(s.Max-s.Min))
	}

	if exam == "PTE" {
		return PTEEstimateFromRawMarks(correct, total)
	}

	return s.EstimateFromAccuracy(float64(correct) / float64(total))
}

// EstimateOverall computes the overall mock exam score according to official exam rules:
// - IELTS: Arithmetic mean of communicative skill bands, rounded using official IELTS
//   rules (.25 -> .5, .75 -> next whole band).
// - PTE Academic: Arithmetic mean of communicative skill scores, rounded to the nearest
//   integer on the 10-90 scale.
func (s Scale) EstimateOverall(exam string, skillScores map[models.SkillType]float64, totalCorrect, totalQuestions int) float64 {
	if len(skillScores) == 0 {
		return s.EstimateFromRawMarks(exam, "", totalCorrect, totalQuestions)
	}
	if len(skillScores) == 1 {
		for _, score := range skillScores {
			return score
		}
	}

	var sum float64
	for _, score := range skillScores {
		sum += score
	}
	mean := sum / float64(len(skillScores))

	if exam == "IELTS" {
		return RoundIELTSBand(mean)
	}
	if exam == "PTE" {
		return clamp(math.Round(mean), 10.0, 90.0)
	}

	if s.Step > 0 {
		mean = math.Round(mean/s.Step) * s.Step
	}
	return clamp(mean, s.Min, s.Max)
}

// RoundIELTSBand implements the official Cambridge/IDP IELTS rounding rule:
// - If the fractional part is < 0.25, it rounds down to .0 (e.g. 6.125 -> 6.0)
// - If the fractional part is >= 0.25 and < 0.75, it rounds to .5 (e.g. 6.25 -> 6.5, 6.625 -> 6.5)
// - If the fractional part is >= 0.75, it rounds up to the next whole band (e.g. 6.75 -> 7.0)
func RoundIELTSBand(raw float64) float64 {
	raw = clamp(raw, 0, 9.0)
	floor := math.Floor(raw)
	fraction := raw - floor

	switch {
	case fraction < 0.25:
		return clamp(floor, 0, 9.0)
	case fraction < 0.75:
		return clamp(floor+0.5, 0, 9.0)
	default:
		return clamp(floor+1.0, 0, 9.0)
	}
}

// PTEEstimateFromAccuracy maps an accuracy percentage (0.0 to 1.0) onto the
// official PTE Academic 10-90 scale with integer rounding.
func PTEEstimateFromAccuracy(accuracy float64) float64 {
	accuracy = clamp(accuracy, 0, 1)
	raw := 10.0 + accuracy*80.0
	return clamp(math.Round(raw), 10.0, 90.0)
}

// PTEEstimateFromRawMarks converts correct answers out of total questions into
// a PTE Academic scaled score (10 to 90).
func PTEEstimateFromRawMarks(correct, total int) float64 {
	if total <= 0 {
		return 10.0
	}
	return PTEEstimateFromAccuracy(float64(correct) / float64(total))
}

// PTEToIELTSConcordance maps a PTE Academic score to the official IELTS Band
// according to Pearson's published research concordance table.
func PTEToIELTSConcordance(pteScore float64) float64 {
	switch {
	case pteScore >= 86:
		return 9.0
	case pteScore >= 83:
		return 8.5
	case pteScore >= 79:
		return 8.0
	case pteScore >= 73:
		return 7.5
	case pteScore >= 65:
		return 7.0
	case pteScore >= 58:
		return 6.5
	case pteScore >= 50:
		return 6.0
	case pteScore >= 42:
		return 5.5
	case pteScore >= 36:
		return 5.0
	case pteScore >= 29:
		return 4.5
	default:
		return 4.0
	}
}

// IELTSToPTEConcordance maps an IELTS Band to the equivalent PTE Academic score median
// according to Pearson's published research concordance table.
func IELTSToPTEConcordance(ieltsBand float64) float64 {
	switch {
	case ieltsBand >= 9.0:
		return 88.0
	case ieltsBand >= 8.5:
		return 84.0
	case ieltsBand >= 8.0:
		return 80.0
	case ieltsBand >= 7.5:
		return 76.0
	case ieltsBand >= 7.0:
		return 68.0
	case ieltsBand >= 6.5:
		return 61.0
	case ieltsBand >= 6.0:
		return 54.0
	case ieltsBand >= 5.5:
		return 46.0
	case ieltsBand >= 5.0:
		return 39.0
	case ieltsBand >= 4.5:
		return 32.0
	default:
		return 20.0
	}
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

