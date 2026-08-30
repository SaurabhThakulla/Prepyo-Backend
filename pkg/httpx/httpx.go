// Package httpx holds the request/response conventions every handler follows.
//
// Success:  {"success": true,  ...payload}
// Failure:  {"success": false, "error": {"code": "not_found", "message": "..."}}
//
// The code is a stable string the frontend can branch on. The message is for
// humans and may change wording at any time. Provider or database errors are
// never passed through to the client.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

// Error codes. Add to this list rather than inventing codes at a call site.
const (
	CodeBadRequest     = "bad_request"
	CodeValidation     = "validation_failed"
	CodeUnauthorized   = "unauthorized"
	CodeForbidden      = "forbidden"
	CodeNotFound       = "not_found"
	CodeConflict       = "conflict"
	CodeLimitReached   = "limit_reached"
	CodeAIUnavailable  = "ai_unavailable"
	CodeInternal       = "internal_error"
	CodeTooManyRequest = "too_many_requests"
	// The two below belong to issue reporting, which is the one feature that
	// can be switched off by configuration and the one that talks to a mail
	// server. Both need wording a reporter can act on, so neither collapses
	// into CodeInternal.
	CodeNotConfigured = "not_configured"
	CodeSendFailed    = "send_failed"
)

// maxBodyBytes caps request bodies. Everything an ordinary endpoint accepts is
// text a person typed, which never comes near this.
const maxBodyBytes = 1 << 20 // 1 MiB

// MaxAudioBodyBytes is the cap for the one kind of body that is not typed text:
// a base64 recording. Two minutes of 16 kHz mono PCM is about 3.8 MB, and
// base64 adds a third, so this leaves room for the longest task plus the JSON
// around it without letting an upload run away.
const MaxAudioBodyBytes = 12 << 20 // 12 MiB

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// JSON writes a success response. Payload keys are merged with "success": true.
func JSON(w http.ResponseWriter, status int, payload map[string]any) {
	body := make(map[string]any, len(payload)+1)
	body["success"] = true
	for k, v := range payload {
		body[k] = v
	}
	write(w, status, body)
}

// Error writes a failure response.
func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, map[string]any{
		"success": false,
		"error":   errorBody{Code: code, Message: message},
	})
}

// ValidationError reports which fields were rejected and why.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	write(w, http.StatusUnprocessableEntity, map[string]any{
		"success": false,
		"error": errorBody{
			Code:    CodeValidation,
			Message: "Some fields need attention.",
			Fields:  fields,
		},
	})
}

// Internal logs the real cause and returns a generic message. Use it for
// anything the caller cannot act on: database failures, encoding bugs, and so
// on. The client never sees the underlying error text.
func Internal(w http.ResponseWriter, log *slog.Logger, op string, err error) {
	log.Error("request failed", "op", op, "error", err)
	Error(w, http.StatusInternalServerError, CodeInternal, "Something went wrong on our side. Please try again.")
}

// RateLimited is the response for a throttled request.
//
// The limiter's own default writes plain text, which reaches the browser as an
// unparseable body and shows a learner "the server returned an unreadable
// response". Every failure the client can see should arrive in the same
// envelope as every other.
func RateLimited(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusTooManyRequests, CodeTooManyRequest,
		"Too many attempts in a short time. Please wait a minute and try again.")
}

// Decode reads a JSON body into dst. It rejects unknown fields so a typo in a
// client payload fails loudly in the log instead of being silently ignored.
//
// The client is told only that the request could not be read. A malformed body
// is a bug in the caller, never something a learner can act on, and the
// specifics — which field was unexpected, which one had the wrong type — are a
// description of the API's own shape. Handing those back turns this endpoint
// into a way to enumerate the request schema, so they go to the log instead,
// where a developer can read them and a stranger cannot.
func Decode(w http.ResponseWriter, r *http.Request, dst any, log *slog.Logger, op string) bool {
	return DecodeLimit(w, r, dst, log, op, maxBodyBytes)
}

// DecodeLimit is Decode with the body cap named at the call site, for the
// endpoints that carry a recording rather than typed text.
func DecodeLimit(w http.ResponseWriter, r *http.Request, dst any, log *slog.Logger, op string, maxBytes int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		rejectBody(w, log, op, err)
		return false
	}
	// A second value in the body means the client sent something we did not
	// expect; treat it as malformed rather than reading only the first.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		rejectBody(w, log, op, errors.New("body contained more than one JSON value"))
		return false
	}
	return true
}

func rejectBody(w http.ResponseWriter, log *slog.Logger, op string, err error) {
	// Size is the one thing the sender can act on, and saying so reveals
	// nothing about the request shape. A learner who has written a very long
	// essay needs to know that is why it bounced.
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		Error(w, http.StatusBadRequest, CodeBadRequest,
			"That is longer than we can accept in one go. Please shorten it and try again.")
		return
	}

	log.Warn("rejected request body", "op", op, "error", err)
	Error(w, http.StatusBadRequest, CodeBadRequest,
		"We could not read that request. Please refresh the page and try again.")
}

func write(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// The status line is already sent, so a failure here can only be logged by
	// the caller's middleware. Encoding a map of plain values does not fail in
	// practice.
	_ = json.NewEncoder(w).Encode(body)
}
