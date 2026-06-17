import type { BackendErrorCode } from "./errorTypes";

export function mapErrorToMessage(code?: BackendErrorCode, fallback?: string) {
  switch (code) {
    case "CLIENT_ALREADY_EXISTS":
      return "A client with this ID already exists";

    case "USER_ALREADY_EXISTS":
      return "A user with this username already exists";

    case "INVALID_CLIENT_ID":
      return "Client ID contains invalid characters";

    case "VALIDATION_FAILED":
      return "Some fields need your attention";

    case "CLIENT_DISABLED":
      return "This client is currently disabled";

    case "USER_DISABLED":
      return "This user account is disabled";

    case "RATE_LIMITED":
      return "Too many attempts — please try again later";

    case "UNAUTHORIZED":
      return "You are not signed in";

    case "FORBIDDEN":
      return "You do not have permission to perform this action";

    case "INTERNAL_ERROR":
      return "Something went wrong on our side";

    default:
      return fallback ?? "Operation failed";
  }
}
