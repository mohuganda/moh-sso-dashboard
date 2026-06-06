/**
 * ---------------------------------------------------
 * Domain Severity (Business Layer)
 * ---------------------------------------------------
 * This represents backend/domain severity.
 */
export type Severity = "info" | "success" | "warning" | "critical" | "danger";

/**
 * ---------------------------------------------------
 * Carbon Tag Type (Inferred from Carbon)
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
 * Severity → Carbon Tag Mapping
 * ---------------------------------------------------
 */
const severityTagMap: Record<Severity, CarbonTagType> = {
  info: "blue",
  success: "green",
  warning: "magenta", // Carbon warning tone
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

/** Get Carbon Tag type */
export const getSeverityTagType = (severity: Severity): CarbonTagType => severityTagMap[severity];

/** Get formatted label */
export const getSeverityLabel = (severity: Severity): string => severityLabelMap[severity];

/**
 * Optional: numeric ranking (useful for sorting)
 */
export const getSeverityRank = (severity: Severity): number => {
  switch (severity) {
    case "danger":
      return 5;
    case "critical":
      return 4;
    case "warning":
      return 3;
    case "success":
      return 2;
    case "info":
      return 1;
  }
};
