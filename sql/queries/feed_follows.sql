-- name: CreateFeedFollow :one
with inserted AS (
INSERT INTO feed_follows (user_id, feed_id)
VALUES ($1, $2)
RETURNING *
)
SELECT inserted.*, users.name AS user_name, feeds.name AS feed_name
FROM inserted
JOIN users ON inserted.user_id = users.id
JOIN feeds ON inserted.feed_id = feeds.id;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;;
