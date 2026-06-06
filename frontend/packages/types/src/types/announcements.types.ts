export type AnnouncementLevel = "INFO" | "SUCCESS" | "WARNING" | "CRITICAL";

export type AnnouncementStatus = "DRAFT" | "SCHEDULED" | "PUBLISHED" | "ARCHIVED";

export type AnnouncementAudienceType =
  | "ALL_USERS"
  | "ADMINS_ONLY"
  | "SPECIFIC_CLIENTS"
  | "SPECIFIC_ROLES"
  | "SPECIFIC_USERS";

export interface Announcement {
  id: string;
  title: string;
  message: string;
  summary: string | null;
  level: AnnouncementLevel;
  tag: string | null;
  link_url: string | null;
  link_label: string | null;
  priority: number;
  is_pinned: boolean;
  status: AnnouncementStatus;
  publish_at: string | null;
  expires_at: string | null;
  audience_type: AnnouncementAudienceType;
  created_by: string | null;
  updated_by: string | null;
  published_by: string | null;
  archived_by: string | null;
  created_at: string;
  updated_at: string;
  published_at: string | null;
  archived_at: string | null;
  deleted_at: string | null;
  deleted_by: string | null;
  version: number;
}

export interface ListAnnouncementsParams {
  limit?: number;
  offset?: number;
  role?: string;
  clientId?: string;
}

export interface CreateAnnouncementRequest {
  title: string;
  message: string;
  summary?: string | null;
  level: AnnouncementLevel;
  tag?: string | null;
  link_url?: string | null;
  link_label?: string | null;
  priority?: number;
  is_pinned?: boolean;
  status?: AnnouncementStatus;
  publish_at?: string | null;
  expires_at?: string | null;
  audience_type?: AnnouncementAudienceType;
  client_ids?: string[];
  role_names?: string[];
  user_ids?: string[];
}

export interface UpdateAnnouncementRequest {
  title: string;
  message: string;
  summary?: string | null;
  level: AnnouncementLevel;
  tag?: string | null;
  link_url?: string | null;
  link_label?: string | null;
  priority?: number;
  is_pinned?: boolean;
  status?: AnnouncementStatus;
  publish_at?: string | null;
  expires_at?: string | null;
  audience_type?: AnnouncementAudienceType;
  client_ids?: string[];
  role_names?: string[];
  user_ids?: string[];
}

export interface ScheduleAnnouncementRequest {
  publish_at: string;
}

export interface SetAnnouncementPinnedRequest {
  is_pinned: boolean;
}

export interface SetAnnouncementPriorityRequest {
  priority: number;
}

export interface AnnouncementStats {
  total: number;
  draft_count: number;
  scheduled_count: number;
  published_count: number;
  archived_count: number;
  active_count: number;
}
