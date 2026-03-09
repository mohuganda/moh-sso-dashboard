package announcements

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type announcementsRepository struct {
	db     db.Store
	logger *logger.Logger
}

func NewAnnouncementRepository(
	db db.Store,
	log logger.Logger) AnnouncementRepository {
	return &announcementsRepository{
		db:     db,
		logger: &log,
	}
}

func (r *announcementsRepository) List(ctx context.Context, limit int32) ([]db.ListAnnouncementsRow, error) {
	items, err := r.db.ListAnnouncements(ctx, limit)
	if err != nil {
		r.logger.Error("failed to list announcements", err)
		return nil, fmt.Errorf("list announcements: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) Create(
	ctx context.Context,
	params db.CreateAnnouncementParams,
) (db.Announcement, error) {

	item, err := r.db.CreateAnnouncement(ctx, params)
	if err != nil {
		r.logger.Error("failed to create announcement", err)
		return db.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) Delete(
	ctx context.Context,
	announcementID uuid.UUID,
) error {

	err := r.db.DeleteAnnouncement(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to delete announcement", err)
		return fmt.Errorf("delete announcement: %w", err)
	}

	return nil
}
