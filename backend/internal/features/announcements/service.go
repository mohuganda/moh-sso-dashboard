package announcements

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	emailutil "github.com/moh-sso-dashboard/internal/email"
	userRepository "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/utils"
)

type Service struct {
	repo          Repository
	userRepo      userRepository.UserRepository
	notifications sharedservice.NotificationsService
	storage       storage.Storage
	cfg           *config.Config
}

type AnnouncementEmailOptions struct {
	Attachments               []models.Attachment
	AttachmentLinks           []AnnouncementEmailAttachmentLink
	IncludeAttachmentsInEmail bool
	ScheduledAt               *time.Time
}

type AnnouncementEmailAttachmentLink struct {
	FileName    string
	ContentType string
	FileSize    int64
	URL         string
}

func NewService(
	repo Repository,
	userRepo userRepository.UserRepository,
	notifications sharedservice.NotificationsService,
	fileStorage storage.Storage,
	cfg ...*config.Config,
) *Service {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &Service{
		repo:          repo,
		userRepo:      userRepo,
		notifications: notifications,
		storage:       fileStorage,
		cfg:           appConfig,
	}
}

// ---------------------------------
// Core CRUD
// ---------------------------------

func (s *Service) CreateAnnouncement(
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
			"notify_by_email": item.NotifyByEmail,
		}),
	})

	return item, nil
}

func (s *Service) CreateAnnouncementFromInput(
	ctx context.Context,
	input CreateAnnouncementInput,
) (db.Announcement, error) {
	return s.CreateAnnouncement(ctx, db.CreateAnnouncementParams{
		Title:         strings.TrimSpace(input.Title),
		Message:       strings.TrimSpace(input.Message),
		Summary:       input.Summary,
		Level:         dbAnnouncementLevel(input.Level),
		Tag:           input.Tag,
		LinkUrl:       input.LinkURL,
		Priority:      input.Priority,
		IsPinned:      input.IsPinned,
		Status:        dbAnnouncementStatus(input.Status),
		PublishAt:     input.PublishAt,
		ExpiresAt:     input.ExpiresAt,
		AudienceType:  dbAnnouncementAudienceType(input.AudienceType),
		NotifyByEmail: input.NotifyByEmail,
		CreatedBy:     input.CreatedBy,
	})
}

func (s *Service) GetAnnouncementByID(
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

func (s *Service) UpdateAnnouncement(
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
			"notify_by_email": item.NotifyByEmail,
		}),
	})

	return item, nil
}

func (s *Service) UpdateAnnouncementFromInput(
	ctx context.Context,
	input UpdateAnnouncementInput,
) (db.Announcement, error) {
	return s.UpdateAnnouncement(ctx, db.UpdateAnnouncementParams{
		ID:            input.ID,
		Title:         strings.TrimSpace(input.Title),
		Message:       strings.TrimSpace(input.Message),
		Summary:       input.Summary,
		Level:         dbAnnouncementLevel(input.Level),
		Tag:           input.Tag,
		LinkUrl:       input.LinkURL,
		Priority:      input.Priority,
		IsPinned:      input.IsPinned,
		PublishAt:     input.PublishAt,
		ExpiresAt:     input.ExpiresAt,
		AudienceType:  dbAnnouncementAudienceType(input.AudienceType),
		NotifyByEmail: input.NotifyByEmail,
		UpdatedBy: uuid.NullUUID{
			UUID:  input.UpdatedBy,
			Valid: input.UpdatedBy != uuid.Nil,
		},
	})
}

