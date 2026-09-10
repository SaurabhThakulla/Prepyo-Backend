package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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
	// styleBlanks: the answer key is per gap inside a gapped text the question
	// carries itself, rather than one answer for the question as a whole.
	styleBlanks
	styleResource
)

var (
	examsBoth  = []string{"PTE", "IELTS"}
	examsPTE   = []string{"PTE"}
	examsIELTS = []string{"IELTS"}
)

type readingTypeSpec struct {
	name string
	// The type ids the frontend sends, from src/lib/readingTypes.ts.
	style answerStyle
	// multi allows more than one correct answer.
	multi bool
	// fixedOptions is the answer set for verdict tasks, in display order.
	fixedOptions []questionOption
	// wordBank says the gaps are filled from one list shared by the whole
	// group, with more words in it than gaps, rather than from choices of
	// their own.
	wordBank bool
	exams    []string
	display  string
	shuffle  bool
}

func (s readingTypeSpec) setsExam(exam string) bool {
	for _, allowed := range s.exams {
		if allowed == exam {
			return true
		}
	}
	return false
}

type questionOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// supportedTypes is every reading task this endpoint can author.
//
// Re-order Paragraphs is deliberately absent: an item is not a passage task —
// it lives in reading_reorder_items with its own boxes — so it is authored
// through createReorderItem rather than here.
var supportedTypes = map[string]readingTypeSpec{
	"fill-in-blanks-rw": {
		name: "Reading & Writing: Fill in the Blanks", style: styleBlanks,
		exams: examsPTE, display: "hidden",
	},
	"fill-in-blanks-r": {
		name: "Reading: Fill in the Blanks", style: styleBlanks, wordBank: true,
		exams: examsPTE, display: "hidden",
	},
	"reading-mcq-single": {
		name: "Multiple Choice, Single Answer", style: styleChoice,
		exams: examsBoth, display: "full", shuffle: true,
	},
	"reading-mcq-multiple": {
		name: "Multiple Choice, Multiple Answers", style: styleChoice, multi: true,
		exams: examsBoth, display: "full", shuffle: true,
	},
	"reading-sentence-completion": {
		name: "Sentence Completion", style: styleText, multi: true,
		exams: examsIELTS, display: "full",
	},
	"reading-find-the-paragraph": {
		name: "Find the Paragraph", style: styleParagraph,
		exams: examsBoth, display: "full", shuffle: true,
	},
	"reading-matching-information": {
		name: "Matching Information", style: styleParagraph,
		exams: examsBoth, display: "full", shuffle: true,
	},
	"reading-arrange-passage": {
		name: "Arrange the Passage", style: styleResource,
		exams: examsIELTS, display: "full",
	},
	"reading-true-false": {
		name: "True / False / Not Given", style: styleVerdict,
		exams: examsBoth, display: "full", shuffle: true,
		fixedOptions: []questionOption{
			{ID: "TRUE", Text: "True"},
			{ID: "FALSE", Text: "False"},
			{ID: "NOT_GIVEN", Text: "Not Given"},
		},
	},
	"reading-yes-no-not-given": {
		name: "Yes / No / Not Given", style: styleVerdict,
		exams: examsBoth, display: "full", shuffle: true,
		fixedOptions: []questionOption{
			{ID: "YES", Text: "Yes"},
			{ID: "NO", Text: "No"},
			{ID: "NOT_GIVEN", Text: "Not Given"},
		},
	},
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
	// ContextPassage is the gapped text a fill-in-the-blanks question carries,
	// with each gap written as [[b1]], [[b2]] …
	ContextPassage string `json:"contextPassage"`
	// Blanks is the answer key for those gaps, one entry per marker.
	Blanks []newBlank `json:"blanks"`
}

