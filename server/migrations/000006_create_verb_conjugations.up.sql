CREATE TABLE verb_conjugations (
  conjugation_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  verb_id INTEGER REFERENCES verbs(verb_id) NOT NULL,
  tense tense NOT NULL,
  pronoun pronoun NOT NULL,
  correct_form VARCHAR NOT NULL,
  UNIQUE(verb_id, tense, pronoun)
);