package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type fieldUse string

const (
	fieldUnused   fieldUse = ""
	fieldOptional fieldUse = "optional"
	fieldRequired fieldUse = "required"
)

type answerKind string

const (
	answerRubric    answerKind = "rubric"
	answerSingle    answerKind = "single"
	answerMultiple  answerKind = "multiple"
	answerBlanks    answerKind = "blanks"
	answerDictation answerKind = "dictation"
)

type questionTypeSpec struct {
	Exam             string     `json:"exam"`
	Skill            string     `json:"skill"`
	TypeID           string     `json:"typeId"`
	TypeName         string     `json:"typeName"`
	Answer           answerKind `json:"answer"`
	Passage          fieldUse   `json:"passage"`
	Audio            fieldUse   `json:"audio"`
	Image            fieldUse   `json:"image"`
	Figure           fieldUse   `json:"figure"`
	Prompt           string     `json:"prompt"`
	PrepSeconds      int        `json:"prepSeconds"`
	TimeLimitSeconds int        `json:"timeLimitSeconds"`
	Points           int        `json:"points"`
}

var authorableTypes = []questionTypeSpec{
	{Exam: "IELTS", Skill: "writing", TypeID: "ielts-writing-task1-figure", TypeName: "Describe the figure",
		Answer: answerRubric, Image: fieldRequired, Figure: fieldRequired,
		Prompt:           "Summarise the information by selecting and reporting the main features, and make comparisons where relevant.\n\nWrite at least 150 words.",
		TimeLimitSeconds: 1200, Points: 20},
	{Exam: "IELTS", Skill: "writing", TypeID: "ielts-writing-task2-opinion", TypeName: "Opinion / Agree or Disagree",
		Answer:           answerRubric,
		Prompt:           "To what extent do you agree or disagree? Give reasons for your answer and include any relevant examples from your own knowledge or experience.\n\nWrite at least 250 words.",
		TimeLimitSeconds: 2400, Points: 25},
	{Exam: "PTE", Skill: "writing", TypeID: "summarize-written-text", TypeName: "Summarize Written Text",
		Answer: answerRubric, Passage: fieldRequired,
		Prompt:           "Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.",
		TimeLimitSeconds: 600, Points: 10},
	{Exam: "PTE", Skill: "writing", TypeID: "pte-write-essay", TypeName: "Write Essay",
		Answer:           answerRubric,
		Prompt:           "Do you agree or disagree with this statement? Support your point of view with reasons and examples from your own experience or observation.\n\nWrite 200-300 words.",
		TimeLimitSeconds: 1200, Points: 15},

	{Exam: "IELTS", Skill: "speaking", TypeID: "ielts-speaking-part1", TypeName: "Introduction",
		Answer: answerRubric, TimeLimitSeconds: 60, Points: 10},
	{Exam: "IELTS", Skill: "speaking", TypeID: "ielts-speaking-part2", TypeName: "Speaking Part 2 (Cue Card)",
		Answer: answerRubric, PrepSeconds: 60, TimeLimitSeconds: 120, Points: 15},
	{Exam: "PTE", Skill: "speaking", TypeID: "read-aloud", TypeName: "Read Aloud",
		Answer: answerRubric, Passage: fieldRequired,
		Prompt:      "Look at the text below. In 35 seconds, read this text aloud as naturally and clearly as possible.",
		PrepSeconds: 35, TimeLimitSeconds: 40, Points: 15},
	{Exam: "PTE", Skill: "speaking", TypeID: "pte-repeat-sentence", TypeName: "Repeat Sentence",
		Answer: answerRubric, Audio: fieldRequired,
		Prompt:           "You will hear a sentence. Repeat the sentence exactly as you hear it.",
		TimeLimitSeconds: 15, Points: 10},
	{Exam: "PTE", Skill: "speaking", TypeID: "pte-describe-image", TypeName: "Describe Image",
		Answer: answerRubric, Image: fieldRequired,
		Prompt:      "Look at the image below. In 25 seconds, describe in detail what the image is showing.",
		PrepSeconds: 25, TimeLimitSeconds: 40, Points: 15},
	{Exam: "PTE", Skill: "speaking", TypeID: "pte-retell-lecture", TypeName: "Re-tell Lecture",
		Answer: answerRubric, Audio: fieldRequired,
		Prompt:      "You will hear a lecture. After listening, retell what you have just heard in your own words.",
		PrepSeconds: 10, TimeLimitSeconds: 40, Points: 15},
	{Exam: "PTE", Skill: "speaking", TypeID: "pte-answer-short-question", TypeName: "Answer Short Question",
		Answer: answerRubric, Audio: fieldRequired,
		Prompt:           "You will hear a question. Give a simple and short answer. Often just one or a few words is enough.",
		TimeLimitSeconds: 10, Points: 5},

	{Exam: "IELTS", Skill: "listening", TypeID: "ielts-listening-fill-blanks", TypeName: "Fill in the blanks",
		Answer: answerBlanks, Audio: fieldRequired,
		Prompt:           "Complete the notes below. Write NO MORE THAN TWO WORDS AND/OR A NUMBER for each answer.",
		TimeLimitSeconds: 300, Points: 10},
	{Exam: "IELTS", Skill: "listening", TypeID: "ielts-listening-mcq", TypeName: "Choose the correct answer",
		Answer: answerSingle, Audio: fieldRequired,
		Prompt:           "Choose the correct letter, A, B or C.",
		TimeLimitSeconds: 120, Points: 10},
	{Exam: "IELTS", Skill: "listening", TypeID: "ielts-listening-map", TypeName: "Navigating according to audio",
		Answer: answerSingle, Audio: fieldRequired, Image: fieldOptional,
		Prompt:           "Listen to the directions and choose the correct place on the map.",
		TimeLimitSeconds: 180, Points: 10},
	{Exam: "PTE", Skill: "listening", TypeID: "pte-listening-mcma", TypeName: "Multiple Choice, Multiple Answers",
		Answer: answerMultiple, Audio: fieldRequired,
		Prompt:           "Listen to the recording and answer the question by selecting all the correct responses. You will need to select more than one response.",
		TimeLimitSeconds: 120, Points: 10},
	{Exam: "PTE", Skill: "listening", TypeID: "pte-listening-fib", TypeName: "Fill in the Blanks",
		Answer: answerBlanks, Audio: fieldRequired,
		Prompt:           "You will hear a recording. Type the missing word in each blank.",
		TimeLimitSeconds: 120, Points: 10},
	{Exam: "PTE", Skill: "listening", TypeID: "pte-highlight-correct-summary", TypeName: "Highlight Correct Summary",
		Answer: answerSingle, Audio: fieldRequired,
		Prompt:           "You will hear a recording. Choose the paragraph that best summarises it.",
		TimeLimitSeconds: 120, Points: 10},
	{Exam: "PTE", Skill: "listening", TypeID: "pte-select-missing-word", TypeName: "Select Missing Word",
		Answer: answerSingle, Audio: fieldRequired,
		Prompt:           "You will hear a recording. The last word or group of words has been replaced by a beep. Choose the option that completes it.",
		TimeLimitSeconds: 90, Points: 10},
	{Exam: "PTE", Skill: "listening", TypeID: "write-from-dictation", TypeName: "Write from Dictation",
		Answer: answerDictation, Audio: fieldRequired,
		Prompt:           "You will hear a sentence. Type the sentence exactly as you hear it.",
		TimeLimitSeconds: 60, Points: 10},
}

