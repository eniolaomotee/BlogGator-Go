-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;


-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetUserFeed :one
SELECT users.name
FROM users
JOIN feeds ON users.id = feeds.user_id
WHERE feeds.id = $1;

