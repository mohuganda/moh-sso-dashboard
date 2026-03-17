import type {
  Announcement,
  AnnouncementLevel,
  AnnouncementStatus,
} from "../store/types/announcements.types";

export function getTagType(
  level: AnnouncementLevel,
): "blue" | "green" | "red" | "purple" | "warm-gray" {
  switch (level) {
    case "SUCCESS":
      return "green";
    case "WARNING":
      return "purple";
    case "CRITICAL":
      return "red";
    case "INFO":
    default:
      return "blue";
  }
}

export function getStatusTagType(
  status: AnnouncementStatus,
): "blue" | "green" | "red" | "purple" | "warm-gray" {
  switch (status) {
    case "PUBLISHED":
      return "green";
    case "DRAFT":
      return "warm-gray";
    case "SCHEDULED":
      return "purple";
    case "ARCHIVED":
      return "warm-gray";
    default:
      return "blue";
  }
}

export function formatRelativeTime(dateString: string) {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();

  const minutes = Math.floor(diffMs / (1000 * 60));
  const hours = Math.floor(diffMs / (1000 * 60 * 60));
  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (minutes < 1) return "Just now";
  if (minutes < 60) return `${minutes} min ago`;
  if (hours < 24) return `${hours} hr ago`;
  if (days === 1) return "Yesterday";
  return `${days} days ago`;
}

export function formatDateTime(dateString?: string | null) {
  if (!dateString) return "—";

  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) return "—";

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function truncateText(text: string, limit = 110) {
  if (!text) return "";
  return text.length > limit ? `${text.slice(0, limit)}...` : text;
}

export function isAnnouncementActive(item: Announcement) {
  if (item.status !== "PUBLISHED") return false;

  const now = Date.now();
  const publishAt = item.publish_at ? new Date(item.publish_at).getTime() : null;
  const expiresAt = item.expires_at ? new Date(item.expires_at).getTime() : null;

  if (publishAt && publishAt > now) return false;
  if (expiresAt && expiresAt <= now) return false;

  return true;
}