var authorableByID = func() map[string]questionTypeSpec {
	byID := make(map[string]questionTypeSpec, len(authorableTypes))
	for _, spec := range authorableTypes {
		byID[spec.TypeID] = spec
	}
	return byID
}()

var questionIDPrefix = map[string]string{"speaking": "spk", "writing": "wrt", "listening": "lis"}

const maxChoiceOptions = 10

func authorableSkill(skill string) bool {
	_, ok := questionIDPrefix[skill]
	return ok
}

const authoredScope = `skill IN ('speaking', 'writing', 'listening') AND passage_id IS NULL AND reorder_item_id IS NULL`

type newAuthoredQuestion struct {
	Exam             string     `json:"exam"`
	TypeID           string     `json:"typeId"`
	Title            string     `json:"title"`
	Prompt           string     `json:"prompt"`
	ContextPassage   string     `json:"contextPassage"`
	AudioURL         string     `json:"audioUrl"`
	AudioTranscript  string     `json:"audioTranscript"`
	ImageURL         string     `json:"imageUrl"`
	FigureData       string     `json:"figureData"`
	Options          []string   `json:"options"`
	CorrectAnswers   []string   `json:"correctAnswers"`
	Blanks           []newBlank `json:"blanks"`
	ModelAnswer      string     `json:"modelAnswer"`
	Explanation      string     `json:"explanation"`
	Difficulty       string     `json:"difficulty"`
	Tags             []string   `json:"tags"`
	PrepTimeSeconds  int        `json:"prepTimeSeconds"`
	TimeLimitSeconds int        `json:"timeLimitSeconds"`
	Points           int        `json:"points"`
	Publish          bool       `json:"publish"`
}

