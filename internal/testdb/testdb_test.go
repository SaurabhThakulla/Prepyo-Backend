package testdb

import "testing"

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		url   string
		valid bool
	}{
		{"postgres://postgres@127.0.0.1:55439/prepyo_writing_test?sslmode=disable", true},
		{"postgres://postgres@localhost/prepyo_test", true},
		{"postgres://postgres@[::1]/prepyo_test", true},
		{"postgres://postgres@localhost/prepyo", false},
		{"postgres://postgres@localhost:6543/prepyo_test", false},
		{"postgres://postgres@db.example.com/prepyo_test", false},
		{"postgres://postgres@127.0.0.1,db.example.com/prepyo_test", false},
		{"not a connection string", false},
	} {
		t.Run(tc.url, func(t *testing.T) {
			if got := Validate(tc.url) == nil; got != tc.valid {
				t.Fatalf("valid=%v, want %v", got, tc.valid)
			}
		})
	}
}
