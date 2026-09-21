package scoring

import (
	"fmt"
	"strings"
)

// Verbatim scoring: the tasks where the learner is told exactly what to say.
//
// Read Aloud gives them a passage; Repeat Sentence gives them a sentence. In
// both the right answer is known word for word, so content does not need a
// model's opinion - it needs an alignment. Counting which words survived is
// reproducible, instant, free, and closer to how Pearson scores these than any
// language model guess: the same recording always earns the same content mark,
// and the learner can see exactly which words were missed.
//
// What this does not do is judge pronunciation. Word matching runs on a
// transcript, and a transcript is words.

// VerbatimTypes are the task types with a fixed text to say. Re-tell Lecture
// and Summarise Group Discussion also carry a transcript, but of the material
// the learner listens to, not of what they are meant to say: aligning against
// those would punish a learner for using their own words, which is the point of
// the task.
var VerbatimTypes = map[string]bool{
	"read-aloud":          true,
	"pte-read-aloud":      true,
	"repeat-sentence":     true,
	"pte-repeat-sentence": true,
}

// IsVerbatimTask reports whether a task type is scored by alignment.
func IsVerbatimTask(typeID string) bool {
	return VerbatimTypes[strings.ToLower(strings.TrimSpace(typeID))]
}

// VerbatimResult is one alignment, in the terms the feedback is written in.
type VerbatimResult struct {
	// Accuracy is the share of expected words the learner actually said, in
	// order. 1 means every word, in sequence.
	Accuracy float64
	Expected int
	Matched  int
	// Missed are expected words that never arrived, in the order they were
	// meant to be said. Added are words the learner said that were not in the
	// text. Both are capped for display; Accuracy is over everything.
	Missed []string
	Added  []string
}

// maxListedWords keeps feedback readable. A learner who missed forty words does
// not need all forty listed back at them.
const maxListedWords = 8

// ScoreVerbatim aligns what was said against what was given.
//
// The alignment is a longest common subsequence rather than a bag of words: a
// learner who says every word in the wrong order has not read the passage
// aloud, and order is exactly what these tasks test. Speech recognition errors
// count against the learner here, which is why the caller keeps this to content
// and leaves the rest of the band to other criteria.
func ScoreVerbatim(expected, spoken string) VerbatimResult {
	expectedWords := words(expected)
	spokenWords := words(spoken)

	result := VerbatimResult{Expected: len(expectedWords)}
	if len(expectedWords) == 0 {
		return result
	}

	matchedExpected, matchedSpoken := longestCommonSubsequence(expectedWords, spokenWords)
	result.Matched = len(matchedExpected)
	result.Accuracy = float64(result.Matched) / float64(len(expectedWords))

	for i, word := range expectedWords {
		if !matchedExpected[i] && len(result.Missed) < maxListedWords {
			result.Missed = append(result.Missed, word)
		}
	}
	for i, word := range spokenWords {
		if !matchedSpoken[i] && len(result.Added) < maxListedWords {
			result.Added = append(result.Added, word)
		}
	}
	return result
}

// longestCommonSubsequence marks which positions on each side are part of the
// longest in-order match. The tables are O(n*m); the inputs here are one
// sentence or one short passage, so that is a few thousand cells at worst.
func longestCommonSubsequence(expected, spoken []string) (map[int]bool, map[int]bool) {
	rows, cols := len(expected), len(spoken)
	table := make([][]int, rows+1)
	for i := range table {
		table[i] = make([]int, cols+1)
	}

	for i := rows - 1; i >= 0; i-- {
		for j := cols - 1; j >= 0; j-- {
			if expected[i] == spoken[j] {
				table[i][j] = table[i+1][j+1] + 1
				continue
			}
			table[i][j] = max(table[i+1][j], table[i][j+1])
		}
	}

	inExpected := make(map[int]bool, rows)
	inSpoken := make(map[int]bool, cols)
	i, j := 0, 0
	for i < rows && j < cols {
		switch {
		case expected[i] == spoken[j]:
			inExpected[i], inSpoken[j] = true, true
			i, j = i+1, j+1
		case table[i+1][j] >= table[i][j+1]:
			i++
		default:
			j++
		}
	}
	return inExpected, inSpoken
}

// VerbatimFeedback is the sentence a learner reads about their content mark.
func (r VerbatimResult) Feedback() string {
	var b strings.Builder
	fmt.Fprintf(&b, "You said %d of the %d words given, in order (%.0f%%).",
		r.Matched, r.Expected, r.Accuracy*100)

	if len(r.Missed) > 0 {
		fmt.Fprintf(&b, " Words missed or changed: %s.", strings.Join(r.Missed, ", "))
	}
	if len(r.Added) > 0 {
		fmt.Fprintf(&b, " Words added that were not in the text: %s.", strings.Join(r.Added, ", "))
	}
	return b.String()
}