type authoredQuestion struct {
	spec             questionTypeSpec
	title            string
	prompt           string
	contextPassage   string
	audioURL         string
	audioTranscript  string
	imageURL         string
	figureData       string
	options          []questionOption
	correctAnswers   []string
	blanks           []storedBlank
	modelAnswer      string
	explanation      string
	difficulty       string
	tags             []string
	prepSeconds      int
	timeLimitSeconds int
	points           int
	publish          bool
}

func (req newAuthoredQuestion) normalise() (authoredQuestion, map[string]string) {
	problems := map[string]string{}

	spec, ok := authorableByID[req.TypeID]
	if !ok {
		problems["typeId"] = "Pick a task type."
		return authoredQuestion{}, problems
	}
	if req.Exam != spec.Exam {
		problems["exam"] = fmt.Sprintf("%s is a %s task.", spec.TypeName, spec.Exam)
	}

	q := authoredQuestion{
		spec:             spec,
		title:            strings.TrimSpace(req.Title),
		prompt:           strings.TrimSpace(req.Prompt),
		contextPassage:   kept(spec.Passage, req.ContextPassage),
		audioURL:         kept(spec.Audio, req.AudioURL),
		audioTranscript:  kept(spec.Audio, req.AudioTranscript),
		imageURL:         kept(spec.Image, req.ImageURL),
		figureData:       kept(spec.Figure, req.FigureData),
		explanation:      strings.TrimSpace(req.Explanation),
		difficulty:       req.Difficulty,
		prepSeconds:      orDefault(req.PrepTimeSeconds, spec.PrepSeconds),
		timeLimitSeconds: orDefault(req.TimeLimitSeconds, spec.TimeLimitSeconds),
		points:           orDefault(req.Points, spec.Points),
		publish:          req.Publish,
	}
	if spec.Skill != "speaking" {
		q.prepSeconds = 0
	}

	if q.title == "" {
		problems["title"] = "Give the question a title."
	}
	if q.prompt == "" {
		q.prompt = spec.Prompt
	}
	if q.prompt == "" {
		problems["prompt"] = "Write the instructions the learner sees."
	}

	switch q.difficulty {
	case "":
		q.difficulty = "medium"
	case "easy", "medium", "hard":
	default:
		problems["difficulty"] = "Difficulty must be easy, medium or hard."
	}

	if spec.Passage == fieldRequired && q.contextPassage == "" {
		problems["contextPassage"] = "Write the text the learner works from."
	}
	if spec.Audio == fieldRequired && q.audioTranscript == "" && q.audioURL == "" {
		problems["audioTranscript"] = "Write the script, or give a recording URL."
	}
	if spec.Image == fieldRequired && q.imageURL == "" {
		problems["imageUrl"] = "Give the image URL."
	}
	if spec.Figure == fieldRequired && q.figureData == "" {
		problems["figureData"] = "Write out what the figure shows, so the evaluator can check the answer against it."
	}
	if q.audioURL != "" && !isWebURL(q.audioURL) {
		problems["audioUrl"] = "Use a full http:// or https:// link."
	}
	if q.imageURL != "" && !isWebURL(q.imageURL) {
		problems["imageUrl"] = "Use a full http:// or https:// link."
	}

	switch spec.Answer {
	case answerRubric:
		q.modelAnswer = strings.TrimSpace(req.ModelAnswer)
	case answerSingle, answerMultiple:
		options, correct, choiceProblems := choiceKey(req.Options, req.CorrectAnswers, spec.Answer == answerMultiple)
		q.options, q.correctAnswers = options, correct
		for field, problem := range choiceProblems {
			problems[field] = problem
		}
	case answerBlanks:
		blanks, problem := blanksKey(req.Blanks)
		if problem != "" {
			problems["blanks"] = problem
		}
		q.blanks = blanks
	case answerDictation:
		if q.audioTranscript == "" {
			problems["audioTranscript"] = "Write the sentence the learner hears. It is also the answer."
		} else {
			q.correctAnswers = []string{q.audioTranscript}
		}
	}

	if q.timeLimitSeconds > 7200 {
		problems["timeLimitSeconds"] = "Keep the time limit under two hours."
	}
	if q.prepSeconds > 600 {
		problems["prepTimeSeconds"] = "Keep preparation under ten minutes."
	}
	if q.points > 100 {
		problems["points"] = "A question is worth at most 100 points."
	}

	q.tags = normaliseTags(req.Tags)
	if len(q.tags) == 0 {
		q.tags = normaliseTags([]string{
			spec.Exam + " " + strings.ToUpper(spec.Skill[:1]) + spec.Skill[1:],
			spec.TypeName,
		})
	}

	return q, problems
}

