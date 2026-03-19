-- name: CreateImportBatch :one
INSERT INTO surveillance_import_batches (
  source_name,
  file_name,
  dataset_type,
  imported_by,
  status,
  notes
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetImportBatchByID :one
SELECT *
FROM surveillance_import_batches
WHERE id = $1
LIMIT 1;

-- name: ListImportBatches :many
SELECT *
FROM surveillance_import_batches
ORDER BY imported_at DESC;

-- name: UpdateImportBatchStatus :one
UPDATE surveillance_import_batches
SET
  status = $2,
  notes = $3
WHERE id = $1
RETURNING *;

-- name: CreateImportRawRow :one
INSERT INTO surveillance_import_raw_rows (
  batch_id,
  row_number,
  payload
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: ListImportRawRowsByBatch :many
SELECT *
FROM surveillance_import_raw_rows
WHERE batch_id = $1
ORDER BY row_number ASC;