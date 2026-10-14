package scoring

import (
	"fmt"
	"strings"
)

// Delivery is what the browser measured while the learner was speaking.
//
// None of it needs a model: the recorder already watches the microphone level
// to draw its meter, so counting the silences costs nothing and turns fluency
// feedback from a guess into a measurement. It is not pronunciation - it is
// pace and hesitation, which is the part of delivery a waveform can show.
type Delivery struct {
	DurationSeconds int
	WordsSpoken     int
	// PauseCount is silences longer than PauseThresholdSeconds. Ordinary
	// sentence breaks are shorter than that and are not counted.
	PauseCount          int
	LongestPauseSeconds float64
	// SpeakingRatio is the share of the recording with voice in it, 0 to 1.
	SpeakingRatio float64
}

// PauseThresholdSeconds is the gap at which a breath becomes a hesitation.
// Natural phrase breaks in fluent speech sit well under half a second.
const PauseThresholdSeconds = 0.6

// Speech-rate band for exam speaking. Below the floor a learner is groping for
// words; above the ceiling they are rushing, which examiners mark down for
// intelligibility rather than reward for confidence.
const (
	slowWordsPerMinute = 100
	fastWordsPerMinute = 180
)

// WordsPerMinute is the measured speech rate, or 0 when nothing was timed.
func (d Delivery) WordsPerMinute() int {
	if d.DurationSeconds <= 0 {
		return 0
	}
	return d.WordsSpoken * 60 / d.DurationSeconds
}

// Measured reports whether the browser supplied delivery data at all. Without
// it the caller falls back to judging content alone rather than inventing a
// fluency mark from nothing.
func (d Delivery) Measured() bool {
	return d.DurationSeconds > 0 && d.SpeakingRatio > 0
}

// FluencyPoints rates pace and hesitation on Pearson's Oral Fluency scale (0-5).
func (d Delivery) FluencyPoints() int {
	// Hesitations, per the ladder: none 5, one 4, more than one 3, two or three 2, many 1.
	var points int
	switch {
	case d.PauseCount == 0:
		points = 5
	case d.PauseCount == 1:
		points = 4
	case d.PauseCount <= 3:
		points = 3
	case d.PauseCount <= 5:
		points = 2
	default:
		points = 1
	}
	if d.LongestPauseSeconds >= longPauseSeconds {
		points = min(points, 2)
	}
	// Pace is judged only on recordings long enough for lead-in silence not to dominate.
	if rate := d.WordsPerMinute(); d.DurationSeconds >= 10 && rate > 0 {
		switch {
		case rate < 70:
			points = min(points, 1)
		case rate < slowWordsPerMinute, rate > fastWordsPerMinute:
			points--
		}
	}
	if d.SpeakingRatio < 0.5 {
		points--
	}
	return max(points, 0)
}

// longPauseSeconds is the silence Pearson's ladder calls a long pause.
const longPauseSeconds = 2.0

// Feedback describes the delivery in the terms it was measured in, so a learner
// can check it against their own recording.
func (d Delivery) Feedback() string {
	if !d.Measured() {
		return "There was not enough timing information to judge your pace."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "You spoke about %d words a minute over %d seconds",
		d.WordsPerMinute(), d.DurationSeconds)

	switch rate := d.WordsPerMinute(); {
	case rate < slowWordsPerMinute:
		b.WriteString(", which is slower than an examiner expects")
	case rate > fastWordsPerMinute:
		b.WriteString(", which is fast enough to cost you clarity")
	default:
		b.WriteString(", a comfortable exam pace")
	}

	if d.PauseCount > 0 {
		fmt.Fprintf(&b, ", with %d pause", d.PauseCount)
		if d.PauseCount != 1 {
			b.WriteString("s")
		}
		fmt.Fprintf(&b, " longer than %.1fs", PauseThresholdSeconds)
		if d.LongestPauseSeconds >= 1 {
			fmt.Fprintf(&b, " and a longest silence of %.1fs", d.LongestPauseSeconds)
		}
	} else {
		b.WriteString(", with no long hesitations")
	}

	b.WriteString(".")
	return b.String()
}

// Summary is the one-line delivery note handed to a model as evidence, so it
// comments on measured pace instead of guessing at it.
func (d Delivery) Summary() string {
	if !d.Measured() {
		return ""
	}
	return fmt.Sprintf(
		"Delivery measured from the recording: %d words in %d seconds (about %d words per minute), "+
			"%d pauses longer than %.1fs, longest silence %.1fs, speaking for %.0f%% of the time.",
		d.WordsSpoken, d.DurationSeconds, d.WordsPerMinute(),
		d.PauseCount, PauseThresholdSeconds, d.LongestPauseSeconds, d.SpeakingRatio*100)
}
