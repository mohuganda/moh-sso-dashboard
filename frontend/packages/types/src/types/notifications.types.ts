import type { Severity } from "@moh-sso/ui";

export type Notification = {
  id: string;
  title: string;
  message: string;
  severity: Severity;
  created_at: string;
  read: boolean;
};

export type GetNotificationsParams = {
  unread?: boolean;
  limit?: number;
  offset?: number;
};
