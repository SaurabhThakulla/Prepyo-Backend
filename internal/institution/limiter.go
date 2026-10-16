package institution

import (
	"sync"
	"time"
)

const (
	lockWindow            = 10 * time.Minute
	maxFailuresPerNetwork = 5     // wrong passwords for one email from one network
	maxFailuresPerEmail   = 30    // wrong passwords for one email from anywhere
	maxTrackedKeys        = 10000 // cap on memory used by made-up emails
)

// failureLimiter counts recent wrong passwords per key and locks a key that has too many.
type failureLimiter struct {
	limit int

	mu       sync.Mutex
	failures map[string][]time.Time
}

func newFailureLimiter(limit int) *failureLimiter {
	return &failureLimiter{limit: limit, failures: map[string][]time.Time{}}
}

// locked reports whether the key has used up its attempts.
func (l *failureLimiter) locked(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recent(key, now)) >= l.limit
}

// record notes one more wrong password for the key.
func (l *failureLimiter) record(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, known := l.failures[key]; !known && len(l.failures) >= maxTrackedKeys {
		l.forgetOldest(now)
	}
	l.failures[key] = append(l.recent(key, now), now)
}

// clear forgets the key's failures after a correct sign-in.
func (l *failureLimiter) clear(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

// recent returns the key's failures still inside the window. Caller holds mu.
func (l *failureLimiter) recent(key string, now time.Time) []time.Time {
	cutoff := now.Add(-lockWindow)
	kept := l.failures[key][:0]
	for _, at := range l.failures[key] {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	if len(kept) == 0 {
		delete(l.failures, key)
		return nil
	}
	l.failures[key] = kept
	return kept
}

// forgetOldest drops expired keys, then the stalest one, to make room. Caller holds mu.
func (l *failureLimiter) forgetOldest(now time.Time) {
	var oldestKey string
	var oldestAt time.Time
	for key := range l.failures {
		kept := l.recent(key, now)
		if len(kept) == 0 {
			continue
		}
		if last := kept[len(kept)-1]; oldestKey == "" || last.Before(oldestAt) {
			oldestKey, oldestAt = key, last
		}
	}
	if len(l.failures) >= maxTrackedKeys {
		delete(l.failures, oldestKey)
	}
}
