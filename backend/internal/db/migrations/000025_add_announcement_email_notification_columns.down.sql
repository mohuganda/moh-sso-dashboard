-- 20260608_add_announcement_email_notification_fields.down.sql

DROP INDEX IF EXISTS idx_announcements_email_notification_pending;

DROP INDEX IF EXISTS idx_announcements_notify_by_email;

ALTER TABLE announcements
DROP COLUMN IF EXISTS email_notification_sent_at;

ALTER TABLE announcements
DROP COLUMN IF EXISTS notify_by_email;