package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// Reading content authoring.
//
// Until now every passage arrived through a migration. This is the first write
// path, so it is deliberately strict: a passage that grades wrongly is worse
// than one that was never added, and a learner is the last person who should
// discover that an answer key does not match its options.
//
// Everything lands in one transaction. A half-written passage — groups with no
// questions, questions pointing at a group that failed — would be invisible
// until someone practised it.

// answerStyle says how a task's answer key is expressed, which is what decides
// how it can be validated.
type answerStyle int

const (
	// styleChoice: the author supplies the options; answers name them.
	styleChoice answerStyle = iota
	// styleVerdict: the options are fixed by the task itself.
	styleVerdict
	// styleParagraph: the answer is a paragraph label from the passage.
	styleParagraph
	// styleText: free text, optionally with word choices.
	styleText
)

type readingTypeSpec struct {
	name string
	// The type ids the frontend sends, from src/lib/readingTypes.ts.
	style answerStyle
	// multi allows more than one correct answer.
	multi bool
	// fixedOptions is the answer set for verdict tasks, in display order.
	fixedOptions []questionOption
}

type questionOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// supportedTypes is every reading task this endpoint can author.
//
// Re-order Paragraphs and the two PTE gap-fills are deliberately absent: their
// answers live in shapes this form does not carry — ordered item rows and
// per-blank keys — so accepting them here would store something the grader
// cannot read. They stay with the migrations until the form grows to match.
var supportedTypes = map[string]readingTypeSpec{
	"reading-mcq-single":          {name: "Multiple Choice, Single Answer", style: styleChoice},
	"reading-mcq-multiple":        {name: "Multiple Choice, Multiple Answers", style: styleChoice, multi: true},
	"reading-sentence-completion": {name: "Sentence Completion", style: styleText, multi: true},
	"reading-find-the-paragraph":  {name: "Find the Paragraph", style: styleParagraph},
	"reading-matching-information": {
		name: "Matching Information", style: styleParagraph,
	},
	"reading-true-false": {name: "True / False / Not Given", style: styleVerdict, fixedOptions: []questionOption{
		{ID: "TRUE", Text: "True"},
		{ID: "FALSE", Text: "False"},
		{ID: "NOT_GIVEN", Text: "Not Given"},
	}},
	"reading-yes-no-not-given": {name: "Yes / No / Not Given", style: styleVerdict, fixedOptions: []questionOption{
		{ID: "YES", Text: "Yes"},
		{ID: "NO", Text: "No"},
		{ID: "NOT_GIVEN", Text: "Not Given"},
	}},
}

type newQuestion struct {
	Prompt string `json:"prompt"`
	// Options are the author's own wording. Ids are derived from the text, the
	// convention the seeded content already uses, so an answer key reads as
	// words rather than as opaque handles.
	Options []string `json:"options"`
	// CorrectAnswers name option texts for a chooser, a paragraph label for a
	// matching task, or the accepted spellings for a typed answer.
	CorrectAnswers []string `json:"correctAnswers"`
	Explanation    string   `json:"explanation"`
	Points         int      `json:"points"`
}

type newGroup struct {
	TypeID           string        `json:"typeId"`
	Instructions     string        `json:"instructions"`
	TimeLimitSeconds int           `json:"timeLimitSeconds"`
	Questions        []newQuestion `json:"questions"`
}

