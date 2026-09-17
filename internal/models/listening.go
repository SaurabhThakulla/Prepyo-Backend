package models

// CanonicalPTEListeningType maps legacy seed IDs to the authoring catalogue IDs.
// Question IDs are unchanged, so attempt and mock references remain valid.
func CanonicalPTEListeningType(id string) string {
	switch id {
	case "multiple-choice-multiple":
		return "pte-listening-mcma"
	case "fill-in-the-blanks":
		return "pte-listening-fib"
	case "highlight-correct-summary":
		return "pte-highlight-correct-summary"
	case "select-missing-word":
		return "pte-select-missing-word"
	default:
		return id
	}
}

// PTEListeningTypeIDs accepts both generations of filter IDs, including before
// the seed migration has been applied. Callers must scope these aliases to PTE listening.
func PTEListeningTypeIDs(id string) []string {
	canonical := CanonicalPTEListeningType(id)
	ids := []string{canonical}
	switch canonical {
	case "pte-listening-mcma":
		ids = append(ids, "multiple-choice-multiple")
	case "pte-listening-fib":
		ids = append(ids, "fill-in-the-blanks")
	case "pte-highlight-correct-summary":
		ids = append(ids, "highlight-correct-summary")
	case "pte-select-missing-word":
		ids = append(ids, "select-missing-word")
	}
	return ids
}
