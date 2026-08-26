package referrals

import (
	"strings"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode failed: %v", err)
		}

		if !strings.HasPrefix(code, "PREP-") {
			t.Errorf("Code %q missing PREP- prefix", code)
		}

		if len(code) != 11 {
			t.Errorf("Expected code length 11, got %d for %q", len(code), code)
		}

		if !IsValidFormat(code) {
			t.Errorf("Code %q failed IsValidFormat", code)
		}

		if seen[code] {
			t.Fatalf("Duplicate code generated in 1000 iterations: %q", code)
		}
		seen[code] = true
	}
}

func TestNormalizeCode(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"prep-x7k4m9", "PREP-X7K4M9"},
		{"  PREP-ABC234  ", "PREP-ABC234"},
		{"prep-234567", "PREP-234567"},
	}

	for _, c := range cases {
		got := NormalizeCode(c.input)
		if got != c.want {
			t.Errorf("NormalizeCode(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}

func TestIsValidFormat(t *testing.T) {
	cases := []struct {
		code  string
		valid bool
	}{
		{"PREP-X7K4M9", true},
		{"PREP-234567", true},
		{"prep-x7k4m9", true}, // normalized
		{"PREP-123456", false}, // contains '1', which is excluded from charset
		{"PREP-OOOOOO", false}, // contains 'O', which is excluded
		{"X7K4M9", false},      // missing prefix
		{"PREP-TOOLONG123", false},
		{"", false},
	}

	for _, c := range cases {
		got := IsValidFormat(c.code)
		if got != c.valid {
			t.Errorf("IsValidFormat(%q) = %v; want %v", c.code, got, c.valid)
		}
	}
}

func TestMaskName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Aprapya Rana", "A*** R***"},
		{"John Doe", "J*** D***"},
		{"Student", "S***"},
		{"A B C", "A* B* C*"},
		{"", "Learner"},
	}

	for _, c := range cases {
		got := maskName(c.input)
		if got != c.want {
			t.Errorf("maskName(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}
