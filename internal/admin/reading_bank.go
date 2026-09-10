package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// The passage bank an admin works in: what exists, what is on it, and whether
// it has been published.
//
// Authoring is not one sitting. A passage is written first and its task sets
// are added to it afterwards, one type at a time, which is why creating a
// passage no longer demands questions and why the sets have an endpoint of
// their own. Nothing here is visible to a learner until it is published: every
// learner-facing query filters on is_published.

// passageSummary is one row of the list: enough to see what a passage holds
// without loading its text and every question.
type passageSummary struct {
	ID          string   `json:"id"`
	Exam        string   `json:"exam"`
	Exams       []string `json:"exams"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Topic       string   `json:"topic,omitempty"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	WordCount   int      `json:"wordCount"`
	Paragraphs  int      `json:"paragraphs"`
	Groups      int      `json:"groups"`
	Questions   int      `json:"questions"`
	TaskTypes   []string `json:"taskTypes"`
	IsPublished bool     `json:"isPublished"`
	CreatedAt   string   `json:"createdAt"`
}

func (h *Handler) passages(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `
		SELECT p.id, v.exam, p.title, p.subtitle, p.topic, p.difficulty, p.tags,
		       p.word_count, jsonb_array_length(p.paragraphs), p.is_published, p.created_at,
		       count(DISTINCT g.id),
		       count(q.id),
		       coalesce(array_agg(DISTINCT g.type_name) FILTER (WHERE g.id IS NOT NULL), '{}'),
		       coalesce((
		           SELECT array_agg(DISTINCT ex ORDER BY ex)
		           FROM reading_question_groups g2
		           JOIN questions q2 ON q2.group_id = g2.id
		           CROSS JOIN LATERAL unnest(q2.supported_exams) AS ex
		           WHERE g2.passage_id = p.id
		       ), '{}')
		FROM reading_passages p
		JOIN exam_versions v ON v.id = p.exam_version_id
		LEFT JOIN reading_question_groups g ON g.passage_id = p.id
		LEFT JOIN questions q ON q.group_id = g.id
		GROUP BY p.id, v.exam
		ORDER BY p.created_at DESC`)
	if err != nil {
		httpx.Internal(w, h.log, "admin.passages", err)
		return
	}
	defer rows.Close()

	list := []passageSummary{}
	for rows.Next() {
		var p passageSummary
		var created time.Time
		if err := rows.Scan(&p.ID, &p.Exam, &p.Title, &p.Subtitle, &p.Topic, &p.Difficulty,
			&p.Tags, &p.WordCount, &p.Paragraphs, &p.IsPublished, &created,
			&p.Groups, &p.Questions, &p.TaskTypes, &p.Exams); err != nil {
			httpx.Internal(w, h.log, "admin.passages.scan", err)
			return
		}
		p.CreatedAt = created.Format(time.RFC3339)
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "admin.passages.rows", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"passages": list})
}

// passageDetail is the authoring view of one passage: its text, its task sets
// and their answer keys, which the learner-facing endpoint never returns.
type passageDetail struct {
	passageSummary
	Body []paragraph      `json:"body"`
	Sets []map[string]any `json:"sets"`
}

func (h *Handler) passage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var out passageDetail
	var created time.Time
	var paragraphsJSON []byte

	err := h.db.QueryRow(r.Context(), `
		SELECT p.id, v.exam, p.title, p.subtitle, p.topic, p.difficulty, p.tags,
		       p.word_count, p.paragraphs, p.is_published, p.created_at
		FROM reading_passages p
		JOIN exam_versions v ON v.id = p.exam_version_id
		WHERE p.id = $1`, id).
		Scan(&out.ID, &out.Exam, &out.Title, &out.Subtitle, &out.Topic, &out.Difficulty,
			&out.Tags, &out.WordCount, &paragraphsJSON, &out.IsPublished, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No passage has that id.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.passage", err)
		return
	}

	out.CreatedAt = created.Format(time.RFC3339)
	if err := json.Unmarshal(paragraphsJSON, &out.Body); err != nil {
		httpx.Internal(w, h.log, "admin.passage.paragraphs", err)
		return
	}
	out.Paragraphs = len(out.Body)

	rows, err := h.db.Query(r.Context(), `
		SELECT g.id, g.position, g.type_id, g.type_name, g.instructions, g.box_title,
		       g.resources, g.passage_display,
		       coalesce((
		           SELECT array_agg(DISTINCT ex ORDER BY ex)
		           FROM questions q2
		           CROSS JOIN LATERAL unnest(q2.supported_exams) AS ex
		           WHERE q2.group_id = g.id
		       ), '{}'),
		       coalesce(jsonb_agg(jsonb_build_object(
		           'id', q.id,
		           'prompt', q.prompt,
		           'options', q.options,
		           'correctAnswers', q.correct_answers,
		           'blanks', q.blanks,
		           'contextPassage', q.context_passage,
		           'explanation', q.explanation,
		           'points', q.points
		       ) ORDER BY q.group_position) FILTER (WHERE q.id IS NOT NULL), '[]'::jsonb)
		FROM reading_question_groups g
		LEFT JOIN questions q ON q.group_id = g.id
		WHERE g.passage_id = $1
		GROUP BY g.id
		ORDER BY g.position`, id)
	if err != nil {
		httpx.Internal(w, h.log, "admin.passage.sets", err)
		return
	}
	defer rows.Close()

	out.Sets = []map[string]any{}
	out.Exams = []string{}
	for rows.Next() {
		var (
			groupID, typeID, typeName    string
			instructions, boxTitle, disp string
			position                     int
			setExams                     []string
			resources, questions         json.RawMessage
		)
		if err := rows.Scan(&groupID, &position, &typeID, &typeName, &instructions, &boxTitle,
			&resources, &disp, &setExams, &questions); err != nil {
			httpx.Internal(w, h.log, "admin.passage.sets.scan", err)
			return
		}
		for _, exam := range setExams {
			if !seenExam(out.Exams, exam) {
				out.Exams = append(out.Exams, exam)
			}
		}
		out.Sets = append(out.Sets, map[string]any{
			"id":             groupID,
			"position":       position,
			"typeId":         typeID,
			"typeName":       typeName,
			"instructions":   instructions,
			"boxTitle":       boxTitle,
			"exams":          setExams,
			"resources":      resources,
			"passageDisplay": disp,
			"questions":      questions,
		})
		out.Groups++
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "admin.passage.sets.rows", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"passage": out})
}

