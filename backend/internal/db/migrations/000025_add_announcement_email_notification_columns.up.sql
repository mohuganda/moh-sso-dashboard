
ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS notify_by_email BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS email_notification_sent_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_notify_by_email
    ON announcements (notify_by_email);

CREATE INDEX IF NOT EXISTS idx_announcements_email_notification_pending
    ON announcements (publish_at, created_at)
    WHERE
        deleted_at IS NULL
        AND status = 'PUBLISHED'
        AND notify_by_email = TRUE
        AND email_notification_sent_at IS NULL;