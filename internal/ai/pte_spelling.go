package ai

import (
	_ "embed"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

// words_en.txt is SCOWL size 60, US + British spellings, from
// https://app.aspell.net/create (licence notice kept at the top of the file).
//
//go:embed words_en.txt
var wordListText string

var (
	wordListOnce sync.Once
	wordList     map[string]bool
)

func englishWords() map[string]bool {
	wordListOnce.Do(func() {
		_, body, ok := strings.Cut(wordListText, "\n---")
		if !ok {
			return
		}
		wordList = map[string]bool{}
		for _, w := range strings.Fields(body) {
			wordList[normWord(w)] = true
		}
	})
	return wordList
}

func normWord(w string) string {
	return strings.Trim(strings.ToLower(strings.ReplaceAll(w, "’", "'")), "'")
}

var wordRe = regexp.MustCompile(`[\p{L}'’]+`)

// spellingErrors counts misspelled words in text: words not in the word list
// and not used in the task's source (prompt or transcript). A capitalised word
// mid-sentence is taken as a name. Without the word list it falls back to the
// words the model listed.
func spellingErrors(text, source string, listed []string) int {
	dict := englishWords()
	if len(dict) == 0 {
		return listedErrors(text, listed)
	}
	known := map[string]bool{}
	for _, w := range wordRe.FindAllString(source, -1) {
		known[normWord(w)] = true
	}
	errors := 0
	for _, loc := range wordRe.FindAllStringIndex(text, -1) {
		raw := text[loc[0]:loc[1]]
		w := normWord(raw)
		if w == "" || dict[w] || dict[strings.TrimSuffix(w, "'s")] || known[w] {
			continue
		}
		if unicode.IsUpper([]rune(raw)[0]) && !sentenceStart(text[:loc[0]]) {
			continue
		}
		errors++
	}
	return errors
}

func sentenceStart(before string) bool {
	before = strings.TrimSpace(before)
	return before == "" || strings.ContainsAny(before[len(before)-1:], ".!?\n")
}

// listedErrors counts each occurrence in the text of a word the model listed
// as misspelled; listed words that are not in the text are ignored.
func listedErrors(text string, listed []string) int {
	set := map[string]bool{}
	for _, m := range listed {
		set[normWord(m)] = true
	}
	errors := 0
	for _, w := range wordRe.FindAllString(text, -1) {
		if set[normWord(w)] {
			errors++
		}
	}
	return errors
}
