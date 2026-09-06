package auth

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidPhone is returned for anything that is not a Nepali mobile number.
var ErrInvalidPhone = errors.New("invalid phone number")

// nepaliMobile is the national part: ten digits opening 96, 97 or 98, which
// covers NTC, Ncell and Smart Cell. Landlines cannot receive an SMS, so they
// are rejected rather than accepted and silently failed later.
var nepaliMobile = regexp.MustCompile(`^9[678]\d{8}$`)

var notDigits = regexp.MustCompile(`\D`)

// NormalisePhone turns anything a learner might type into E.164.
//
// 9801234567, 977-980-123-4567, +977 9801234567 and 00977 9801234567 all
// become +9779801234567, so one person cannot hold several accounts by varying
// the punctuation.
func NormalisePhone(raw string) (string, error) {
	digits := notDigits.ReplaceAllString(strings.TrimSpace(raw), "")

	digits = strings.TrimPrefix(digits, "00")
	if len(digits) > 10 {
		digits = strings.TrimPrefix(digits, "977")
	}

	if !nepaliMobile.MatchString(digits) {
		return "", ErrInvalidPhone
	}
	return "+977" + digits, nil
}
