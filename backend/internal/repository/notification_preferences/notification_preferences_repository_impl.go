package notification_preferences

import (
	"context"
	"database/sql"
	"errors"

	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type repository struct {
	db *sql.DB
}

func NewNotificationPreferencesRepository(store storepkg.Store) NotificationPreferencesRepository {
	var primaryDB *sql.DB
	if sqlStore, ok := store.(*storepkg.SQLStore); ok && sqlStore != nil {
		primaryDB = sqlStore.DB()
	}

	return &repository{db: primaryDB}
}

func (r *repository) GetByUserID(ctx context.Context, userID string) (NotificationPreference, error) {
	if r == nil || r.db == nil {
		return NotificationPreference{}, errors.New("notification preferences database is nil")
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			email_enabled,
			sms_enabled,
			phone_number,
			phone_verified,
			quiet_hours_start,
			quiet_hours_end,
			created_at,
			updated_at
		FROM notification_preferences
		WHERE user_id = $1
		LIMIT 1
	`, userID)

	return scanNotificationPreference(row)
}

func (r *repository) Upsert(
	ctx context.Context,
	arg UpsertNotificationPreferencesParams,
) (NotificationPreference, error) {
	if r == nil || r.db == nil {
		return NotificationPreference{}, errors.New("notification preferences database is nil")
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO notification_preferences (
			user_id,
			email_enabled,
			sms_enabled,
			phone_number,
			phone_verified,
			quiet_hours_start,
			quiet_hours_end
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		ON CONFLICT (user_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			sms_enabled = EXCLUDED.sms_enabled,
			phone_number = EXCLUDED.phone_number,
			phone_verified = CASE
				WHEN COALESCE(notification_preferences.phone_number, '') IS DISTINCT FROM COALESCE(EXCLUDED.phone_number, '')
				THEN FALSE
				ELSE notification_preferences.phone_verified
			END,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			updated_at = now()
		RETURNING
			id,
			user_id,
			email_enabled,
			sms_enabled,
			phone_number,
			phone_verified,
			quiet_hours_start,
			quiet_hours_end,
			created_at,
			updated_at
	`,
		arg.UserID,
		arg.EmailEnabled,
		arg.SMSEnabled,
		arg.PhoneNumber,
		arg.PhoneVerified,
		arg.QuietHoursStart,
		arg.QuietHoursEnd,
	)

	return scanNotificationPreference(row)
}

func (r *repository) ListForSMS(
	ctx context.Context,
	arg ListNotificationPreferencesForSMSParams,
) ([]NotificationPreference, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("notification preferences database is nil")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			email_enabled,
			sms_enabled,
			phone_number,
			phone_verified,
			quiet_hours_start,
			quiet_hours_end,
			created_at,
			updated_at
		FROM notification_preferences
		WHERE sms_enabled = TRUE
			AND phone_number IS NOT NULL
			AND trim(phone_number) <> ''
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	preferences := make([]NotificationPreference, 0)
	for rows.Next() {
		preference, err := scanNotificationPreference(rows)
		if err != nil {
			return nil, err
		}
		preferences = append(preferences, preference)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return preferences, nil
}

type notificationPreferenceScanner interface {
	Scan(dest ...any) error
}

func scanNotificationPreference(scanner notificationPreferenceScanner) (NotificationPreference, error) {
	var preference NotificationPreference
	err := scanner.Scan(
		&preference.ID,
		&preference.UserID,
		&preference.EmailEnabled,
		&preference.SMSEnabled,
		&preference.PhoneNumber,
		&preference.PhoneVerified,
		&preference.QuietHoursStart,
		&preference.QuietHoursEnd,
		&preference.CreatedAt,
		&preference.UpdatedAt,
	)
	if err != nil {
		return NotificationPreference{}, err
	}

	return preference, nil
}
