CREATE TYPE process_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED',
    'CANCELLED'
);

-- name: CreateProcess :one
INSERT INTO processes (
    id,
    document_id,
    process_type,
    created_by
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetProcessByID :one
SELECT *
FROM processes
WHERE id = $1;

-- name: ListProcesses :many
SELECT *
FROM processes
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: ClaimNextPendingProcess :one
UPDATE processes
SET status = 'PROCESSING',
    started_at = NOW(),
    attempts = attempts + 1,
    updated_at = NOW()
WHERE id = (
    SELECT id
    FROM processes
    WHERE status = 'PENDING'
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: UpdateProcessProgress :exec
UPDATE processes
SET progress = $2,
    message = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteProcess :exec
UPDATE processes
SET status = 'COMPLETED',
    progress = 100,
    finished_at = NOW(),
    updated_at = NOW()
WHERE id = $1;


-- name: FailProcess :exec
UPDATE processes
SET status = 'FAILED',
    error = $2,
    finished_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CancelProcess :exec
UPDATE processes
SET status = 'CANCELLED',
    finished_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- -- name: InsertProcessEvent :one
-- INSERT INTO process_events (
--     id,
--     process_id,
--     status,
--     message
-- ) VALUES (
--     $1, $2, $3, $4
-- )
-- RETURNING *;


-- -- name: ListProcessEvents :many
-- SELECT *
-- FROM process_events
-- WHERE process_id = $1
-- ORDER BY created_at ASC;

-- name: ListProcessesByDocument :many
SELECT *
FROM processes
WHERE document_id = $1
ORDER BY created_at DESC;


CREATE INDEX IF NOT EXISTS idx_process_status_created
ON processes(status, created_at);

CREATE INDEX idx_process_type_status
ON processes(process_type, status);

-- CREATE INDEX IF NOT EXISTS idx_process_events_process
-- ON process_events(process_id);


