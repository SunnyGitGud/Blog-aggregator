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


-- name: GetAllFeeds :many
SELECT 
    feeds.name AS feed_name,
    feeds.url AS feed_url,
    users.name AS user_name
FROM feeds
INNER JOIN users ON feeds.user_id = users.id;

-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, 
    (SELECT id FROM feeds WHERE feeds.url = $2)
  )  -- Pass feed_id directly from application
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT 
    inserted.id,
    inserted.created_at,
    inserted.updated_at,
    inserted.user_id,
    inserted.feed_id,
    users.name AS user_name,
    feeds.name AS feed_name
FROM inserted
JOIN users ON inserted.user_id = users.id
JOIN feeds ON inserted.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT 
  feed_follows.id,
  feed_follows.created_at,
  feed_follows.updated_at,
  feed_follows.user_id,
  feed_follows.feed_id,
  users.name AS user_name,
  feeds.name AS feed_name
FROM feed_follows
JOIN users ON feed_follows.user_id = users.id
JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;

-- name: GetFeedByURL :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at
FROM feeds
WHERE url = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;



-- name: MarkFeedFetched :exec
UPDATE feeds
SET 
    last_fetched_at = NOW(),
    updated_at = NOW()
WHERE id = $1;


-- name: GetNextFeedToFetch :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.*
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
WHERE feeds.user_id = $1
ORDER BY posts.published_at DESC NULLS LAST
LIMIT $2;

