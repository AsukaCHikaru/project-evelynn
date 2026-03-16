CREATE TABLE anki_words (
  word_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  word VARCHAR NOT NULL,
  part_of_speech part_of_speech NOT NULL,
  difficulty difficulty NOT NULL,
  gender gender,
  en_translation VARCHAR NOT NULL,
  example_sentence VARCHAR,
  en_example_sentence VARCHAR,
  created_at TIMESTAMP DEFAULT NOW()
);