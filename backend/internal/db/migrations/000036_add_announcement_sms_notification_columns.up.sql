ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS notify_by_sms BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS sms_message TEXT;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS sms_notification_queued_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_notify_by_sms
    ON announcements (notify_by_sms);

CREATE INDEX IF NOT EXISTS idx_announcements_sms_notification_pending
    ON announcements (publish_at, created_at)
    WHERE
        deleted_at IS NULL
        AND status = 'PUBLISHED'
        AND notify_by_sms = TRUE
        AND sms_notification_queued_at IS NULL;
