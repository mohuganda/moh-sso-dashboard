-- name: CreateDocumentTemplateColumn :one
INSERT INTO document_template_columns (
    id,
    sheet_id,
    column_key,
    column_name,
    display_name,
    data_type,
    required,
    is_unique,
    column_order,
    default_value,
    configuration,
    allowed_values,
    aliases
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
    $10,
    $11,
    $12,
    $13
)
RETURNING *;


-- name: GetDocumentTemplateColumnByID :one
SELECT *
FROM document_template_columns
WHERE id = $1
LIMIT 1;


-- name: GetDocumentTemplateColumnByKey :one
SELECT *
FROM document_template_columns
WHERE sheet_id = $1
  AND column_key = $2
  AND archived_at IS NULL
LIMIT 1;


-- name: ListDocumentTemplateColumns :many
SELECT *
FROM document_template_columns
WHERE sheet_id = $1
  AND archived_at IS NULL
ORDER BY column_order NULLS LAST, created_at;

-- name: ListRequiredDocumentTemplateColumns :many
SELECT *
FROM document_template_columns
WHERE sheet_id = $1
  AND required = TRUE
  AND archived_at IS NULL
ORDER BY column_order NULLS LAST;


-- name: ListUniqueDocumentTemplateColumns :many
SELECT *
FROM document_template_columns
WHERE sheet_id = $1
  AND is_unique = TRUE
  AND archived_at IS NULL
ORDER BY column_order NULLS LAST;

-- name: UpdateDocumentTemplateColumn :one
UPDATE document_template_columns
SET
    column_name = $2,
    display_name = $3,
    data_type = $4,
    required = $5,
    is_unique = $6,
    column_order = $7,
    default_value = $8,
    configuration = $9,
    allowed_values = $10,
    aliases = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchiveDocumentTemplateColumn :exec
UPDATE document_template_columns
SET
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ExistsDocumentTemplateColumnKey :one
SELECT EXISTS (
    SELECT 1
    FROM document_template_columns
    WHERE sheet_id = $1
      AND column_key = $2
      AND archived_at IS NULL
);

-- name: ExistsDocumentTemplateColumnName :one
SELECT EXISTS (
    SELECT 1
    FROM document_template_columns
    WHERE sheet_id = $1
      AND column_name = $2
      AND archived_at IS NULL
);

-- name: ReorderDocumentTemplateColumns :exec
UPDATE document_template_columns
SET
    column_order = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteDocumentTemplateColumn :exec
DELETE FROM document_template_columns
WHERE id = $1;