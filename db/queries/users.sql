-- name: InsertUser :exec
INSERT INTO users (id, email, password_hash, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetUserByEmail :one
SELECT id, email, password_hash
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash
FROM users
WHERE id = $1;
