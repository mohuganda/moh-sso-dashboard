export type AnnouncementLevel = "INFO" | "SUCCESS" | "WARNING" | "CRITICAL";

export type AnnouncementStatus = "DRAFT" | "SCHEDULED" | "PUBLISHED" | "ARCHIVED";

export type Announcement = {
  id: string;
  title: string;
  message: string;
  summary: string | null;
  level: AnnouncementLevel;
  tag: string | null;
  link_url?: string | null;
  link_label?: string | null;
  status: AnnouncementStatus;
  publish_at: string | null;
  expires_at: string | null;
  is_pinned?: boolean;
};
