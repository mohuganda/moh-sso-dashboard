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

// ---------------------------------
// Core CRUD
// ---------------------------------

func (s *AnnouncementService) CreateAnnouncement(
	ctx context.Context,
	params db.CreateAnnouncementParams,
) (db.Announcement, error) {
	item, err := s.repo.Create(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	nt := models.AnnouncementCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q created", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) GetAnnouncementByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Announcement, error) {
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
	item, err := s.repo.Update(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("update announcement: %w", err)
	}

	nt := models.AnnouncementUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q updated", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) DeleteAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	deletedBy uuid.UUID,
) error {
	if err := s.repo.SoftDelete(ctx, id, deletedBy); err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}

	nt := models.AnnouncementDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Announcement deleted",
		TargetRole: "admin",
	})

	return nil
}

func (s *AnnouncementService) RestoreAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := s.repo.Restore(ctx, id, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("restore announcement: %w", err)
	}

	nt := models.AnnouncementUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q restored", item.Title),
		TargetRole: "admin",
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
	items, err := s.repo.ListAdmin(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list announcements admin: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountAnnouncementsAdmin(
	ctx context.Context,
) (int64, error) {
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
	count, err := s.repo.CountSearchAdmin(ctx, search)
	if err != nil {
		return 0, fmt.Errorf("count search announcements admin: %w", err)
	}

	return count, nil
}

func (s *AnnouncementService) GetAnnouncementStats(
	ctx context.Context,
) (db.GetAnnouncementStatsRow, error) {
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
	items, err := s.repo.ListActivePublished(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list active published announcements: %w", err)
	}

	return items, nil
}

func (s *AnnouncementService) CountActivePublishedAnnouncements(
	ctx context.Context,
) (int64, error) {
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
	item, err := s.repo.PublishNow(ctx, id, publishedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("publish announcement: %w", err)
	}

	nt := models.AnnouncementPublished
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q published", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) ScheduleAnnouncement(
	ctx context.Context,
	params db.ScheduleAnnouncementParams,
) (db.Announcement, error) {
	item, err := s.repo.Schedule(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("schedule announcement: %w", err)
	}

	nt := models.AnnouncementScheduled
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q scheduled", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) ArchiveAnnouncement(
	ctx context.Context,
	id uuid.UUID,
	archivedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := s.repo.Archive(ctx, id, archivedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("archive announcement: %w", err)
	}

	nt := models.AnnouncementArchived
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q archived", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) UnarchiveAnnouncementToDraft(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := s.repo.UnarchiveToDraft(ctx, id, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("unarchive announcement to draft: %w", err)
	}

	nt := models.AnnouncementUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q moved back to draft", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) UpdateAnnouncementStatus(
	ctx context.Context,
	params db.UpdateAnnouncementStatusParams,
) (db.Announcement, error) {
	item, err := s.repo.UpdateStatus(ctx, params)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("update announcement status: %w", err)
	}

	nt := models.AnnouncementUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q status updated", item.Title),
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) SetAnnouncementPinned(
	ctx context.Context,
	id uuid.UUID,
	isPinned bool,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := s.repo.SetPinned(ctx, id, isPinned, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("set announcement pinned: %w", err)
	}

	nt := models.AnnouncementUpdated
	msg := fmt.Sprintf("Announcement %q unpinned", item.Title)
	if isPinned {
		msg = fmt.Sprintf("Announcement %q pinned", item.Title)
	}

	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
	})

	return item, nil
}

func (s *AnnouncementService) SetAnnouncementPriority(
	ctx context.Context,
	id uuid.UUID,
	priority int32,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := s.repo.SetPriority(ctx, id, priority, updatedBy)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("set announcement priority: %w", err)
	}

	nt := models.AnnouncementUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    fmt.Sprintf("Announcement %q priority updated to %d", item.Title, priority),
		TargetRole: "admin",
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
	if err := s.repo.ReplaceClientAudience(ctx, announcementID, clientIDs); err != nil {
		return fmt.Errorf("replace client audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
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
	if err := s.repo.ReplaceRoleAudience(ctx, announcementID, roleNames); err != nil {
		return fmt.Errorf("replace role audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]string, error) {
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
	if err := s.repo.ReplaceUserAudience(ctx, announcementID, userIDs); err != nil {
		return fmt.Errorf("replace user audience: %w", err)
	}

	return nil
}

func (s *AnnouncementService) ListUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	items, err := s.repo.ListUserAudience(ctx, announcementID)
	if err != nil {
		return nil, fmt.Errorf("list user audience: %w", err)
	}

	return items, nil
}
