package ai

import "testing"

// A base URL that is missing or doubling its version segment is the classic way
// a provider swap fails, and it fails at runtime rather than at compile time.
func TestProviderCompletionsURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"codecraft", "https://codecraftapi.com/v1", "https://codecraftapi.com/v1/chat/completions"},
		{"openrouter", "https://openrouter.ai/api/v1", "https://openrouter.ai/api/v1/chat/completions"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := newProvider(tc.baseURL, "key").completionsURL(); got != tc.want {
				t.Errorf("completionsURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProviderNameIsHost(t *testing.T) {
	if got := newProvider("https://codecraftapi.com/v1", "key").name; got != "codecraftapi.com" {
		t.Errorf("name = %q, want %q", got, "codecraftapi.com")
	}
}

func TestProviderConfigured(t *testing.T) {
	if newProvider("https://codecraftapi.com/v1", "").configured() {
		t.Error("configured() = true without an api key")
	}
	if newProvider("", "key").configured() {
		t.Error("configured() = true without a base url")
	}
	if !newProvider("https://codecraftapi.com/v1", "key").configured() {
		t.Error("configured() = false with both set")
	}
}
