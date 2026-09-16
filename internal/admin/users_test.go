package admin

import "testing"

func TestParseTargetExam(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
		want  string
	}{
		{value: "PTE", valid: true, want: "PTE"},
		{value: " ielts ", valid: true, want: "IELTS"},
		{value: "TOEFL", valid: false},
	} {
		exam, ok := parseTargetExam(tc.value)
		if ok != tc.valid || (tc.valid && string(exam) != tc.want) {
			t.Errorf("parseTargetExam(%q) = (%q, %t), want (%q, %t)", tc.value, exam, ok, tc.want, tc.valid)
		}
	}
}
