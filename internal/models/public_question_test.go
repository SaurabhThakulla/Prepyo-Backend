package models

import "testing"

// With a recording, the transcript is the Write from Dictation answer and must
// not reach the learner. Without one, the browser reads it aloud, so it stays.
func TestPublicQuestionTranscript(t *testing.T) {
	const script = "The library closes at nine."

	withAudio := Question{AudioURL: "/api/v1/questions/assets/a1", AudioTranscript: script}
	if got := withAudio.PublicQuestion().AudioTranscript; got != "" {
		t.Fatalf("transcript sent alongside a recording: %q", got)
	}
	if got := withAudio.ForReview().AudioTranscript; got != script {
		t.Fatalf("review lost the transcript: %q", got)
	}

	spokenByBrowser := Question{AudioTranscript: script}
	if got := spokenByBrowser.PublicQuestion().AudioTranscript; got != script {
		t.Fatalf("transcript dropped with no recording to play: %q", got)
	}
}

// An IELTS listening script is every answer on the item. It is served at play
// time, not in the question payload; review restores it.
func TestIELTSListeningScriptIsServedOnRequest(t *testing.T) {
	const script = "Receptionist: The class used to be on Tuesday; from next month it will be Thursday."
	q := Question{Exam: ExamIELTS, SupportedExams: []ExamType{ExamIELTS}, Skill: SkillListening, AudioTranscript: script}

	public := q.PublicQuestion()
	if public.AudioTranscript != "" || !public.ScriptOnRequest {
		t.Fatalf("public = transcript %q, scriptOnRequest %v; want no transcript and on-request", public.AudioTranscript, public.ScriptOnRequest)
	}
	if q.PlaybackScript() != script {
		t.Fatalf("playback script = %q", q.PlaybackScript())
	}
	if q.ForReview().AudioTranscript != script {
		t.Fatal("review lost the transcript")
	}
}

// "Choose TWO letters" tells an IELTS learner how many to pick; PTE does not.
func TestSelectCountOnlyForIELTSMultipleAnswers(t *testing.T) {
	options := []QuestionOption{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	ielts := Question{Exam: ExamIELTS, SupportedExams: []ExamType{ExamIELTS}, Options: options, CorrectAnswers: []string{"A", "C"}}
	if got := ielts.PublicQuestion().SelectCount; got != 2 {
		t.Fatalf("IELTS select count = %d, want 2", got)
	}
	pte := Question{Exam: ExamPTE, SupportedExams: []ExamType{ExamPTE}, Options: options, CorrectAnswers: []string{"A", "C"}}
	if got := pte.PublicQuestion().SelectCount; got != 0 {
		t.Fatalf("PTE select count = %d, want 0", got)
	}
}
