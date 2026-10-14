package scoring

import (
	"regexp"
	"strconv"
	"strings"
)

// Speech recognition writes numbers as digits ("£180", "1-2", "9th") where a
// passage may spell them out, and the reverse. spellNumbers writes every
// number as words so the two sides of a Read Aloud comparison agree.
var (
	numberToken = regexp.MustCompile(`([£$€])?(\d[\d,]*(?:\.\d+)?)(st|nd|rd|th)?(%)?(?:\s*[-–]\s*(\d+))?`)
	ones        = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
		"eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens          = []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	ordinalEnding = map[string]string{"one": "first", "two": "second", "three": "third", "five": "fifth",
		"eight": "eighth", "nine": "ninth", "twelve": "twelfth"}
	currencyName = map[string]string{"£": "pounds", "$": "dollars", "€": "euros"}
)

func spellNumbers(text string) string {
	return numberToken.ReplaceAllStringFunc(text, func(token string) string {
		m := numberToken.FindStringSubmatch(token)
		symbol, digits, suffix, percent, rangeEnd := m[1], strings.ReplaceAll(m[2], ",", ""), m[3], m[4], m[5]
		whole, fraction, _ := strings.Cut(digits, ".")
		n, err := strconv.Atoi(whole)
		if err != nil || n > 999_999_999 {
			return token
		}
		var out string
		switch {
		case suffix != "":
			out = ordinal(cardinal(n))
		case symbol == "" && percent == "" && fraction == "" && rangeEnd == "" && len(whole) == 4 && n >= 1100 && n < 2100:
			out = year(n)
		default:
			out = cardinal(n)
		}
		if fraction != "" && symbol == "" {
			out += " point " + digitsByOne(fraction)
		}
		if rangeEnd != "" {
			if end, err := strconv.Atoi(rangeEnd); err == nil {
				out += " to " + cardinal(end)
			}
		}
		if symbol != "" {
			out += " " + currencyName[symbol]
			if fraction != "" {
				if cents, err := strconv.Atoi(fraction); err == nil && cents > 0 {
					out += " " + cardinal(cents)
				}
			}
		}
		if percent != "" {
			out += " per cent"
		}
		return " " + out + " "
	})
}

// cardinal writes n in British English: 180 is "one hundred and eighty".
func cardinal(n int) string {
	switch {
	case n < 20:
		return ones[n]
	case n < 100:
		if n%10 == 0 {
			return tens[n/10]
		}
		return tens[n/10] + "-" + ones[n%10]
	}
	for _, unit := range []struct {
		size int
		name string
	}{{1_000_000, "million"}, {1000, "thousand"}, {100, "hundred"}} {
		if n >= unit.size {
			out := cardinal(n/unit.size) + " " + unit.name
			switch rest := n % unit.size; {
			case rest == 0:
			case rest < 100:
				out += " and " + cardinal(rest)
			default:
				out += " " + cardinal(rest)
			}
			return out
		}
	}
	return ""
}

// ordinal turns the last word of a cardinal into its ordinal form.
func ordinal(words string) string {
	i := strings.LastIndexAny(words, " -") + 1
	last := words[i:]
	switch {
	case ordinalEnding[last] != "":
		last = ordinalEnding[last]
	case strings.HasSuffix(last, "y"):
		last = strings.TrimSuffix(last, "y") + "ieth"
	default:
		last += "th"
	}
	return words[:i] + last
}

// year reads 1995 as "nineteen ninety-five" and 2005 as "two thousand and five".
func year(n int) string {
	if n >= 2000 && n < 2010 {
		return cardinal(n)
	}
	head, tail := n/100, n%100
	switch {
	case tail == 0:
		return cardinal(head) + " hundred"
	case tail < 10:
		return cardinal(head) + " oh " + cardinal(tail)
	}
	return cardinal(head) + " " + cardinal(tail)
}

func digitsByOne(s string) string {
	parts := make([]string, 0, len(s))
	for _, r := range s {
		parts = append(parts, ones[r-'0'])
	}
	return strings.Join(parts, " ")
}