// newBlank is one gap in a gapped text. Options belong to the dropdown kinds;
// a word-bank task leaves them empty and answers from the group's list.
type newBlank struct {
	ID            string   `json:"id"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
}

type newGroup struct {
	TypeID       string `json:"typeId"`
	Instructions string `json:"instructions"`
	// BoxTitle heads the box a summary set is printed in, the way a real paper
	// titles one. Empty for tasks that are not printed in a box.
	BoxTitle         string        `json:"boxTitle"`
	TimeLimitSeconds int           `json:"timeLimitSeconds"`
	Questions        []newQuestion `json:"questions"`
	// WordBank is the draggable list a Reading gap-fill answers from, shared
	// by every question in the group.
	WordBank []string `json:"wordBank"`
	Exams    []string `json:"exams"`
	Boxes    []string `json:"boxes"`
}

func resolveExams(spec readingTypeSpec, wanted []string, fallback string) ([]string, string) {
	chosen := make([]string, 0, 2)
	seen := map[string]bool{}
	for _, raw := range wanted {
		exam := strings.ToUpper(strings.TrimSpace(raw))
		if exam == "" || seen[exam] {
			continue
		}
		if exam != string(models.ExamPTE) && exam != string(models.ExamIELTS) {
			return nil, fmt.Sprintf("%q is not an exam. Use PTE or IELTS.", raw)
		}
		seen[exam] = true
		chosen = append(chosen, exam)
	}
	if len(chosen) == 0 {
		if spec.setsExam(fallback) {
			chosen = []string{fallback}
		} else {
			chosen = spec.exams
		}
	}
	for _, exam := range chosen {
		if !spec.setsExam(exam) {
			return nil, fmt.Sprintf("%s does not set %s.", exam, spec.name)
		}
	}
	ordered := make([]string, 0, len(chosen))
	for _, exam := range examsBoth {
		if seenExam(chosen, exam) {
			ordered = append(ordered, exam)
		}
	}
	return ordered, ""
}

func seenExam(list []string, exam string) bool {
	for _, entry := range list {
		if entry == exam {
			return true
		}
	}
	return false
}

func boxesOf(names []string) ([]resource, map[string]bool, string) {
	boxes := make([]resource, 0, len(names))
	labels := map[string]bool{}
	for _, name := range names {
		text := strings.TrimSpace(name)
		if text == "" {
			continue
		}
		label := paragraphLabel(len(boxes))
		labels[label] = true
		boxes = append(boxes, resource{Label: "Paragraph " + label, Text: text})
	}
	if len(boxes) < 2 {
		return nil, nil, "Write at least two boxes for the learner to order."
	}
	return boxes, labels, ""
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

	// A passage may be created empty and filled with task sets afterwards, which
	// is how the admin screen works: write the text, then add sets to it one at
	// a time. Nothing reaches a learner until it is published.

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

	versions, err := examVersions(ctx, tx)
	if err != nil {
		return "", err
	}

	if _, err := writeGroups(ctx, tx, groupContext{
		passageID:     passageID,
		versionID:     versionID,
		exam:          req.Exam,
		versions:      versions,
		title:         req.Title,
		difficulty:    difficulty,
		publish:       req.Publish,
		labels:        labels,
		startPosition: 1,
	}, req.Groups); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit passage tx: %w", err)
	}
	return passageID, nil
}

// blankMarker matches the [[b1]] gaps a gapped text is written with.
var blankMarker = regexp.MustCompile(`\[\[(b[0-9]+)\]\]`)

// storedBlank is one entry of the per-gap key questions.blanks holds.
type storedBlank struct {
	ID            string   `json:"id"`
	Options       []string `json:"options,omitempty"`
	CorrectAnswer string   `json:"correctAnswer"`
}

// resource is one labelled item belonging to a task rather than to the passage,
// the shape reading_question_groups.resources holds.
type resource struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// wordBankOf turns the author's list into the group's own material, and the
// lookup the answer keys are checked against.
//
// Lettered when the paper prints the list to be chosen from by letter — a
// summary completed from a list of words A-H — and numbered when it is a heap
// of draggable words with no letters on them, which is the PTE gap-fill.
func wordBankOf(words []string, lettered bool) ([]resource, map[string]bool, string) {
	bank := make([]resource, 0, len(words))
	known := map[string]bool{}
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if known[strings.ToLower(word)] {
			return nil, nil, fmt.Sprintf("Word %q is listed twice.", word)
		}
		known[strings.ToLower(word)] = true

		label := fmt.Sprintf("w%d", len(bank)+1)
		if lettered {
			label = paragraphLabel(len(bank))
		}
		bank = append(bank, resource{Label: label, Text: word})
	}
	return bank, known, ""
}

// resolveBlanks checks a gapped text against its answer key.
//
// The rule the grader depends on is that the two match exactly: a key entry
// with no marker grades a gap the learner never saw, and a marker with no key
// can never be answered correctly.
func resolveBlanks(spec readingTypeSpec, q newQuestion, bank map[string]bool) (string, []storedBlank, string) {
	if strings.TrimSpace(q.Prompt) == "" {
		return "", nil, "Write the instruction line for this text."
	}

	text := strings.TrimSpace(q.ContextPassage)
	if text == "" {
		return "", nil, "Paste the gapped text, writing each gap as [[b1]], [[b2]] …"
	}

	markers := blankMarker.FindAllStringSubmatch(text, -1)
	if len(markers) == 0 {
		return "", nil, "Mark at least one gap, written as [[b1]]."
	}

	order := make([]string, 0, len(markers))
	seen := map[string]bool{}
	for _, match := range markers {
		if seen[match[1]] {
			return "", nil, fmt.Sprintf("Gap [[%s]] appears twice in the text.", match[1])
		}
		seen[match[1]] = true
		order = append(order, match[1])
	}

	keyed := make(map[string]newBlank, len(q.Blanks))
	for _, blank := range q.Blanks {
		id := strings.TrimSpace(blank.ID)
		if !seen[id] {
			return "", nil, fmt.Sprintf("There is an answer for [[%s]], but no such gap in the text.", id)
		}
		keyed[id] = blank
	}

	blanks := make([]storedBlank, 0, len(order))
	for _, id := range order {
		blank, ok := keyed[id]
		if !ok {
			return "", nil, fmt.Sprintf("Gap [[%s]] has no answer.", id)
		}
		answer := strings.TrimSpace(blank.CorrectAnswer)
		if answer == "" {
			return "", nil, fmt.Sprintf("Gap [[%s]] has no answer.", id)
		}

		if spec.wordBank {
			if !bank[strings.ToLower(answer)] {
				return "", nil, fmt.Sprintf("Gap [[%s]] is answered %q, which is not in the word list.", id, answer)
			}
			blanks = append(blanks, storedBlank{ID: id, CorrectAnswer: answer})
			continue
		}

		options := make([]string, 0, len(blank.Options))
		known := map[string]bool{}
		for _, option := range blank.Options {
			option = strings.TrimSpace(option)
			if option == "" {
				continue
			}
			if known[option] {
				return "", nil, fmt.Sprintf("Gap [[%s]] lists %q twice.", id, option)
			}
			known[option] = true
			options = append(options, option)
		}
		if len(options) < 2 {
			return "", nil, fmt.Sprintf("Gap [[%s]] needs at least two choices.", id)
		}
		if !known[answer] {
			return "", nil, fmt.Sprintf("Gap [[%s]] is answered %q, which is not one of its choices.", id, answer)
		}
		blanks = append(blanks, storedBlank{ID: id, Options: options, CorrectAnswer: answer})
	}

	return text, blanks, ""
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

	case styleParagraph, styleResource:
		for _, answer := range answers {
			if !labels[strings.ToUpper(answer)] {
				if spec.style == styleResource {
					return nil, nil, fmt.Sprintf("Box %q is not one of this set's boxes.", answer)
				}
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

// ---------------------------------------------------------------------------
// Re-order Paragraphs
// ---------------------------------------------------------------------------

// A re-order item is not a passage task. It carries its own boxes, and the
// order they are written in is the answer key, so it is authored here rather
// than as a group on a passage.
type newReorderItem struct {
	Exam  string `json:"exam"`
	Title string `json:"title"`
	// Boxes in their correct order, which is what makes them the key.
	Boxes            []string `json:"boxes"`
	Prompt           string   `json:"prompt"`
	Explanation      string   `json:"explanation"`
	SourcePassageID  string   `json:"sourcePassageId"`
	Topic            string   `json:"topic"`
	Difficulty       string   `json:"difficulty"`
	Tags             []string `json:"tags"`
	TimeLimitSeconds int      `json:"timeLimitSeconds"`
	Publish          bool     `json:"publish"`
}

// minReorderBoxes mirrors reading_reorder_items_enough_boxes: the scorer counts
// adjacent pairs, so fewer than three boxes is not a task.
const minReorderBoxes = 3

func (i newReorderItem) normalise() ([]paragraph, map[string]string) {
	problems := map[string]string{}

	if i.Exam != string(models.ExamPTE) && i.Exam != string(models.ExamIELTS) {
		problems["exam"] = "Choose PTE or IELTS."
	}
	if strings.TrimSpace(i.Title) == "" {
		problems["title"] = "Give the item a title."
	}
	switch i.Difficulty {
	case "", "easy", "medium", "hard":
	default:
		problems["difficulty"] = "Difficulty must be easy, medium or hard."
	}

	boxes := make([]paragraph, 0, len(i.Boxes))
	for _, box := range i.Boxes {
		text := strings.TrimSpace(box)
		if text == "" {
			continue
		}
		boxes = append(boxes, paragraph{Label: paragraphLabel(len(boxes)), Text: text})
	}
	if len(boxes) < minReorderBoxes {
		problems["boxes"] = fmt.Sprintf("Write at least %d boxes, in their correct order.", minReorderBoxes)
	}

	return boxes, problems
}

func (h *Handler) createReorderItem(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	var req newReorderItem
	if !httpx.DecodeLimit(w, r, &req, h.log, "admin.createReorderItem", 1<<20) {
		return
	}

	boxes, problems := req.normalise()
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	id, err := h.insertReorderItem(r.Context(), req, boxes)
	if err != nil {
		var invalid validationError
		if errors.As(err, &invalid) {
			httpx.ValidationError(w, invalid.fields)
			return
		}
		httpx.Internal(w, h.log, "admin.createReorderItem", err)
		return
	}

	h.log.Info("admin created a re-order item",
		"actor", actor.Email, "item", id, "exam", req.Exam,
		"boxes", len(boxes), "published", req.Publish)

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"item": map[string]any{
			"id":        id,
			"title":     strings.TrimSpace(req.Title),
			"exam":      req.Exam,
			"boxes":     len(boxes),
			"published": req.Publish,
		},
	})
}

func (h *Handler) insertReorderItem(ctx context.Context, req newReorderItem, boxes []paragraph) (string, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin reorder tx: %w", err)
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

	// A source passage is what stops an item being dealt to someone who has
	// already read the text it was cut from, so a wrong id is worth catching
	// here rather than as a foreign key error.
	var sourcePassage *string
	if trimmed := strings.TrimSpace(req.SourcePassageID); trimmed != "" {
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM reading_passages WHERE id = $1)`, trimmed).Scan(&exists); err != nil {
			return "", fmt.Errorf("check source passage: %w", err)
		}
		if !exists {
			return "", validationError{fields: map[string]string{
				"sourcePassageId": "No passage has that id.",
			}}
		}
		sourcePassage = &trimmed
	}

	boxesJSON, err := json.Marshal(boxes)
	if err != nil {
		return "", fmt.Errorf("encode boxes: %w", err)
	}

	options := make([]questionOption, len(boxes))
	order := make([]string, len(boxes))
	for i, box := range boxes {
		options[i] = questionOption{ID: box.Label, Text: box.Text}
		order[i] = box.Label
	}
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("encode boxes as options: %w", err)
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return "", fmt.Errorf("encode box order: %w", err)
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = "medium"
	}
	timeLimit := req.TimeLimitSeconds
	if timeLimit <= 0 {
		timeLimit = 300
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = "The text boxes below have been placed in a random order. Restore the original order."
	}

	itemID := newContentID("ro", req.Title)
	if _, err := tx.Exec(ctx, `
		INSERT INTO reading_reorder_items
			(id, exam_version_id, exam, title, paragraphs, source_passage_id,
			 topic, word_count, difficulty, tags, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		itemID, versionID, req.Exam, strings.TrimSpace(req.Title), boxesJSON, sourcePassage,
		strings.TrimSpace(req.Topic), wordCount(boxes), difficulty,
		normaliseTags(req.Tags), req.Publish,
	); err != nil {
		return "", fmt.Errorf("insert reorder item: %w", err)
	}

	// Points are the number of adjacent pairs, matching the seeded items:
	// scoring.gradeReorder marks each neighbour a learner gets right.
	if _, err := tx.Exec(ctx, `
		INSERT INTO questions
			(id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
			 options, correct_answers, explanation, difficulty, points,
			 time_limit_seconds, is_published, reorder_item_id)
		VALUES ($1, $2, $3, $4, 'reading', 'reorder-paragraphs', 'Re-order Paragraphs', $5, $6,
		        $7, $8, $9, $10, $11, $12, $13, $14)`,
		"q-"+itemID, versionID, req.Exam, []string{req.Exam},
		strings.TrimSpace(req.Title), prompt,
		optionsJSON, orderJSON, strings.TrimSpace(req.Explanation), difficulty,
		len(boxes)-1, timeLimit, req.Publish, itemID,
	); err != nil {
		return "", fmt.Errorf("insert reorder question: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit reorder tx: %w", err)
	}
	return itemID, nil
}

// groupContext is what writing a task set needs to know about the passage it
// hangs on: everything a group and its questions inherit rather than carry.
type groupContext struct {
	passageID     string
	versionID     string
	exam          string
	versions      map[string]string
	title         string
	difficulty    string
	publish       bool
	labels        map[string]bool
	startPosition int
}

func (gc groupContext) versionFor(exam string) string {
	if id, ok := gc.versions[exam]; ok {
		return id
	}
	return gc.versionID
}

func examVersions(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	rows, err := tx.Query(ctx, `SELECT exam, id FROM exam_versions WHERE is_current`)
	if err != nil {
		return nil, fmt.Errorf("load exam versions: %w", err)
	}
	defer rows.Close()

	versions := map[string]string{}
	for rows.Next() {
		var exam, id string
		if err := rows.Scan(&exam, &id); err != nil {
			return nil, fmt.Errorf("scan exam version: %w", err)
		}
		versions[exam] = id
	}
	return versions, rows.Err()
}

// errNoPassage is returned when task sets are addressed to a passage that is
// not there.
var errNoPassage = errors.New("no such passage")

// writeGroups inserts task sets and their questions, and returns how many
// questions it wrote.
//
// Shared by the two ways sets arrive: with a passage as it is created, and
// added to one that already exists.
func writeGroups(ctx context.Context, tx pgx.Tx, gc groupContext, groups []newGroup) (int, error) {
	written := 0
	for gi, group := range groups {
		spec, ok := supportedTypes[group.TypeID]
		if !ok {
			return 0, validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.typeId", gi): "This task type cannot be authored here yet.",
			}}
		}
		if len(group.Questions) == 0 {
			return 0, validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.questions", gi): "Add at least one question.",
			}}
		}

		exams, problem := resolveExams(spec, group.Exams, gc.exam)
		if problem != "" {
			return 0, validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.exams", gi): problem,
			}}
		}

		// A summary completed from a list is chosen from by letter; a PTE
		// gap-fill drags unlettered words.
		bank, bankLookup, problem := wordBankOf(group.WordBank, spec.style == styleText)
		if problem != "" {
			return 0, validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.wordBank", gi): problem,
			}}
		}
		if spec.wordBank && len(bank) == 0 {
			return 0, validationError{fields: map[string]string{
				fmt.Sprintf("groups.%d.wordBank", gi): "List the words that can be dragged into the gaps.",
			}}
		}

		var boxes []resource
		boxLabels := map[string]bool{}
		if spec.style == styleResource {
			boxes, boxLabels, problem = boxesOf(group.Boxes)
			if problem != "" {
				return 0, validationError{fields: map[string]string{
					fmt.Sprintf("groups.%d.boxes", gi): problem,
				}}
			}
		}

		// The list belongs to the set, not to each gap: the paper prints it once
		// above the box. Dropping it was silent before — the author's list
		// vanished and every gap became a typed answer.
		resources := []resource{}
		switch {
		case spec.style == styleResource:
			resources = boxes
		case spec.wordBank || len(bank) > 0:
			resources = bank
		}
		resourcesJSON, err := json.Marshal(resources)
		if err != nil {
			return 0, fmt.Errorf("encode group resources: %w", err)
		}

		// A gap-fill group that rendered the passage would show the learner the
		// ungapped original, which is the answer key in prose.
		display := spec.display
		if display == "" {
			display = "full"
		}

		groupID := newContentID("group", fmt.Sprintf("%s-%d", gc.passageID, gc.startPosition+gi))
		if _, err := tx.Exec(ctx, `
			INSERT INTO reading_question_groups
				(id, passage_id, position, type_id, type_name, instructions, box_title,
				 resources, passage_display, shuffle_questions, time_limit_seconds)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			groupID, gc.passageID, gc.startPosition+gi, group.TypeID, spec.name,
			strings.TrimSpace(group.Instructions), strings.TrimSpace(group.BoxTitle),
			resourcesJSON, display,
			// Authored sets read in the order they were written; shuffling a
			// sentence-completion paragraph, or gapped texts that track the
			// passage top to bottom, would scramble the prose.
			spec.shuffle, group.TimeLimitSeconds,
		); err != nil {
			return 0, fmt.Errorf("insert question group: %w", err)
		}

		for qi, question := range group.Questions {
			var optionsJSON, answersJSON, blanksJSON []byte
			var contextPassage *string

			if spec.style == styleBlanks {
				text, blanks, problem := resolveBlanks(spec, question, bankLookup)
				if problem != "" {
					return 0, validationError{fields: map[string]string{
						fmt.Sprintf("groups.%d.questions.%d", gi, qi): problem,
					}}
				}
				blanksJSON, err = json.Marshal(blanks)
				if err != nil {
					return 0, fmt.Errorf("encode blanks: %w", err)
				}
				contextPassage = &text
			} else {
				// Gaps completed from the set's list offer that list, so the key
				// is checked against it and the learner chooses rather than
				// guessing the exact wording.
				if spec.style == styleText && len(bank) > 0 && len(question.Options) == 0 {
					for _, entry := range bank {
						question.Options = append(question.Options, entry.Text)
					}
				}
				answerLabels := gc.labels
				if spec.style == styleResource {
					answerLabels = boxLabels
				}
				options, answers, problem := resolveAnswers(spec, question, answerLabels)
				if problem != "" {
					return 0, validationError{fields: map[string]string{
						fmt.Sprintf("groups.%d.questions.%d", gi, qi): problem,
					}}
				}

				optionsJSON, err = json.Marshal(options)
				if err != nil {
					return 0, fmt.Errorf("encode options: %w", err)
				}
				answersJSON, err = json.Marshal(answers)
				if err != nil {
					return 0, fmt.Errorf("encode answers: %w", err)
				}
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
					 options, correct_answers, blanks, context_passage, explanation, difficulty, points,
					 time_limit_seconds, is_published, passage_id, group_id, group_position)
				VALUES ($1, $2, $3, $4, 'reading', $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 0, $16, $17, $18, $19)`,
				newContentID("q", fmt.Sprintf("%s-%d", groupID, qi+1)), gc.versionFor(exams[0]), exams[0],
				exams,
				group.TypeID, spec.name, strings.TrimSpace(gc.title), strings.TrimSpace(question.Prompt),
				optionsJSON, answersJSON, blanksJSON, contextPassage,
				strings.TrimSpace(question.Explanation), gc.difficulty, points,
				gc.publish, gc.passageID, groupID, qi+1,
			); err != nil {
				return 0, fmt.Errorf("insert question: %w", err)
			}
			written++
		}
	}

	return written, nil
}

// insertGroups adds task sets to a passage that already exists, continuing its
// numbering rather than restarting it.
//
// Everything a set inherits — the exam version, the difficulty, whether it is
// published, the paragraph labels a matching answer names — is read from the
// passage rather than sent again, so a set cannot contradict the passage it
// hangs on.
func (h *Handler) insertGroups(ctx context.Context, passageID string, groups []newGroup) (int, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin groups tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var gc groupContext
	var paragraphsJSON []byte
	err = tx.QueryRow(ctx, `
		SELECT p.id, p.exam_version_id, v.exam, p.title, p.difficulty, p.is_published, p.paragraphs,
		       coalesce((SELECT max(position) FROM reading_question_groups WHERE passage_id = p.id), 0) + 1
		FROM reading_passages p
		JOIN exam_versions v ON v.id = p.exam_version_id
		WHERE p.id = $1`, passageID).
		Scan(&gc.passageID, &gc.versionID, &gc.exam, &gc.title, &gc.difficulty, &gc.publish,
			&paragraphsJSON, &gc.startPosition)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errNoPassage
	}
	if err != nil {
		return 0, fmt.Errorf("load passage for groups: %w", err)
	}

	var paragraphs []paragraph
	if err := json.Unmarshal(paragraphsJSON, &paragraphs); err != nil {
		return 0, fmt.Errorf("decode paragraphs: %w", err)
	}
	gc.labels = make(map[string]bool, len(paragraphs))
	for _, para := range paragraphs {
		gc.labels[para.Label] = true
	}

	if gc.versions, err = examVersions(ctx, tx); err != nil {
		return 0, err
	}

	written, err := writeGroups(ctx, tx, gc, groups)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit groups tx: %w", err)
	}
	return written, nil
}
