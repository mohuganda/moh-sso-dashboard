package ratelimit

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	PolicyAuthenticatedDefault = "authenticated-default"
	PolicyAuthLogin            = "auth-login"
	PolicyAuthCallback         = "auth-callback"
	PolicyAuthRefresh          = "auth-refresh"
	PolicyAuditExport          = "audit-export"
	PolicyUsersEmailAction     = "users-email-action"
	PolicyUsersPasswordAction  = "users-password-action"
	PolicyEmailSend            = "email-send"
	PolicyEmailManage          = "email-manage"
	PolicyDocumentsWrite       = "documents-write"
	PolicyDocumentsProcess     = "documents-process"
	PolicySurveillanceImport   = "surveillance-import"
	PolicySurveillanceManage   = "surveillance-manage"
	PolicyRBACSensitiveWrite   = "rbac-sensitive-write"
	PolicyRBACBulkWrite        = "rbac-bulk-write"
	PolicyNotificationsWrite   = "notifications-write"
	PolicyAnnouncementsWrite   = "announcements-write"
	PolicyAnnouncementsPublish = "announcements-publish"
)

func PerUserPolicy(name string, limit int) Policy {
	return Policy{
		Name:    name,
		Limit:   limit,
		Window:  time.Minute,
		KeyFunc: ByUser,
	}
}

func PerIPPolicy(name string, limit int) Policy {
	return Policy{
		Name:    name,
		Limit:   limit,
		Window:  time.Minute,
		KeyFunc: ByIP,
	}
}

func LoginPolicy(limit int) Policy {
	return Policy{
		Name:    PolicyAuthLogin,
		Limit:   limit,
		Window:  time.Minute,
		KeyFunc: ByIPAndUsername,
	}
}

func AuthenticatedDefaultPolicy(limit int) Policy {
	return PerUserPolicy(PolicyAuthenticatedDefault, limit)
}

func AuditExportPolicy() Policy {
	return PerUserPolicy(PolicyAuditExport, 10)
}

func UsersEmailActionPolicy() Policy {
	return PerUserPolicy(PolicyUsersEmailAction, 20)
}

func UsersPasswordActionPolicy() Policy {
	return PerUserPolicy(PolicyUsersPasswordAction, 10)
}

func EmailSendPolicy() Policy {
	return PerUserPolicy(PolicyEmailSend, 30)
}

func EmailManagePolicy() Policy {
	return PerUserPolicy(PolicyEmailManage, 30)
}

func DocumentsWritePolicy() Policy {
	return PerUserPolicy(PolicyDocumentsWrite, 30)
}

func DocumentsProcessPolicy() Policy {
	return PerUserPolicy(PolicyDocumentsProcess, 10)
}

func SurveillanceImportPolicy() Policy {
	return PerUserPolicy(PolicySurveillanceImport, 10)
}

func SurveillanceManagePolicy() Policy {
	return PerUserPolicy(PolicySurveillanceManage, 10)
}

func RBACSensitiveWritePolicy() Policy {
	return PerUserPolicy(PolicyRBACSensitiveWrite, 20)
}

func RBACBulkWritePolicy() Policy {
	return PerUserPolicy(PolicyRBACBulkWrite, 10)
}

func NotificationsWritePolicy() Policy {
	return PerUserPolicy(PolicyNotificationsWrite, 30)
}

func AnnouncementsWritePolicy() Policy {
	return PerUserPolicy(PolicyAnnouncementsWrite, 60)
}

func AnnouncementsPublishPolicy() Policy {
	return PerUserPolicy(PolicyAnnouncementsPublish, 30)
}

func MiddlewareForUserPolicy(limiter *Limiter, name string, limit int) gin.HandlerFunc {
	return MiddlewareForPolicy(limiter, PerUserPolicy(name, limit))
}
