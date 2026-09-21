package evaluations

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A request whose deadline runs out while the provider is answering is not a
// fault on our side, and telling the learner it was sent them to support for
// something a retry fixes.
func TestDeadlineExceededReturnsUnavailable(t *testing.T) {
	for _, err := range []error{context.DeadlineExceeded, context.Canceled} {
		h := &Handler{}
		response := httptest.NewRecorder()
		h.writeError(response, fmt.Errorf("save evaluation: %w", err), "test", nil)

		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%v: status=%d body=%s", err, response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), "ai_unavailable") {
			t.Fatalf("%v: body=%s, want the ai_unavailable code", err, response.Body.String())
		}
	}
}
