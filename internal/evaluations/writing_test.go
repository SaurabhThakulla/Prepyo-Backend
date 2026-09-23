package evaluations

import (
	"errors"
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func TestWritingResponseThreshold(t *testing.T) {
	for _, tc := range []struct {
		name   string
		exam   models.ExamType
		skill  models.SkillType
		typeID string
		words  int
		want   error
	}{
		{"summary below minimum", models.ExamPTE, models.SkillWriting, "summarize-written-text", 4, ErrEmptyResponse},
		{"five word summary", models.ExamPTE, models.SkillWriting, "summarize-written-text", 5, nil},
		{"nineteen word summary", models.ExamPTE, models.SkillWriting, "summarize-written-text", 19, nil},
		{"maximum summary", models.ExamPTE, models.SkillWriting, "summarize-written-text", 75, nil},
		{"overlong summary still gets feedback", models.ExamPTE, models.SkillWriting, "summarize-written-text", 76, nil},
		{"short essay", models.ExamPTE, models.SkillWriting, "pte-write-essay", 19, ErrEmptyResponse},
		{"essay feedback minimum", models.ExamPTE, models.SkillWriting, "pte-write-essay", 20, nil},
		{"IELTS figure", models.ExamIELTS, models.SkillWriting, "ielts-writing-task1-figure", 20, nil},
		{"IELTS short response is rated, not refused", models.ExamIELTS, models.SkillWriting, "ielts-writing-task2-opinion", 5, nil},
		{"IELTS empty response", models.ExamIELTS, models.SkillWriting, "ielts-writing-task2-opinion", 0, ErrEmptyResponse},
		{"reject other skill", models.ExamPTE, models.SkillReading, "summarize-written-text", 20, ErrWrongWritingSkill},
		{"empty", models.ExamPTE, models.SkillWriting, "summarize-written-text", 0, ErrEmptyResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := models.Question{Exam: tc.exam, Skill: tc.skill, TypeID: tc.typeID}
			err := validateWritingResponse(q, " \n"+strings.Repeat("word\t", tc.words))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
