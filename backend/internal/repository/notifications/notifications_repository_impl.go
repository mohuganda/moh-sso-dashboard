package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/sqlc-dev/pqtype"
)

type notificationsRepository struct {
	db     db.Store
	logger *logger.Logger
}

func NewNotificationsRepository(
	db db.Store,
	log logger.Logger,
) NotificationsRepository {
	return &notificationsRepository{
		db:     db,
		logger: &log,
	}
}

func (r *notificationsRepository) Notify(
	ctx context.Context,
	notification model.Notification,
) (*model.Notification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("notifications repository or database is nil")
	}

	params := db.CreateNotificationParams{
		Type:    notification.Type,
		Title:   notification.Title,
		Message: notification.Message,
		Severity: sql.NullString{
			String: notification.Severity,
			Valid:  notification.Severity != "",
		},
		TargetRole: notification.TargetRole,
		Metadata: pqtype.NullRawMessage{
			RawMessage: notification.Metadata,
			Valid:      notification.Metadata != nil,
		},
	}

	n, err := r.db.CreateNotification(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return mapDBNotification(n), nil
}

func (r *notificationsRepository) ListNotifications(
	ctx context.Context,
	targetRole string,
	unread *bool,
	offset, limit int32,
) ([]model.Notification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("notifications repository or database is nil")
	}

	params := db.ListNotificationsParams{
		TargetRole: targetRole,
		Limit:      limit,
		Offset:     offset,
		Unread: sql.NullBool{
			Valid: unread != nil,
			Bool:  unread != nil && *unread,
		},
	}

	rows, err := r.db.ListNotifications(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	out := make([]model.Notification, 0, len(rows))
	for _, n := range rows {
		out = append(out, *mapDBNotification(n))
	}

	return out, nil
}

func (r *notificationsRepository) GetNotificationByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Notification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("notifications repository or database is nil")
	}

	n, err := r.db.GetNotificationByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification by id: %w", err)
	}

	return mapDBNotification(n), nil
}

func (r *notificationsRepository) MarkNotificationAsRead(
	ctx context.Context,
	id uuid.UUID,
) error {
	if r == nil || r.db == nil {
		return errors.New("notifications repository or database is nil")
	}
	return r.db.MarkNotificationRead(ctx, id)
}

func (r *notificationsRepository) DeleteNotification(
	ctx context.Context,
	id uuid.UUID,
) error {
	if r == nil || r.db == nil {
		return errors.New("notifications repository or database is nil")
	}
	return r.db.DeleteNotificationByID(ctx, id)
}

func (r *notificationsRepository) DeleteOldNotifications(
	ctx context.Context,
) error {
	if r == nil || r.db == nil {
		return errors.New("notifications repository or database is nil")
	}
	return r.db.DeleteOldNotifications(ctx)
}

func (r *notificationsRepository) CountNotifications(
	ctx context.Context,
	targetRole string,
) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("notifications repository or database is nil")
	}
	return r.db.CountNotifications(ctx, targetRole)
}

func (r *notificationsRepository) CountUnreadNotifications(
	ctx context.Context,
	targetRole string,
) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("notifications repository or database is nil")
	}
	return r.db.CountUnreadNotifications(ctx, targetRole)
}

func mapDBNotification(n db.Notification) *model.Notification {
	var metadata json.RawMessage
	if n.Metadata.Valid {
		metadata = n.Metadata.RawMessage
	}

	return &model.Notification{
		ID:         n.ID,
		Type:       n.Type,
		Title:      n.Title,
		Message:    n.Message,
		Severity:   n.Severity.String,
		TargetRole: n.TargetRole,
		Metadata:   metadata,
		Read:       n.Read.Bool,
		CreatedAt:  n.CreatedAt.Time,
	}
}
