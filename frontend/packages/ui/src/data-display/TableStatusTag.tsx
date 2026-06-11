import { Tag } from "@carbon/react";

type TagKind =
  | "red"
  | "magenta"
  | "purple"
  | "blue"
  | "cyan"
  | "teal"
  | "green"
  | "gray"
  | "cool-gray"
  | "warm-gray"
  | "high-contrast"
  | "outline";

type TableStatusTagProps = {
  status: string;
  kind?: TagKind;
  mapStatusToKind?: (status: string) => TagKind;
};

function defaultStatusKind(status: string): TagKind {
  const normalized = status.trim().toLowerCase();

  if (["active", "enabled", "completed", "success", "verified"].includes(normalized)) {
    return "green";
  }

  if (["disabled", "failed", "failure", "error", "not verified"].includes(normalized)) {
    return "red";
  }

  if (["pending", "processing", "warning"].includes(normalized)) {
    return "blue";
  }

  return "gray";
}

export function TableStatusTag({ status, kind, mapStatusToKind }: TableStatusTagProps) {
  return <Tag type={kind ?? mapStatusToKind?.(status) ?? defaultStatusKind(status)}>{status}</Tag>;
}
