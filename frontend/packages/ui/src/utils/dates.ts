export function formatDateTime(value?: string | number | null): string {
  if (!value) return "—";

  const timestamp = typeof value === "number" && value < 10_000_000_000 ? value * 1000 : value;

  const date = new Date(timestamp);

  if (Number.isNaN(date.getTime())) return "—";

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function formatRelativeTime(value?: string | number | null): string {
  if (!value) return "Recently";

  const timestamp = typeof value === "number" && value < 10_000_000_000 ? value * 1000 : value;

  const date = new Date(timestamp);

  if (Number.isNaN(date.getTime())) return "Recently";

  const diffMs = Date.now() - date.getTime();

  if (diffMs < 0) return formatDateTime(value);

  const seconds = Math.floor(diffMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (seconds < 60) return "Just now";
  if (minutes < 60) return `${minutes} minute${minutes === 1 ? "" : "s"} ago`;
  if (hours < 24) return `${hours} hour${hours === 1 ? "" : "s"} ago`;
  if (days < 7) return `${days} day${days === 1 ? "" : "s"} ago`;

  return formatDateTime(value);
}
