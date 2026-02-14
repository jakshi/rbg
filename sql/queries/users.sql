-- name: CreateUser :one
INSERT INTO users (name)
VALUES ($1)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByName :one
SELECT *
FROM users
WHERE name = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;
