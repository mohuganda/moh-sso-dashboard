import type {
  Announcement,
  AnnouncementLevel,
  AnnouncementStatus,
} from "../store/types/announcements.types";
import type { DocumentProcessType } from "../store/types/documents.types";

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

export function validateProcessTypeAgainstFile(
  file: File,
  processType: DocumentProcessType,
): string | null {
  const fileName = file.name.toLowerCase();

  if (processType === "SURVEILLANCE_CSV_IMPORT" && !fileName.endsWith(".csv")) {
    return "Surveillance CSV import requires a .csv file.";
  }

  if (processType === "SURVEILLANCE_EXCEL_IMPORT" && !fileName.endsWith(".xlsx")) {
    return "Surveillance Excel import requires an .xlsx file.";
  }

  if (processType === "EXCEL_IMPORT" && !fileName.endsWith(".xlsx")) {
    return "Excel import requires an .xlsx file.";
  }

  return null;
}

export function getSuggestedProcessType(file: File): DocumentProcessType | "" {
  const fileName = file.name.toLowerCase();

  if (fileName.endsWith(".xlsx")) {
    return "EXCEL_IMPORT";
  }

  if (fileName.endsWith(".csv")) {
    return "CSV_IMPORT";
  }

  return "";
}

export function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(2)} MB`;
}

export const ACCEPTED_EXTENSIONS = [".csv", ".xlsx", ".xls"] as const;
export const ACCEPTED_MIME_TYPES = [
  "text/csv",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
] as const;

export function getExtension(name: string): string {
  const index = name.lastIndexOf(".");
  return index >= 0 ? name.slice(index).toLowerCase() : "";
}

export function isAcceptedFile(file: File): boolean {
  const extension = getExtension(file.name);
  const hasValidExtension = ACCEPTED_EXTENSIONS.includes(
    extension as (typeof ACCEPTED_EXTENSIONS)[number],
  );
  const hasValidMimeType = ACCEPTED_MIME_TYPES.includes(
    file.type as (typeof ACCEPTED_MIME_TYPES)[number],
  );

  return hasValidExtension || hasValidMimeType;
}



