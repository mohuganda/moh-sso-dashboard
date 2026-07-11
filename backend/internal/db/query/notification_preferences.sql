-- name: GetNotificationPreferencesByUserID :one
SELECT *
FROM notification_preferences
WHERE user_id = $1
LIMIT 1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (
    user_id,
    email_enabled,
    sms_enabled,
    phone_number,
    phone_verified,
    quiet_hours_start,
    quiet_hours_end
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
ON CONFLICT (user_id) DO UPDATE SET
    email_enabled = EXCLUDED.email_enabled,
    sms_enabled = EXCLUDED.sms_enabled,
    phone_number = EXCLUDED.phone_number,
    phone_verified = CASE
        WHEN COALESCE(notification_preferences.phone_number, '') IS DISTINCT FROM COALESCE(EXCLUDED.phone_number, '')
        THEN FALSE
        ELSE notification_preferences.phone_verified
    END,
    quiet_hours_start = EXCLUDED.quiet_hours_start,
    quiet_hours_end = EXCLUDED.quiet_hours_end,
    updated_at = now()
RETURNING *;

-- name: ListNotificationPreferencesForSMS :many
SELECT *
FROM notification_preferences
WHERE sms_enabled = TRUE
    AND phone_number IS NOT NULL
    AND trim(phone_number) <> ''
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;
