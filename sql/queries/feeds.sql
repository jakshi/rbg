-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListFeeds :many
SELECT *
FROM feeds
ORDER BY created_at DESC;

-- name: GetFeedByURL :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT ff.id, ff.user_id, ff.feed_id, ff.created_at, ff.updated_at,
       feeds.name AS feed_name, feeds.url AS feed_url
FROM feed_follows ff
JOIN feeds ON ff.feed_id = feeds.id
WHERE ff.user_id = $1
ORDER BY ff.created_at DESC;
