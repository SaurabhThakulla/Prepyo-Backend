package httpx

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Page holds validated pagination values.
type Page struct {
	Limit  int
	Offset int
}

// ReadPage parses ?limit and ?offset. Out-of-range or unparseable values fall
// back to the defaults rather than erroring, and limit is capped so one request
// cannot ask for the whole table.
func ReadPage(r *http.Request) Page {
	q := r.URL.Query()

	limit := defaultLimit
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		limit = min(v, maxLimit)
	}

	offset := 0
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v > 0 {
		offset = v
	}

	return Page{Limit: limit, Offset: offset}
}

// Meta is the pagination block included with every collection response.
func (p Page) Meta(total int) map[string]any {
	return map[string]any{
		"limit":   p.Limit,
		"offset":  p.Offset,
		"total":   total,
		"hasMore": p.Offset+p.Limit < total,
	}
}
