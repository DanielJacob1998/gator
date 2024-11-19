-- name: UnfollowFeed :exec
DELETE FROM feed_follows ff
WHERE ff.user_id = $1
AND ff.feed_id IN (SELECT id FROM feeds WHERE url = $2);