func choiceKey(texts, answers []string, multi bool) ([]questionOption, []string, map[string]string) {
	problems := map[string]string{}

	options := []questionOption{}
	idByText := map[string]string{}
	for _, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		key := strings.ToLower(text)
		if _, dup := idByText[key]; dup {
			problems["options"] = "Two options have the same wording."
			continue
		}
		id := string(rune('A' + len(options)))
		idByText[key] = id
		options = append(options, questionOption{ID: id, Text: text})
	}
	switch {
	case len(options) < 2:
		problems["options"] = "Write at least two options."
	case len(options) > maxChoiceOptions:
		problems["options"] = fmt.Sprintf("Keep it to %d options.", maxChoiceOptions)
	}

	correct := []string{}
	seen := map[string]bool{}
	for _, answer := range answers {
		answer = strings.TrimSpace(answer)
		if answer == "" {
			continue
		}
		id, ok := idByText[strings.ToLower(answer)]
		if !ok {
			problems["correctAnswers"] = "Every correct answer must be one of the options."
			continue
		}
		if !seen[id] {
			seen[id] = true
			correct = append(correct, id)
		}
	}
	switch {
	case len(correct) == 0:
		problems["correctAnswers"] = "Mark the correct option."
	case !multi && len(correct) > 1:
		problems["correctAnswers"] = "This task has exactly one correct option."
	case multi && len(correct) == len(options):
		problems["correctAnswers"] = "At least one option must be wrong."
	}

	return options, correct, problems
}

