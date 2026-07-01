DROP INDEX IF EXISTS idx_notification_deliveries_provider;

ALTER TABLE notification_deliveries
    DROP COLUMN IF EXISTS provider_response,
    DROP COLUMN IF EXISTS provider_status,
    DROP COLUMN IF EXISTS provider_message_id,
    DROP COLUMN IF EXISTS provider;