type newPassage struct {
	Exam     string `json:"exam"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Body is the passage as typed, split into paragraphs on blank lines.
	// Paragraphs, when given, wins and is used as-is.
	Body       string     `json:"body"`
	Paragraphs []string   `json:"paragraphs"`
	Difficulty string     `json:"difficulty"`
	Topic      string     `json:"topic"`
	Tags       []string   `json:"tags"`
	Publish    bool       `json:"publish"`
	Groups     []newGroup `json:"groups"`
}

// createPassage writes a passage, its question groups and their questions.
func (h *Handler) createPassage(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	var req newPassage
	// A passage is long, and a set of questions longer. The default decode cap
	// is meant for short forms and would reject real content.
	if !httpx.DecodeLimit(w, r, &req, h.log, "admin.createPassage", 2<<20) {
		return
	}

	paragraphs, problems := req.normalise()
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	id, err := h.insertPassage(r.Context(), req, paragraphs)
	if err != nil {
		var invalid validationError
		if errors.As(err, &invalid) {
			httpx.ValidationError(w, invalid.fields)
			return
		}
		httpx.Internal(w, h.log, "admin.createPassage", err)
		return
	}

	h.log.Info("admin created a reading passage",
		"actor", actor.Email, "passage", id, "exam", req.Exam,
		"groups", len(req.Groups), "published", req.Publish)

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"passage": map[string]any{
			"id":         id,
			"title":      strings.TrimSpace(req.Title),
			"exam":       req.Exam,
			"paragraphs": len(paragraphs),
			"groups":     len(req.Groups),
			"questions":  req.questionCount(),
			"published":  req.Publish,
		},
	})
}

func (p newPassage) questionCount() int {
	total := 0
	for _, group := range p.Groups {
		total += len(group.Questions)
	}
	return total
}

// paragraph is one labelled block of the passage, the shape reading_passages
// stores and the reader renders.
type paragraph struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// normalise validates the passage itself and returns its labelled paragraphs.
// Question-level checks need the paragraph labels, so they happen later.
func (p newPassage) normalise() ([]paragraph, map[string]string) {
	problems := map[string]string{}

	if p.Exam != string(models.ExamPTE) && p.Exam != string(models.ExamIELTS) {
		problems["exam"] = "Choose PTE or IELTS."
	}
	if strings.TrimSpace(p.Title) == "" {
		problems["title"] = "Give the passage a title."
	}
	switch p.Difficulty {
	case "", "easy", "medium", "hard":
	default:
		problems["difficulty"] = "Difficulty must be easy, medium or hard."
	}

	blocks := p.Paragraphs
	if len(blocks) == 0 {
		// Blank lines are how anyone pasting prose already separates
		// paragraphs, so that is what splits them.
		for _, block := range strings.Split(strings.ReplaceAll(p.Body, "\r\n", "\n"), "\n\n") {
			if strings.TrimSpace(block) != "" {
				blocks = append(blocks, block)
			}
		}
	}

	paragraphs := make([]paragraph, 0, len(blocks))
	for i, block := range blocks {
		text := strings.TrimSpace(block)
		if text == "" {
			continue
		}
		paragraphs = append(paragraphs, paragraph{Label: paragraphLabel(i), Text: text})
	}
	if len(paragraphs) == 0 {
		problems["body"] = "Paste the passage text."
	}

	if len(p.Groups) == 0 {
		problems["groups"] = "Add at least one question set."
	}

	return paragraphs, problems
}

// paragraphLabel numbers paragraphs A, B, C … then AA, AB for a passage long
// enough to run past Z, which the matching tasks answer with.
func paragraphLabel(index int) string {
	if index < 26 {
		return string(rune('A' + index))
	}
	return string(rune('A'+index/26-1)) + string(rune('A'+index%26))
}

// minPassageWords mirrors reading_passages_word_count_floor from 000009.
const minPassageWords = 700

type validationError struct{ fields map[string]string }

func (e validationError) Error() string { return "reading passage failed validation" }

func (h *Handler) insertPassage(ctx context.Context, req newPassage, paragraphs []paragraph) (string, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin passage tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var versionID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM exam_versions WHERE exam = $1 AND is_current`, req.Exam).Scan(&versionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", validationError{fields: map[string]string{
			"exam": "That exam has no current version configured.",
		}}
	}
	if err != nil {
		return "", fmt.Errorf("find current exam version: %w", err)
	}

	labels := make(map[string]bool, len(paragraphs))
	for _, para := range paragraphs {
		labels[para.Label] = true
	}

	// 000009 puts a floor under passage length: shorter than this and there is
	// not enough text to carry three task types. Checked here rather than left
	// to the constraint so the author is told the number, not shown a 500.
	words := wordCount(paragraphs)
	if words < minPassageWords {
		return "", validationError{fields: map[string]string{
			"body": fmt.Sprintf("A reading passage needs at least %d words; this one has %d.", minPassageWords, words),
		}}
	}

	passageID := newContentID("passage", req.Title)
	paragraphJSON, err := json.Marshal(paragraphs)
	if err != nil {
		return "", fmt.Errorf("encode paragraphs: %w", err)
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = "medium"
	}

	// The passage's exam is carried by its version — 000028 dropped the column
	// that repeated it here.
	if _, err := tx.Exec(ctx, `
		INSERT INTO reading_passages
			(id, exam_version_id, title, subtitle, paragraphs, word_count,
			 difficulty, topic, tags, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		passageID, versionID, strings.TrimSpace(req.Title), strings.TrimSpace(req.Subtitle),
		paragraphJSON, words, difficulty, strings.TrimSpace(req.Topic),
		normaliseTags(req.Tags), req.Publish,
	); err != nil {
		return "", fmt.Errorf("insert passage: %w", err)
	}

	for gi, group := range req.Groups {
		spec, ok := supportedTypes[group.TypeID]
		if !ok {
			return "", validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.typeId", gi): "This task type cannot be authored here yet.",
			}}
		}
		if len(group.Questions) == 0 {
			return "", validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.questions", gi): "Add at least one question.",
			}}
		}

		groupID := newContentID("group", fmt.Sprintf("%s-%d", passageID, gi+1))
		if _, err := tx.Exec(ctx, `
			INSERT INTO reading_question_groups
				(id, passage_id, position, type_id, type_name, instructions,
				 shuffle_questions, time_limit_seconds)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			groupID, passageID, gi+1, group.TypeID, spec.name,
			strings.TrimSpace(group.Instructions),
			// Authored sets read in the order they were written; shuffling a
			// sentence-completion paragraph would scramble the prose.
			spec.style != styleText, group.TimeLimitSeconds,
		); err != nil {
			return "", fmt.Errorf("insert question group: %w", err)
		}

		for qi, question := range group.Questions {
			options, answers, problem := resolveAnswers(spec, question, labels)
			if problem != "" {
				return "", validationError{fields: map[string]string{
					fmt.Sprintf("groups.%d.questions.%d", gi, qi): problem,
				}}
			}

			optionsJSON, err := json.Marshal(options)
			if err != nil {
				return "", fmt.Errorf("encode options: %w", err)
			}
			answersJSON, err := json.Marshal(answers)
			if err != nil {
				return "", fmt.Errorf("encode answers: %w", err)
			}

			points := question.Points
			if points <= 0 {
				points = 10
			}

			// supported_exams is how one passage can serve both papers. An
			// authored question belongs to the exam it was written for, and the
			// column must not be empty.
			if _, err := tx.Exec(ctx, `
				INSERT INTO questions
					(id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
					 options, correct_answers, explanation, difficulty, points,
					 time_limit_seconds, is_published, passage_id, group_id, group_position)
				VALUES ($1, $2, $3, $4, 'reading', $5, $6, $7, $8, $9, $10, $11, $12, $13, 0, $14, $15, $16, $17)`,
				newContentID("q", fmt.Sprintf("%s-%d", groupID, qi+1)), versionID, req.Exam,
				[]string{req.Exam},
				group.TypeID, spec.name, strings.TrimSpace(req.Title), strings.TrimSpace(question.Prompt),
				optionsJSON, answersJSON, strings.TrimSpace(question.Explanation), difficulty, points,
				req.Publish, passageID, groupID, qi+1,
			); err != nil {
				return "", fmt.Errorf("insert question: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit passage tx: %w", err)
	}
	return passageID, nil
}

