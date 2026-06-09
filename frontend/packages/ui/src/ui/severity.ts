/**
 * ---------------------------------------------------
 * Domain Severity (Business Layer)
 * ---------------------------------------------------
 */
export type Severity = "info" | "success" | "warning" | "critical" | "danger";

/**
 * ---------------------------------------------------
 * Carbon Tag Type
 * ---------------------------------------------------
 */
export type CarbonTagType =
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

/**
 * ---------------------------------------------------
 * Severity guards / normalization
 * ---------------------------------------------------
 */
const SEVERITIES: readonly Severity[] = [
  "info",
  "success",
  "warning",
  "critical",
  "danger",
] as const;

export function isSeverity(value: unknown): value is Severity {
  return typeof value === "string" && SEVERITIES.includes(value.toLowerCase() as Severity);
}

export function normalizeSeverity(value?: string | null): Severity {
  const normalized = value?.trim().toLowerCase();

  if (isSeverity(normalized)) {
    return normalized;
  }

  return "info";
}

/**
 * ---------------------------------------------------
 * Severity → Carbon Tag Mapping
 * ---------------------------------------------------
 */
const severityTagMap: Record<Severity, CarbonTagType> = {
  info: "blue",
  success: "green",
  warning: "magenta",
  critical: "red",
  danger: "red",
};

/**
 * ---------------------------------------------------
 * Severity → Label Formatting
 * ---------------------------------------------------
 */
const severityLabelMap: Record<Severity, string> = {
  info: "Info",
  success: "Success",
  warning: "Warning",
  critical: "Critical",
  danger: "Danger",
};

/**
 * ---------------------------------------------------
 * Public Helpers
 * ---------------------------------------------------
 */

export const getSeverityTagType = (severity?: string | null): CarbonTagType => {
  return severityTagMap[normalizeSeverity(severity)];
};

export const getSeverityLabel = (severity?: string | null): string => {
  return severityLabelMap[normalizeSeverity(severity)];
};

export const getSeverityRank = (severity?: string | null): number => {
  switch (normalizeSeverity(severity)) {
    case "danger":
      return 5;
    case "critical":
      return 4;
    case "warning":
      return 3;
    case "success":
      return 2;
    case "info":
    default:
      return 1;
  }
};
