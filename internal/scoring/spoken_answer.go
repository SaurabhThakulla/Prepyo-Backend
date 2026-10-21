package scoring

import "strings"

// MatchSpokenAnswer reports whether a spoken reply gives one of the accepted
// answers anywhere in it ("it's a thermometer"), ignoring case, punctuation,
// articles and possessives, simple plurals and numbers said as words or digits.
func MatchSpokenAnswer(accepted []string, spoken string) bool {
	said := answerWords(spoken)
	if len(said) == 0 {
		return false
	}
	for _, answer := range accepted {
		if key := answerWords(answer); len(key) > 0 && containsRun(said, key) {
			return true
		}
	}
	return false
}

// determiners are left out of an answer: "your teeth" is "teeth".
var determiners = map[string]bool{"a": true, "an": true, "the": true,
	"my": true, "your": true, "his": true, "her": true, "their": true, "our": true, "its": true}

// answerWords is phrase() with numbers spelled out and determiners dropped.
func answerWords(s string) []string {
	ws := strings.Fields(phrase(spellNumbers(s)))
	out := []string{}
	for i, w := range ws {
		// "a hundred" is one hundred, so it cannot match "two hundred".
		if (w == "a" || w == "an") && i+1 < len(ws) && (ws[i+1] == "hundred" || ws[i+1] == "thousand" || ws[i+1] == "million") {
			out = append(out, "one")
			continue
		}
		// "twenty four" stays one number, so it cannot match "four".
		if n := len(out); n > 0 && isUnit(w) && isTens(out[n-1]) {
			out[n-1] += "-" + w
			continue
		}
		if !determiners[w] {
			out = append(out, w)
		}
	}
	return out
}

func isTens(w string) bool {
	for _, t := range tens[2:] {
		if w == t {
			return true
		}
	}
	return false
}

func isUnit(w string) bool {
	for _, o := range ones[1:10] {
		if w == o {
			return true
		}
	}
	return false
}

// containsRun reports whether key appears in words as consecutive words.
func containsRun(words, key []string) bool {
	for i := 0; i+len(key) <= len(words); i++ {
		match := true
		for j, k := range key {
			if !sameWord(words[i+j], k) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// sameWord is equality up to a regular plural: -s, -es, or -ies for -y.
func sameWord(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	// A base under three letters is too short to tell a plural apart ("i" and "is").
	if len(a) < 3 {
		return false
	}
	return b == a+"s" || b == a+"es" ||
		(strings.HasSuffix(a, "y") && b == strings.TrimSuffix(a, "y")+"ies")
}
