export interface Announcement {
  id: string;
  title: string;
  message: string;
  tag: string;
  priority: number | null;
  link_url: string | null;
  created_at: string;
}

export interface ListAnnouncementsParams {
  limit?: number;
}

export interface CreateAnnouncementRequest {
  title: string;
  message: string;
  tag: string;
  priority?: number | null;
  link_url?: string | null;
}
