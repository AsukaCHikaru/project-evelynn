CREATE TABLE quiz_attempts (
  attempt_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INTEGER REFERENCES users(user_id) NOT NULL,
  verb_id INTEGER REFERENCES verbs(verb_id) NOT NULL,
  tense tense NOT NULL,
  perfect BOOLEAN NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);