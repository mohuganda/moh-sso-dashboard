package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/announcements"
	"github.com/moh-sso-dashboard/internal/utils"
)

type AnnouncementService struct {
	repo          repository.AnnouncementRepository
	notifications NotificationsService
	cfg           *config.Config
}

func NewAnnouncementService(
	repo repository.AnnouncementRepository,
	notifications NotificationsService,
	cfg ...*config.Config,
) *AnnouncementService {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &AnnouncementService{
		repo:          repo,
		notifications: notifications,
		cfg:           appConfig,
	}
}

// ---------------------------------
// Core CRUD
// ---------------------------------

func (s *AnnouncementService) CreateAnnouncement(
	ctx context.Context,
	params db.CreateAnnouncementParams,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	item, err := s.repo.Create(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementCreated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q created", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"created_by":      item.CreatedBy.String(),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) GetAnnouncementByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("get announcement by id: %w", err)
	}

	return item, nil
}

func (s *AnnouncementService) UpdateAnnouncement(
	ctx context.Context,
	params db.UpdateAnnouncementParams,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	item, err := s.repo.Update(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("update announcement: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementUpdated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q updated", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"updated_by":      item.UpdatedBy,
		}),
	})

	return item, nil
}

func (s *AnnouncementService) DeleteAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	deletedBy uuid.UUID,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("announcement id is required")
	}

	item, _ := s.repo.GetByID(ctx, id)

	if err := s.repo.SoftDelete(ctx, id, deletedBy); err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}

	nt := models.AnnouncementDeleted
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Announcement deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": id.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"deleted_by":      deletedBy.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"admin-alert",
		"Announcement deleted",
		fmt.Sprintf("Announcement %q was deleted.", item.Title),
		map[string]any{
			"Name":           s.systemAdminName(),
			"Platform":       s.platformName(),
			"Message":        "An announcement was deleted.",
			"AnnouncementID": id.String(),
			"Title":          item.Title,
			"Status":         announcementStatusString(item.Status),
			"ActionURL":      s.adminAnnouncementsURL(),
			"Details": fmt.Sprintf(
				"Announcement ID: %s\nTitle: %s\nStatus: %s\nDeleted By: %s",
				id.String(),
				item.Title,
				announcementStatusString(item.Status),
				deletedBy.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *AnnouncementService) RestoreAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.Restore(ctx, id, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("restore announcement: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementUpdated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q restored", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"updated_by":      updatedBy.String(),
		}),
	})

	return item, nil
}

// ---------------------------------
// Admin listing / search / stats
// ---------------------------------

