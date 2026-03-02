-- name: CreateDocument :one
INSERT INTO documents (
    id,
    original_filename,
    content_type,
    size_bytes,
    checksum_sha256,
    storage_location_id,
    object_key,
    uploaded_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;


-- name: GetDocumentByID :one
SELECT *
FROM documents
WHERE id = $1;


-- name: ListDocuments :many
SELECT *
FROM documents
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: ListDocumentsByUser :many
SELECT *
FROM documents
WHERE uploaded_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: DeleteDocument :exec
DELETE FROM documents
WHERE id = $1;


-- name: UpdateDocument :one
UPDATE documents
SET original_filename = $2,
    content_type = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;