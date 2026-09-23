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
