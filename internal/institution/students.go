package institution

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/pkg/httpx"
)

// LearnerSignIn starts an ordinary learner session, the kind the rest of the app understands.
type LearnerSignIn func(w http.ResponseWriter, r *http.Request, userID string) error

// freeUsername makes a username from the student's first name: ram, then ram2, ram3...
func freeUsername(name string, taken map[string]bool) string {
	var base strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			base.WriteRune(r)
		} else if base.Len() > 0 {
			break // stop at the end of the first word
		}
	}
	if base.Len() < 2 {
		base.Reset()
		base.WriteString("student")
	}

	username := base.String()
	for n := 2; taken[username]; n++ {
		username = fmt.Sprintf("%s%d", base.String(), n)
	}
	return username
}

// passwordLetters leaves out characters that look alike (0/o, 1/l/i).
const passwordLetters = "abcdefghjkmnpqrstuvwxyz23456789"

// newStudentPassword makes a random password the admin can read out or write down.
func newStudentPassword() (string, error) {
	password := make([]byte, 10)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordLetters))))
		if err != nil {
			return "", fmt.Errorf("make student password: %w", err)
		}
		password[i] = passwordLetters[n.Int64()]
	}
	return string(password), nil
}

// lookup tells the student sign-in page whose page it is.
func (h *Handler) lookup(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(chi.URLParam(r, "code"))
	name, color, logoUpdatedAt, err := h.store.brand(r.Context(), code)
	if errors.Is(err, errNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No institution uses this address.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.lookup", err)
		return
	}
	body := map[string]any{"name": name, "code": code, "brandColor": color}
	if logoUpdatedAt != nil {
		body["logoUrl"] = logoURL(code, *logoUpdatedAt)
	}
	httpx.JSON(w, http.StatusOK, body)
}

type studentLoginRequest struct {
	Code     string `json:"code"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) studentLogIn(w http.ResponseWriter, r *http.Request) {
	var req studentLoginRequest
	if !httpx.Decode(w, r, &req, h.log, "institution.studentLogIn") {
		return
	}

	code := strings.ToLower(strings.TrimSpace(req.Code))
	username := strings.ToLower(strings.TrimSpace(req.Username))
	now := time.Now()
	studentKey := "student|" + code + "|" + username
	networkKey := studentKey + "|" + clientIP(r)

	if h.failuresByNetwork.locked(networkKey, now) || h.failuresByEmail.locked(studentKey, now) {
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeTooManyRequest,
			"Too many wrong attempts. Please wait a few minutes and try again.")
		return
	}

	userID, passwordHash, disabled, err := h.store.studentForLogin(r.Context(), code, username)
	if err != nil && !errors.Is(err, errNotFound) {
		httpx.Internal(w, h.log, "institution.studentLogIn", err)
		return
	}

	if !passwordMatches(passwordHash, req.Password) {
		h.failuresByNetwork.record(networkKey, now)
		h.failuresByEmail.record(studentKey, now)
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized,
			"That username and password do not match a student of this institution.")
		return
	}
	h.failuresByNetwork.clear(networkKey)
	h.failuresByEmail.clear(studentKey)

	// Only someone who knows the password learns that the account is turned off.
	if disabled {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden,
			"Your institution has turned off this account. Please contact them.")
		return
	}

	if err := h.signInLearner(w, r, userID); err != nil {
		httpx.Internal(w, h.log, "institution.studentLogIn", err)
		return
	}
	httpx.JSON(w, http.StatusOK, nil)
}

func (h *Handler) listStudents(w http.ResponseWriter, r *http.Request) {
	days, since, ok := readPeriod(r)
	if !ok {
		httpx.ValidationError(w, map[string]string{"days": "Choose 7, 30, 90 or 0 (all time)."})
		return
	}
	students, err := h.store.students(r.Context(), signedInAccount(r).Institution.ID, since)
	if err != nil {
		httpx.Internal(w, h.log, "institution.listStudents", err)
		return
	}

	list := make([]map[string]any, 0, len(students))
	for _, student := range students {
		list = append(list, studentBody(student))
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"days": days, "students": list})
}

type addStudentRequest struct {
	Name string `json:"name"`
}

// addStudent creates a student and answers with the password, which is shown this once only.
func (h *Handler) addStudent(w http.ResponseWriter, r *http.Request) {
	var req addStudentRequest
	if !httpx.Decode(w, r, &req, h.log, "institution.addStudent") {
		return
	}
	name := strings.Join(strings.Fields(req.Name), " ")
	if n := utf8.RuneCountInString(name); n < 2 || n > 80 {
		httpx.ValidationError(w, map[string]string{"name": "Enter the student's name."})
		return
	}

	password, passwordHash, err := newPasswordAndHash()
	if err != nil {
		httpx.Internal(w, h.log, "institution.addStudent", err)
		return
	}

	student, err := h.store.addStudent(r.Context(), signedInAccount(r).Institution.ID, name, passwordHash)
	if errors.Is(err, errSeatsFull) {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"All your student seats are in use. Remove a student or contact Prepyo for a bigger plan.")
		return
	}
	if errors.Is(err, errNoPlan) {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"Your institution has no plan yet. Contact Prepyo to choose Basic, Extended or Ultimate.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.addStudent", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"student": studentBody(student), "password": password})
}

func (h *Handler) resetStudentPassword(w http.ResponseWriter, r *http.Request) {
	password, passwordHash, err := newPasswordAndHash()
	if err != nil {
		httpx.Internal(w, h.log, "institution.resetStudentPassword", err)
		return
	}

	err = h.store.setStudentPassword(r.Context(), signedInAccount(r).Institution.ID, chi.URLParam(r, "id"), passwordHash)
	if errors.Is(err, errNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That student is not in your institution.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.resetStudentPassword", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"password": password})
}

func (h *Handler) removeStudent(w http.ResponseWriter, r *http.Request) {
	err := h.store.removeStudent(r.Context(), signedInAccount(r).Institution.ID, chi.URLParam(r, "id"))
	if errors.Is(err, errNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That student is not in your institution.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.removeStudent", err)
		return
	}
	httpx.JSON(w, http.StatusOK, nil)
}

func newPasswordAndHash() (password, hash string, err error) {
	if password, err = newStudentPassword(); err != nil {
		return "", "", err
	}
	if hash, err = hashPassword(password); err != nil {
		return "", "", err
	}
	return password, hash, nil
}

func studentBody(student Student) map[string]any {
	return map[string]any{
		"id":        student.ID,
		"name":      student.Name,
		"username":  student.Username,
		"createdAt": student.CreatedAt.Format(time.RFC3339),
		"disabled":  student.Disabled,
		"usage":     student.Usage,
	}
}

type disableRequest struct {
	Disabled bool `json:"disabled"`
}

// setStudentDisabled turns a student's account off or back on; history and seat are kept.
func (h *Handler) setStudentDisabled(w http.ResponseWriter, r *http.Request) {
	var req disableRequest
	if !httpx.Decode(w, r, &req, h.log, "institution.setStudentDisabled") {
		return
	}
	err := h.store.setStudentDisabled(r.Context(), signedInAccount(r).Institution.ID, chi.URLParam(r, "id"), req.Disabled)
	if errors.Is(err, errNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That student is not in your institution.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.setStudentDisabled", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"disabled": req.Disabled})
}
