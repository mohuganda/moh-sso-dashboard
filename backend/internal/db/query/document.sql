-- name: CreateDocument :one
INSERT INTO documents (
    id,
    original_filename,
    content_type,
    size_bytes,
    checksum_sha256,
    storage_location_id,
    object_key,
    uploaded_by,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
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


-- name: UpdateDocumentStatus :one
UPDATE documents
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MarkDocumentCompleted :one
UPDATE documents
SET status = 'COMPLETED',
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MarkDocumentProcessing :one
UPDATE documents
SET status = 'PROCESSING',
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MarkDocumentFailed :one
UPDATE documents
SET status = 'FAILED',
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MarkDocumentPending :one
UPDATE documents
SET status = 'PENDING',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListDocumentsByStatus :many
SELECT *
FROM documents
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;