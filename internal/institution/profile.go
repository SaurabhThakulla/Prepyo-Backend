package institution

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/pkg/httpx"
)

// Profile is what an institution answers in its setup after sign-up.
type Profile struct {
	Exams           []string `json:"exams"`
	City            string   `json:"city"`
	Branches        int      `json:"branches"`
	MonthlyStudents string   `json:"monthlyStudents"`
	BrandColor      string   `json:"brandColor"`
}

var (
	profileExams    = []string{"IELTS", "PTE", "EPS_TOPIK"}
	studentRanges   = map[string]bool{"under-25": true, "25-50": true, "50-100": true, "100-plus": true}
	brandColorShape = regexp.MustCompile(`^#[0-9a-f]{6}$`)
)

const maxBranches = 50

// Clean trims the answers the way they are stored, keeping exams in a fixed order.
func (p *Profile) Clean() {
	chosen := map[string]bool{}
	for _, exam := range p.Exams {
		chosen[strings.ToUpper(strings.TrimSpace(exam))] = true
	}
	p.Exams = []string{}
	for _, exam := range profileExams {
		if chosen[exam] {
			p.Exams = append(p.Exams, exam)
		}
	}
	p.City = strings.Join(strings.Fields(p.City), " ")
	p.BrandColor = strings.ToLower(strings.TrimSpace(p.BrandColor))
}

// Problems returns a message per rejected field, keyed by its JSON name.
func (p Profile) Problems() map[string]string {
	problems := map[string]string{}
	if len(p.Exams) == 0 {
		problems["exams"] = "Choose at least one exam you teach."
	}
	if n := utf8.RuneCountInString(p.City); n < 2 || n > 60 {
		problems["city"] = "Enter the city your institution is in."
	}
	if p.Branches < 1 || p.Branches > maxBranches {
		problems["branches"] = fmt.Sprintf("Enter between 1 and %d branches.", maxBranches)
	}
	if !studentRanges[p.MonthlyStudents] {
		problems["monthlyStudents"] = "Choose roughly how many students you prepare each month."
	}
	if !brandColorShape.MatchString(p.BrandColor) {
		problems["brandColor"] = "Choose a brand colour."
	}
	return problems
}

// saveProfile stores the setup answers and unlocks the dashboard.
func (h *Handler) saveProfile(w http.ResponseWriter, r *http.Request) {
	var profile Profile
	if !httpx.Decode(w, r, &profile, h.log, "institution.saveProfile") {
		return
	}
	profile.Clean()
	if problems := profile.Problems(); len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	account := signedInAccount(r)
	onboardedAt, err := h.store.saveProfile(r.Context(), account.Institution.ID, profile)
	if err != nil {
		httpx.Internal(w, h.log, "institution.saveProfile", err)
		return
	}
	account.Institution.Profile = profile
	account.Institution.OnboardedAt = &onboardedAt
	httpx.JSON(w, http.StatusOK, accountBody(account))
}

// A logo is shown small; anything bigger is a print file nobody needs here.
const maxLogoBytes = 2 << 20

// The bytes decide the type, never the request header. SVG is left out: it can carry script.
var logoTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}

// putLogo stores a logo sent as the raw request body.
func (h *Handler) putLogo(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxLogoBytes))
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		httpx.Error(w, http.StatusRequestEntityTooLarge, httpx.CodeBadRequest, "That logo is larger than 2 MB. Please choose a smaller one.")
		return
	case err != nil:
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "We could not read that upload.")
		return
	case len(data) == 0:
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "That file is empty.")
		return
	}

	contentType := http.DetectContentType(data)
	if !logoTypes[contentType] {
		httpx.Error(w, http.StatusUnsupportedMediaType, httpx.CodeBadRequest, "Please upload a PNG, JPEG or WebP image.")
		return
	}

	account := signedInAccount(r)
	updatedAt, err := h.store.setLogo(r.Context(), account.Institution.ID, contentType, data)
	if err != nil {
		httpx.Internal(w, h.log, "institution.putLogo", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"logoUrl": logoURL(account.Institution.Code, updatedAt)})
}

// getLogo serves an institution's logo to anyone: it is shown on its public sign-in page.
func (h *Handler) getLogo(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(chi.URLParam(r, "code"))
	contentType, data, err := h.store.logo(r.Context(), code)
	if errors.Is(err, errNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No logo set.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.getLogo", err)
		return
	}

	// The address carries ?v=<updated time>, so a new logo is a new address.
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := w.Write(data); err != nil {
		h.log.Warn("institution.getLogo write", "error", err)
	}
}

func logoURL(code string, updatedAt time.Time) string {
	return fmt.Sprintf("/api/v1/institution/logo/%s?v=%d", code, updatedAt.Unix())
}
