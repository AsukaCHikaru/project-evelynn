CREATE TABLE anki_user_words (
  entry_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INTEGER REFERENCES users(user_id) NOT NULL,
  word_id INTEGER REFERENCES anki_words(word_id) NOT NULL,
  ease_factor REAL NOT NULL,
  review_interval INTEGER NOT NULL,
  last_reviewed TIMESTAMP NOT NULL,
  rep_count SMALLINT NOT NULL DEFAULT 0,
  next_review_date DATE NOT NULL,
  revealed BOOLEAN NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);