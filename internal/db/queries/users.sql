-- name: CreateUser :one
INSERT INTO users (username, password_hash, first_name, last_name, avatar_url, phone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS exists;

-- name: UpdateUser :one
UPDATE users
SET username      = $2,
    password_hash = $3,
    first_name    = $4,
    last_name     = $5,
    avatar_url    = $6,
    phone         = $7
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
