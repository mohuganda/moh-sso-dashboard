-- name: ListAnnouncements :many
SELECT
    id,
    title,
    message,
    tag,
    priority,
    link_url,
    created_at
FROM announcements
ORDER BY created_at DESC
LIMIT $1;


-- name: CreateAnnouncement :one
INSERT INTO announcements (
    title,
    message,
    tag,
    priority,
    link_url,
    created_by
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: DeleteAnnouncement :exec
DELETE FROM announcements
WHERE id = $1;