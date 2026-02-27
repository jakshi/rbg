-- name: CreatePost :one
INSERT INTO posts (feed_id, title, url, published_at, description)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (url) DO NOTHING
RETURNING *;

-- name: GetPostsForUser :many
SELECT p.*, f.name AS feed_name, f.url AS feed_url
FROM posts p
JOIN feeds f ON p.feed_id = f.id
JOIN feed_follows ff ON f.id = ff.feed_id
WHERE ff.user_id = $1
ORDER BY p.published_at DESC, p.created_at DESC
LIMIT $2 OFFSET $3;