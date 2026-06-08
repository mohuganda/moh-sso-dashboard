-- name: CreateAnnouncement :one
INSERT INTO announcements (
    title,
    message,
    summary,
    level,
    tag,
    link_url,
    priority,
    is_pinned,
    status,
    publish_at,
    expires_at,
    audience_type,
    notify_by_email,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14
)
RETURNING *;


-- name: GetAnnouncementByID :one
SELECT *
FROM announcements
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: GetAnnouncementByIDForUpdate :one
SELECT *
FROM announcements
WHERE id = $1
  AND deleted_at IS NULL
FOR UPDATE;


-- name: ListAnnouncementsAdmin :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
ORDER BY is_pinned DESC, priority DESC, created_at DESC
LIMIT $1 OFFSET $2;


-- name: CountAnnouncementsAdmin :one
SELECT COUNT(*)::bigint
FROM announcements
WHERE deleted_at IS NULL;


-- name: ListAnnouncementsByStatus :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND status = $1
ORDER BY is_pinned DESC, priority DESC, created_at DESC
LIMIT $2 OFFSET $3;


-- name: CountAnnouncementsByStatus :one
SELECT COUNT(*)::bigint
FROM announcements
WHERE deleted_at IS NULL
  AND status = $1;


-- name: ListActivePublishedAnnouncements :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND status = 'PUBLISHED'
  AND (publish_at IS NULL OR publish_at <= now())
  AND (expires_at IS NULL OR expires_at > now())
ORDER BY is_pinned DESC, priority DESC, publish_at DESC NULLS LAST, created_at DESC
LIMIT $1 OFFSET $2;


-- name: CountActivePublishedAnnouncements :one
SELECT COUNT(*)::bigint
FROM announcements
WHERE deleted_at IS NULL
  AND status = 'PUBLISHED'
  AND (publish_at IS NULL OR publish_at <= now())
  AND (expires_at IS NULL OR expires_at > now());


-- name: SearchAnnouncementsAdmin :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND (
      title ILIKE '%' || sqlc.arg(search_text) || '%'
      OR message ILIKE '%' || sqlc.arg(search_text) || '%'
      OR COALESCE(summary, '') ILIKE '%' || sqlc.arg(search_text) || '%'
      OR COALESCE(tag, '') ILIKE '%' || sqlc.arg(search_text) || '%'
  )
ORDER BY is_pinned DESC, priority DESC, created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);


-- name: CountSearchAnnouncementsAdmin :one
SELECT COUNT(*)::bigint
FROM announcements
WHERE deleted_at IS NULL
  AND (
      title ILIKE '%' || sqlc.arg(search_text) || '%'
      OR message ILIKE '%' || sqlc.arg(search_text) || '%'
      OR COALESCE(summary, '') ILIKE '%' || sqlc.arg(search_text) || '%'
      OR COALESCE(tag, '') ILIKE '%' || sqlc.arg(search_text) || '%'
  );


-- name: UpdateAnnouncement :one
UPDATE announcements
SET
    title = $2,
    message = $3,
    summary = $4,
    level = $5,
    tag = $6,
    link_url = $7,
    priority = $8,
    is_pinned = $9,
    publish_at = $10,
    expires_at = $11,
    audience_type = $12,
    notify_by_email = $13,
    updated_by = $14
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: UpdateAnnouncementStatus :one
UPDATE announcements
SET
    status = $2,
    updated_by = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: PublishAnnouncementNow :one
UPDATE announcements
SET
    status = 'PUBLISHED',
    publish_at = COALESCE(publish_at, now()),
    published_at = now(),
    published_by = $2,
    updated_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: DraftAnnouncement :one
UPDATE announcements
SET
    status = 'DRAFT',
    updated_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: ScheduleAnnouncement :one
UPDATE announcements
SET
    status = 'SCHEDULED',
    publish_at = $2,
    updated_by = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: ArchiveAnnouncement :one
UPDATE announcements
SET
    status = 'ARCHIVED',
    archived_at = now(),
    archived_by = $2,
    updated_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: UnarchiveAnnouncementToDraft :one
UPDATE announcements
SET
    status = 'DRAFT',
    archived_at = NULL,
    archived_by = NULL,
    updated_by = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SoftDeleteAnnouncement :exec
UPDATE announcements
SET
    deleted_at = now(),
    deleted_by = $2,
    updated_by = $2
WHERE id = $1
  AND deleted_at IS NULL;


-- name: RestoreAnnouncement :one
UPDATE announcements
SET
    deleted_at = NULL,
    deleted_by = NULL,
    updated_by = $2
WHERE id = $1
RETURNING *;


