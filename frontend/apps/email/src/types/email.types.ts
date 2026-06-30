export type EmailAddress = {
  name?: string;
  email: string;
};

export type EmailAttachment = {
  file_name: string;
  content_type?: string;
  path?: string;
  data_base64?: string;
  content_id?: string;
  inline?: boolean;
};

export type SendEmailRequest = {
  id?: string;
  from?: EmailAddress;
  to: EmailAddress[];
  cc?: EmailAddress[];
  bcc?: EmailAddress[];
  reply_to?: EmailAddress[];
  subject: string;
  text_body?: string;
  html_body?: string;
  template_name?: string;
  template_data?: Record<string, unknown>;
  attachments?: EmailAttachment[];
  headers?: Record<string, string>;
  metadata?: Record<string, string>;
  scheduled_at?: string;
};

export type EmailStatus = "PENDING" | "PROCESSING" | "SENT" | "FAILED" | "RETRY";

export type EmailMessage = {
  id?: string;
  from?: EmailAddress;
  to?: EmailAddress[];
  cc?: EmailAddress[];
  bcc?: EmailAddress[];
  reply_to?: EmailAddress[];
  subject?: string;
  text_body?: string;
  html_body?: string;
  template_name?: string;
  template_data?: Record<string, unknown>;
  attachments?: EmailAttachment[];
  headers?: Record<string, string>;
  metadata?: Record<string, string>;
  scheduled_at?: string | null;
  attachment_count?: number;
};

export type EmailOutboxItemResponse = {
  id: string;
  tenant_id?: string | null;
  message_id?: string | null;
  message: EmailMessage;
  status: EmailStatus | string;
  attempts: number;
  max_attempts: number;
  last_error?: string | null;
  scheduled_at?: string | null;
  locked_at?: string | null;
  sent_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type EmailOutboxItem = {
  id: string;
  tenant_id?: string | null;
  message_id?: string | null;
  message: EmailMessage;
  status: EmailStatus | string;
  attempts: number;
  max_attempts: number;
  last_error?: string | null;
  scheduled_at?: string | null;
  locked_at?: string | null;
  sent_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type EmailListParams = {
  limit?: number;
  offset?: number;
};

export type EmailListByStatusParams = EmailListParams & {
  status: EmailStatus | string;
};
