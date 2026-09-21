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

// FluencyAccuracy scores pace and hesitation between 0 and 1.
//
// It starts from full marks and takes away for what was actually measured, so
// every deduction can be pointed at in the feedback. A learner who speaks at a
// steady pace without long silences keeps the lot.
func (d Delivery) FluencyAccuracy() float64 {
	if !d.Measured() {
		return 0
	}

	score := 1.0

	// Pace. The penalty grows with the distance outside the band rather than
	// stepping, so 95 words a minute is not treated like 40.
	if rate := d.WordsPerMinute(); rate > 0 {
		switch {
		case rate < slowWordsPerMinute:
			score -= 0.4 * clamp(float64(slowWordsPerMinute-rate)/slowWordsPerMinute, 0, 1)
		case rate > fastWordsPerMinute:
			score -= 0.3 * clamp(float64(rate-fastWordsPerMinute)/float64(fastWordsPerMinute), 0, 1)
		}
	}

	// Hesitation. Two pauses in a forty-second answer is human; ten is not.
	if d.DurationSeconds > 0 {
		pausesPerMinute := float64(d.PauseCount) * 60 / float64(d.DurationSeconds)
		if pausesPerMinute > 4 {
			score -= 0.3 * clamp((pausesPerMinute-4)/8, 0, 1)
		}
	}

	// One long silence costs on its own: it is what a listener remembers.
	if d.LongestPauseSeconds > 2 {
		score -= 0.2 * clamp((d.LongestPauseSeconds-2)/4, 0, 1)
	}

	// Mostly-silent recordings are not fluent answers, whatever the words say.
	if d.SpeakingRatio < 0.5 {
		score -= 0.3 * clamp((0.5-d.SpeakingRatio)/0.5, 0, 1)
	}

	return clamp(score, 0, 1)
}

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
