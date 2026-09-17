package models

import (
	"reflect"
	"testing"
)

func TestPTEListeningTypeAliases(t *testing.T) {
	pairs := map[string]string{
		"multiple-choice-multiple":  "pte-listening-mcma",
		"fill-in-the-blanks":        "pte-listening-fib",
		"highlight-correct-summary": "pte-highlight-correct-summary",
		"select-missing-word":       "pte-select-missing-word",
	}
	for legacy, canonical := range pairs {
		for _, input := range []string{legacy, canonical} {
			if got := CanonicalPTEListeningType(input); got != canonical {
				t.Errorf("canonical(%q) = %q", input, got)
			}
			if got := PTEListeningTypeIDs(input); !reflect.DeepEqual(got, []string{canonical, legacy}) {
				t.Errorf("aliases(%q) = %v", input, got)
			}
		}
	}
	for _, input := range []string{"", "reading-mcq-single", "write-from-dictation", "unknown"} {
		if got := CanonicalPTEListeningType(input); got != input {
			t.Errorf("unrelated type changed: %q -> %q", input, got)
		}
	}
}
