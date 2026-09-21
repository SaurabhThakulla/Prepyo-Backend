package evaluations

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/scoring"
)

var (
	// ErrEmptyResponse means there is nothing worth evaluating.
	ErrEmptyResponse = errors.New("response is empty")
	// Skill errors reject submissions sent to the wrong evaluator endpoint.
	ErrWrongWritingSkill = errors.New("question is not a writing task")
	ErrWrongSkill        = errors.New("question is not a speaking task")
	ErrLimitReached      = billing.ErrLimitReached
)

// minWordsToEvaluate is a feedback threshold, not the full exam word limit.
// PTE summaries explicitly permit a single sentence of as few as five words.
func minWordsToEvaluate(q models.Question) int {
	if q.Exam == models.ExamPTE && q.TypeID == "summarize-written-text" {
		return 5
	}
	return 20
}

func validateWritingResponse(q models.Question, text string) error {
	if q.Skill != models.SkillWriting {
		return ErrWrongWritingSkill
	}
	if len(strings.Fields(text)) < minWordsToEvaluate(q) {
		return ErrEmptyResponse
	}
	return nil
}

// minRecordingSeconds is the speaking equivalent: below this there is not
// enough speech to say anything useful about, and a learner should not spend an
// evaluation to be told so.
const minRecordingSeconds = 2

type Service struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	exams     *exams.Repository
	billing   *billing.Service
	gateway   *ai.Gateway
	xp        *gamification.Service
}

func NewService(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	examRepo *exams.Repository,
	billingService *billing.Service,
	gateway *ai.Gateway,
	xp *gamification.Service,
) *Service {
	return &Service{
		db: db, repo: repo, questions: questionRepo, exams: examRepo,
		billing: billingService, gateway: gateway, xp: xp,
	}
}

// SpeakingTranscriptAvailable reports whether a spoken answer can be scored
// from a transcript the learner's device produced. That needs only the text
// provider, which is what a provider without audio models leaves us.
func (s *Service) SpeakingTranscriptAvailable() bool {
	return s.gateway.Available()
}

// SpeakingAvailable reports whether a recording can be scored at all. Handlers
// check this before accepting an upload, so a learner is told straight away
// rather than after a provider round trip that was never going to work.
func (s *Service) SpeakingAvailable() bool { return s.gateway.SpeakingAvailable() }

type Request struct {
	User       models.User
	QuestionID string
	Text       string
}

type Outcome struct {
	Evaluation models.Evaluation
	// Reused is true when an identical submission had already been evaluated,
	// so no provider call was made and no allowance was spent.
	Reused       bool
	XPAwarded    int
	Streak       int
	Missions     []models.DailyMission
	Subscription models.SubscriptionState
}

// EvaluateWriting runs the full flow: allowance, deduplication, provider call,
// validation, persistence and rewards.
func (s *Service) EvaluateWriting(ctx context.Context, req Request) (Outcome, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return Outcome{}, ErrEmptyResponse
	}

	question, err := s.questions.ByID(ctx, req.QuestionID)
	if err != nil {
		return Outcome{}, err
	}

	if err := validateWritingResponse(question, text); err != nil {
		return Outcome{}, err
	}

	// Same learner, same question, same words means the same feedback. Return
	// the stored result rather than paying for an identical call.
	// Do not reuse feedback produced before source passages were supplied.
	fingerprint := fingerprintOf(req.User.ID, question.ID+":"+ai.WritingPromptVersion, text)
	if existing, err := s.repo.ByFingerprint(ctx, req.User.ID, fingerprint); err == nil {
		state, err := s.billing.State(ctx, s.db, req.User)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{Evaluation: existing, Reused: true, Subscription: state}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Outcome{}, err
	}

	// The session already paid at start; submission needs no remaining credit.
	if err := s.billing.RequireStartedSubTest(ctx, s.db, req.User, question, question.Exam); err != nil {
		return Outcome{}, err
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return Outcome{}, err
	}

	evaluation, usage, err := s.gateway.EvaluateWriting(ctx, ai.WritingRequest{
		Exam:           question.Exam,
		TaskName:       question.TypeName,
		Prompt:         question.Prompt,
		FigureData:     question.FigureData,
		ContextPassage: question.ContextPassage,
		LearnerText:    text,
		MinScore:       version.MinScore,
		MaxScore:       version.MaxScore,
	})
	if err != nil {
		return Outcome{}, err
	}

	return s.persist(ctx, persistParams{
		User:        req.User,
		Question:    question,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage:       usage,
		XPReason:    "Writing evaluated: " + question.TypeName,
	})
}

