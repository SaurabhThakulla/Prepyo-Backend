// Package report handles issue reporting by emailing user feedback.
package report

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

const (
	maxMessage  = 4000
	maxPath     = 200
	smtpHost    = "smtp.gmail.com"
	smtpAddr    = smtpHost + ":465"
	sendTimeout = 15 * time.Second
)

type Handler struct {
	smtpUser     string
	smtpPassword string
	to           string
	log          *slog.Logger
}

func NewHandler(smtpUser, smtpPassword, to string, log *slog.Logger) *Handler {
	return &Handler{smtpUser: smtpUser, smtpPassword: smtpPassword, to: to, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.submit)
	return r
}

type submitRequest struct {
	Message string `json:"message"`
	Path    string `json:"path"`
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	var req submitRequest
	if !httpx.Decode(w, r, &req, h.log, "report.submit") {
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"Please write a little about what went wrong.")
		return
	}
	if len(message) > maxMessage {
		message = message[:maxMessage]
	}

	path := req.Path
	if path == "" {
		path = "unknown"
	}
	if len(path) > maxPath {
		path = path[:maxPath]
	}

	if h.smtpUser == "" || h.smtpPassword == "" {
		h.log.Error("issue report dropped: GMAIL_USER / GMAIL_APP_PASSWORD are not set")
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeNotConfigured,
			"Reporting is not switched on in this environment yet.")
		return
	}

	if err := h.send(user.Name, user.Email, path, message); err != nil {
		h.log.Error("failed to send issue report", "op", "report.submit", "error", err)
		httpx.Error(w, http.StatusBadGateway, httpx.CodeSendFailed,
			"We could not send that just now. Please try again in a moment.")
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

// send composes and delivers the issue report email.
func (h *Handler) send(name, email, path, message string) error {
	name = sanitizeHeader(name)
	email = sanitizeHeader(email)
	if name == "" {
		name = "Learner"
	}

	subject := mime.QEncoding.Encode("utf-8", "Prepyo issue report from "+name)

	headers := []string{
		fmt.Sprintf("From: Prepyo Reports <%s>", h.smtpUser),
		fmt.Sprintf("To: %s", h.to),
		fmt.Sprintf("Reply-To: %s", email),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
	}

	body := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", name, email),
		"Page: " + path,
		"Time: " + time.Now().UTC().Format(time.RFC3339),
		"",
		message,
	}, "\r\n")

	// SMTP wants CRLF line endings, and a blank line between headers and body.
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n")

	return h.deliver([]byte(msg))
}

// deliver establishes a TLS SMTP connection with timeout and transmits the message.
func (h *Handler) deliver(msg []byte) error {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: sendTimeout},
		"tcp", smtpAddr,
		&tls.Config{ServerName: smtpHost},
	)
	if err != nil {
		return fmt.Errorf("dial %s: %w", smtpAddr, err)
	}
	// Bounds the rest of the exchange, not just the dial.
	_ = conn.SetDeadline(time.Now().Add(sendTimeout))

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", h.smtpUser, h.smtpPassword, smtpHost)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(h.smtpUser); err != nil {
		return fmt.Errorf("smtp from: %w", err)
	}
	if err := client.Rcpt(h.to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close body: %w", err)
	}

	return client.Quit()
}

// sanitizeHeader strips carriage returns and newlines to prevent header injection.
func sanitizeHeader(v string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(v))
}
