package mocks

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func bank() map[string]models.Question {
	return map[string]models.Question{
		"r1": {
			ID: "r1", Skill: models.SkillReading, TypeID: "fill-in-blanks-rw", Points: 10,
			Blanks: []models.Blank{{ID: "b1", CorrectAnswer: "debunked"}},
		},
		"r2": {
			ID: "r2", Skill: models.SkillReading, TypeID: "ielts-reading-tfng", Points: 10,
			CorrectAnswers: []string{"TRUE", "FALSE"},
		},
		"l1": {
			ID: "l1", Skill: models.SkillListening, TypeID: "write-from-dictation", Points: 10,
			CorrectAnswers: []string{"the library closes at eight"},
		},
		"w1": {
			ID: "w1", Skill: models.SkillWriting, TypeID: "ielts-writing-task2", Points: 25,
		},
	}
}

func TestGradeAllUsesSubmittedAnswers(t *testing.T) {
	got := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "r2", SelectedOptions: []string{"TRUE", "FALSE"}},
		{QuestionID: "l1", TextResponse: "the library closes at eight"},
	})

	if got.total != 3 {
		t.Errorf("total = %d, want 3", got.total)
	}
	if got.correct != 3 {
		t.Errorf("correct = %d, want 3", got.correct)
	}
	if got.accuracy() != 1 {
		t.Errorf("accuracy = %v, want 1", got.accuracy())
	}
}

func TestGradeAllWrongAnswersLowerTheScore(t *testing.T) {
	perfect := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "r2", SelectedOptions: []string{"TRUE", "FALSE"}},
	})
	wrong := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "reinforced"}},
		{QuestionID: "r2", SelectedOptions: []string{"FALSE", "TRUE"}},
	})

	if wrong.accuracy() >= perfect.accuracy() {
		t.Errorf("wrong answers scored %v, perfect scored %v; wrong must be lower",
			wrong.accuracy(), perfect.accuracy())
	}
	if wrong.correct != 0 {
		t.Errorf("correct = %d, want 0", wrong.correct)
	}
}

// A client must not be able to pad its score with answers to questions that are
// not part of this mock.
func TestGradeAllIgnoresQuestionsOutsideTheMock(t *testing.T) {
	got := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "not-in-this-mock", TextResponse: "free marks please"},
	})

	if got.total != 1 {
		t.Errorf("total = %d, want 1: the unknown question must be ignored", got.total)
	}
}

// Submitting the same question twice must count once.
func TestGradeAllIgnoresDuplicateAnswers(t *testing.T) {
	got := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
	})

	if got.total != 1 {
		t.Errorf("total = %d, want 1", got.total)
	}
}

// Writing and speaking are reported separately rather than counted as correct.
func TestGradeAllSeparatesUngradableSkills(t *testing.T) {
	got := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "w1", TextResponse: "an essay that needs a human or a model to judge"},
	})

	if got.total != 1 {
		t.Errorf("total = %d, want 1: only the reading question is auto-graded", got.total)
	}
	if got.ungraded != 1 {
		t.Errorf("ungraded = %d, want 1", got.ungraded)
	}
	if got.correct != 1 {
		t.Errorf("correct = %d, want 1: the essay must not be counted as correct", got.correct)
	}
}

func TestGradeAllTracksSkillsSeparately(t *testing.T) {
	got := gradeAll(bank(), []models.AnswerSubmission{
		{QuestionID: "r1", BlankResponses: map[string]string{"b1": "debunked"}},
		{QuestionID: "l1", TextResponse: "completely different words entirely"},
	})

	reading := got.bySkill[models.SkillReading]
	listening := got.bySkill[models.SkillListening]

	if reading.accuracy() != 1 {
		t.Errorf("reading accuracy = %v, want 1", reading.accuracy())
	}
	if listening.accuracy() >= reading.accuracy() {
		t.Errorf("listening accuracy = %v, want less than reading %v",
			listening.accuracy(), reading.accuracy())
	}
}

// An empty submission must produce no graded questions, so the handler can
// reject it rather than storing a zero as if it were a real result.
func TestGradeAllWithNoAnswers(t *testing.T) {
	got := gradeAll(bank(), nil)

	if got.total != 0 {
		t.Errorf("total = %d, want 0", got.total)
	}
}
