-- name: CreateUser :one
INSERT INTO users (
  user_hash_id,
  display_name
) 
VALUES ($1, $2) 
RETURNING *;