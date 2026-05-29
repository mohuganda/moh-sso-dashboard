-- =====================================================
-- Notification Deliveries
-- =====================================================

CREATE TABLE notification_deliveries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  notification_id UUID NOT NULL
    REFERENCES notifications(id)
    ON DELETE CASCADE,

  channel TEXT NOT NULL CHECK (
    channel IN ('in_app', 'email', 'sms', 'webhook')
  ),

  status TEXT NOT NULL DEFAULT 'PENDING' CHECK (
    status IN ('PENDING', 'PROCESSING', 'SENT', 'FAILED', 'RETRY', 'CANCELLED')
  ),

  recipient JSONB,
  template_name TEXT,
  template_data JSONB,
  payload JSONB,

  scheduled_at TIMESTAMPTZ,
  locked_at TIMESTAMPTZ,
  sent_at TIMESTAMPTZ,

  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  last_error TEXT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_deliveries_notification_id
ON notification_deliveries(notification_id);

CREATE INDEX idx_notification_deliveries_status_scheduled
ON notification_deliveries(status, scheduled_at, created_at);

CREATE INDEX idx_notification_deliveries_channel_status
ON notification_deliveries(channel, status);

CREATE INDEX idx_notification_deliveries_locked_at
ON notification_deliveries(locked_at);