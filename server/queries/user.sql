-- name: CreateUser :one
INSERT INTO users (
  user_hash_id,
  display_name
) 
VALUES ($1, $2) 
RETURNING *;

-- name: GetUser :one
SELECT display_name, daily_word_limit
FROM users
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET 
  display_name = COALESCE(sqlc.narg('display_name')::text, display_name),
  daily_word_limit = COALESCE(sqlc.narg('daily_word_limit')::int, daily_word_limit)
WHERE user_id = (
  SELECT user_id
  FROM users
  LIMIT 1
)
RETURNING display_name, daily_word_limit;