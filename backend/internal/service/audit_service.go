package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/cache"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
	"github.com/sqlc-dev/pqtype"
)

type AuditService struct {
	store         db.Store
	cache         *cache.RedisCache
	notifications NotificationsService
}

func NewAuditService(
	store db.Store,
	cache *cache.RedisCache,
	notifications ...NotificationsService,
) *AuditService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &AuditService{
		store:         store,
		cache:         cache,
		notifications: notificationSvc,
	}
}

//
// ----------------------------------------------------
// Internal helpers
// ----------------------------------------------------
//

// Ensures we never violate FK constraints.
// If the user does not exist locally, we downgrade to NULL (system).
func (a *AuditService) safeUserID(
	ctx context.Context,
	userID uuid.NullUUID,
) uuid.NullUUID {
	if a == nil || a.store == nil {
		return uuid.NullUUID{}
	}

	// Anonymous / system event
	if !userID.Valid {
		return userID
	}

	exists, err := a.store.UserExists(ctx, userID.UUID)
	if err != nil || !exists {
		return uuid.NullUUID{}
	}

	return userID
}

func (a *AuditService) write(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) error {
	if a == nil {
		return errors.New("audit service is nil")
	}

	if a.store == nil {
		return errors.New("audit store is nil")
	}

	action = strings.TrimSpace(action)
	if action == "" {
		return errors.New("audit action is required")
	}

	safeID := a.safeUserID(ctx, userID)

	if err := a.store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID: safeID,
		Action: action,
		Metadata: pqtype.NullRawMessage{
			RawMessage: utils.Encode(metadata),
			Valid:      metadata != nil,
		},
	}); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	if a.cache != nil {
		_ = a.cache.DeletePattern(ctx, "audit_logs:*")
		_ = a.cache.DeletePattern(ctx, "audit_actions:*")
		_ = a.cache.DeletePattern(ctx, "audit_metrics:*")
	}

	a.notifyForAuditEvent(ctx, safeID, action, metadata)

	return nil
}

func (a *AuditService) notifyForAuditEvent(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) {
	if a == nil || a.notifications == nil {
		return
	}

	notification := auditNotificationFromAction(userID, action, metadata)
	if notification == nil {
		return
	}

	if _, err := a.notifications.Notify(ctx, *notification); err != nil {
		fmt.Printf("audit notification failed action=%s error=%v\n", action, err)
	}
}

func auditNotificationFromAction(
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) *model.Notification {
	switch action {
	case "auth.login":
		success, _ := metadataBool(metadata, "success")
		if success {
			return nil
		}

		return &model.Notification{
			Type:       "AUTH_LOGIN_FAILED",
			Title:      "Failed login attempt",
			Severity:   "warning",
			Message:    "A failed login attempt was recorded",
			TargetRole: "admin",
			UserID:     nullUUIDStringValue(userID),
			Metadata: utils.MustJSON(map[string]any{
				"action":     action,
				"user_id":    nullUUIDStringValue(userID),
				"client_id":  metadataString(metadata, "client_id"),
				"ip":         metadataString(metadata, "ip"),
				"user_agent": metadataString(metadata, "user_agent"),
				"country":    metadataString(metadata, "country"),
				"city":       metadataString(metadata, "city"),
				"success":    success,
			}),
		}

	case "auth.refresh_failed":
		return &model.Notification{
			Type:       "AUTH_TOKEN_REFRESH_FAILED",
			Title:      "Token refresh failed",
			Severity:   "warning",
			Message:    "A token refresh attempt failed",
			TargetRole: "admin",
			UserID:     nullUUIDStringValue(userID),
			Metadata: utils.MustJSON(map[string]any{
				"action":     action,
				"user_id":    nullUUIDStringValue(userID),
				"ip":         metadataString(metadata, "ip"),
				"user_agent": metadataString(metadata, "user_agent"),
				"success":    false,
			}),
		}

	case "auth.password_reset":
		success, _ := metadataBool(metadata, "success")
		severity := "info"
		title := "Password reset recorded"
		message := "A password reset event was recorded"

		if !success {
			severity = "warning"
			title = "Password reset failed"
			message = "A password reset attempt failed"
		}

		return &model.Notification{
			Type:       "AUTH_PASSWORD_RESET",
			Title:      title,
			Severity:   severity,
			Message:    message,
			TargetRole: "admin",
			UserID:     nullUUIDStringValue(userID),
			Metadata: utils.MustJSON(map[string]any{
				"action":  action,
				"user_id": nullUUIDStringValue(userID),
				"success": success,
				"ip":      metadataString(metadata, "ip"),
			}),
		}

	case "client.create":
		return &model.Notification{
			Type:       "CLIENT_CREATED",
			Title:      "Client created",
			Severity:   "info",
			Message:    "Client application created",
			TargetRole: "admin",
			UserID:     nullUUIDStringValue(userID),
			ClientID:   metadataString(metadata, "client_id"),
			Metadata: utils.MustJSON(map[string]any{
				"action":     action,
				"admin_id":   nullUUIDStringValue(userID),
				"client_id":  metadataString(metadata, "client_id"),
				"ip":         metadataString(metadata, "ip"),
				"user_agent": metadataString(metadata, "user_agent"),
			}),
		}

	case "client.delete":
		return &model.Notification{
			Type:       "CLIENT_DELETED",
			Title:      "Client deleted",
			Severity:   "critical",
			Message:    "Client application deleted",
			TargetRole: "admin",
			UserID:     nullUUIDStringValue(userID),
			ClientID:   metadataString(metadata, "client_id"),
			Metadata: utils.MustJSON(map[string]any{
				"action":     action,
				"admin_id":   nullUUIDStringValue(userID),
				"client_id":  metadataString(metadata, "client_id"),
				"ip":         metadataString(metadata, "ip"),
				"user_agent": metadataString(metadata, "user_agent"),
			}),
		}

	default:
		if isSensitiveAdminAction(action) {
			return &model.Notification{
				Type:       "ADMIN_ACTION_RECORDED",
				Title:      "Admin action recorded",
				Severity:   adminActionSeverity(action),
				Message:    "A sensitive admin action was recorded",
				TargetRole: "admin",
				UserID:     nullUUIDStringValue(userID),
				Metadata: utils.MustJSON(map[string]any{
					"action":   action,
					"admin_id": nullUUIDStringValue(userID),
					"metadata": metadata,
				}),
			}
		}

		return nil
	}
}

