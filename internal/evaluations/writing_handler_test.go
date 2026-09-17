package evaluations

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrongWritingSkillReturnsBadRequest(t *testing.T) {
	h := &Handler{}
	response := httptest.NewRecorder()
	h.writeError(response, fmt.Errorf("validate writing: %w", ErrWrongWritingSkill), "test", nil)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "not a writing task") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