-- name: SetAnnouncementPinned :one
UPDATE announcements
SET
    is_pinned = $2,
    updated_by = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SetAnnouncementPriority :one
UPDATE announcements
SET
    priority = $2,
    updated_by = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: ListAnnouncementsCreatedByUser :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: CountAnnouncementsCreatedByUser :one
SELECT COUNT(*)::bigint
FROM announcements
WHERE deleted_at IS NULL
  AND created_by = $1;


-- name: MarkAnnouncementEmailNotificationSent :one
UPDATE announcements
SET email_notification_sent_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: ListPendingAnnouncementEmailNotifications :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND status = 'PUBLISHED'
  AND notify_by_email = TRUE
  AND email_notification_sent_at IS NULL
ORDER BY published_at ASC NULLS LAST, created_at ASC
LIMIT $1;


-- name: InsertAnnouncementClient :exec
INSERT INTO announcement_clients (
    announcement_id,
    client_id
) VALUES ($1, $2)
ON CONFLICT (announcement_id, client_id) DO NOTHING;


-- name: DeleteAnnouncementClients :exec
DELETE FROM announcement_clients
WHERE announcement_id = $1;


-- name: ListAnnouncementClients :many
SELECT client_id
FROM announcement_clients
WHERE announcement_id = $1
ORDER BY client_id;


-- name: InsertAnnouncementRole :exec
INSERT INTO announcement_roles (
    announcement_id,
    role_name
) VALUES ($1, $2)
ON CONFLICT (announcement_id, role_name) DO NOTHING;


-- name: DeleteAnnouncementRoles :exec
DELETE FROM announcement_roles
WHERE announcement_id = $1;


-- name: ListAnnouncementRoles :many
SELECT role_name
FROM announcement_roles
WHERE announcement_id = $1
ORDER BY role_name;


-- name: InsertAnnouncementUser :exec
INSERT INTO announcement_users (
    announcement_id,
    user_id
) VALUES ($1, $2)
ON CONFLICT (announcement_id, user_id) DO NOTHING;


-- name: DeleteAnnouncementUsers :exec
DELETE FROM announcement_users
WHERE announcement_id = $1;


-- name: ListAnnouncementUsers :many
SELECT user_id
FROM announcement_users
WHERE announcement_id = $1
ORDER BY user_id;


-- name: ListAnnouncementsForClient :many
SELECT DISTINCT a.*
FROM announcements a
LEFT JOIN announcement_clients ac
    ON ac.announcement_id = a.id
WHERE a.deleted_at IS NULL
  AND a.status = 'PUBLISHED'
  AND (a.publish_at IS NULL OR a.publish_at <= now())
  AND (a.expires_at IS NULL OR a.expires_at > now())
  AND (
      a.audience_type = 'ALL_USERS'
      OR (a.audience_type = 'SPECIFIC_CLIENTS' AND ac.client_id = $1)
  )
ORDER BY a.is_pinned DESC, a.priority DESC, a.publish_at DESC NULLS LAST, a.created_at DESC
LIMIT $2 OFFSET $3;


-- name: ListAnnouncementsForRole :many
SELECT DISTINCT a.*
FROM announcements a
LEFT JOIN announcement_roles ar
    ON ar.announcement_id = a.id
WHERE a.deleted_at IS NULL
  AND a.status = 'PUBLISHED'
  AND (a.publish_at IS NULL OR a.publish_at <= now())
  AND (a.expires_at IS NULL OR a.expires_at > now())
  AND (
      a.audience_type = 'ALL_USERS'
      OR (a.audience_type = 'ADMINS_ONLY' AND LOWER(sqlc.arg(role_name)::text) = 'admin')
      OR (a.audience_type = 'SPECIFIC_ROLES' AND LOWER(ar.role_name) = LOWER(sqlc.arg(role_name)::text))
  )
ORDER BY a.is_pinned DESC, a.priority DESC, a.publish_at DESC NULLS LAST, a.created_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);


-- name: ListAnnouncementsForUser :many
SELECT DISTINCT a.*
FROM announcements a
LEFT JOIN announcement_users au
    ON au.announcement_id = a.id
WHERE a.deleted_at IS NULL
  AND a.status = 'PUBLISHED'
  AND (a.publish_at IS NULL OR a.publish_at <= now())
  AND (a.expires_at IS NULL OR a.expires_at > now())
  AND (
      a.audience_type = 'ALL_USERS'
      OR (a.audience_type = 'SPECIFIC_USERS' AND au.user_id = $1)
  )
ORDER BY a.is_pinned DESC, a.priority DESC, a.publish_at DESC NULLS LAST, a.created_at DESC
LIMIT $2 OFFSET $3;