// SpeakingRequest is one recording to evaluate. Audio arrives already encoded
// in a format the provider accepts; see ai.AudioFormats.
type SpeakingRequest struct {
	User            models.User
	QuestionID      string
	Audio           []byte
	AudioFormat     string
	DurationSeconds int
}

// EvaluateSpeaking runs the same flow as EvaluateWriting for a recording.
//
// The recording itself is not stored. It is sent to the provider, transcribed,
// scored, and dropped; what persists is the transcript and the feedback. That
// keeps a learner's voice out of the database entirely, and the transcript is
// the part the feedback actually quotes.
func (s *Service) EvaluateSpeaking(ctx context.Context, req SpeakingRequest) (Outcome, error) {
	if len(req.Audio) == 0 || req.DurationSeconds < minRecordingSeconds {
		return Outcome{}, ErrEmptyResponse
	}

	question, err := s.questions.ByID(ctx, req.QuestionID)
	if err != nil {
		return Outcome{}, err
	}
	// A recording only means something against a speaking task. Without this a
	// reading question could be sent here and scored on criteria it has none of.
	if question.Skill != models.SkillSpeaking {
		return Outcome{}, ErrWrongSkill
	}

	// Two identical uploads are a double-click or a retry, not two attempts.
	// Distinct recordings of the same words hash differently, which is right:
	// they are different performances and deserve their own feedback.
	fingerprint := fingerprintOfAudio(req.User.ID, question.ID, req.Audio)
	if existing, err := s.repo.ByFingerprint(ctx, req.User.ID, fingerprint); err == nil {
		state, err := s.billing.State(ctx, s.db, req.User)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{Evaluation: existing, Reused: true, Subscription: state}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Outcome{}, err
	}

	// The session already paid at start; submission needs no remaining credit.
	if err := s.billing.RequireStartedSubTest(ctx, s.db, req.User, question, question.Exam); err != nil {
		return Outcome{}, err
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return Outcome{}, err
	}

	material := speakingMaterialOf(question)
	evaluation, usage, err := s.gateway.EvaluateSpeaking(ctx, ai.SpeakingRequest{
		Exam:            question.Exam,
		TaskName:        question.TypeName,
		Prompt:          question.Prompt,
		ExpectedText:    material.expected,
		SourceText:      material.source,
		ReferenceAnswer: material.reference,
		AudioBase64:     base64.StdEncoding.EncodeToString(req.Audio),
		AudioFormat:     req.AudioFormat,
		DurationSeconds: req.DurationSeconds,
		MinScore:        version.MinScore,
		MaxScore:        version.MaxScore,
	})
	if err != nil {
		return Outcome{}, err
	}

	return s.persist(ctx, persistParams{
		User:        req.User,
		Question:    question,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage:       usage,
		XPReason:    "Speaking evaluated: " + question.TypeName,
	})
}

// TranscriptRequest is a spoken answer that was transcribed on the device.
type TranscriptRequest struct {
	User            models.User
	QuestionID      string
	Transcript      string
	DurationSeconds int
	// Delivery is what the browser measured while recording: pace and the
	// silences in between. Empty when the browser could not measure it, in
	// which case nothing here judges pace at all.
	Delivery scoring.Delivery
}

// minTranscriptWords is the point below which there is nothing to give feedback
// on. Two or three recognised words is a learner who was not heard, and they
// should be told that rather than handed a band for it.
const minTranscriptWords = 8

