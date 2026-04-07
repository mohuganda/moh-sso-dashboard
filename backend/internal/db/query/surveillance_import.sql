-- name: CreateImportBatch :one
INSERT INTO surveillance_import_batches (
  source_name,
  file_name,
  dataset_type,
  imported_by,
  status,
  notes,
  document_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetImportBatchByID :one
SELECT *
FROM surveillance_import_batches
WHERE id = $1
LIMIT 1;


-- name: GetLatestImportBatchByDocumentID :one
SELECT *
FROM surveillance_import_batches
WHERE document_id = $1
ORDER BY imported_at DESC, id DESC
LIMIT 1;

-- name: ListImportBatches :many
SELECT *
FROM surveillance_import_batches
ORDER BY imported_at DESC, id DESC;

-- name: UpdateImportBatchStatus :one
UPDATE surveillance_import_batches
SET
  status = $2,
  notes = $3
WHERE id = $1
RETURNING *;

-- name: UpdateImportBatchProgress :exec
UPDATE surveillance_import_batches
SET
  total_rows = $2,
  success_rows = $3,
  failed_rows = $4,
  notes = $5
WHERE id = $1;

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


-- name: MarkSurveillanceImportRawRowProcessed :exec
UPDATE surveillance_import_raw_rows
SET
  status = 'PROCESSED',
  error_message = NULL
WHERE id = $1;

-- name: MarkSurveillanceImportRawRowFailed :exec
UPDATE surveillance_import_raw_rows
SET
  status = 'FAILED',
  error_message = $2
WHERE id = $1;

-- name: CompleteSurveillanceImportBatch :exec
UPDATE surveillance_import_batches
SET
  status = 'COMPLETED',
  success_rows = $2,
  failed_rows = $3,
  completed_at = now(),
  notes = $4
WHERE id = $1;

-- name: FailSurveillanceImportBatch :exec
UPDATE surveillance_import_batches
SET
  status = 'FAILED',
  notes = $2,
  completed_at = now()
WHERE id = $1;

-- name: DeleteProcessedSurveillanceImportRawRows :exec
DELETE FROM surveillance_import_raw_rows
WHERE batch_id = $1
  AND status = 'PROCESSED';