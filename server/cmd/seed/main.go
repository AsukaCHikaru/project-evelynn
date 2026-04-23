package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to open db: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to db: ", err)
	}

	file, err := os.OpenFile("./db/seed_words.csv", os.O_RDONLY, 0)
	if err != nil {
		log.Fatal("Failed to open seed file: ", err)
	}
	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read seed file: ", err)
	}

	col_id_map := make(map[string]int)
	for i, col := range rows[0] {
		col_id_map[col] = i
	}

	log.Printf("Start seeding, csv has %d rows", len(rows)-1)

	var rows_affected_sum int = 0

	for _, row := range rows[1:] {

		tx, err := db.Begin()
		if err != nil {
			log.Fatal("Failed to begin transaction: ", err)
		}
		defer tx.Rollback()

		var gender sql.NullString
		if raw := row[col_id_map["gender"]]; raw != "" {
			gender = sql.NullString{String: raw, Valid: true}
		}

		result, err := tx.Exec(`
		INSERT INTO anki_words (word, part_of_speech, difficulty, gender, en_translation, example_sentence, en_example_sentence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (word, part_of_speech) DO UPDATE SET
		difficulty = EXCLUDED.difficulty,
		gender = EXCLUDED.gender,
		en_translation = EXCLUDED.en_translation,
		example_sentence = EXCLUDED.example_sentence,
		en_example_sentence = EXCLUDED.en_example_sentence
		`,
			row[col_id_map["word"]],
			row[col_id_map["part_of_speech"]],
			row[col_id_map["difficulty"]],
			gender,
			row[col_id_map["en_translation"]],
			row[col_id_map["example_sentence"]],
			row[col_id_map["en_example_sentence"]],
		)
		if err != nil {
			log.Fatal("Failed to insert word: ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatal("Failed to commit transaction: ", err)
		}

		rows_affected, err := result.RowsAffected()
		if err != nil {
			log.Fatal("Failed to get rows affected: ", err)
		}
		rows_affected_sum += int(rows_affected)
	}

	log.Printf("Seeded %d words", rows_affected_sum)
}
