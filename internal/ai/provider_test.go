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
		{"openrouter", "https://openrouter.ai/api/v1", "https://openrouter.ai/api/v1/chat/completions"},
		{"custom", "https://api.example.com/v1", "https://api.example.com/v1/chat/completions"},
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
	if got := newProvider("https://openrouter.ai/api/v1", "key").name; got != "openrouter.ai" {
		t.Errorf("name = %q, want %q", got, "openrouter.ai")
	}
}

func TestProviderConfigured(t *testing.T) {
	if newProvider("https://openrouter.ai/api/v1", "").configured() {
		t.Error("configured() = true without an api key")
	}
	if newProvider("", "key").configured() {
		t.Error("configured() = true without a base url")
	}
	if !newProvider("https://openrouter.ai/api/v1", "key").configured() {
		t.Error("configured() = false with both set")
	}
}

func TestFallbackCandidates(t *testing.T) {
	candidates := fallbackCandidates("gemini-3.8-flash-high", false)
	for _, c := range candidates {
		if c == "gemini-3.8-flash-high" {
			t.Errorf("expected gemini-3.8-flash-high to be excluded from candidates, got %v", candidates)
		}
	}
	if len(candidates) == 0 {
		t.Fatal("expected fallback candidates, got empty")
	}
	if candidates[0] != "anthropic/claude-haiku-5.5" {
		t.Errorf("expected first fallback to be anthropic/claude-haiku-5.5, got %q", candidates[0])
	}

	if got := fallbackCandidates("anthropic/claude-haiku-5.5", false); len(got) != 1 {
		t.Errorf("expected Haiku alone, got %v", got)
	}
}
