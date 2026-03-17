package announcements

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
)

type AnnouncementRepository interface {
	// ---------------------------------
	// Core CRUD
	// ---------------------------------

	Create(
		ctx context.Context,
		params db.CreateAnnouncementParams,
	) (db.Announcement, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (db.Announcement, error)

	Update(
		ctx context.Context,
		params db.UpdateAnnouncementParams,
	) (db.Announcement, error)

	SoftDelete(
		ctx context.Context,
		id uuid.UUID,
		deletedBy uuid.UUID,
	) error

	Restore(
		ctx context.Context,
		id uuid.UUID,
		updatedBy uuid.UUID,
	) (db.Announcement, error)

	// ---------------------------------
	// Listing / search
	// ---------------------------------

	ListAdmin(
		ctx context.Context,
		params db.ListAnnouncementsAdminParams,
	) ([]db.Announcement, error)

	CountAdmin(
		ctx context.Context,
	) (int64, error)

	ListByStatus(
		ctx context.Context,
		params db.ListAnnouncementsByStatusParams,
	) ([]db.Announcement, error)

	CountByStatus(
		ctx context.Context,
		status model.AnnouncementStatus,
	) (int64, error)

	SearchAdmin(
		ctx context.Context,
		params db.SearchAnnouncementsAdminParams,
	) ([]db.Announcement, error)

	CountSearchAdmin(
		ctx context.Context,
		search string,
	) (int64, error)

	ListActivePublished(
		ctx context.Context,
		params db.ListActivePublishedAnnouncementsParams,
	) ([]db.Announcement, error)

	CountActivePublished(
		ctx context.Context,
	) (int64, error)

	ListCreatedByUser(
		ctx context.Context,
		params db.ListAnnouncementsCreatedByUserParams,
	) ([]db.Announcement, error)

	CountCreatedByUser(
		ctx context.Context,
		userID uuid.UUID,
	) (int64, error)

	// ---------------------------------
	// Lifecycle actions
	// ---------------------------------

	PublishNow(
		ctx context.Context,
		id uuid.UUID,
		publishedBy uuid.UUID,
	) (db.Announcement, error)

	Schedule(
		ctx context.Context,
		params db.ScheduleAnnouncementParams,
	) (db.Announcement, error)

	Archive(
		ctx context.Context,
		id uuid.UUID,
		archivedBy uuid.UUID,
	) (db.Announcement, error)

	UnarchiveToDraft(
		ctx context.Context,
		id uuid.UUID,
		updatedBy uuid.UUID,
	) (db.Announcement, error)

	UpdateStatus(
		ctx context.Context,
		params db.UpdateAnnouncementStatusParams,
	) (db.Announcement, error)

	SetPinned(
		ctx context.Context,
		id uuid.UUID,
		isPinned bool,
		updatedBy uuid.UUID,
	) (db.Announcement, error)

	SetPriority(
		ctx context.Context,
		id uuid.UUID,
		priority int32,
		updatedBy uuid.UUID,
	) (db.Announcement, error)

	// ---------------------------------
	// Audience mappings
	// ---------------------------------

	AddClientAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		clientID uuid.UUID,
	) error

	ReplaceClientAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		clientIDs []uuid.UUID,
	) error

	ListClientAudience(
		ctx context.Context,
		announcementID uuid.UUID,
	) ([]uuid.UUID, error)

	AddRoleAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		roleName string,
	) error

	ReplaceRoleAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		roleNames []string,
	) error

	ListRoleAudience(
		ctx context.Context,
		announcementID uuid.UUID,
	) ([]string, error)

	AddUserAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		userID uuid.UUID,
	) error

	ReplaceUserAudience(
		ctx context.Context,
		announcementID uuid.UUID,
		userIDs []uuid.UUID,
	) error

	ListUserAudience(
		ctx context.Context,
		announcementID uuid.UUID,
	) ([]uuid.UUID, error)

	// ---------------------------------
	// End-user targeting queries
	// ---------------------------------

	ListForClient(
		ctx context.Context,
		params db.ListAnnouncementsForClientParams,
	) ([]db.Announcement, error)

	ListForRole(
		ctx context.Context,
		params db.ListAnnouncementsForRoleParams,
	) ([]db.Announcement, error)

	ListForUser(
		ctx context.Context,
		params db.ListAnnouncementsForUserParams,
	) ([]db.Announcement, error)

	// ---------------------------------
	// Reporting
	// ---------------------------------

	GetStats(
		ctx context.Context,
	) (db.GetAnnouncementStatsRow, error)
}
