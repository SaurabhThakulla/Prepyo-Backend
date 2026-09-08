// Package httpx provides standard HTTP request and response helpers.
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
	CodeNotConfigured = "not_configured"
	CodeSendFailed    = "send_failed"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// MaxAudioBodyBytes is the maximum allowed request body size for audio uploads.
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

// Internal logs the internal error and returns a 500 response.
func Internal(w http.ResponseWriter, log *slog.Logger, op string, err error) {
	log.Error("request failed", "op", op, "error", err)
	Error(w, http.StatusInternalServerError, CodeInternal, "Something went wrong on our side. Please try again.")
}

// RateLimited writes a 429 response for rate-limited requests.
func RateLimited(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusTooManyRequests, CodeTooManyRequest,
		"Too many attempts in a short time. Please wait a minute and try again.")
}

// Decode decodes a JSON request body into dst with standard size limits and unknown field rejection.
func Decode(w http.ResponseWriter, r *http.Request, dst any, log *slog.Logger, op string) bool {
	return DecodeLimit(w, r, dst, log, op, maxBodyBytes)
}

// DecodeLimit decodes a JSON request body into dst with a custom byte limit.
func DecodeLimit(w http.ResponseWriter, r *http.Request, dst any, log *slog.Logger, op string, maxBytes int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		rejectBody(w, log, op, err)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		rejectBody(w, log, op, errors.New("body contained more than one JSON value"))
		return false
	}
	return true
}

func rejectBody(w http.ResponseWriter, log *slog.Logger, op string, err error) {
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
