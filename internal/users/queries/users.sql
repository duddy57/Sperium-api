-- name: CreateUserQuery :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByIDQuery :one
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmailQuery :one
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateUserQuery :exec
UPDATE users
SET
    name = $2,
    updated_at = now()
WHERE id = $1;

-- name: DeleteUserQuery :exec
DELETE
FROM users
WHERE id = $1;

