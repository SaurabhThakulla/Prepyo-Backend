package admin

import (
	"fmt"
	"strings"
)

func summaryKeywords(answers []string) []string {
	keywords := []string{}
	seen := map[string]bool{}
	for _, answer := range answers {
		key := strings.ToLower(strings.TrimSpace(answer))
		if key != "" && !seen[key] {
			keywords = append(keywords, key)
			seen[key] = true
		}
	}
	return keywords
}

// Word IDs identify occurrences, not spellings: repeated words are distinct.
// The client renders the returned options in order and submits their IDs.
// Punctuation stays attached to its whitespace-delimited word.
func highlightedWordKey(passage string, answers []string) ([]questionOption, []string, map[string]string) {
	problems := map[string]string{}
	words := strings.Fields(passage)
	if len(words) < 2 || len(words) > 500 {
		problems["contextPassage"] = "Write a displayed transcript of 2-500 words."
		return nil, nil, problems
	}
	options := make([]questionOption, len(words))
	valid := make(map[string]bool, len(words))
	for i, word := range words {
		id := fmt.Sprintf("w%d", i+1)
		options[i] = questionOption{ID: id, Text: word}
		valid[id] = true
	}
	correct := []string{}
	seen := map[string]bool{}
	for _, answer := range answers {
		id := strings.TrimSpace(answer)
		if !valid[id] {
			problems["correctAnswers"] = "Identify each incorrect word by its position ID (w1, w2, ...)."
			continue
		}
		if !seen[id] {
			correct = append(correct, id)
			seen[id] = true
		}
	}
	if len(correct) == 0 || len(correct) == len(words) {
		problems["correctAnswers"] = "Mark at least one incorrect word and leave at least one word unmarked."
	}
	return options, correct, problems
}
