-- name: CreateEmailOutbox :one
INSERT INTO email_outbox (
  message_id,
  from_address,
  to_addresses,
  cc_addresses,
  bcc_addresses,
  reply_to_addresses,
  subject,
  text_body,
  html_body,
  template_name,
  template_data,
  attachments,
  headers,
  metadata,
  status,
  attempts,
  max_attempts,
  scheduled_at
)
VALUES (
  sqlc.arg(message_id),
  sqlc.arg(from_address),
  sqlc.arg(to_addresses),
  sqlc.arg(cc_addresses),
  sqlc.arg(bcc_addresses),
  sqlc.arg(reply_to_addresses),
  sqlc.arg(subject),
  sqlc.arg(text_body),
  sqlc.arg(html_body),
  sqlc.arg(template_name),
  sqlc.arg(template_data),
  sqlc.arg(attachments),
  sqlc.arg(headers),
  sqlc.arg(metadata),
  COALESCE(sqlc.narg(status), 'PENDING'),
  COALESCE(sqlc.narg(attempts), 0),
  COALESCE(sqlc.narg(max_attempts), 5),
  sqlc.arg(scheduled_at)
)
RETURNING *;


-- name: GetEmailOutboxByID :one
SELECT *
FROM email_outbox
WHERE id = $1
LIMIT 1;


-- name: ListEmailOutbox :many
SELECT *
FROM email_outbox
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: ListEmailOutboxByStatus :many
SELECT *
FROM email_outbox
WHERE status = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: ClaimEmailOutboxBatch :many
WITH picked AS (
  SELECT id
  FROM email_outbox
  WHERE status IN ('PENDING', 'RETRYING')
    AND (scheduled_at IS NULL OR scheduled_at <= now())
    AND attempts < max_attempts
    AND (locked_at IS NULL OR locked_at < now() - interval '5 minutes')
  ORDER BY created_at ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
UPDATE email_outbox eo
SET status = 'PROCESSING',
    locked_at = now(),
    updated_at = now()
FROM picked
WHERE eo.id = picked.id
RETURNING eo.*;

-- name: MarkEmailOutboxSent :exec
UPDATE email_outbox
SET status = 'SENT',
    sent_at = now(),
    locked_at = NULL,
    updated_at = now()
WHERE id = $1;

-- name: MarkEmailOutboxRetry :exec
UPDATE email_outbox
SET
  status = 'RETRYING',
  attempts = sqlc.arg(attempts),
  last_error = sqlc.arg(last_error),
  locked_at = NULL,
  updated_at = now()
WHERE id = sqlc.arg(id);

-- name: MarkEmailOutboxFailed :exec
UPDATE email_outbox
SET
  status = 'FAILED',
  attempts = sqlc.arg(attempts),
  last_error = sqlc.arg(last_error),
  locked_at = NULL,
  updated_at = now()
WHERE id = sqlc.arg(id);

-- name: MarkEmailOutboxProcessing :exec
UPDATE email_outbox
SET
  status = 'PROCESSING',
  locked_at = now(),
  updated_at = now()
WHERE id = $1;

-- name: ResetStuckEmailOutboxJobs :execrows
UPDATE email_outbox
SET
  status = 'RETRYING',
  locked_at = NULL,
  updated_at = now()
WHERE status = 'PROCESSING'
  AND locked_at IS NOT NULL
  AND locked_at < now() - interval '15 minutes';


-- name: DeleteEmailOutboxByID :exec
DELETE FROM email_outbox
WHERE id = $1;

-- name: DeleteSentEmailOutboxOlderThan :execrows
DELETE FROM email_outbox
WHERE status = 'SENT'
  AND sent_at IS NOT NULL
  AND sent_at < $1;

-- name: CountEmailOutboxByStatus :one
SELECT COUNT(*)::bigint
FROM email_outbox
WHERE status = $1;

-- name: ListEmailOutboxPendingDue :many
SELECT *
FROM email_outbox
WHERE status IN ('PENDING', 'RETRYING')
  AND attempts < max_attempts
  AND (scheduled_at IS NULL OR scheduled_at <= now())
ORDER BY created_at ASC
LIMIT $1;

-- name: ListEmailOutboxFailures :many
SELECT *
FROM email_outbox
WHERE status = 'FAILED'
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;

-- name: ListEmailOutboxByRecipient :many
SELECT *
FROM email_outbox
WHERE to_addresses::text ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;