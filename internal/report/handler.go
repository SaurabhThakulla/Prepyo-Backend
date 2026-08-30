// Package report takes an issue report from the in-app dialog and emails it.
//
// This used to be a Next.js route handler in the frontend repo, back when that
// repo shipped a server. The frontend is a static build now, so the endpoint
// moved here — which is where it always belonged: it is the only part of the
// product that sends mail, and mail is not something a browser bundle can do.
//
// Two things got simpler in the move. The Next version held an opaque session
// cookie and had to call /profile to learn who was reporting; here the auth
// middleware has already resolved the user. And rate limiting was a map of IPs
// swept by hand; here it is the same chi middleware every other route uses.
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
	// Long enough for a real report, short enough that nobody can post a book.
	maxMessage = 4000
	// The path is context, not content. It only ever holds an app route.
	maxPath = 200

	smtpHost = "smtp.gmail.com"
	// Implicit TLS. Port 587 would work too, but it starts in the clear and
	// upgrades, and there is nothing to gain from the extra round trip.
	smtpAddr = smtpHost + ":465"

	// A report is a background courtesy, not something the learner waits on
	// happily. If Gmail has not answered by now it is not going to.
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
		// Deliberately not a 200. See Config.ReportingEnabled.
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

// send composes the mail and hands it to Gmail.
//
// The reporter's name and address end up in headers, so both are stripped of
// CR and LF first. Without that, a name containing a newline could append
// headers of its own and turn this endpoint into an open relay for whoever
// picked the name. The subject is Q-encoded because plenty of learners have
// names that are not ASCII.
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
		// Replying in the mail client goes straight back to the learner.
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

// deliver opens the connection itself rather than calling smtp.SendMail, which
// dials without a timeout: one unresponsive mail server would otherwise hang a
// goroutine for as long as the process lives.
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

// sanitizeHeader removes what would let a value break out of its header line.
func sanitizeHeader(v string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(v))
}
