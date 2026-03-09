package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/announcements"
)

type AnnouncementService struct {
	repo          repository.AnnouncementRepository
	notifications NotificationsService
}

func NewAnnouncementService(
	repo repository.AnnouncementRepository,
	notifications NotificationsService,
) *AnnouncementService {
	return &AnnouncementService{
		repo:          repo,
		notifications: notifications,
	}
}

func (s *AnnouncementService) ListAnnouncements(
	ctx context.Context,
	limit int32,
) ([]db.ListAnnouncementsRow, error) {

	items, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CreateAnnouncement(
	ctx context.Context,
	params db.CreateAnnouncementParams,
) (db.Announcement, error) {

	item, err := s.repo.Create(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}
	// Notify users about new announcement
	nt := models.AnnouncementCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New announcement created",
		TargetRole: "admin",
	})

	return item, nil
}
func (s *AnnouncementService) DeleteAnnouncement(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}

	nt := models.AnnouncementDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New announcement deleted",
		TargetRole: "admin",
	})

	return nil
}