func blanksKey(blanks []newBlank) ([]storedBlank, string) {
	stored := []storedBlank{}
	for _, blank := range blanks {
		answer := strings.TrimSpace(blank.CorrectAnswer)
		choices := []string{}
		for _, choice := range blank.Options {
			if choice = strings.TrimSpace(choice); choice != "" {
				choices = append(choices, choice)
			}
		}
		if answer == "" && len(choices) == 0 {
			continue
		}

		number := len(stored) + 1
		if answer == "" {
			return nil, fmt.Sprintf("Blank %d has no answer.", number)
		}
		if len(choices) > 0 {
			match := ""
			for _, choice := range choices {
				if strings.EqualFold(choice, answer) {
					match = choice
					break
				}
			}
			if match == "" {
				return nil, fmt.Sprintf("The answer to blank %d is not one of its choices.", number)
			}
			answer = match
		}

		stored = append(stored, storedBlank{
			ID:            fmt.Sprintf("b%d", number),
			Options:       choices,
			CorrectAnswer: answer,
		})
	}
	if len(stored) == 0 {
		return nil, "Add at least one blank with its answer."
	}
	return stored, ""
}

func kept(use fieldUse, value string) string {
	if use == fieldUnused {
		return ""
	}
	return strings.TrimSpace(value)
}

func orDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func isWebURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func textOrNull(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func jsonOrNull[T any](items []T) (any, error) {
	if len(items) == 0 {
		return nil, nil
	}
	return json.Marshal(items)
}

func (h *Handler) questionTypes(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"types": authorableTypes})
}

type authoredQuestionSummary struct {
	ID          string    `json:"id"`
	Exam        string    `json:"exam"`
	TypeID      string    `json:"typeId"`
	TypeName    string    `json:"typeName"`
	Title       string    `json:"title"`
	Prompt      string    `json:"prompt"`
	Difficulty  string    `json:"difficulty"`
	Points      int       `json:"points"`
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
	Answers     int       `json:"answers"`
	InMock      bool      `json:"inMock"`
}

