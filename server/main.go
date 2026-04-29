package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/asukachikaru/project-evelynn/server/api"
	"github.com/asukachikaru/project-evelynn/server/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to open db: ", err)
	}
	err = conn.Ping()
	if err != nil {
		log.Fatal("Failed to connect to db: ", err)
	}
	log.Println("Connected to db.")
	defer conn.Close()

	q := db.New(conn)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	var srv = api.NewServer(q)

	router.Get("/api/user/profile", srv.GetUserProfile)
	router.Patch("/api/user/profile", srv.UpdateUser)
	router.Post("/api/user/profile", srv.CreateUser)

	log.Println("server starting on :8080")
	http.ListenAndServe(":8080", router)
}
