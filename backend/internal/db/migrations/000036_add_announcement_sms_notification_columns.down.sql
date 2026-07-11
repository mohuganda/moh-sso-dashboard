DROP INDEX IF EXISTS idx_announcements_sms_notification_pending;
DROP INDEX IF EXISTS idx_announcements_notify_by_sms;

ALTER TABLE announcements
DROP COLUMN IF EXISTS sms_notification_queued_at;

ALTER TABLE announcements
DROP COLUMN IF EXISTS sms_message;

ALTER TABLE announcements
DROP COLUMN IF EXISTS notify_by_sms;