func (h *Handler) authoredQuestions(w http.ResponseWriter, r *http.Request) {
	skill := r.URL.Query().Get("skill")
	if !authorableSkill(skill) {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Skill must be speaking, writing or listening.")
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT q.id, q.exam, q.type_id, q.type_name, q.title, q.prompt, q.difficulty, q.points,
		       q.is_published, q.created_at,
		       (SELECT count(*) FROM practice_attempts pa WHERE pa.question_id = q.id)
		         + (SELECT count(*) FROM ai_evaluations e WHERE e.question_id = q.id),
		       EXISTS (SELECT 1 FROM mock_sections ms WHERE q.id = ANY(ms.question_ids))
		FROM questions q
		WHERE q.skill = $1 AND q.passage_id IS NULL AND q.reorder_item_id IS NULL
		ORDER BY q.created_at DESC, q.id`, skill)
	if err != nil {
		httpx.Internal(w, h.log, "admin.authoredQuestions", err)
		return
	}
	defer rows.Close()

	list := []authoredQuestionSummary{}
	for rows.Next() {
		var q authoredQuestionSummary
		if err := rows.Scan(&q.ID, &q.Exam, &q.TypeID, &q.TypeName, &q.Title, &q.Prompt, &q.Difficulty,
			&q.Points, &q.IsPublished, &q.CreatedAt, &q.Answers, &q.InMock); err != nil {
			httpx.Internal(w, h.log, "admin.authoredQuestions.scan", err)
			return
		}
		list = append(list, q)
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "admin.authoredQuestions.rows", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"questions": list})
}

func (h *Handler) createQuestion(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	var req newAuthoredQuestion
	if !httpx.DecodeLimit(w, r, &req, h.log, "admin.createQuestion", 1<<20) {
		return
	}

	q, problems := req.normalise()
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	id, err := h.insertAuthoredQuestion(r.Context(), q)
	if err != nil {
		var invalid validationError
		if errors.As(err, &invalid) {
			httpx.ValidationError(w, invalid.fields)
			return
		}
		httpx.Internal(w, h.log, "admin.createQuestion", err)
		return
	}

	h.log.Info("admin created a question",
		"actor", actor.Email, "question", id, "exam", q.spec.Exam,
		"type", q.spec.TypeID, "published", q.publish)

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"question": map[string]any{
			"id":          id,
			"exam":        q.spec.Exam,
			"typeName":    q.spec.TypeName,
			"title":       q.title,
			"isPublished": q.publish,
		},
	})
}

func (h *Handler) insertAuthoredQuestion(ctx context.Context, q authoredQuestion) (string, error) {
	var versionID string
	err := h.db.QueryRow(ctx,
		`SELECT id FROM exam_versions WHERE exam = $1 AND is_current`, q.spec.Exam).Scan(&versionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", validationError{fields: map[string]string{
			"exam": "That exam has no current version configured.",
		}}
	}
	if err != nil {
		return "", fmt.Errorf("find current exam version: %w", err)
	}

	options, err := jsonOrNull(q.options)
	if err != nil {
		return "", fmt.Errorf("encode options: %w", err)
	}
	correct, err := jsonOrNull(q.correctAnswers)
	if err != nil {
		return "", fmt.Errorf("encode answers: %w", err)
	}
	blanks, err := jsonOrNull(q.blanks)
	if err != nil {
		return "", fmt.Errorf("encode blanks: %w", err)
	}

	id := newContentID(questionIDPrefix[q.spec.Skill], q.title)
	if _, err := h.db.Exec(ctx, `
		INSERT INTO questions
			(id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
			 context_passage, audio_url, audio_transcript, image_url, figure_data,
			 prep_time_seconds, time_limit_seconds, options, correct_answers, blanks,
			 model_answer, explanation, difficulty, tags, points, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		        $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)`,
		id, versionID, q.spec.Exam, []string{q.spec.Exam}, q.spec.Skill, q.spec.TypeID, q.spec.TypeName,
		q.title, q.prompt,
		textOrNull(q.contextPassage), textOrNull(q.audioURL), textOrNull(q.audioTranscript),
		textOrNull(q.imageURL), textOrNull(q.figureData),
		q.prepSeconds, q.timeLimitSeconds, options, correct, blanks,
		textOrNull(q.modelAnswer), textOrNull(q.explanation), q.difficulty, q.tags, q.points, q.publish,
	); err != nil {
		return "", fmt.Errorf("insert question: %w", err)
	}
	return id, nil
}

func (h *Handler) setQuestionPublished(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		Publish bool `json:"publish"`
	}
	if !httpx.Decode(w, r, &req, h.log, "admin.setQuestionPublished") {
		return
	}

	tag, err := h.db.Exec(r.Context(),
		`UPDATE questions SET is_published = $2 WHERE id = $1 AND `+authoredScope, id, req.Publish)
	if err != nil {
		httpx.Internal(w, h.log, "admin.setQuestionPublished", err)
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No authored question has that id.")
		return
	}

	h.log.Info("admin changed a question's visibility", "actor", actor.Email, "question", id, "published", req.Publish)
	httpx.JSON(w, http.StatusOK, map[string]any{"id": id, "isPublished": req.Publish})
}

func (h *Handler) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "admin.deleteQuestion.begin", err)
		return
	}
	defer tx.Rollback(r.Context())

	var exists, inMock bool
	if err := tx.QueryRow(r.Context(), `
		SELECT EXISTS (SELECT 1 FROM questions WHERE id = $1 AND `+authoredScope+`),
		       EXISTS (SELECT 1 FROM mock_sections WHERE $1 = ANY(question_ids))`, id).Scan(&exists, &inMock); err != nil {
		httpx.Internal(w, h.log, "admin.deleteQuestion.exists", err)
		return
	}
	if !exists {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No authored question has that id.")
		return
	}
	if inMock {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"This question is part of a mock exam. Unpublish it instead of deleting it.")
		return
	}

	for _, statement := range []string{
		`DELETE FROM ai_evaluations WHERE question_id = $1`,
		`DELETE FROM mistakes WHERE question_id = $1`,
		`DELETE FROM practice_attempts WHERE question_id = $1`,
		`DELETE FROM questions WHERE id = $1`,
	} {
		if _, err := tx.Exec(r.Context(), statement, id); err != nil {
			httpx.Internal(w, h.log, "admin.deleteQuestion", err)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		httpx.Internal(w, h.log, "admin.deleteQuestion.commit", err)
		return
	}

	h.log.Info("admin deleted a question", "actor", actor.Email, "question", id)
	httpx.JSON(w, http.StatusOK, map[string]any{"deleted": id})
}

type authoredQuestionDetail struct {
	ID               string     `json:"id"`
	Exam             string     `json:"exam"`
	Skill            string     `json:"skill"`
	TypeID           string     `json:"typeId"`
	TypeName         string     `json:"typeName"`
	Title            string     `json:"title"`
	Prompt           string     `json:"prompt"`
	ContextPassage   string     `json:"contextPassage"`
	AudioURL         string     `json:"audioUrl"`
	AudioTranscript  string     `json:"audioTranscript"`
	ImageURL         string     `json:"imageUrl"`
	FigureData       string     `json:"figureData"`
	Options          []string   `json:"options"`
	CorrectAnswers   []string   `json:"correctAnswers"`
	Blanks           []newBlank `json:"blanks"`
	ModelAnswer      string     `json:"modelAnswer"`
	Explanation      string     `json:"explanation"`
	Difficulty       string     `json:"difficulty"`
	Tags             []string   `json:"tags"`
	PrepTimeSeconds  int        `json:"prepTimeSeconds"`
	TimeLimitSeconds int        `json:"timeLimitSeconds"`
	Points           int        `json:"points"`
	IsPublished      bool       `json:"isPublished"`
	InMock           bool       `json:"inMock"`
}

func (h *Handler) authoredQuestion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var (
		q                      authoredQuestionDetail
		optionsRaw, correctRaw []byte
		blanksRaw              []byte
	)
	err := h.db.QueryRow(r.Context(), `
		SELECT q.id, q.exam, q.skill, q.type_id, q.type_name, q.title, q.prompt,
		       coalesce(q.context_passage, ''), coalesce(q.audio_url, ''), coalesce(q.audio_transcript, ''),
		       coalesce(q.image_url, ''), coalesce(q.figure_data, ''),
		       q.options, q.correct_answers, q.blanks,
		       coalesce(q.model_answer, ''), coalesce(q.explanation, ''), q.difficulty, q.tags,
		       q.prep_time_seconds, q.time_limit_seconds, q.points, q.is_published,
		       EXISTS (SELECT 1 FROM mock_sections ms WHERE q.id = ANY(ms.question_ids))
		FROM questions q
		WHERE q.id = $1 AND `+authoredScope, id).
		Scan(&q.ID, &q.Exam, &q.Skill, &q.TypeID, &q.TypeName, &q.Title, &q.Prompt,
			&q.ContextPassage, &q.AudioURL, &q.AudioTranscript, &q.ImageURL, &q.FigureData,
			&optionsRaw, &correctRaw, &blanksRaw,
			&q.ModelAnswer, &q.Explanation, &q.Difficulty, &q.Tags,
			&q.PrepTimeSeconds, &q.TimeLimitSeconds, &q.Points, &q.IsPublished, &q.InMock)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No authored question has that id.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.authoredQuestion", err)
		return
	}

	stored := []questionOption{}
	if len(optionsRaw) > 0 {
		if err := json.Unmarshal(optionsRaw, &stored); err != nil {
			httpx.Internal(w, h.log, "admin.authoredQuestion.options", err)
			return
		}
	}
	textByID := make(map[string]string, len(stored))
	q.Options = []string{}
	for _, option := range stored {
		textByID[option.ID] = option.Text
		q.Options = append(q.Options, option.Text)
	}

	answerIDs := []string{}
	if len(correctRaw) > 0 {
		if err := json.Unmarshal(correctRaw, &answerIDs); err != nil {
			httpx.Internal(w, h.log, "admin.authoredQuestion.answers", err)
			return
		}
	}
	q.CorrectAnswers = []string{}
	for _, answer := range answerIDs {
		if text, ok := textByID[answer]; ok {
			q.CorrectAnswers = append(q.CorrectAnswers, text)
			continue
		}
		q.CorrectAnswers = append(q.CorrectAnswers, answer)
	}

	blanks := []storedBlank{}
	if len(blanksRaw) > 0 {
		if err := json.Unmarshal(blanksRaw, &blanks); err != nil {
			httpx.Internal(w, h.log, "admin.authoredQuestion.blanks", err)
			return
		}
	}
	q.Blanks = []newBlank{}
	for _, blank := range blanks {
		q.Blanks = append(q.Blanks, newBlank{
			ID:            blank.ID,
			Options:       blank.Options,
			CorrectAnswer: blank.CorrectAnswer,
		})
	}
	if q.Tags == nil {
		q.Tags = []string{}
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"question": q})
}

func (h *Handler) updateQuestion(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	var req newAuthoredQuestion
	if !httpx.DecodeLimit(w, r, &req, h.log, "admin.updateQuestion", 1<<20) {
		return
	}

	q, problems := req.normalise()
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	if err := h.saveAuthoredQuestion(r.Context(), id, q); err != nil {
		var invalid validationError
		if errors.As(err, &invalid) {
			httpx.ValidationError(w, invalid.fields)
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No authored question has that id.")
			return
		}
		httpx.Internal(w, h.log, "admin.updateQuestion", err)
		return
	}

	h.log.Info("admin edited a question",
		"actor", actor.Email, "question", id, "exam", q.spec.Exam,
		"type", q.spec.TypeID, "published", q.publish)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"question": map[string]any{
			"id":          id,
			"exam":        q.spec.Exam,
			"typeName":    q.spec.TypeName,
			"title":       q.title,
			"isPublished": q.publish,
		},
	})
}

func (h *Handler) saveAuthoredQuestion(ctx context.Context, id string, q authoredQuestion) error {
	var versionID string
	err := h.db.QueryRow(ctx,
		`SELECT id FROM exam_versions WHERE exam = $1 AND is_current`, q.spec.Exam).Scan(&versionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return validationError{fields: map[string]string{
			"exam": "That exam has no current version configured.",
		}}
	}
	if err != nil {
		return fmt.Errorf("find current exam version: %w", err)
	}

	options, err := jsonOrNull(q.options)
	if err != nil {
		return fmt.Errorf("encode options: %w", err)
	}
	correct, err := jsonOrNull(q.correctAnswers)
	if err != nil {
		return fmt.Errorf("encode answers: %w", err)
	}
	blanks, err := jsonOrNull(q.blanks)
	if err != nil {
		return fmt.Errorf("encode blanks: %w", err)
	}

	tag, err := h.db.Exec(ctx, `
		UPDATE questions SET
			exam_version_id = $2, exam = $3, supported_exams = $4, skill = $5, type_id = $6, type_name = $7,
			title = $8, prompt = $9, context_passage = $10, audio_url = $11, audio_transcript = $12,
			image_url = $13, figure_data = $14, prep_time_seconds = $15, time_limit_seconds = $16,
			options = $17, correct_answers = $18, blanks = $19, model_answer = $20, explanation = $21,
			difficulty = $22, tags = $23, points = $24, is_published = $25
		WHERE id = $1 AND `+authoredScope,
		id, versionID, q.spec.Exam, []string{q.spec.Exam}, q.spec.Skill, q.spec.TypeID, q.spec.TypeName,
		q.title, q.prompt,
		textOrNull(q.contextPassage), textOrNull(q.audioURL), textOrNull(q.audioTranscript),
		textOrNull(q.imageURL), textOrNull(q.figureData),
		q.prepSeconds, q.timeLimitSeconds, options, correct, blanks,
		textOrNull(q.modelAnswer), textOrNull(q.explanation), q.difficulty, q.tags, q.points, q.publish,
	)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