// resolveAnswers turns what the author typed into the options and answer key
// the grader reads, and refuses anything it cannot grade.
//
// The rule that matters is the last one: every answer must name something that
// exists. An answer key pointing at an option nobody can pick, or a paragraph
// the passage does not have, marks every attempt wrong forever.
func resolveAnswers(spec readingTypeSpec, q newQuestion, labels map[string]bool) ([]questionOption, []string, string) {
	if strings.TrimSpace(q.Prompt) == "" {
		return nil, nil, "Write the question."
	}

	answers := make([]string, 0, len(q.CorrectAnswers))
	for _, answer := range q.CorrectAnswers {
		if trimmed := strings.TrimSpace(answer); trimmed != "" {
			answers = append(answers, trimmed)
		}
	}
	if len(answers) == 0 {
		return nil, nil, "Mark the correct answer."
	}
	if !spec.multi && len(answers) > 1 {
		return nil, nil, "This task takes exactly one correct answer."
	}

	switch spec.style {
	case styleVerdict:
		valid := map[string]string{}
		for _, option := range spec.fixedOptions {
			valid[strings.ToLower(option.Text)] = option.ID
			valid[strings.ToLower(option.ID)] = option.ID
		}
		id, ok := valid[strings.ToLower(answers[0])]
		if !ok {
			return nil, nil, "Choose one of the answers this task allows."
		}
		return spec.fixedOptions, []string{id}, ""

	case styleParagraph:
		for _, answer := range answers {
			if !labels[strings.ToUpper(answer)] {
				return nil, nil, fmt.Sprintf("Paragraph %q is not in this passage.", answer)
			}
		}
		options := make([]questionOption, 0, len(labels))
		for label := range labels {
			options = append(options, questionOption{ID: label, Text: "Paragraph " + label})
		}
		sortOptions(options)
		upper := make([]string, len(answers))
		for i, answer := range answers {
			upper[i] = strings.ToUpper(answer)
		}
		return options, upper, ""

	default:
		options := make([]questionOption, 0, len(q.Options))
		known := map[string]bool{}
		for _, text := range q.Options {
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			if known[text] {
				return nil, nil, fmt.Sprintf("Option %q is listed twice.", text)
			}
			known[text] = true
			// id == text, the convention the seeded content uses, so a review
			// screen can print the answer key as words.
			options = append(options, questionOption{ID: text, Text: text})
		}

		if spec.style == styleChoice {
			if len(options) < 2 {
				return nil, nil, "Give at least two options."
			}
		}
		// A typed answer may have no options at all; when it does have them,
		// the key still has to name one.
		if len(options) > 0 {
			for _, answer := range answers {
				if !known[answer] {
					return nil, nil, fmt.Sprintf("%q is not one of the options.", answer)
				}
			}
		}
		return options, answers, ""
	}
}

func sortOptions(options []questionOption) {
	for i := 1; i < len(options); i++ {
		for j := i; j > 0 && options[j].ID < options[j-1].ID; j-- {
			options[j], options[j-1] = options[j-1], options[j]
		}
	}
}

func wordCount(paragraphs []paragraph) int {
	total := 0
	for _, para := range paragraphs {
		total += len(strings.Fields(para.Text))
	}
	return total
}

func normaliseTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	return out
}

// newContentID builds a readable, unique id. Content ids are TEXT and appear in
// URLs and logs, so they carry a slug of the title; the random suffix is what
// makes re-adding a passage with the same name safe.
func newContentID(prefix, seed string) string {
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32
		case r == ' ', r == '-', r == '_':
			return '-'
		default:
			return -1
		}
	}, seed)

	slug = strings.Trim(slug, "-")
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	if slug == "" {
		slug = prefix
	}

	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s-%s-%s", prefix, slug, hex.EncodeToString(buf))
}
