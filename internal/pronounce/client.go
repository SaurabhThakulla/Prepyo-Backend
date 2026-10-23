// Package pronounce calls the self-hosted pronunciation scorer (prepyo-pronounce on
// AWS Lambda): a recording and the text it should say in, a 0-100 score and a
// good / average / bad label per word out. No paid AI model is involved.
package pronounce

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Word is one word of the given text as the scorer heard it.
type Word struct {
	Word  string  `json:"word"`
	Score float64 `json:"score"`
	Label string  `json:"label"` // good, average or bad
}

// Result is the scorer's reply.
type Result struct {
	Words         []Word  `json:"words"`
	Pronunciation int     `json:"pronunciation"`
	Seconds       float64 `json:"seconds"`
}

// Client is nil-safe: a nil client means the scorer is not configured.
type Client struct {
	url, key string
	http     *http.Client
	log      *slog.Logger
}

// timeout is how long a learner waits for the scorer, retries included, before the
// answer falls back to the Gemini rating. It covers a cold start (about 25-30 s).
const timeout = 35 * time.Second

// New returns nil unless both the URL and the key are set.
func New(url, key string, log *slog.Logger) *Client {
	if strings.TrimSpace(url) == "" || strings.TrimSpace(key) == "" {
		return nil
	}
	return &Client{url: strings.TrimSpace(url), key: strings.TrimSpace(key), http: &http.Client{Timeout: timeout}, log: log}
}

// Score sends one recording (WAV or MP3, 16 kHz mono) with the text it should say.
// When every scorer copy is busy (429) or one is briefly unavailable, it tries
// again after 1, 2, 3... seconds until the timeout: a short wait, not a queue to manage.
func (c *Client) Score(ctx context.Context, audio []byte, text string) (Result, error) {
	if c == nil {
		return Result{}, fmt.Errorf("pronunciation scorer not configured")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for attempt := 1; ; attempt++ {
		res, retry, err := c.score(ctx, audio, text)
		wait := time.Duration(attempt) * time.Second
		if err == nil || !retry {
			return res, err
		}
		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < wait+3*time.Second {
			return res, err
		}
		select {
		case <-ctx.Done():
			return res, err
		case <-time.After(wait):
		}
	}
}

// score makes one attempt; retry reports whether trying again may succeed.
func (c *Client) score(ctx context.Context, audio []byte, text string) (Result, bool, error) {
	body, _ := json.Marshal(map[string]string{"audio": base64.StdEncoding.EncodeToString(audio), "text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return Result{}, false, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-prepyo-key", c.key)
	res, err := c.http.Do(req)
	if err != nil {
		return Result{}, ctx.Err() == nil, err // a dropped connection is worth one more try
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		busy := res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500
		return Result{}, busy, fmt.Errorf("pronunciation scorer returned %d: %.200s", res.StatusCode, raw)
	}
	var out Result
	if err := json.Unmarshal(raw, &out); err != nil {
		return Result{}, false, fmt.Errorf("pronunciation scorer reply: %w", err)
	}
	if len(out.Words) == 0 {
		return Result{}, false, fmt.Errorf("pronunciation scorer heard too little speech")
	}
	return out, false, nil
}

// KeepWarm pings the scorer every few minutes so Lambda keeps a copy loaded and
// learners do not wait for a cold start. The ping carries no key: it is answered
// with 401 in milliseconds and costs almost nothing.
func (c *Client) KeepWarm(ctx context.Context, every time.Duration) {
	if c == nil {
		return
	}
	ping := func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, strings.NewReader("{}"))
		if err != nil {
			return
		}
		if res, err := (&http.Client{Timeout: 60 * time.Second}).Do(req); err == nil {
			res.Body.Close()
		} else if ctx.Err() == nil {
			c.log.Warn("pronunciation scorer keep-warm ping failed", "error", err)
		}
	}
	go func() {
		ping()
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				ping()
			}
		}
	}()
}