func isSensitiveAdminAction(action string) bool {
	action = strings.ToLower(strings.TrimSpace(action))

	if action == "" {
		return false
	}

	sensitivePrefixes := []string{
		"user.",
		"role.",
		"client.",
		"storage.",
		"document.delete",
		"template.delete",
		"template.archive",
		"surveillance.",
		"system.",
		"backup.",
	}

	for _, prefix := range sensitivePrefixes {
		if strings.HasPrefix(action, prefix) {
			return true
		}
	}

	return false
}

func adminActionSeverity(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))

	switch {
	case strings.Contains(action, "delete"):
		return "critical"
	case strings.Contains(action, "disable"):
		return "warning"
	case strings.Contains(action, "archive"):
		return "warning"
	case strings.Contains(action, "failed"):
		return "warning"
	default:
		return "info"
	}
}

func metadataString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}

	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}

	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func metadataBool(metadata map[string]interface{}, key string) (bool, bool) {
	if metadata == nil {
		return false, false
	}

	value, ok := metadata[key]
	if !ok || value == nil {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "y":
			return true, true
		case "false", "0", "no", "n":
			return false, true
		default:
			return false, false
		}
	default:
		switch strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value))) {
		case "true", "1", "yes", "y":
			return true, true
		case "false", "0", "no", "n":
			return false, true
		default:
			return false, false
		}
	}
}

func nullUUIDStringValue(value uuid.NullUUID) string {
	if !value.Valid {
		return ""
	}

	return value.UUID.String()
}

//
// ----------------------------------------------------
// Public API (handlers use ONLY these)
// ----------------------------------------------------
//

// Generic logger (fallback)
func (a *AuditService) Log(
	ctx context.Context,
	userID uuid.NullUUID,
	action string,
	metadata map[string]interface{},
) error {
	return a.write(ctx, userID, action, metadata)
}

// --------------------
// AUTH EVENTS
// --------------------

func (a *AuditService) LoginInitiated(
	ctx context.Context,
	ip string,
	userAgent string,
	clientID string,
) error {
	return a.write(ctx, uuid.NullUUID{}, "auth.login_initiated", map[string]interface{}{
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
		"client_id":  strings.TrimSpace(clientID),
	})
}

func (a *AuditService) LoginResult(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	clientID string,
	ip string,
	userAgent string,
	country string,
	city string,
) error {
	return a.write(ctx, userID, "auth.login", map[string]interface{}{
		"success":    success,
		"client_id":  strings.TrimSpace(clientID),
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
		"country":    strings.TrimSpace(country),
		"city":       strings.TrimSpace(city),
	})
}

func (a *AuditService) TokenRefresh(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	ip string,
	userAgent string,
) error {
	action := "auth.refresh_success"
	if !success {
		action = "auth.refresh_failed"
	}

	return a.write(ctx, userID, action, map[string]interface{}{
		"success":    success,
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
	})
}

func (a *AuditService) Logout(
	ctx context.Context,
	userID uuid.NullUUID,
	ip string,
	userAgent string,
) error {
	return a.write(ctx, userID, "auth.logout", map[string]interface{}{
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
	})
}

func (a *AuditService) PasswordReset(
	ctx context.Context,
	userID uuid.NullUUID,
	success bool,
	ip string,
) error {
	return a.write(ctx, userID, "auth.password_reset", map[string]interface{}{
		"success": success,
		"ip":      strings.TrimSpace(ip),
	})
}

// --------------------
// CLIENT / ADMIN EVENTS
// --------------------

func (a *AuditService) ClientCreated(
	ctx context.Context,
	adminID uuid.NullUUID,
	clientID string,
	ip string,
	userAgent string,
) error {
	return a.write(ctx, adminID, "client.create", map[string]interface{}{
		"client_id":  strings.TrimSpace(clientID),
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
	})
}

func (a *AuditService) ClientDeleted(
	ctx context.Context,
	adminID uuid.NullUUID,
	clientID string,
	ip string,
	userAgent string,
) error {
	return a.write(ctx, adminID, "client.delete", map[string]interface{}{
		"client_id":  strings.TrimSpace(clientID),
		"ip":         strings.TrimSpace(ip),
		"user_agent": strings.TrimSpace(userAgent),
	})
}

func (a *AuditService) AdminAction(
	ctx context.Context,
	adminID uuid.NullUUID,
	action string, // e.g. user.disable, role.assign
	metadata map[string]interface{},
) error {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	return a.write(ctx, adminID, strings.TrimSpace(action), metadata)
}
