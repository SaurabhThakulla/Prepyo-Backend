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