-- name: ListPublicAnnouncements :many
SELECT *
FROM announcements
WHERE deleted_at IS NULL
  AND status = 'PUBLISHED'
  AND (publish_at IS NULL OR publish_at <= now())
  AND (expires_at IS NULL OR expires_at > now())
  AND audience_type = 'ALL_USERS'
ORDER BY is_pinned DESC, priority DESC, publish_at DESC NULLS LAST, created_at DESC
LIMIT $1 OFFSET $2;


-- =====================================================
-- Announcement email recipient resolution
-- =====================================================

-- name: ListAnnouncementEmailRecipientsAllUsers :many
SELECT DISTINCT
    u.id,
    u.email,
    u.username,
    COALESCE(
        NULLIF(btrim(CONCAT_WS(' ', u.first_name, u.last_name)), ''),
        NULLIF(btrim(u.username), ''),
        u.email
    ) AS full_name
FROM users u
WHERE u.deleted_at IS NULL
  AND u.enabled = TRUE
  AND u.email IS NOT NULL
  AND btrim(u.email) <> ''
ORDER BY u.email;


-- name: ListAnnouncementEmailRecipientsAdmins :many
SELECT DISTINCT
    u.id,
    u.email,
    u.username,
    COALESCE(
        NULLIF(btrim(CONCAT_WS(' ', u.first_name, u.last_name)), ''),
        NULLIF(btrim(u.username), ''),
        u.email
    ) AS full_name
FROM users u
JOIN user_roles ur
    ON ur.user_id = u.id
WHERE u.deleted_at IS NULL
  AND u.enabled = TRUE
  AND LOWER(ur.role_name) = 'admin'
  AND u.email IS NOT NULL
  AND btrim(u.email) <> ''
ORDER BY u.email;


-- name: ListAnnouncementEmailRecipientsByClients :many
SELECT DISTINCT
    u.id,
    u.email,
    u.username,
    COALESCE(
        NULLIF(btrim(CONCAT_WS(' ', u.first_name, u.last_name)), ''),
        NULLIF(btrim(u.username), ''),
        u.email
    ) AS full_name
FROM announcement_clients ac
JOIN user_clients uc
    ON uc.client_id = ac.client_id
JOIN users u
    ON u.id = uc.user_id
WHERE ac.announcement_id = $1
  AND u.deleted_at IS NULL
  AND u.enabled = TRUE
  AND u.email IS NOT NULL
  AND btrim(u.email) <> ''
ORDER BY u.email;


-- name: ListAnnouncementEmailRecipientsByRoles :many
SELECT DISTINCT
    u.id,
    u.email,
    u.username,
    COALESCE(
        NULLIF(btrim(CONCAT_WS(' ', u.first_name, u.last_name)), ''),
        NULLIF(btrim(u.username), ''),
        u.email
    ) AS full_name
FROM announcement_roles ar
JOIN user_roles ur
    ON LOWER(ur.role_name) = LOWER(ar.role_name)
JOIN users u
    ON u.id = ur.user_id
WHERE ar.announcement_id = $1
  AND u.deleted_at IS NULL
  AND u.enabled = TRUE
  AND u.email IS NOT NULL
  AND btrim(u.email) <> ''
ORDER BY u.email;


-- name: ListAnnouncementEmailRecipientsByUsers :many
SELECT DISTINCT
    u.id,
    u.email,
    u.username,
    COALESCE(
        NULLIF(btrim(CONCAT_WS(' ', u.first_name, u.last_name)), ''),
        NULLIF(btrim(u.username), ''),
        u.email
    ) AS full_name
FROM announcement_users au
JOIN users u
    ON u.id = au.user_id
WHERE au.announcement_id = $1
  AND u.deleted_at IS NULL
  AND u.enabled = TRUE
  AND u.email IS NOT NULL
  AND btrim(u.email) <> ''
ORDER BY u.email;


-- name: GetAnnouncementStats :one
SELECT
    COUNT(*)::bigint AS total,
    COUNT(*) FILTER (WHERE status = 'DRAFT')::bigint AS draft_count,
    COUNT(*) FILTER (WHERE status = 'SCHEDULED')::bigint AS scheduled_count,
    COUNT(*) FILTER (WHERE status = 'PUBLISHED')::bigint AS published_count,
    COUNT(*) FILTER (WHERE status = 'ARCHIVED')::bigint AS archived_count,
    COUNT(*) FILTER (
        WHERE status = 'PUBLISHED'
          AND (publish_at IS NULL OR publish_at <= now())
          AND (expires_at IS NULL OR expires_at > now())
    )::bigint AS active_count
FROM announcements
WHERE deleted_at IS NULL;