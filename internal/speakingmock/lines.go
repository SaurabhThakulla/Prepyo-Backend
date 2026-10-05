package speakingmock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ExaminerLines is every piece of speech the examiner says for a set, in test
// order and without repeats: each step's lead and its question, the two
// halves the player says (Part 2 says them apart, a minute of preparation in
// between). Stored recordings are made from exactly these texts and found
// again by them, so a line that changes simply goes back to the live voice
// until it is recorded again.
func ExaminerLines(raw []byte) ([]string, error) {
	var content setContent
	if err := json.Unmarshal(raw, &content); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var lines []string
	for _, step := range StepsFor(content) {
		for _, line := range []string{step.Lead, step.Question} {
			line = strings.TrimSpace(line)
			if line != "" && !seen[line] {
				seen[line] = true
				lines = append(lines, line)
			}
		}
	}
	return lines, nil
}

// attachRecordings fills in the stored recording of each step's lead and
// question, where voice_clips has one for that exact text.
func (s *Service) attachRecordings(ctx context.Context, steps []Step) error {
	var texts []string
	for _, step := range steps {
		texts = append(texts, strings.TrimSpace(step.Lead), strings.TrimSpace(step.Question))
	}
	rows, err := s.db.Query(ctx, `SELECT text, asset_id FROM voice_clips WHERE text = ANY($1)`, texts)
	if err != nil {
		return fmt.Errorf("read examiner recordings: %w", err)
	}
	defer rows.Close()
	urls := map[string]string{}
	for rows.Next() {
		var text, asset string
		if err := rows.Scan(&text, &asset); err != nil {
			return err
		}
		urls[text] = "/api/v1/questions/assets/" + asset
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range steps {
		steps[i].LeadAudioURL = urls[strings.TrimSpace(steps[i].Lead)]
		steps[i].QuestionAudioURL = urls[strings.TrimSpace(steps[i].Question)]
	}
	return nil
}
