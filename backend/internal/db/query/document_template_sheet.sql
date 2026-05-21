-- name: CreateDocumentTemplateSheet :one
INSERT INTO document_template_sheets (
    id,
    template_id,
    code,
    name,
    display_name,
    required,
    sheet_order,
    header_row,
    start_row,
    allow_extra_columns,
    allow_duplicate_headers,
    configuration
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
    $12
)
RETURNING *;


-- name: GetDocumentTemplateSheetByID :one
SELECT *
FROM document_template_sheets
WHERE id = $1
LIMIT 1;


-- name: GetDocumentTemplateSheetByCode :one
SELECT *
FROM document_template_sheets
WHERE template_id = $1
  AND code = $2
  AND archived_at IS NULL
LIMIT 1;


-- name: ListDocumentTemplateSheets :many
SELECT *
FROM document_template_sheets
WHERE template_id = $1
  AND archived_at IS NULL
ORDER BY sheet_order NULLS LAST, created_at;


-- name: ListRequiredDocumentTemplateSheets :many
SELECT *
FROM document_template_sheets
WHERE template_id = $1
  AND required = TRUE
  AND archived_at IS NULL
ORDER BY sheet_order NULLS LAST;


-- name: UpdateDocumentTemplateSheet :one
UPDATE document_template_sheets
SET
    name = $2,
    display_name = $3,
    required = $4,
    sheet_order = $5,
    header_row = $6,
    start_row = $7,
    allow_extra_columns = $8,
    allow_duplicate_headers = $9,
    configuration = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: ArchiveDocumentTemplateSheet :exec
UPDATE document_template_sheets
SET
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1;


-- name: DeleteDocumentTemplateSheet :exec
DELETE FROM document_template_sheets
WHERE id = $1;

-- name: ExistsDocumentTemplateSheetCode :one
SELECT EXISTS (
    SELECT 1
    FROM document_template_sheets
    WHERE template_id = $1
      AND code = $2
      AND archived_at IS NULL
);

-- name: ExistsDocumentTemplateSheetName :one
SELECT EXISTS (
    SELECT 1
    FROM document_template_sheets
    WHERE template_id = $1
      AND name = $2
      AND archived_at IS NULL
);

-- name: ReorderDocumentTemplateSheets :exec
UPDATE document_template_sheets
SET
    sheet_order = $2,
    updated_at = NOW()
WHERE id = $1;
