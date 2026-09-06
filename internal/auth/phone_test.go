package auth

import "testing"

func TestNormalisePhoneCollapsesFormatting(t *testing.T) {
	want := "+9779801234567"
	for _, raw := range []string{
		"9801234567",
		"980 123 4567",
		"980-123-4567",
		"+9779801234567",
		"9779801234567",
		"009779801234567",
		"  +977 980-123 4567 ",
	} {
		got, err := NormalisePhone(raw)
		if err != nil {
			t.Errorf("NormalisePhone(%q) returned %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("NormalisePhone(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestNormalisePhoneRejectsNonMobile(t *testing.T) {
	for _, raw := range []string{
		"",
		"12345",
		"9601234567890",
		"014567890",
		"9501234567",
		"98012345",
		"not a number",
	} {
		if got, err := NormalisePhone(raw); err == nil {
			t.Errorf("NormalisePhone(%q) accepted it as %q", raw, got)
		}
	}
}
