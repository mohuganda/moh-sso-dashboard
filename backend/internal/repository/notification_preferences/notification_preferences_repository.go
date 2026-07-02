package notification_preferences

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type NotificationPreference struct {
	ID              uuid.UUID
	UserID          string
	EmailEnabled    bool
	SMSEnabled      bool
	PhoneNumber     sql.NullString
	PhoneVerified   bool
	QuietHoursStart sql.NullString
	QuietHoursEnd   sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpsertNotificationPreferencesParams struct {
	UserID          string
	EmailEnabled    bool
	SMSEnabled      bool
	PhoneNumber     sql.NullString
	PhoneVerified   bool
	QuietHoursStart sql.NullString
	QuietHoursEnd   sql.NullString
}

type ListNotificationPreferencesForSMSParams struct {
	Limit  int32
	Offset int32
}

type NotificationPreferencesRepository interface {
	GetByUserID(ctx context.Context, userID string) (NotificationPreference, error)
	Upsert(ctx context.Context, arg UpsertNotificationPreferencesParams) (NotificationPreference, error)
	ListForSMS(ctx context.Context, arg ListNotificationPreferencesForSMSParams) ([]NotificationPreference, error)
}
