-- name: CreateDocumentTemplate :one
INSERT INTO document_templates (
    id,
    document_id,
    code,
    name,
    description,
    file_type,
    version,
    is_active,
    configuration,
    created_by
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;


-- name: GetDocumentTemplateByID :one
SELECT *
FROM document_templates
WHERE id = $1
LIMIT 1;

-- name: GetDocumentTemplateByCode :one
SELECT *
FROM document_templates
WHERE code = $1
  AND is_active = TRUE
ORDER BY version DESC
LIMIT 1;


-- name: ListDocumentTemplates :many
SELECT *
FROM document_templates
WHERE archived_at IS NULL
ORDER BY created_at DESC;


-- name: ListActiveDocumentTemplates :many
SELECT *
FROM document_templates
WHERE is_active = TRUE
  AND archived_at IS NULL
ORDER BY name;

-- name: UpdateDocumentTemplate :one
UPDATE document_templates
SET
    name = $2,
    description = $3,
    file_type = $4,
    configuration = $5,
    is_active = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: ArchiveDocumentTemplate :exec
UPDATE document_templates
SET archived_at = NOW()
WHERE id = $1;

-- name: DeleteDocumentTemplate :exec
DELETE FROM document_templates
WHERE id = $1;

-- name: ListDocumentTemplateVersions :many
SELECT *
FROM document_templates
WHERE code = $1
ORDER BY version DESC;


-- name: ExistsDocumentTemplateCode :one
SELECT EXISTS (
    SELECT 1
    FROM document_templates
    WHERE code = $1
);