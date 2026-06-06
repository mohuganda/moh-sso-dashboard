import type { BackendErrorCode } from "./errorTypes";

export type ToastSeverity = "error" | "warning" | "info";

export function mapErrorToSeverity(code?: BackendErrorCode): ToastSeverity {
  switch (code) {
    /* -------------------------
     * Hard failures
     * ------------------------- */
    case "INTERNAL_ERROR":
    case "UNAUTHORIZED":
    case "FORBIDDEN":
      return "error";

    /* -------------------------
     * User-correctable
     * ------------------------- */
    case "CLIENT_ALREADY_EXISTS":
    case "USER_ALREADY_EXISTS":
    case "INVALID_CLIENT_ID":
    case "VALIDATION_FAILED":
      return "warning";

    /* -------------------------
     * Informational / policy
     * ------------------------- */
    case "CLIENT_DISABLED":
    case "USER_DISABLED":
    case "RATE_LIMITED":
      return "info";

    default:
      return "error";
  }
}