// EvaluateSpeakingTranscript scores a spoken answer from its transcript.
//
// It runs the same flow as the recording path - allowance, deduplication,
// provider call, validation, persistence, rewards - with one difference that
// matters to the learner: pronunciation is not judged, because nothing in this
// path ever heard them speak.
func (s *Service) EvaluateSpeakingTranscript(ctx context.Context, req TranscriptRequest) (Outcome, error) {
	transcript := strings.TrimSpace(req.Transcript)
	words := len(strings.Fields(transcript))
	if words < minTranscriptWords {
		return Outcome{}, ErrEmptyResponse
	}

	question, err := s.questions.ByID(ctx, req.QuestionID)
	if err != nil {
		return Outcome{}, err
	}
	if question.Skill != models.SkillSpeaking {
		return Outcome{}, ErrWrongSkill
	}

	// The same words spoken again are the same submission to judge, so the
	// fingerprint is over the transcript rather than the audio bytes.
	fingerprint := fingerprintOf(req.User.ID, question.ID+":"+ai.SpeakingTranscriptPromptVersion, transcript)
	if existing, err := s.repo.ByFingerprint(ctx, req.User.ID, fingerprint); err == nil {
		state, err := s.billing.State(ctx, s.db, req.User)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{Evaluation: existing, Reused: true, Subscription: state}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Outcome{}, err
	}

	if err := s.billing.RequireStartedSubTest(ctx, s.db, req.User, question, question.Exam); err != nil {
		return Outcome{}, err
	}

	version, err := s.exams.ByID(ctx, question.ExamVersionID)
	if err != nil {
		return Outcome{}, err
	}

	material := speakingMaterialOf(question)

	// Read Aloud and Repeat Sentence have one right answer, word for word. That
	// is an alignment, not an opinion: scoring it here is instant, costs no
	// provider call, and gives the same answer every time, which a model does
	// not. See scoring.ScoreVerbatim.
	if strings.TrimSpace(material.expected) != "" {
		evaluation := verbatimEvaluation(question, version, material.expected, transcript, req.Delivery)
		return s.persist(ctx, persistParams{
			User:        req.User,
			Question:    question,
			Fingerprint: fingerprint,
			Evaluation:  evaluation,
			Usage: ai.Usage{
				Provider:      "prepyo",
				Model:         "verbatim-alignment",
				PromptVersion: verbatimScoringVersion,
			},
			XPReason: "Speaking evaluated: " + question.TypeName,
		})
	}

	evaluation, usage, err := s.gateway.EvaluateSpokenTranscript(ctx, ai.SpokenTranscriptRequest{
		Exam:            question.Exam,
		TaskName:        question.TypeName,
		Prompt:          question.Prompt,
		ExpectedText:    material.expected,
		SourceText:      material.source,
		ReferenceAnswer: material.reference,
		Transcript:      transcript,
		DurationSeconds: req.DurationSeconds,
		WordCount:       words,
		Delivery:        req.Delivery.Summary(),
		MinScore:        version.MinScore,
		MaxScore:        version.MaxScore,
	})
	if err != nil {
		return Outcome{}, err
	}

	return s.persist(ctx, persistParams{
		User:        req.User,
		Question:    question,
		Fingerprint: fingerprint,
		Evaluation:  evaluation,
		Usage:       usage,
		XPReason:    "Speaking evaluated: " + question.TypeName,
	})
}

type persistParams struct {
	User        models.User
	Question    models.Question
	Fingerprint []byte
	Evaluation  models.Evaluation
	Usage       ai.Usage
	XPReason    string
}

// persistTimeout bounds the writes below on their own, now that they no longer
// inherit whatever is left of the request's deadline.
const persistTimeout = 20 * time.Second

