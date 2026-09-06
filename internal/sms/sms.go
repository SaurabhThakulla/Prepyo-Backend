// Package sms delivers short messages to Nepali mobile numbers.
//
// The Sender interface exists so the auth flow can be exercised without an
// account at a provider: LogSender writes the message to the log instead of
// sending it, which is what local development and the tests use.
package sms

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotConfigured is returned by a sender with no credentials. Callers must
// surface it rather than pretending a code was sent.
var ErrNotConfigured = errors.New("sms provider is not configured")

type Sender interface {
	Send(ctx context.Context, toE164, text string) error
}

// LogSender writes messages to the log. Anything it "sends" reaches nobody, so
// it must never be selected when the environment is production.
type LogSender struct {
	Log *slog.Logger
}

func (s LogSender) Send(_ context.Context, toE164, text string) error {
	s.Log.Warn("sms not sent: no provider configured, logging instead",
		"to", toE164, "text", text)
	return nil
}

// Sparrow talks to Sparrow SMS, https://api.sparrowsms.com/v2/sms/.
//
// The API takes the national ten-digit number rather than E.164, so Send
// strips the +977 prefix before posting.
type Sparrow struct {
	Token    string
	From     string
	Endpoint string
	Client   *http.Client
}

func NewSparrow(token, from string) *Sparrow {
	return &Sparrow{
		Token:    token,
		From:     from,
		Endpoint: "https://api.sparrowsms.com/v2/sms/",
		Client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Sparrow) Send(ctx context.Context, toE164, text string) error {
	if s.Token == "" || s.From == "" {
		return ErrNotConfigured
	}

	form := url.Values{
		"token": {s.Token},
		"from":  {s.From},
		"to":    {nationalNumber(toE164)},
		"text":  {text},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Endpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build sparrow request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send via sparrow: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sparrow returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func nationalNumber(e164 string) string {
	return strings.TrimPrefix(strings.TrimPrefix(e164, "+977"), "977")
}