func (s *Service) DeleteAnnouncement(
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

func (s *Service) RestoreAnnouncement(
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

func (s *Service) ListAnnouncementsAdmin(
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

func (s *Service) ListAnnouncementsAdminPage(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return s.ListAnnouncementsAdmin(ctx, db.ListAnnouncementsAdminParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Service) CountAnnouncementsAdmin(
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

func (s *Service) ListAnnouncementsByStatus(
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

func (s *Service) CountAnnouncementsByStatus(
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

func (s *Service) SearchAnnouncementsAdmin(
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

func (s *Service) CountSearchAnnouncementsAdmin(
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

func (s *Service) GetAnnouncementStats(
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

func (s *Service) ListActivePublishedAnnouncements(
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

func (s *Service) ListActivePublishedAnnouncementsPage(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return s.ListActivePublishedAnnouncements(ctx, db.ListActivePublishedAnnouncementsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Service) CountActivePublishedAnnouncements(
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

func (s *Service) ListAnnouncementsForClient(
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

func (s *Service) ListAnnouncementsForClientPage(
	ctx context.Context,
	clientID uuid.UUID,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return s.ListAnnouncementsForClient(ctx, db.ListAnnouncementsForClientParams{
		ClientID: clientID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *Service) ListAnnouncementsForRole(
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

func (s *Service) ListAnnouncementsForRolePage(
	ctx context.Context,
	roleName string,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return s.ListAnnouncementsForRole(ctx, db.ListAnnouncementsForRoleParams{
		RoleName:   strings.TrimSpace(roleName),
		PageLimit:  limit,
		PageOffset: offset,
	})
}

func (s *Service) ListAnnouncementsForUser(
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

func (s *Service) ListAnnouncementsForUserPage(
	ctx context.Context,
	userID uuid.UUID,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return s.ListAnnouncementsForUser(ctx, db.ListAnnouncementsForUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Service) ListMyAnnouncements(
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

func (s *Service) ListPublicAnnouncements(
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

func (s *Service) PublishAnnouncementNow(
	ctx context.Context,
	id uuid.UUID,
	publishedBy uuid.UUID,
	options ...AnnouncementEmailOptions,
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
			"announcement_id":                    item.ID.String(),
			"title":                              item.Title,
			"status":                             announcementStatusString(item.Status),
			"published_by":                       publishedBy.String(),
			"notify_by_email":                    item.NotifyByEmail,
			"email_notification_sent_at_present": item.EmailNotificationSentAt.Valid,
		}),
	}

	if s.shouldSendAnnouncementEmail(item) {
		recipients, err := s.resolveAnnouncementEmailRecipients(ctx, item)
		if err != nil {
			return item, fmt.Errorf("resolve announcement email recipients: %w", err)
		}

		if len(recipients) > 0 {
			emailOptions, err := s.normalizeAnnouncementEmailOptions(options...)
			if err != nil {
				return item, err
			}
			emailOptions, err = s.withPersistedAnnouncementEmailAttachments(ctx, item.ID, emailOptions)
			if err != nil {
				return item, err
			}
			s.attachAnnouncementEmailDelivery(&notification, item, recipients, emailOptions)

			markedItem, err := s.repo.MarkEmailNotificationSent(ctx, item.ID)
			if err != nil {
				return item, fmt.Errorf("mark announcement email notification sent: %w", err)
			}

			item = markedItem
		}
	}

	s.notify(ctx, notification)

	return item, nil
}

func (s *Service) MoveAnnouncementToDraft(
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

func (s *Service) MoveAnnouncementToDraftByUser(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	return s.MoveAnnouncementToDraft(ctx, db.DraftAnnouncementParams{
		ID: id,
		UpdatedBy: uuid.NullUUID{
			UUID:  updatedBy,
			Valid: updatedBy != uuid.Nil,
		},
	})
}

func (s *Service) ScheduleAnnouncement(
	ctx context.Context,
	params db.ScheduleAnnouncementParams,
	options ...AnnouncementEmailOptions,
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

	nt := models.AnnouncementScheduled
	notification := models.Notification{
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
	}

	if s.shouldSendAnnouncementEmail(item) {
		recipients, err := s.resolveAnnouncementEmailRecipients(ctx, item)
		if err != nil {
			return item, fmt.Errorf("resolve announcement email recipients: %w", err)
		}

		if len(recipients) > 0 {
			emailOptions, err := s.normalizeAnnouncementEmailOptions(options...)
			if err != nil {
				return item, err
			}
			emailOptions, err = s.withPersistedAnnouncementEmailAttachments(ctx, item.ID, emailOptions)
			if err != nil {
				return item, err
			}
			if emailOptions.ScheduledAt == nil && item.PublishAt.Valid {
				publishAt := item.PublishAt.Time
				emailOptions.ScheduledAt = &publishAt
			}
			s.attachAnnouncementEmailDelivery(&notification, item, recipients, emailOptions)

			markedItem, err := s.repo.MarkEmailNotificationSent(ctx, item.ID)
			if err != nil {
				return item, fmt.Errorf("mark announcement email notification scheduled: %w", err)
			}

			item = markedItem
		}
	}

	s.notify(ctx, notification)

	return item, nil
}

func (s *Service) ScheduleAnnouncementByUser(
	ctx context.Context,
	id uuid.UUID,
	publishAt time.Time,
	updatedBy uuid.UUID,
	options ...AnnouncementEmailOptions,
) (db.Announcement, error) {
	return s.ScheduleAnnouncement(
		ctx,
		db.ScheduleAnnouncementParams{
			ID: id,
			PublishAt: sql.NullTime{
				Time:  publishAt,
				Valid: true,
			},
			UpdatedBy: uuid.NullUUID{
				UUID:  updatedBy,
				Valid: updatedBy != uuid.Nil,
			},
		},
		options...,
	)
}

func (s *Service) ArchiveAnnouncement(
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

func (s *Service) UnarchiveAnnouncementToDraft(
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

func (s *Service) UpdateAnnouncementStatus(
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

func (s *Service) SetAnnouncementPinned(
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

func (s *Service) SetAnnouncementPriority(
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

func (s *Service) AddClientAudience(
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

func (s *Service) ReplaceClientAudience(
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

func (s *Service) ListClientAudience(
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

func (s *Service) AddRoleAudience(
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

func (s *Service) ReplaceRoleAudience(
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

func (s *Service) ListRoleAudience(
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

func (s *Service) AddUserAudience(
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

func (s *Service) ReplaceUserAudience(
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

func (s *Service) ListUserAudience(
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

// ---------------------------------
// Announcement email recipient resolution
// ---------------------------------

func (s *Service) resolveAnnouncementEmailRecipients(
	ctx context.Context,
	item db.Announcement,
) ([]AnnouncementEmailRecipient, error) {
	if s == nil {
		return nil, errors.New("announcement service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("announcement repository is nil")
	}

	if s.userRepo == nil {
		return nil, errors.New("user repository is nil")
	}

	audienceType := announcementAudienceTypeString(item.AudienceType)

	switch audienceType {
	case "ALL_USERS":
		return s.resolveAllUsersEmailRecipients(ctx)

	case "ADMINS_ONLY":
		return s.resolveUsersByRealmRoles(ctx, []string{"admin"})

	case "SPECIFIC_ROLES":
		roleNames, err := s.repo.ListRoleAudience(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("list announcement role audience: %w", err)
		}

		return s.resolveUsersByRealmRoles(ctx, roleNames)

	case "SPECIFIC_USERS":
		userIDs, err := s.repo.ListUserAudience(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("list announcement user audience: %w", err)
		}

		return s.resolveSpecificUsersEmailRecipients(ctx, userIDs)

	case "SPECIFIC_CLIENTS":
		clientIDs, err := s.repo.ListClientAudience(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("list announcement client audience: %w", err)
		}

		return s.resolveUsersByClientAccess(ctx, clientIDs)

	default:
		return nil, fmt.Errorf("unsupported announcement audience type: %s", audienceType)
	}
}

func (s *Service) resolveAllUsersEmailRecipients(
	ctx context.Context,
) ([]AnnouncementEmailRecipient, error) {
	users, err := s.userRepo.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return uniqueAnnouncementRecipientsFromUsers(users), nil
}

func (s *Service) resolveSpecificUsersEmailRecipients(
	ctx context.Context,
	userIDs []uuid.UUID,
) ([]AnnouncementEmailRecipient, error) {
	seen := make(map[string]AnnouncementEmailRecipient)

	for _, userID := range userIDs {
		if userID == uuid.Nil {
			continue
		}

		user, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			return nil, fmt.Errorf("get user by id %s: %w", userID.String(), err)
		}

		if user == nil {
			continue
		}

		recipient, ok := announcementRecipientFromUser(*user)
		if !ok {
			continue
		}

		seen[strings.ToLower(recipient.Email)] = recipient
	}

	return announcementRecipientMapToSlice(seen), nil
}

func (s *Service) resolveUsersByRealmRoles(
	ctx context.Context,
	roleNames []string,
) ([]AnnouncementEmailRecipient, error) {
	seen := make(map[string]AnnouncementEmailRecipient)

	for _, roleName := range roleNames {
		roleName = strings.TrimSpace(roleName)
		if roleName == "" {
			continue
		}

		users, err := s.userRepo.GetUsersByRealmRole(ctx, roleName)
		if err != nil {
			return nil, fmt.Errorf("get users by realm role %q: %w", roleName, err)
		}

		for _, user := range users {
			recipient, ok := announcementRecipientFromUser(user)
			if !ok {
				continue
			}

			seen[strings.ToLower(recipient.Email)] = recipient
		}
	}

	return announcementRecipientMapToSlice(seen), nil
}

func (s *Service) resolveUsersByClientAccess(
	ctx context.Context,
	clientIDs []uuid.UUID,
) ([]AnnouncementEmailRecipient, error) {
	allUsers, err := s.userRepo.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("list users for client access resolution: %w", err)
	}

	clientIDSet := make(map[string]struct{}, len(clientIDs))
	for _, clientID := range clientIDs {
		if clientID == uuid.Nil {
			continue
		}

		clientIDSet[strings.ToLower(clientID.String())] = struct{}{}
	}

	if len(clientIDSet) == 0 {
		return nil, nil
	}

	seen := make(map[string]AnnouncementEmailRecipient)

	for _, user := range allUsers {
		userID := strings.TrimSpace(user.ID)
		if userID == "" {
			continue
		}

		assignments, err := s.userRepo.GetUserClientRoles(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get user client roles for user %s: %w", userID, err)
		}

		if !userHasAnyAnnouncementClientAccess(assignments, clientIDSet) {
			continue
		}

		recipient, ok := announcementRecipientFromUser(user)
		if !ok {
			continue
		}

		seen[strings.ToLower(recipient.Email)] = recipient
	}

	return announcementRecipientMapToSlice(seen), nil
}

func announcementRecipientFromUser(user models.User) (AnnouncementEmailRecipient, bool) {
	if !user.Enabled {
		return AnnouncementEmailRecipient{}, false
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		return AnnouncementEmailRecipient{}, false
	}

	fullName := strings.TrimSpace(user.FullName)
	if fullName == "" {
		fullName = repositorySafeFullName(user.FirstName, user.LastName)
	}

	if fullName == "" {
		fullName = strings.TrimSpace(user.Username)
	}

	if fullName == "" {
		fullName = email
	}

	userID, err := uuid.Parse(strings.TrimSpace(user.ID))
	if err != nil {
		userID = uuid.Nil
	}

	return AnnouncementEmailRecipient{
		ID:       userID,
		Email:    email,
		Username: strings.TrimSpace(user.Username),
		FullName: fullName,
	}, true
}

func uniqueAnnouncementRecipientsFromUsers(
	users []models.User,
) []AnnouncementEmailRecipient {
	seen := make(map[string]AnnouncementEmailRecipient)

	for _, user := range users {
		recipient, ok := announcementRecipientFromUser(user)
		if !ok {
			continue
		}

		seen[strings.ToLower(recipient.Email)] = recipient
	}

	return announcementRecipientMapToSlice(seen)
}

func announcementRecipientMapToSlice(
	items map[string]AnnouncementEmailRecipient,
) []AnnouncementEmailRecipient {
	out := make([]AnnouncementEmailRecipient, 0, len(items))

	for _, item := range items {
		out = append(out, item)
	}

	return out
}

func userHasAnyAnnouncementClientAccess(
	assignments []keycloak.UserClientRoleAssignment,
	clientIDSet map[string]struct{},
) bool {
	for _, assignment := range assignments {
		candidates := []string{
			assignment.Id,
			assignment.ClientID,
		}

		for _, candidate := range candidates {
			candidate = strings.ToLower(strings.TrimSpace(candidate))
			if candidate == "" {
				continue
			}

			if _, ok := clientIDSet[candidate]; ok {
				return true
			}
		}
	}

	return false
}

func repositorySafeFullName(firstName string, lastName string) string {
	return strings.Join(
		strings.Fields(
			strings.TrimSpace(firstName)+" "+strings.TrimSpace(lastName),
		),
		" ",
	)
}

// ---------------------------------
// Notification helpers
// ---------------------------------

func (s *Service) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	_, _ = s.notifications.Notify(ctx, notification)
}

func (s *Service) shouldSendAnnouncementEmail(item db.Announcement) bool {
	return item.NotifyByEmail && !item.EmailNotificationSentAt.Valid
}

func (s *Service) attachAnnouncementEmailDelivery(
	notification *models.Notification,
	item db.Announcement,
	recipients []AnnouncementEmailRecipient,
	options AnnouncementEmailOptions,
) {
	if notification == nil {
		return
	}

	deliveries := []models.NotificationDeliveryRequest{
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
	}

	for _, recipient := range recipients {
		email := strings.TrimSpace(recipient.Email)
		if email == "" {
			continue
		}

		name := strings.TrimSpace(recipient.FullName)
		if name == "" {
			name = strings.TrimSpace(recipient.Username)
		}

		if name == "" {
			name = email
		}

		attachments := []models.Attachment(nil)
		if options.IncludeAttachmentsInEmail {
			attachments = options.Attachments
		}
		attachmentNames := announcementEmailAttachmentNames(attachments)
		attachmentLinks := options.AttachmentLinks
		actionURL := s.announcementActionURL(item)
		relatedLinkURL := s.announcementRelatedLinkURL(item)

		deliveries = append(deliveries, models.NotificationDeliveryRequest{
			Channel: models.NotificationChannelEmail,
			Recipient: map[string]any{
				"user_id": recipient.ID,
				"name":    name,
				"email":   email,
			},
			TemplateName: "announcement",
			TemplateData: map[string]any{
				"Name":                  name,
				"Platform":              s.platformName(),
				"Title":                 item.Title,
				"Summary":               nullStringValue(item.Summary),
				"Message":               item.Message,
				"Level":                 announcementLevelString(item.Level),
				"Status":                announcementStatusString(item.Status),
				"AnnouncementID":        item.ID.String(),
				"HasAttachments":        len(attachments) > 0,
				"AttachmentCount":       len(attachments),
				"AttachmentNames":       attachmentNames,
				"AttachmentLinks":       attachmentLinks,
				"AnnouncementLinkURL":   relatedLinkURL,
				"AnnouncementLinkLabel": "Open related link",
				"ActionURL":             actionURL,
				"Details": fmt.Sprintf(
					"Title: %s\nLevel: %s\nStatus: %s\nMessage: %s",
					item.Title,
					announcementLevelString(item.Level),
					announcementStatusString(item.Status),
					item.Message,
				),
			},
			Payload: map[string]any{
				"subject":   fmt.Sprintf("[Announcement] %s", item.Title),
				"text_body": item.Message,
			},
			Attachments: attachments,
			ScheduledAt: options.ScheduledAt,
			MaxAttempts: 5,
		})
	}

	notification.Deliveries = deliveries
}

func announcementEmailAttachmentNames(attachments []models.Attachment) []string {
	names := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		name := strings.TrimSpace(attachment.FileName)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func (s *Service) announcementAttachmentDownloadURL(
	announcementID uuid.UUID,
	attachmentID uuid.UUID,
) string {
	if announcementID == uuid.Nil || attachmentID == uuid.Nil {
		return ""
	}

	base := ""
	if s != nil && s.cfg != nil {
		base = strings.TrimRight(strings.TrimSpace(s.cfg.AppBaseURL), "/")
	}

	if base == "" {
		base = "http://localhost:9000"
	}

	return fmt.Sprintf(
		"%s/api/v1/announcements/%s/attachments/%s/download",
		base,
		announcementID.String(),
		attachmentID.String(),
	)
}

func (s *Service) normalizeAnnouncementEmailOptions(options ...AnnouncementEmailOptions) (AnnouncementEmailOptions, error) {
	out := AnnouncementEmailOptions{IncludeAttachmentsInEmail: true}
	if len(options) > 0 {
		out = options[0]
		if len(out.Attachments) > 0 && !out.IncludeAttachmentsInEmail {
			return out, nil
		}
	}
	if len(out.Attachments) == 0 {
		return out, nil
	}
	attachments, err := emailutil.NormalizeAttachments(s.cfg, out.Attachments)
	if err != nil {
		return AnnouncementEmailOptions{}, fmt.Errorf("validate announcement email attachments: %w", err)
	}
	out.Attachments = attachments
	return out, nil
}

func (s *Service) attachAdminEmailDelivery(
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

// ---------------------------------
// URL / config helpers
// ---------------------------------

func (s *Service) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *Service) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *Service) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *Service) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *Service) adminAnnouncementsURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/admin/announcements"
	}

	return base + "/announcements"
}

func (s *Service) portalAnnouncementsURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/apps/news"
	}

	return base + "/apps/news"
}

func (s *Service) announcementActionURL(item db.Announcement) string {
	return s.absolutePortalURL(announcementLinkOrDefault(item, s.portalAnnouncementsURL()))
}

func (s *Service) announcementRelatedLinkURL(item db.Announcement) string {
	if !item.LinkUrl.Valid || strings.TrimSpace(item.LinkUrl.String) == "" {
		return ""
	}

	return s.absolutePortalURL(item.LinkUrl.String)
}

func (s *Service) absolutePortalURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	if parsed, err := url.Parse(value); err == nil && parsed.IsAbs() {
		return value
	}

	base := ""
	if s != nil && s.cfg != nil {
		base = strings.TrimSpace(s.cfg.FrontendBaseURL)
		if base == "" {
			base = strings.TrimSpace(s.cfg.FrontendRedirectURI)
		}
	}
	if base == "" {
		base = s.adminDashboardURL()
	}
	if base == "" {
		return value
	}

	parsedBase, err := url.Parse(base)
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return value
	}

	if strings.HasPrefix(value, "/") {
		return parsedBase.Scheme + "://" + parsedBase.Host + value
	}

	basePath := strings.TrimRight(parsedBase.Path, "/")
	if basePath == "" {
		return parsedBase.Scheme + "://" + parsedBase.Host + "/" + strings.TrimLeft(value, "/")
	}

	return parsedBase.Scheme + "://" + parsedBase.Host + basePath + "/" + strings.TrimLeft(value, "/")
}

// ---------------------------------
// Value helpers
// ---------------------------------

func nullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}

	return strings.TrimSpace(ns.String)
}

func announcementStatusString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return strings.TrimSpace(string(v))
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

func announcementLevelString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return strings.TrimSpace(string(v))
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case models.AnnouncementLevel:
		return strings.TrimSpace(string(v))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func announcementAudienceTypeString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return strings.TrimSpace(string(v))
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case models.AnnouncementAudienceType:
		return strings.TrimSpace(string(v))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func announcementLinkOrDefault(item db.Announcement, fallback string) string {
	if item.LinkUrl.Valid && strings.TrimSpace(item.LinkUrl.String) != "" {
		return strings.TrimSpace(item.LinkUrl.String)
	}

	return fallback
}
