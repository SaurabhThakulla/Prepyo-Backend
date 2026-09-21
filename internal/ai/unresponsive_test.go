package ai

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prepyo/backend/pkg/config"
)

// A provider that accepts the connection and then never answers is down for
// every model on it. Trying the whole candidate list costs one client timeout
// each and ends in the same failure, so the learner waits minutes for it.
func TestUnresponsiveProviderIsTriedOnce(t *testing.T) {
	var calls int32
	// Sleeps past the client's timeout rather than blocking forever, so the
	// handler finishes on its own and the test server can shut down.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(900 * time.Millisecond)
	}))
	defer server.Close()

	cfg := &config.Config{
		AIBaseURL:        server.URL,
		AIAPIKey:         "key",
		AIModels:         config.AIModels{Writing: "model-a"},
		AIRequestTimeout: 300 * time.Millisecond,
		AIMaxTokens:      100,
	}
	g := NewGateway(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

	started := time.Now()
	_, _, err := g.complete(context.Background(), g.text, "model-a", "test", []chatMessage{{Role: "user", Content: "hello"}}, false)
	elapsed := time.Since(started)

	if err == nil {
		t.Fatal("complete() succeeded against a provider that never answers")
	}
	// fallbackCandidates offers five text models; before this, each one was
	// tried and the learner paid a full timeout for every entry.
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("provider was called %d times, want 1", got)
	}
	if elapsed > time.Second {
		t.Fatalf("took %s to give up on an unresponsive provider", elapsed)
	}
}
