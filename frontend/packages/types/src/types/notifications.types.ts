import type { Severity } from "@moh-sso/ui";

export type Notification = {
  id: string;
  type?: string;
  title: string;
  message: string;
  severity: Severity;
  target_role?: string;
  client_id?: string;
  user_id?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  read: boolean;
};

export type GetNotificationsParams = {
  unread?: boolean;
  limit?: number;
  offset?: number;
};

export type NotificationDeliveryStatus =
  | "PENDING"
  | "PROCESSING"
  | "SENT"
  | "FAILED"
  | "RETRY"
  | "CANCELLED";

export type NotificationDelivery = {
  id: string;
  notification_id: string;
  channel: string;
  status: NotificationDeliveryStatus | string;
  recipient?: Record<string, unknown>;
  template_name?: string;
  template_data?: Record<string, unknown>;
  payload?: Record<string, unknown>;
  scheduled_at?: string;
  locked_at?: string;
  sent_at?: string;
  attempts: number;
  max_attempts: number;
  last_error?: string;
  provider?: string;
  provider_message_id?: string;
  provider_status?: string;
  provider_metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type NotificationDeliveryListParams = {
  channel?: "in_app" | "email" | "sms" | "webhook" | string;
  status?: NotificationDeliveryStatus | string;
  search?: string;
  limit?: number;
  offset?: number;
};

export type NotificationDeliveryList = {
  items: NotificationDelivery[];
  total: number;
  limit: number;
  offset: number;
};

export type NotificationDeliveryMetric = {
  channel: string;
  provider?: string;
  total: number;
  pending: number;
  processing: number;
  sent: number;
  failed: number;
  retry: number;
  cancelled: number;
  avg_processing_seconds: number;
};

export type NotificationDeliveryMetrics = {
  items: NotificationDeliveryMetric[];
};

export type TestSMSRequest = {
  to: string;
  message: string;
};

export type TestSMSResponse = {
  notification_id: string;
  status: string;
};

export type NotificationPreferences = {
  user_id: string;
  email_enabled: boolean;
  sms_enabled: boolean;
  phone_number?: string;
  phone_verified: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
  created_at?: string;
  updated_at?: string;
};

export type UpdateNotificationPreferencesRequest = {
  email_enabled?: boolean;
  sms_enabled?: boolean;
  phone_number?: string;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
};