func (h *Handler) addGroups(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		Groups []newGroup `json:"groups"`
	}
	if !httpx.DecodeLimit(w, r, &req, h.log, "admin.addGroups", 2<<20) {
		return
	}
	if len(req.Groups) == 0 {
		httpx.ValidationError(w, map[string]string{"groups": "Add at least one question set."})
		return
	}

	questions, err := h.insertGroups(r.Context(), id, req.Groups)
	if err != nil {
		var invalid validationError
		if errors.As(err, &invalid) {
			httpx.ValidationError(w, invalid.fields)
			return
		}
		if errors.Is(err, errNoPassage) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No passage has that id.")
			return
		}
		httpx.Internal(w, h.log, "admin.addGroups", err)
		return
	}

	h.log.Info("admin added reading task sets",
		"actor", actor.Email, "passage", id, "sets", len(req.Groups), "questions", questions)

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"passage": map[string]any{"id": id, "setsAdded": len(req.Groups), "questionsAdded": questions},
	})
}

func (h *Handler) setPublished(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	var req struct {
		Publish bool `json:"publish"`
	}
	if !httpx.Decode(w, r, &req, h.log, "admin.setPublished") {
		return
	}

	// Publishing a passage with nothing on it would put an empty paper in front
	// of a learner: practice deals a task set, and there would be none.
	if req.Publish {
		var questions int
		if err := h.db.QueryRow(r.Context(), `
			SELECT count(q.id)
			FROM reading_question_groups g
			JOIN questions q ON q.group_id = g.id
			WHERE g.passage_id = $1`, id).Scan(&questions); err != nil {
			httpx.Internal(w, h.log, "admin.setPublished.count", err)
			return
		}
		if questions == 0 {
			httpx.ValidationError(w, map[string]string{
				"publish": "Add at least one question set before publishing this passage.",
			})
			return
		}
	}

	tag, err := h.db.Exec(r.Context(),
		`UPDATE reading_passages SET is_published = $2 WHERE id = $1`, id, req.Publish)
	if err != nil {
		httpx.Internal(w, h.log, "admin.setPublished", err)
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No passage has that id.")
		return
	}

	// Questions carry their own flag and practice checks both, so a passage
	// published while its questions were not would still deal nothing.
	if _, err := h.db.Exec(r.Context(), `
		UPDATE questions SET is_published = $2
		WHERE group_id IN (SELECT id FROM reading_question_groups WHERE passage_id = $1)`,
		id, req.Publish); err != nil {
		httpx.Internal(w, h.log, "admin.setPublished.questions", err)
		return
	}

	h.log.Info("admin changed a passage's publication",
		"actor", actor.Email, "passage", id, "published", req.Publish)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"passage": map[string]any{"id": id, "isPublished": req.Publish},
	})
}

// deletePassage removes a passage, its task sets and their questions.
//
// practice_attempts and mistakes reference questions without a cascade, so a
// passage a learner has answered cannot simply be dropped. Their answers go
// first, in the same transaction, or the delete fails halfway.
func (h *Handler) deletePassage(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	id := chi.URLParam(r, "id")

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "admin.deletePassage.begin", err)
		return
	}
	defer tx.Rollback(r.Context())

	var exists bool
	if err := tx.QueryRow(r.Context(),
		`SELECT EXISTS (SELECT 1 FROM reading_passages WHERE id = $1)`, id).Scan(&exists); err != nil {
		httpx.Internal(w, h.log, "admin.deletePassage.exists", err)
		return
	}
	if !exists {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No passage has that id.")
		return
	}

	const doomed = `
		SELECT q.id FROM questions q
		JOIN reading_question_groups g ON g.id = q.group_id
		WHERE g.passage_id = $1`

	for _, statement := range []string{
		`DELETE FROM ai_evaluations WHERE question_id IN (` + doomed + `)`,
		`DELETE FROM mistakes WHERE question_id IN (` + doomed + `)`,
		`DELETE FROM practice_attempts WHERE question_id IN (` + doomed + `)`,
	} {
		if _, err := tx.Exec(r.Context(), statement, id); err != nil {
			httpx.Internal(w, h.log, "admin.deletePassage.answers", err)
			return
		}
	}

	// Groups, questions and everyone's exposure rows cascade from the passage.
	if _, err := tx.Exec(r.Context(), `DELETE FROM reading_passages WHERE id = $1`, id); err != nil {
		httpx.Internal(w, h.log, "admin.deletePassage", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httpx.Internal(w, h.log, "admin.deletePassage.commit", err)
		return
	}

	h.log.Info("admin deleted a reading passage", "actor", actor.Email, "passage", id)
	httpx.JSON(w, http.StatusOK, map[string]any{"deleted": id})
}
