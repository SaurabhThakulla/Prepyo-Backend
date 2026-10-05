// Command mocklines prints the IELTS speaking mock examiner's lines for the
// voice generator (prepyo-voicegen), built by the same code the mock uses.
//
// It reads a JSON array of speaking_mock_sets content objects on stdin and
// writes a JSON array of distinct lines on stdout. It never connects to a
// database: export the sets with psql, pipe them through, and the generator
// speaks the result.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prepyo/backend/internal/speakingmock"
)

func main() {
	var sets []json.RawMessage
	if err := json.NewDecoder(os.Stdin).Decode(&sets); err != nil {
		fail("read sets: %v", err)
	}
	seen := map[string]bool{}
	lines := []string{}
	for i, set := range sets {
		got, err := speakingmock.ExaminerLines(set)
		if err != nil {
			fail("set %d: %v", i, err)
		}
		for _, line := range got {
			if !seen[line] {
				seen[line] = true
				lines = append(lines, line)
			}
		}
	}
	out := json.NewEncoder(os.Stdout)
	out.SetEscapeHTML(false)
	out.SetIndent("", "  ")
	if err := out.Encode(lines); err != nil {
		fail("write lines: %v", err)
	}
	fmt.Fprintf(os.Stderr, "%d lines from %d sets\n", len(lines), len(sets))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "mocklines: "+format+"\n", args...)
	os.Exit(1)
}