func (s *AnnouncementService) ListAnnouncementsAdmin(
	ctx context.Context,
	params db.ListAnnouncementsAdminParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListAdmin(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements admin: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountAnnouncementsAdmin(
	ctx context.Context,
) (int64, error) {
	if s == nil {
		return 0, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return 0, errors.New("announcement repository is nil")
	}

	count, err := s.repo.CountAdmin(ctx)
	if err != nil {
		return 0, fmt.Errorf("count announcements admin: %w", err)
	}

	return count, nil
}

func (s *AnnouncementService) ListAnnouncementsByStatus(
	ctx context.Context,
	params db.ListAnnouncementsByStatusParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListByStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements by status: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountAnnouncementsByStatus(
	ctx context.Context,
	status models.AnnouncementStatus,
) (int64, error) {
	if s == nil {
		return 0, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return 0, errors.New("announcement repository is nil")
	}

	count, err := s.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("count announcements by status: %w", err)
	}

	return count, nil
}

func (s *AnnouncementService) SearchAnnouncementsAdmin(
	ctx context.Context,
	params db.SearchAnnouncementsAdminParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.SearchAdmin(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("search announcements admin: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountSearchAnnouncementsAdmin(
	ctx context.Context,
	search string,
) (int64, error) {
	if s == nil {
		return 0, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return 0, errors.New("announcement repository is nil")
	}

	count, err := s.repo.CountSearchAdmin(ctx, strings.TrimSpace(search))
	if err != nil {
		return 0, fmt.Errorf("count search announcements admin: %w", err)
	}

	return count, nil
}

func (s *AnnouncementService) GetAnnouncementStats(
	ctx context.Context,
) (db.GetAnnouncementStatsRow, error) {
	if s == nil {
		return db.GetAnnouncementStatsRow{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.GetAnnouncementStatsRow{}, errors.New("announcement repository is nil")
	}

	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return db.GetAnnouncementStatsRow{}, fmt.Errorf("get announcement stats: %w", err)
	}

	return stats, nil
}

// ---------------------------------
// Public / end-user listing
// ---------------------------------

func (s *AnnouncementService) ListActivePublishedAnnouncements(
	ctx context.Context,
	params db.ListActivePublishedAnnouncementsParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListActivePublished(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list active published announcements: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountActivePublishedAnnouncements(
	ctx context.Context,
) (int64, error) {
	if s == nil {
		return 0, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return 0, errors.New("announcement repository is nil")
	}

	count, err := s.repo.CountActivePublished(ctx)
	if err != nil {
		return 0, fmt.Errorf("count active published announcements: %w", err)
	}

	return count, nil
}

func (s *AnnouncementService) ListAnnouncementsForClient(
	ctx context.Context,
	params db.ListAnnouncementsForClientParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListForClient(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements for client: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) ListAnnouncementsForRole(
	ctx context.Context,
	params db.ListAnnouncementsForRoleParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListForRole(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements for role: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) ListAnnouncementsForUser(
	ctx context.Context,
	params db.ListAnnouncementsForUserParams,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	items, err := s.repo.ListForUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements for user: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) ListMyAnnouncements(
	ctx context.Context,
	userID uuid.UUID,
	roleName string,
	clientID uuid.UUID,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	roleName = strings.TrimSpace(roleName)

	results := make(map[uuid.UUID]db.Announcement)

	activeItems, err := s.repo.ListActivePublished(ctx, db.ListActivePublishedAnnouncementsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list active announcements: %w", err)
	}

	for _, item := range activeItems {
		results[item.ID] = item
	}

	if roleName != "" {
		roleItems, err := s.repo.ListForRole(ctx, db.ListAnnouncementsForRoleParams{
			RoleName:   roleName,
			PageLimit:  limit,
			PageOffset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list role announcements: %w", err)
		}

		for _, item := range roleItems {
			results[item.ID] = item
		}
	}

	if clientID != uuid.Nil {
		clientItems, err := s.repo.ListForClient(ctx, db.ListAnnouncementsForClientParams{
			ClientID: clientID,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list client announcements: %w", err)
		}

		for _, item := range clientItems {
			results[item.ID] = item
		}
	}

	if userID != uuid.Nil {
		userItems, err := s.repo.ListForUser(ctx, db.ListAnnouncementsForUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list user announcements: %w", err)
		}

		for _, item := range userItems {
			results[item.ID] = item
		}
	}

	items := make([]db.Announcement, 0, len(results))
	for _, item := range results {
		items = append(items, item)
	}

	return items, nil
}

func (s *AnnouncementService) ListPublicAnnouncements(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)

	return s.repo.ListPublicAnnouncements(ctx, limit, offset)
}

func normalizeLimit(limit int32) int32 {
	switch {
	case limit <= 0:
		return 20
	case limit > 100:
		return 100
	default:
		return limit
	}
}

func normalizeOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}

	return offset
}

// ---------------------------------
// Lifecycle actions
// ---------------------------------

func (s *AnnouncementService) PublishAnnouncementNow(
	ctx context.Context,
	id uuid.UUID,
	publishedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.PublishNow(ctx, id, publishedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("publish announcement: %w", err)
	}

	nt := models.AnnouncementPublished
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q published", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"published_by":    publishedBy.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"notification",
		"Announcement published",
		fmt.Sprintf("Announcement %q was published.", item.Title),
		map[string]any{
			"Name":           s.systemAdminName(),
			"Platform":       s.platformName(),
			"Title":          item.Title,
			"Message":        fmt.Sprintf("Announcement %q was published.", item.Title),
			"AnnouncementID": item.ID.String(),
			"Status":         announcementStatusString(item.Status),
			"ActionURL":      s.adminAnnouncementsURL(),
			"Details": fmt.Sprintf(
				"Announcement ID: %s\nTitle: %s\nStatus: %s\nPublished By: %s",
				item.ID.String(),
				item.Title,
				announcementStatusString(item.Status),
				publishedBy.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return item, nil
}

func (s *AnnouncementService) MoveAnnouncementToDraft(
	ctx context.Context,
	params db.DraftAnnouncementParams,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	item, err := s.repo.Draft(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("drafted announcement: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementDrafted
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q drafted", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) ScheduleAnnouncement(
	ctx context.Context,
	params db.ScheduleAnnouncementParams,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	item, err := s.repo.Schedule(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("schedule announcement: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementScheduled
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q scheduled", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) ArchiveAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	archivedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.Archive(ctx, id, archivedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("archive announcement: %w", err)
	}

	nt := models.AnnouncementArchived
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q archived", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"archived_by":     archivedBy.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"admin-alert",
		"Announcement archived",
		fmt.Sprintf("Announcement %q was archived.", item.Title),
		map[string]any{
			"Name":           s.systemAdminName(),
			"Platform":       s.platformName(),
			"Message":        fmt.Sprintf("Announcement %q was archived.", item.Title),
			"AnnouncementID": item.ID.String(),
			"Title":          item.Title,
			"Status":         announcementStatusString(item.Status),
			"ActionURL":      s.adminAnnouncementsURL(),
			"Details": fmt.Sprintf(
				"Announcement ID: %s\nTitle: %s\nStatus: %s\nArchived By: %s",
				item.ID.String(),
				item.Title,
				announcementStatusString(item.Status),
				archivedBy.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return item, nil
}

func (s *AnnouncementService) UnarchiveAnnouncementToDraft(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.UnarchiveToDraft(ctx, id, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("unarchive announcement to draft: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementUpdated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q moved back to draft", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"updated_by":      updatedBy.String(),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) UpdateAnnouncementStatus(
	ctx context.Context,
	params db.UpdateAnnouncementStatusParams,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	item, err := s.repo.UpdateStatus(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("update announcement status: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementUpdated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q status updated", item.Title),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) SetAnnouncementPinned(
	ctx context.Context,
	id uuid.UUID,
	isPinned bool,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.SetPinned(ctx, id, isPinned, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("set announcement pinned: %w", err)
	}

	nt := models.AnnouncementUpdated
	msg := fmt.Sprintf("Announcement %q unpinned", item.Title)
	if isPinned {
		msg = fmt.Sprintf("Announcement %q pinned", item.Title)
	}

	// In-app only.
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"status":          announcementStatusString(item.Status),
			"is_pinned":       isPinned,
			"updated_by":      updatedBy.String(),
		}),
	})

	return item, nil
}

func (s *AnnouncementService) SetAnnouncementPriority(
	ctx context.Context,
	id uuid.UUID,
	priority int32,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	if s == nil {
		return db.Announcement{}, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return db.Announcement{}, errors.New("announcement repository is nil")
	}

	if id == uuid.Nil {
		return db.Announcement{}, errors.New("announcement id is required")
	}

	item, err := s.repo.SetPriority(ctx, id, priority, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("set announcement priority: %w", err)
	}

	// In-app only.
	nt := models.AnnouncementUpdated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q priority updated to %d", item.Title, priority),
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"announcement_id": item.ID.String(),
			"title":           item.Title,
			"priority":        priority,
			"updated_by":      updatedBy.String(),
		}),
	})

	return item, nil
}

// ---------------------------------
// Audience management
// ---------------------------------

func (s *AnnouncementService) AddClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	clientID uuid.UUID,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	if clientID == uuid.Nil {
		return errors.New("client id is required")
	}

	if err := s.repo.AddClientAudience(ctx, announcementID, clientID); err != nil {
		return fmt.Errorf("add client audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ReplaceClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	clientIDs []uuid.UUID,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	if err := s.repo.ReplaceClientAudience(ctx, announcementID, clientIDs); err != nil {
		return fmt.Errorf("replace client audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return nil, errors.New("announcement id is required")
	}

	items, err := s.repo.ListClientAudience(ctx, announcementID)
	if err != nil {
		return nil, fmt.Errorf("list client audience: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) AddRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	roleName string,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return errors.New("role name is required")
	}

	if err := s.repo.AddRoleAudience(ctx, announcementID, roleName); err != nil {
		return fmt.Errorf("add role audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ReplaceRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	roleNames []string,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	cleaned := make([]string, 0, len(roleNames))
	for _, role := range roleNames {
		role = strings.TrimSpace(role)
		if role != "" {
			cleaned = append(cleaned, role)
		}
	}

	if err := s.repo.ReplaceRoleAudience(ctx, announcementID, cleaned); err != nil {
		return fmt.Errorf("replace role audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]string, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return nil, errors.New("announcement id is required")
	}

	items, err := s.repo.ListRoleAudience(ctx, announcementID)
	if err != nil {
		return nil, fmt.Errorf("list role audience: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) AddUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	userID uuid.UUID,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	if userID == uuid.Nil {
		return errors.New("user id is required")
	}

	if err := s.repo.AddUserAudience(ctx, announcementID, userID); err != nil {
		return fmt.Errorf("add user audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ReplaceUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	userIDs []uuid.UUID,
) error {
	if s == nil {
		return errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return errors.New("announcement id is required")
	}

	if err := s.repo.ReplaceUserAudience(ctx, announcementID, userIDs); err != nil {
		return fmt.Errorf("replace user audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	if announcementID == uuid.Nil {
		return nil, errors.New("announcement id is required")
	}

	items, err := s.repo.ListUserAudience(ctx, announcementID)
	if err != nil {
		return nil, fmt.Errorf("list user audience: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		fmt.Printf("announcement notification failed type=%s error=%v\n", notification.Type, err)
	}
}

func (s *AnnouncementService) attachAdminEmailDelivery(
	notification *models.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminAnnouncementsURL()
	}

	notification.Deliveries = []models.NotificationDeliveryRequest{
		{
			Channel: models.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: models.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *AnnouncementService) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *AnnouncementService) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *AnnouncementService) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *AnnouncementService) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *AnnouncementService) adminAnnouncementsURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/admin/announcements"
	}

	return base + "/announcements"
}

func announcementStatusString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case models.AnnouncementStatus:
		return strings.TrimSpace(string(v))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
