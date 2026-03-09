package announcements

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type AnnouncementRepository interface {
	List(ctx context.Context, limit int32) ([]db.ListAnnouncementsRow, error)

	Create(
		ctx context.Context,
		params db.CreateAnnouncementParams,
	) (db.Announcement, error)

	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error
}
