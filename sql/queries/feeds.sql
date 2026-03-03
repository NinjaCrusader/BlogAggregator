-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: PrintFeeds :many
SELECT users.name AS user, url, feeds.name AS feed
FROM feeds
INNER JOIN users ON feeds.user_id = users.id;