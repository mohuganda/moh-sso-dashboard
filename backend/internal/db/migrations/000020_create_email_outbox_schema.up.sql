CREATE TABLE email_outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  message_id TEXT NULL,

  from_address JSONB,
  to_addresses JSONB NOT NULL,
  cc_addresses JSONB,
  bcc_addresses JSONB,
  reply_to_addresses JSONB,

  subject TEXT NOT NULL,
  text_body TEXT,
  html_body TEXT,
  template_name TEXT,
  template_data JSONB,

  attachments JSONB,
  headers JSONB,
  metadata JSONB,

  status TEXT NOT NULL DEFAULT 'PENDING',
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  last_error TEXT,

  scheduled_at TIMESTAMPTZ,
  locked_at TIMESTAMPTZ,
  sent_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_outbox_status_scheduled
ON email_outbox(status, scheduled_at, created_at);

CREATE INDEX idx_email_outbox_locked_at
ON email_outbox(locked_at);

CREATE INDEX idx_email_outbox_sent_at
ON email_outbox(sent_at);