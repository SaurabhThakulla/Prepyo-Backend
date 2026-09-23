// Package speech speaks listening scripts on the server, for browsers that
// cannot speak them themselves.
//
// A script is split into turns by its speaker labels ("Receptionist: …"), the
// labels are dropped, and each turn is cut into pieces short enough for the
// provider (200 characters for Groq Orpheus) at sentence and then phrase
// boundaries. Each speaker gets their own voice. A clip is made the first time
// it is played and cached by a hash of model, voice and text.
package speech

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
)

// ErrClipNotFound means no clip has that id.
var ErrClipNotFound = errors.New("speech clip not found")

// Synthesizer speaks one short piece of text. ai.Gateway implements it.
type Synthesizer interface {
	SpeechAvailable() bool
	SpeechModel() string
	SpeechVoices() []string
	Synthesize(ctx context.Context, voice, text string) ([]byte, error)
}

// Segment is one clip of a spoken script, in playing order.
type Segment struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Speaker string `json:"speaker,omitempty"`
}

// Turn is what one speaker says before the next speaks.
type Turn struct {
	Speaker string
	Text    string
}

// Piece is text short enough for one request, with its speaker.
type Piece struct {
	Speaker string
	Text    string
}

type Service struct {
	db    database.DB
	voice Synthesizer
	// MaxChars is the longest piece sent in one request.
	MaxChars int
}

func NewService(db database.DB, voice Synthesizer, maxChars int) *Service {
	return &Service{db: db, voice: voice, MaxChars: maxChars}
}

// Available reports whether the server can speak at all.
func (s *Service) Available() bool { return s.voice != nil && s.voice.SpeechAvailable() }

var (
	leadingLabel = regexp.MustCompile(`^([A-Z][A-Za-z]+(?: [A-Z][A-Za-z]+)?):\s+`)
	anyLabel     = regexp.MustCompile(`(?:^|([.!?]["”’)]?)\s+|\n+)([A-Z][A-Za-z]+(?: [A-Z][A-Za-z]+)?):\s+`)
	sentenceEnd  = regexp.MustCompile(`[^.!?]+[.!?]+["”’)]*\s*|[^.!?]+$`)
)

// SplitTurns divides a script into turns. A script that does not open with a
// speaker label is one turn with no speaker.
func SplitTurns(script string) []Turn {
	text := strings.TrimSpace(script)
	if text == "" {
		return nil
	}
	if !leadingLabel.MatchString(text) {
		return []Turn{{Text: text}}
	}
	matches := anyLabel.FindAllStringSubmatchIndex(text, -1)
	var turns []Turn
	for i, m := range matches {
		speaker := text[m[4]:m[5]]
		start := m[1]
		end := len(text)
		if i+1 < len(matches) {
			next := matches[i+1]
			// Keep the sentence-ending punctuation the label match consumed.
			end = next[0]
			if next[2] >= 0 {
				end = next[3]
			}
		}
		if spoken := strings.TrimSpace(text[start:end]); spoken != "" {
			turns = append(turns, Turn{Speaker: speaker, Text: spoken})
		}
	}
	if len(turns) == 0 {
		return []Turn{{Text: text}}
	}
	return turns
}

// SplitPieces cuts turns into pieces of at most max characters, breaking at
// sentence ends, then at commas and semicolons, then at spaces.
func SplitPieces(turns []Turn, max int) []Piece {
	var pieces []Piece
	for _, turn := range turns {
		var current string
		flush := func() {
			if t := strings.TrimSpace(current); t != "" {
				pieces = append(pieces, Piece{Speaker: turn.Speaker, Text: t})
			}
			current = ""
		}
		for _, sentence := range sentenceEnd.FindAllString(turn.Text, -1) {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}
			for _, part := range fit(sentence, max) {
				if len([]rune(current))+len([]rune(part))+1 > max {
					flush()
				}
				if current == "" {
					current = part
				} else {
					current += " " + part
				}
			}
		}
		flush()
	}
	return pieces
}

// fit breaks one sentence into parts no longer than max.
func fit(sentence string, max int) []string {
	if len([]rune(sentence)) <= max {
		return []string{sentence}
	}
	var parts []string
	var current string
	for _, clause := range splitKeep(sentence, ",;:") {
		for _, word := range strings.Fields(clause) {
			if len([]rune(current))+len([]rune(word))+1 > max && current != "" {
				parts = append(parts, strings.TrimSpace(current))
				current = ""
			}
			if current == "" {
				current = word
			} else {
				current += " " + word
			}
		}
	}
	if strings.TrimSpace(current) != "" {
		parts = append(parts, strings.TrimSpace(current))
	}
	return parts
}

func splitKeep(s, seps string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if strings.ContainsRune(seps, r) {
			out = append(out, s[start:i+1])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// VoiceFor gives each speaker their own voice, in order of first appearance.
// fixed pins named speakers to a voice, as the examiner always is.
func VoiceFor(pieces []Piece, voices []string, fixed map[string]string) map[string]string {
	assigned := map[string]string{}
	next := 0
	for _, p := range pieces {
		if _, ok := assigned[p.Speaker]; ok {
			continue
		}
		if v, ok := fixed[p.Speaker]; ok {
			assigned[p.Speaker] = v
			continue
		}
		assigned[p.Speaker] = voices[next%len(voices)]
		next++
	}
	return assigned
}

func clipID(model, voice, text string) string {
	sum := sha256.Sum256([]byte(model + "\x00" + voice + "\x00" + text))
	return hex.EncodeToString(sum[:16])
}

// Segments lists the clips a script is spoken as, registering any that are
// new. Nothing is synthesised here: each clip is made when first played.
func (s *Service) Segments(ctx context.Context, script string, fixed map[string]string) ([]Segment, error) {
	if !s.Available() {
		return nil, fmt.Errorf("server speech is not configured")
	}
	pieces := SplitPieces(SplitTurns(script), s.MaxChars)
	voices := VoiceFor(pieces, s.voice.SpeechVoices(), fixed)
	model := s.voice.SpeechModel()

	segments := make([]Segment, 0, len(pieces))
	for _, p := range pieces {
		voice := voices[p.Speaker]
		id := clipID(model, voice, p.Text)
		if _, err := s.db.Exec(ctx, `
			INSERT INTO speech_clips (id, model, voice, text) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING`, id, model, voice, p.Text); err != nil {
			return nil, fmt.Errorf("register speech clip: %w", err)
		}
		segments = append(segments, Segment{ID: id, URL: "/api/v1/speech/" + id, Speaker: p.Speaker})
	}
	return segments, nil
}

// Clip returns a clip's audio, synthesising and caching it on first request.
func (s *Service) Clip(ctx context.Context, id string) ([]byte, string, error) {
	var model, voice, text, contentType string
	var data []byte
	err := s.db.QueryRow(ctx, `SELECT model, voice, text, content_type, data FROM speech_clips WHERE id = $1`, id).
		Scan(&model, &voice, &text, &contentType, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrClipNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("read speech clip: %w", err)
	}
	if len(data) > 0 {
		return data, contentType, nil
	}

	audio, err := s.voice.Synthesize(ctx, voice, text)
	if err != nil {
		return nil, "", err
	}
	// Saved even if the request is cancelled now: the provider has been paid.
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if _, err := s.db.Exec(saveCtx, `
		UPDATE speech_clips SET data = $2, synthesized_at = now() WHERE id = $1 AND data IS NULL`, id, audio); err != nil {
		return nil, "", fmt.Errorf("cache speech clip: %w", err)
	}
	return audio, "audio/wav", nil
}
