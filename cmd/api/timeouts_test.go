package main

import (
	"testing"
	"time"

	"github.com/prepyo/backend/pkg/config"
)

// The evaluation routes wait on a provider call, so their deadline has to
// outlast one. When it did not, a speaking submission was cancelled mid-flight
// and the learner was shown a server error for a request that was merely slow.
func TestAIRouteTimeoutOutlastsProviderCalls(t *testing.T) {
	cfg := &config.Config{AIRequestTimeout: 45 * time.Second}

	if got := AIRouteTimeout(cfg); got <= cfg.AIRequestTimeout*2 {
		t.Fatalf("AIRouteTimeout = %s, want room for more than two provider calls", got)
	}
	if AIRouteTimeout(cfg) <= requestTimeout {
		t.Fatalf("AIRouteTimeout = %s, want longer than the %s every other route gets",
			AIRouteTimeout(cfg), requestTimeout)
	}
}