// persist stores a validated evaluation and pays out for it, in one transaction
// so a learner cannot end up with XP for feedback that was never saved.
func (s *Service) persist(ctx context.Context, p persistParams) (Outcome, error) {
	// By this point the provider has answered and the allowance has been spent.
	// Saving must not be cancelled because the learner's request ran out of
	// deadline while the model was thinking: that loses feedback they have
	// already paid for and surfaces as a database error, which reads to them as
	// "something went wrong on our side" rather than "that took too long".
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), persistTimeout)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Outcome{}, fmt.Errorf("begin evaluation: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serialize reward writes, but do not charge for saving feedback.
	if err := billing.LockUserForQuota(ctx, tx, p.User.ID); err != nil {
		return Outcome{}, err
	}
	if err := s.billing.RequireStartedSubTest(ctx, tx, p.User, p.Question, p.Question.Exam); err != nil {
		return Outcome{}, err
	}

	saved, err := s.repo.Save(ctx, tx, SaveParams{
		UserID:      p.User.ID,
		QuestionID:  p.Question.ID,
		Fingerprint: p.Fingerprint,
		Evaluation:  p.Evaluation,
		Usage: models.EvaluationUsage{
			Provider:         p.Usage.Provider,
			Model:            p.Usage.Model,
			PromptVersion:    p.Usage.PromptVersion,
			PromptTokens:     p.Usage.PromptTokens,
			CompletionTokens: p.Usage.CompletionTokens,
			LatencyMS:        p.Usage.LatencyMS,
		},
	})
	if err != nil {
		return Outcome{}, err
	}

	awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    p.User.ID,
		Amount:    gamification.XPEvaluation,
		Reason:    p.XPReason,
		SourceKey: "evaluation:" + saved.ID,
	})
	if err != nil {
		return Outcome{}, err
	}

	streak, err := s.xp.TouchStreak(ctx, tx, p.User)
	if err != nil {
		return Outcome{}, err
	}

	missions, err := s.xp.RecordActivity(ctx, tx, p.User, p.Question.Skill)
	if err != nil {
		return Outcome{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Outcome{}, fmt.Errorf("commit evaluation: %w", err)
	}

	// Read the allowance again so the response shows usage including this call.
	state, err := s.billing.State(ctx, s.db, p.User)
	if err != nil {
		return Outcome{}, err
	}

	return Outcome{
		Evaluation:   saved,
		XPAwarded:    awarded,
		Streak:       streak,
		Missions:     missions,
		Subscription: state,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// speakingMaterial is what a speaking task gives the grader to judge against.
type speakingMaterial struct {
	// expected is the text to say word for word. Only Read Aloud and Repeat
	// Sentence have one.
	expected string
	// source is what every other task responds to: a lecture, a discussion, a
	// situation, a question. It must never reach the grader as text to repeat,
	// or a correct retelling is marked down for every word it paraphrases.
	source string
	// reference is the item's model answer, if it has one.
	reference string
}

// speakingMaterialOf sorts a question's text into the part the learner must
// repeat and the part they must respond to. Read Aloud puts the words to say in
// the passage; Repeat Sentence puts them in the transcript.
func speakingMaterialOf(q models.Question) speakingMaterial {
	text := firstNonEmpty(q.ContextPassage, q.AudioTranscript)
	material := speakingMaterial{reference: strings.TrimSpace(q.ModelAnswer)}
	if scoring.IsVerbatimTask(q.TypeID) {
		material.expected = text
	} else {
		material.source = text
	}
	return material
}

// fingerprintOf identifies a submission. Whitespace is normalised so reformatting
// the same essay counts as the same submission.
func fingerprintOf(userID, questionID, text string) []byte {
	normalised := strings.ToLower(strings.Join(strings.Fields(text), " "))
	sum := sha256.Sum256([]byte(userID + "\x00" + questionID + "\x00" + normalised))
	return sum[:]
}

// fingerprintOfAudio is the same idea for a recording. There is nothing to
// normalise: only the identical bytes count as the identical submission.
func fingerprintOfAudio(userID, questionID string, audio []byte) []byte {
	h := sha256.New()
	h.Write([]byte(userID + "\x00" + questionID + "\x00"))
	h.Write(audio)
	return h.Sum(nil)
}
