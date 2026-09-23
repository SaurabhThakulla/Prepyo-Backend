-- The IELTS Speaking mock: all three parts in one sitting.
--
-- Why this exists. Speaking could only be practised one question at a time,
-- and a single answer is not a Speaking band: the test runs 11-14 minutes,
-- Part 3 discusses the topic of the Part 2 cue card, and "a candidate will be
-- rated on their average performance across all parts of the test". A set is
-- one complete test: two Part 1 topics, a cue card with its follow-up
-- question, and Part 3 questions on the cue card's theme. All sets are
-- original Prepyo practice content, not official or recalled test material.

CREATE TABLE speaking_mock_sets (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    -- {"part1":[{"topic":..,"questions":[..]}],"part2":{"topic":..,"cue":..,"points":[..],"rounding":..},
    --  "part3":{"topic":..,"questions":[..]}}
    content      JSONB NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE speaking_mock_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    set_id          TEXT NOT NULL REFERENCES speaking_mock_sets(id),
    status          TEXT NOT NULL DEFAULT 'in_progress'
                        CHECK (status IN ('in_progress', 'submitted', 'abandoned')),
    -- One entry per answered step: {"key","transcript","durationSeconds","source"}.
    answers         JSONB NOT NULL DEFAULT '[]'::jsonb,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at    TIMESTAMPTZ,
    mock_attempt_id UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    evaluation_id   UUID REFERENCES ai_evaluations(id) ON DELETE SET NULL,
    CONSTRAINT speaking_mock_answers_array CHECK (jsonb_typeof(answers) = 'array')
);

CREATE UNIQUE INDEX idx_speaking_mock_live
    ON speaking_mock_sessions(user_id) WHERE status = 'in_progress';

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated, is_available) VALUES
    ('mock-ielts-speaking-gen', 'ielts-2026-01', 'IELTS', 'IELTS Speaking Mock',
     'All three parts in one sitting, with the examiner''s questions read aloud and one rating for the whole test.',
     15, FALSE, TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO speaking_mock_sets (id, title, content) VALUES
('sm-01', 'Skills and learning', $json${
 "part1": [
  {"topic": "your home", "questions": ["Do you live in a house or a flat?", "What do you like most about your home?", "Is there anything you would like to change about it?"]},
  {"topic": "music", "questions": ["What kind of music do you enjoy?", "Did you learn to play an instrument when you were a child?", "Do you prefer listening to music alone or with other people?"]}
 ],
 "part2": {"topic": "a skill you learned", "cue": "Describe a skill you learned that took a long time to master.", "points": ["what the skill was", "why you decided to learn it", "how you learned it"], "explain": "and explain how you felt when you had finally mastered it.", "rounding": "Do you still use this skill today?"},
 "part3": {"topic": "learning skills", "questions": ["Why do some people give up when they are learning something new?", "Is it better to learn a skill from a teacher or by yourself?", "Which skills do you think will be most important for young people in the future?", "Should schools spend more time teaching practical skills?", "How has technology changed the way people learn new skills?"]}
}$json$::jsonb),
('sm-02', 'Places and crowds', $json${
 "part1": [
  {"topic": "your work or studies", "questions": ["Do you work or are you a student?", "What do you enjoy most about it?", "What would you like to change about it?"]},
  {"topic": "the weather", "questions": ["What kind of weather do you like best?", "Does the weather ever affect your plans?", "Has the weather in your country changed since you were a child?"]}
 ],
 "part2": {"topic": "a crowded place", "cue": "Describe a place you visited that was very crowded.", "points": ["where it was", "when you went there", "why it was so crowded"], "explain": "and explain how you felt about being there.", "rounding": "Would you go back there again?"},
 "part3": {"topic": "crowds and cities", "questions": ["Why do people like to live in big cities?", "What problems does overcrowding cause in cities?", "How could governments reduce crowding in popular tourist places?", "Do you think more people will choose to live in the countryside in the future?", "Is it the responsibility of individuals or governments to make cities more pleasant to live in?"]}
}$json$::jsonb),
('sm-03', 'Helping others', $json${
 "part1": [
  {"topic": "your hometown", "questions": ["Where is your hometown?", "What is it best known for?", "How has it changed in recent years?"]},
  {"topic": "reading", "questions": ["Do you enjoy reading?", "What kind of things do you usually read?", "Did you read more when you were a child than you do now?"]}
 ],
 "part2": {"topic": "helping someone", "cue": "Describe a time when you helped someone.", "points": ["who you helped", "what the situation was", "how you helped them"], "explain": "and explain how you felt afterwards.", "rounding": "Do you often help people in this way?"},
 "part3": {"topic": "helping others", "questions": ["Why do some people enjoy helping others more than other people do?", "Should schools teach children to help their community?", "What kinds of volunteer work are common in your country?", "Do you think people help their neighbours less than they used to?", "Should companies give their employees time off to do volunteer work?"]}
}$json$::jsonb),
('sm-04', 'Technology in daily life', $json${
 "part1": [
  {"topic": "food", "questions": ["What is your favourite food?", "Do you prefer eating at home or eating out?", "Have your eating habits changed in the last few years?"]},
  {"topic": "free time", "questions": ["What do you usually do in your free time?", "Do you prefer spending your free time alone or with other people?", "Is there a new hobby you would like to try?"]}
 ],
 "part2": {"topic": "a piece of technology", "cue": "Describe a piece of technology you use every day.", "points": ["what it is", "how long you have had it", "what you use it for"], "explain": "and explain how your life would be different without it.", "rounding": "Would you like to replace it with a newer model?"},
 "part3": {"topic": "technology and daily life", "questions": ["How has technology changed the way families spend time together?", "Do you think older people find it difficult to use new technology?", "What are the disadvantages of relying on technology?", "At what age should children be allowed to have a smartphone?", "Which piece of technology do you think will change our lives most in the future?"]}
}$json$::jsonb)
ON CONFLICT (id) DO NOTHING;
