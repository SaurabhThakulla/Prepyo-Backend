package referrals

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

// referralCharset excludes ambiguous characters like 0, O, 1, I, L.
const referralCharset = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
const codeLength = 6
const codePrefix = "PREP-"

var codeRegex = regexp.MustCompile(`^PREP-[2-9A-HJ-NP-Z]{6}$`)

// GenerateCode produces a cryptographically secure, human-friendly referral code.
// Example: "PREP-X7K4M9"
func GenerateCode() (string, error) {
	var sb strings.Builder
	sb.WriteString(codePrefix)

	max := big.NewInt(int64(len(referralCharset)))
	for i := 0; i < codeLength; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		sb.WriteByte(referralCharset[idx.Int64()])
	}
	return sb.String(), nil
}

// NormalizeCode standardizes case and whitespace.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// IsValidFormat tests if the referral code matches the standard structure.
func IsValidFormat(code string) bool {
	return codeRegex.MatchString(NormalizeCode(code))
}
