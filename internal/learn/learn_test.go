package learn

import (
	"encoding/json"
	"testing"
)

// Only a JSON object is stored: anything else is refused before the database.
func TestIsObject(t *testing.T) {
	cases := map[string]bool{
		`{"completed":{},"xp":10}`: true,
		`{}`:                       true,
		`[]`:                       false,
		`"text"`:                   false,
		`42`:                       false,
		`null`:                     false,
		``:                         false,
	}
	for raw, want := range cases {
		if got := isObject(json.RawMessage(raw)); got != want {
			t.Errorf("isObject(%q) = %v, want %v", raw, got, want)
		}
	}
}
