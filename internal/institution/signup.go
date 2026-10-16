package institution

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// SignupForm is what a new institution fills in.
type SignupForm struct {
	InstitutionName string `json:"institutionName"`
	Name            string `json:"name"` // the admin's own name
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
}

// bcrypt reads at most 72 bytes, so a longer password would be cut short.
const (
	minPasswordLength = 10
	maxPasswordLength = 72
)

// Clean trims the form the way it is stored. The password is left as typed.
func (f *SignupForm) Clean() {
	f.InstitutionName = strings.Join(strings.Fields(f.InstitutionName), " ")
	f.Name = strings.Join(strings.Fields(f.Name), " ")
	f.Email = strings.ToLower(strings.TrimSpace(f.Email))
	f.Phone = strings.TrimSpace(f.Phone)
}

// Problems returns a message per rejected field, keyed by its JSON name.
func (f SignupForm) Problems() map[string]string {
	problems := map[string]string{}

	if n := utf8.RuneCountInString(f.InstitutionName); n < 3 || n > 120 {
		problems["institutionName"] = "Enter your institution's name (3 to 120 characters)."
	}
	if n := utf8.RuneCountInString(f.Name); n < 2 || n > 80 {
		problems["name"] = "Enter your own name."
	}
	if !validEmail(f.Email) {
		problems["email"] = "Enter a valid email address."
	}
	if !validPhone(f.Phone) {
		problems["phone"] = "Enter a phone number we can reach you on."
	}
	switch {
	case len(f.Password) < minPasswordLength:
		problems["password"] = fmt.Sprintf("Use at least %d characters.", minPasswordLength)
	case len(f.Password) > maxPasswordLength:
		problems["password"] = fmt.Sprintf("Use at most %d characters.", maxPasswordLength)
	}
	return problems
}

func validEmail(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return false
	}
	domain := email[strings.LastIndex(email, "@")+1:]
	return strings.Contains(domain, ".")
}

// validPhone accepts digits with spaces, dashes, brackets or a leading plus.
func validPhone(phone string) bool {
	digits := 0
	for _, r := range phone {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case strings.ContainsRune(" -+()", r):
		default:
			return false
		}
	}
	return digits >= 7 && digits <= 15 && len(phone) <= 24
}

const (
	maxCodeLength = 30
	fallbackCode  = "institute" // must not be a reserved word itself
)

// reservedCodes are first words the site already uses, which an institution's address (/<code>) must not take.
// Keep in step with APP_PAGES in the web app's src/institution/paths.ts.
var reservedCodes = map[string]bool{
	"login": true, "signup": true, "home": true, "dashboard": true, "practice": true, "me": true, "mocks": true,
	"mistakes": true, "evaluations": true, "leaderboard": true, "profile": true, "referrals": true,
	"subscription": true, "diagnostic": true, "tutorials": true, "videos": true, "learn": true, "reports": true,
	"admin": true, "terms": true, "privacy": true, "terms-and-conditions": true, "privacy-policy": true,
	"account": true, "progress": true,
	// Not pages, but taken all the same.
	"institution": true, "api": true, "health": true, "assets": true, "images": true,
}

// CodeFromName turns "Bright Future Pvt. Ltd." into "bright-future-pvt-ltd".
func CodeFromName(name string) string {
	var code strings.Builder
	needDash := false
	for _, r := range strings.ToLower(name) {
		usable := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if !usable {
			needDash = true
			continue
		}
		if needDash && code.Len() > 0 {
			code.WriteByte('-')
		}
		needDash = false
		code.WriteRune(r)
	}

	result := code.String()
	if len(result) > maxCodeLength {
		result = result[:maxCodeLength]
		// Cut at the end of a word when there is one.
		if lastDash := strings.LastIndex(result, "-"); lastDash >= 12 {
			result = result[:lastDash]
		}
		result = strings.TrimRight(result, "-")
	}

	// A name with nothing usable, such as one in Devanagari, gets the fallback.
	if len(result) < 3 {
		return fallbackCode
	}
	if reservedCodes[result] {
		return result + "-" + fallbackCode
	}
	return result
}
