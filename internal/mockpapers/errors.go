package mockpapers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/prepyo/backend/pkg/httpx"
)

// CheckOpenPaper is the rule every engine applies before resuming an open
// session: asking for a specific test while another test of the section is
// still open is refused rather than silently resuming the other one. With no
// test asked for, or the open one asked for, it resumes.
func CheckOpenPaper(requested string, open *string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nil
	}
	if open != nil && *open == requested {
		return nil
	}
	return ErrOtherPaperOpen
}

// CheckModule is the IELTS module rule for Reading and Writing papers: a
// paper for one module is not dealt to a learner of the other.
func CheckModule(paperModule, learnerModule string) error {
	paperModule = strings.ToLower(strings.TrimSpace(paperModule))
	if paperModule == "" || paperModule == ModuleAny || paperModule == strings.ToLower(learnerModule) {
		return nil
	}
	return ErrWrongModule
}

// WriteError writes the response for a mock paper error and reports whether
// err was one. Engine handlers call it first in their own error mapping.
func WriteError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, ErrPaperNotFound), errors.Is(err, ErrPaperNotPublished):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That test does not exist.")
	case errors.Is(err, ErrPaperRetired):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "That test has been retired. Choose another test.")
	case errors.Is(err, ErrOtherPaperOpen):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"You have another test in this section open. Finish it or leave it before starting a different one.")
	case errors.Is(err, ErrWrongModule):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "That test is for the other IELTS module.")
	default:
		return false
	}
	return true
}
