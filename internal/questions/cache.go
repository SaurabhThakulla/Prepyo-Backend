package questions

import (
	"slices"
	"sync"
	"time"

	"github.com/prepyo/backend/internal/models"
)

// cacheTTL bounds how stale a question can be after an edit made outside the
// admin API (a migration, a voicegen audio push).
const cacheTTL = 5 * time.Minute

// cacheMaxEntries caps memory; the whole bank fits well under it.
const cacheMaxEntries = 10000

type cachedQuestion struct {
	q         models.Question
	published bool
	expires   time.Time
}

// cache keeps questions by id in memory. Questions are read on every
// practice submission and mock view but change only when an admin edits them.
type cache struct {
	mu    sync.RWMutex
	items map[string]cachedQuestion
}

func newCache() *cache {
	return &cache{items: map[string]cachedQuestion{}}
}

func (c *cache) get(id string) (models.Question, bool, bool) {
	c.mu.RLock()
	e, ok := c.items[id]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		return models.Question{}, false, false
	}
	return cloneQuestion(e.q), e.published, true
}

func (c *cache) put(q models.Question, published bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= cacheMaxEntries {
		c.items = map[string]cachedQuestion{}
	}
	c.items[q.ID] = cachedQuestion{q: cloneQuestion(q), published: published, expires: time.Now().Add(cacheTTL)}
}

func (c *cache) clear() {
	c.mu.Lock()
	c.items = map[string]cachedQuestion{}
	c.mu.Unlock()
}

// cloneQuestion copies the slices too, since callers shuffle options in place.
func cloneQuestion(q models.Question) models.Question {
	q.Options = slices.Clone(q.Options)
	q.Tags = slices.Clone(q.Tags)
	q.CorrectAnswers = slices.Clone(q.CorrectAnswers)
	q.SupportedExams = slices.Clone(q.SupportedExams)
	if q.Blanks != nil {
		blanks := make([]models.Blank, len(q.Blanks))
		for i, b := range q.Blanks {
			b.Options = slices.Clone(b.Options)
			b.AcceptedAnswers = slices.Clone(b.AcceptedAnswers)
			blanks[i] = b
		}
		q.Blanks = blanks
	}
	return q
}
