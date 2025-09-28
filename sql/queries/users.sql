-- name: CreateUser :one
INSERT INTO users (id, username, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users
WHERE username = $1;


-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT id, username, created_at, updated_at
FROM users
ORDER BY username;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeedsWithUsers :many
SELECT 
    f.id, f.name, f.url, f.created_at, f.updated_at,
    u.id AS user_id, u.username AS user_name
FROM feeds f
JOIN users u ON f.user_id = u.id
ORDER BY f.created_at DESC;


-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    iff.id,
    iff.created_at,
    iff.updated_at,
    iff.user_id,
    iff.feed_id,
    f.name AS feed_name,
    u.username AS user_name
FROM inserted_feed_follow iff
INNER JOIN feeds f ON f.id = iff.feed_id
INNER JOIN users u ON u.id = iff.user_id;

-- name: GetFeedFollowsForUser :many
SELECT
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    f.name AS feed_name,
    u.username AS user_name
FROM feed_follows ff
INNER JOIN feeds f ON f.id = ff.feed_id
INNER JOIN users u ON u.id = ff.user_id
WHERE ff.user_id = $1
ORDER BY ff.created_at DESC;


-- name: GetFeedByURL :one
SELECT id, name, url, created_at, updated_at, user_id
FROM feeds
WHERE url = $1;

-- name: DeleteFeedFollowByUserAndFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;
