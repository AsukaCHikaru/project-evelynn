CREATE TABLE quiz_user_verb_progress (
  progress_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INTEGER REFERENCES users(user_id) NOT NULL,
  verb_id INTEGER REFERENCES verbs(verb_id) NOT NULL,
  current_tense tense NOT NULL,
  rep_count SMALLINT NOT NULL DEFAULT 0,
  mastered_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(user_id, verb_id)
);