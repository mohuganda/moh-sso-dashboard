package apierror

import (
	"net/http"
	"time"
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

// StatusCode lets HTTP module adapters consume the error without importing
// dashboard packages.
func (e *APIError) StatusCode() int { return e.HTTPStatus }

/* -----------------------------
 * Constructor
 * ----------------------------- */

func New(status int, code, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

/* =========================================================
 * Generic / Cross-cutting errors
 * ========================================================= */

var (
	ErrInternal = New(
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Something went wrong on our side",
	)

	ErrValidationFailed = New(
		http.StatusBadRequest,
		"VALIDATION_FAILED",
		"Invalid request data",
	)

	ErrUnauthorized = New(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"You are not authenticated",
	)

	ErrForbidden = New(
		http.StatusForbidden,
		"FORBIDDEN",
		"You do not have permission to perform this action",
	)

	ErrNotFound = New(
		http.StatusNotFound,
		"NOT_FOUND",
		"Requested resource was not found",
	)

	ErrInvalidUUID = New(
		http.StatusBadRequest,
		"INVALID_UUID",
		"Invalid UUID format",
	)

	ErrInvalidDateFormat = New(
		http.StatusBadRequest,
		"INVALID_DATE_FORMAT",
		"Invalid date format (RFC3339 required)",
	)

	ErrInvalidFile = New(
		http.StatusBadRequest,
		"INVALID_FILE",
		"Invalid or unreadable file",
	)
)

// RateLimitError exposes the same status and retry information to standalone
// HTTP modules through their structural error interfaces.
type RateLimitError struct {
	Reset time.Time
}

func (e RateLimitError) Error() string             { return "rate limit exceeded" }
func (e RateLimitError) StatusCode() int           { return http.StatusTooManyRequests }
func (e RateLimitError) RetryAfter() time.Duration { return time.Until(e.Reset) }

/* =========================================================
 * Client errors
 * ========================================================= */

var (
	ErrClientAlreadyExists = New(
		http.StatusConflict,
		"CLIENT_ALREADY_EXISTS",
		"Client with this ID already exists",
	)

	ErrInvalidClientID = New(
		http.StatusBadRequest,
		"INVALID_CLIENT_ID",
		"Client ID contains invalid characters",
	)

	ErrClientNotFound = New(
		http.StatusNotFound,
		"CLIENT_NOT_FOUND",
		"Client not found",
	)

	ErrClientDisabled = New(
		http.StatusForbidden,
		"CLIENT_DISABLED",
		"Client is disabled",
	)
)

/* =========================================================
 * User errors
 * ========================================================= */

var (
	ErrUserAlreadyExists = New(
		http.StatusConflict,
		"USER_ALREADY_EXISTS",
		"User already exists",
	)

	ErrUserNotFound = New(
		http.StatusNotFound,
		"USER_NOT_FOUND",
		"User not found",
	)

	ErrUserDisabled = New(
		http.StatusForbidden,
		"USER_DISABLED",
		"User account is disabled",
	)

	ErrEmailAlreadyExists = New(
		http.StatusConflict,
		"EMAIL_ALREADY_EXISTS",
		"Email address already exists",
	)

	ErrUsernameAlreadyExists = New(
		http.StatusConflict,
		"USERNAME_ALREADY_EXISTS",
		"Username already exists",
	)
)

/* =========================================================
 * Authentication / Security
 * ========================================================= */

var (
	ErrInvalidCredentials = New(
		http.StatusUnauthorized,
		"INVALID_CREDENTIALS",
		"Invalid username or password",
	)

	ErrTokenExpired = New(
		http.StatusUnauthorized,
		"TOKEN_EXPIRED",
		"Authentication token has expired",
	)

	ErrTokenInvalid = New(
		http.StatusUnauthorized,
		"TOKEN_INVALID",
		"Invalid authentication token",
	)
)

/* =========================================================
 * Notification errors
 * ========================================================= */

var (
	ErrNotificationNotFound = New(
		http.StatusNotFound,
		"NOTIFICATION_NOT_FOUND",
		"Notification not found",
	)
)

/* =========================================================
 * Import / Bulk operations
 * ========================================================= */

var (
	ErrJobNotFound = New(
		http.StatusNotFound,
		"JOB_NOT_FOUND",
		"Import job not found",
	)

	ErrJobAlreadyProcessed = New(
		http.StatusConflict,
		"JOB_ALREADY_PROCESSED",
		"Import job has already been processed",
	)

	ErrCSVInvalid = New(
		http.StatusBadRequest,
		"INVALID_CSV",
		"CSV file is invalid or malformed",
	)
)

/* =========================================================
 * Audit / Metrics
 * ========================================================= */

var (
	ErrAuditLogNotFound = New(
		http.StatusNotFound,
		"AUDIT_LOG_NOT_FOUND",
		"Audit log not found",
	)

	ErrMetricsUnavailable = New(
		http.StatusServiceUnavailable,
		"METRICS_UNAVAILABLE",
		"Metrics service is unavailable",
	)
)
