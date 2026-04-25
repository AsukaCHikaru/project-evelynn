package api

import "github.com/asukachikaru/project-evelynn/server/db"

type Server struct {
	q *db.Queries
}

func NewServer(q *db.Queries) *Server {
	return &Server{
		q: q,
	}
}
