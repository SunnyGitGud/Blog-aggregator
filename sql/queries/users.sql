-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name) 
VALUES ( 
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetAllUsers :many
SELECT name FROM users;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at, name, url, user_id;

-- name: GetFeed :one
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds
WHERE id = $1;

-- name: GetFeedsByUser :many
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds
WHERE user_id = $1;
