package billing

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookFailsClosedWithoutProviderVerification(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(nil, nil, nil, log)
	router := h.Routes(func(next http.Handler) http.Handler { return next })

	body := `{"userId":"attacker","planId":"elite","status":"success","amountNPR":1}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"applied":true`) {
		t.Fatalf("disabled webhook reported an entitlement mutation: %s", rec.Body.String())
	}
}
