-- =====================================================
-- Notification Deliveries
-- =====================================================

-- name: CreateNotificationDelivery :one
INSERT INTO notification_deliveries (
  notification_id,
  channel,
  status,
  recipient,
  template_name,
  template_data,
  payload,
  scheduled_at,
  attempts,
  max_attempts
)
VALUES (
  $1,
  $2,
  COALESCE(sqlc.narg('delivery_status'), 'PENDING'),
  $3,
  $4,
  $5,
  $6,
  $7,
  0,
  COALESCE(sqlc.narg('delivery_max_attempts'), 5)
)
RETURNING *;


-- name: ListNotificationDeliveriesByNotificationID :many
SELECT *
FROM notification_deliveries
WHERE notification_id = $1
ORDER BY created_at ASC;


-- name: GetNotificationDeliveryByID :one
SELECT *
FROM notification_deliveries
WHERE id = $1;


-- name: ClaimPendingNotificationDeliveries :many
WITH next_deliveries AS (
  SELECT nd.id
  FROM notification_deliveries nd
  WHERE
    nd.channel = $1
    AND nd.status IN ('PENDING', 'RETRY')
    AND (nd.scheduled_at IS NULL OR nd.scheduled_at <= now())
  ORDER BY nd.scheduled_at NULLS FIRST, nd.created_at ASC
  LIMIT $2
  FOR UPDATE SKIP LOCKED
)
UPDATE notification_deliveries nd
SET
  status = 'PROCESSING',
  locked_at = now(),
  updated_at = now()
FROM next_deliveries next
WHERE nd.id = next.id
RETURNING nd.*;


-- name: MarkNotificationDeliverySent :exec
UPDATE notification_deliveries
SET
  status = 'SENT',
  sent_at = now(),
  locked_at = NULL,
  last_error = NULL,
  updated_at = now()
WHERE id = $1;


-- name: MarkNotificationDeliveryRetry :exec
UPDATE notification_deliveries
SET
  status = 'RETRY',
  attempts = attempts + 1,
  locked_at = NULL,
  last_error = $2,
  scheduled_at = now() + ($3::text)::interval,
  updated_at = now()
WHERE id = $1;


-- name: MarkNotificationDeliveryFailed :exec
UPDATE notification_deliveries
SET
  status = 'FAILED',
  attempts = attempts + 1,
  locked_at = NULL,
  last_error = $2,
  updated_at = now()
WHERE id = $1;


-- name: CancelNotificationDelivery :exec
UPDATE notification_deliveries
SET
  status = 'CANCELLED',
  locked_at = NULL,
  updated_at = now()
WHERE id = $1;


-- name: CountPendingNotificationDeliveriesByChannel :one
SELECT COUNT(*)::bigint
FROM notification_deliveries
WHERE
  channel = $1
  AND status IN ('PENDING', 'RETRY')
  AND (scheduled_at IS NULL OR scheduled_at <= now());


-- name: ListNotificationDeliveries :many
SELECT nd.*
FROM notification_deliveries nd
WHERE
  (
    sqlc.narg('filter_channel')::text IS NULL
    OR nd.channel = sqlc.narg('filter_channel')::text
  )
  AND (
    sqlc.narg('filter_status')::text IS NULL
    OR nd.status = sqlc.narg('filter_status')::text
  )
ORDER BY nd.created_at DESC
LIMIT $1 OFFSET $2;


-- name: CountNotificationDeliveries :one
SELECT COUNT(*)::bigint
FROM notification_deliveries nd
WHERE
  (
    sqlc.narg('filter_channel')::text IS NULL
    OR nd.channel = sqlc.narg('filter_channel')::text
  )
  AND (
    sqlc.narg('filter_status')::text IS NULL
    OR nd.status = sqlc.narg('filter_status')::text
  );