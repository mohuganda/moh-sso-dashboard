package announcements

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
)

type announcementsRepository struct {
	db     db.Store
	logger *logger.Logger
}

type rawDBStore interface {
	DB() *sql.DB
}

func NewAnnouncementRepository(
	store db.Store,
	log logger.Logger,
) AnnouncementRepository {
	return &announcementsRepository{
		db:     store,
		logger: &log,
	}
}

type txStore interface {
	ExecTx(ctx context.Context, fn func(db.Querier) error) error
}

func (r *announcementsRepository) execTx(
	ctx context.Context,
	fn func(db.Querier) error,
) error {
	tx, ok := r.db.(txStore)
	if !ok {
		return fmt.Errorf("store does not support transactions")
	}
	return tx.ExecTx(ctx, fn)
}

// ---------------------------------
// Core CRUD
// ---------------------------------

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

func (r *announcementsRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.GetAnnouncementByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get announcement by id", err)
		return db.Announcement{}, fmt.Errorf("get announcement by id: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) Update(
	ctx context.Context,
	params db.UpdateAnnouncementParams,
) (db.Announcement, error) {
	item, err := r.db.UpdateAnnouncement(ctx, params)
	if err != nil {
		r.logger.Error("failed to update announcement", err)
		return db.Announcement{}, fmt.Errorf("update announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) rawDB() (*sql.DB, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("announcement repository database is nil")
	}
	store, ok := r.db.(rawDBStore)
	if !ok || store.DB() == nil {
		return nil, fmt.Errorf("announcement repository store does not expose raw database")
	}
	return store.DB(), nil
}

func (r *announcementsRepository) MarkSMSNotificationQueued(
	ctx context.Context,
	id uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.MarkAnnouncementSMSNotificationQueued(ctx, id)
	if err != nil {
		r.logger.Error("failed to mark announcement SMS notification queued", err)
		return db.Announcement{}, fmt.Errorf("mark announcement sms notification queued: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) ListSMSRecipientsForUsers(
	ctx context.Context,
	userIDs []uuid.UUID,
) ([]AnnouncementSMSRecipient, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	rawDB, err := r.rawDB()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID != uuid.Nil {
			ids = append(ids, userID.String())
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := rawDB.QueryContext(ctx, `
		SELECT user_id, phone_number
		FROM notification_preferences
		WHERE sms_enabled = TRUE
		  AND phone_number IS NOT NULL
		  AND trim(phone_number) <> ''
		  AND user_id = ANY($1)
		ORDER BY updated_at DESC
	`, pq.Array(ids))
	if err != nil {
		r.logger.Error("failed to list announcement SMS recipients", err)
		return nil, fmt.Errorf("list announcement sms recipients: %w", err)
	}
	defer rows.Close()

	recipients := make([]AnnouncementSMSRecipient, 0)
	for rows.Next() {
		var userIDRaw string
		var phoneNumber string
		if err := rows.Scan(&userIDRaw, &phoneNumber); err != nil {
			return nil, err
		}

		userID, err := uuid.Parse(userIDRaw)
		if err != nil {
			continue
		}

		recipients = append(recipients, AnnouncementSMSRecipient{
			ID:          userID,
			PhoneNumber: phoneNumber,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recipients, nil
}

func (r *announcementsRepository) CreateAttachment(
	ctx context.Context,
	params db.CreateAnnouncementAttachmentParams,
) (db.AnnouncementAttachment, error) {
	item, err := r.db.CreateAnnouncementAttachment(ctx, params)
	if err != nil {
		r.logger.Error("failed to create announcement attachment", err)
		return db.AnnouncementAttachment{}, fmt.Errorf("create announcement attachment: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) ListAttachmentsByAnnouncementID(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]db.AnnouncementAttachment, error) {
	items, err := r.db.ListAnnouncementAttachmentsByAnnouncementID(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to list announcement attachments", err)
		return nil, fmt.Errorf("list announcement attachments: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) ListEmailAttachments(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]db.AnnouncementAttachment, error) {
	items, err := r.db.ListAnnouncementEmailAttachments(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to list announcement email attachments", err)
		return nil, fmt.Errorf("list announcement email attachments: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountAttachments(
	ctx context.Context,
	announcementID uuid.UUID,
) (int64, error) {
	count, err := r.db.CountAnnouncementAttachments(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to count announcement attachments", err)
		return 0, fmt.Errorf("count announcement attachments: %w", err)
	}

	return count, nil
}

func (r *announcementsRepository) GetAttachmentByID(
	ctx context.Context,
	params db.GetAnnouncementAttachmentByIDParams,
) (db.AnnouncementAttachment, error) {
	item, err := r.db.GetAnnouncementAttachmentByID(ctx, params)
	if err != nil {
		r.logger.Error("failed to get announcement attachment", err)
		return db.AnnouncementAttachment{}, fmt.Errorf("get announcement attachment: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) UpdateAttachment(
	ctx context.Context,
	params db.UpdateAnnouncementAttachmentParams,
) (db.AnnouncementAttachment, error) {
	item, err := r.db.UpdateAnnouncementAttachment(ctx, params)
	if err != nil {
		r.logger.Error("failed to update announcement attachment", err)
		return db.AnnouncementAttachment{}, fmt.Errorf("update announcement attachment: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) SoftDeleteAttachment(
	ctx context.Context,
	params db.SoftDeleteAnnouncementAttachmentParams,
) (db.AnnouncementAttachment, error) {
	item, err := r.db.SoftDeleteAnnouncementAttachment(ctx, params)
	if err != nil {
		r.logger.Error("failed to delete announcement attachment", err)
		return db.AnnouncementAttachment{}, fmt.Errorf("delete announcement attachment: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) SoftDelete(
	ctx context.Context,
	id uuid.UUID,
	deletedBy uuid.UUID,
) error {
	err := r.db.SoftDeleteAnnouncement(ctx, db.SoftDeleteAnnouncementParams{
		ID:        id,
		DeletedBy: uuid.NullUUID{UUID: deletedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to soft delete announcement", err)
		return fmt.Errorf("soft delete announcement: %w", err)
	}

	return nil
}

func (r *announcementsRepository) Restore(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.RestoreAnnouncement(ctx, db.RestoreAnnouncementParams{
		ID:        id,
		UpdatedBy: uuid.NullUUID{UUID: updatedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to restore announcement", err)
		return db.Announcement{}, fmt.Errorf("restore announcement: %w", err)
	}

	return item, nil
}

// ---------------------------------
// Listing / search
// ---------------------------------

func (r *announcementsRepository) ListAdmin(
	ctx context.Context,
	params db.ListAnnouncementsAdminParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsAdmin(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements for admin", err)
		return nil, fmt.Errorf("list announcements admin: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountAdmin(
	ctx context.Context,
) (int64, error) {
	count, err := r.db.CountAnnouncementsAdmin(ctx)
	if err != nil {
		r.logger.Error("failed to count announcements for admin", err)
		return 0, fmt.Errorf("count announcements admin: %w", err)
	}

	return count, nil
}

func (r *announcementsRepository) ListByStatus(
	ctx context.Context,
	params db.ListAnnouncementsByStatusParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsByStatus(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements by status", err)
		return nil, fmt.Errorf("list announcements by status: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountByStatus(
	ctx context.Context,
	status model.AnnouncementStatus,
) (int64, error) {
	count, err := r.db.CountAnnouncementsByStatus(ctx, status)
	if err != nil {
		r.logger.Error("failed to count announcements by status", err)
		return 0, fmt.Errorf("count announcements by status: %w", err)
	}

	return count, nil
}

func (r *announcementsRepository) SearchAdmin(
	ctx context.Context,
	params db.SearchAnnouncementsAdminParams,
) ([]db.Announcement, error) {
	items, err := r.db.SearchAnnouncementsAdmin(ctx, params)
	if err != nil {
		r.logger.Error("failed to search announcements for admin", err)
		return nil, fmt.Errorf("search announcements admin: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountSearchAdmin(
	ctx context.Context,
	search string,
) (int64, error) {

	searchText := sql.NullString{
		String: search,
		Valid:  true,
	}

	count, err := r.db.CountSearchAnnouncementsAdmin(ctx, searchText)
	if err != nil {
		r.logger.Error("failed to count searched announcements for admin", err)
		return 0, fmt.Errorf("count search announcements admin: %w", err)
	}

	return count, nil
}

func (r *announcementsRepository) ListActivePublished(
	ctx context.Context,
	params db.ListActivePublishedAnnouncementsParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListActivePublishedAnnouncements(ctx, params)
	if err != nil {
		r.logger.Error("failed to list active published announcements", err)
		return nil, fmt.Errorf("list active published announcements: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountActivePublished(
	ctx context.Context,
) (int64, error) {
	count, err := r.db.CountActivePublishedAnnouncements(ctx)
	if err != nil {
		r.logger.Error("failed to count active published announcements", err)
		return 0, fmt.Errorf("count active published announcements: %w", err)
	}

	return count, nil
}

func (r *announcementsRepository) ListCreatedByUser(
	ctx context.Context,
	params db.ListAnnouncementsCreatedByUserParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsCreatedByUser(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements created by user", err)
		return nil, fmt.Errorf("list announcements created by user: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) CountCreatedByUser(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	count, err := r.db.CountAnnouncementsCreatedByUser(ctx, userID)
	if err != nil {
		r.logger.Error("failed to count announcements created by user", err)
		return 0, fmt.Errorf("count announcements created by user: %w", err)
	}

	return count, nil
}

// ---------------------------------
// Lifecycle actions
// ---------------------------------

func (r *announcementsRepository) PublishNow(
	ctx context.Context,
	id uuid.UUID,
	publishedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.PublishAnnouncementNow(ctx, db.PublishAnnouncementNowParams{
		ID:          id,
		PublishedBy: uuid.NullUUID{UUID: publishedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to publish announcement", err)
		return db.Announcement{}, fmt.Errorf("publish announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) Draft(
	ctx context.Context,
	params db.DraftAnnouncementParams,
) (db.Announcement, error) {
	item, err := r.db.DraftAnnouncement(ctx, params)
	if err != nil {
		r.logger.Error("failed to draft announcement", err)
		return db.Announcement{}, fmt.Errorf("draft announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) Schedule(
	ctx context.Context,
	params db.ScheduleAnnouncementParams,
) (db.Announcement, error) {
	item, err := r.db.ScheduleAnnouncement(ctx, params)
	if err != nil {
		r.logger.Error("failed to schedule announcement", err)
		return db.Announcement{}, fmt.Errorf("schedule announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) Archive(
	ctx context.Context,
	id uuid.UUID,
	archivedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.ArchiveAnnouncement(ctx, db.ArchiveAnnouncementParams{
		ID:         id,
		ArchivedBy: uuid.NullUUID{UUID: archivedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to archive announcement", err)
		return db.Announcement{}, fmt.Errorf("archive announcement: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) UnarchiveToDraft(
	ctx context.Context,
	id uuid.UUID,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.UnarchiveAnnouncementToDraft(ctx, db.UnarchiveAnnouncementToDraftParams{
		ID:        id,
		UpdatedBy: uuid.NullUUID{UUID: updatedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to unarchive announcement to draft", err)
		return db.Announcement{}, fmt.Errorf("unarchive announcement to draft: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) UpdateStatus(
	ctx context.Context,
	params db.UpdateAnnouncementStatusParams,
) (db.Announcement, error) {
	item, err := r.db.UpdateAnnouncementStatus(ctx, params)
	if err != nil {
		r.logger.Error("failed to update announcement status", err)
		return db.Announcement{}, fmt.Errorf("update announcement status: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) SetPinned(
	ctx context.Context,
	id uuid.UUID,
	isPinned bool,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.SetAnnouncementPinned(ctx, db.SetAnnouncementPinnedParams{
		ID:        id,
		IsPinned:  isPinned,
		UpdatedBy: uuid.NullUUID{UUID: updatedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to set announcement pinned state", err)
		return db.Announcement{}, fmt.Errorf("set announcement pinned state: %w", err)
	}

	return item, nil
}

func (r *announcementsRepository) SetPriority(
	ctx context.Context,
	id uuid.UUID,
	priority int32,
	updatedBy uuid.UUID,
) (db.Announcement, error) {
	item, err := r.db.SetAnnouncementPriority(ctx, db.SetAnnouncementPriorityParams{
		ID:        id,
		Priority:  priority,
		UpdatedBy: uuid.NullUUID{UUID: updatedBy, Valid: true},
	})
	if err != nil {
		r.logger.Error("failed to set announcement priority", err)
		return db.Announcement{}, fmt.Errorf("set announcement priority: %w", err)
	}

	return item, nil
}

// ---------------------------------
// Audience mappings
// ---------------------------------

func (r *announcementsRepository) AddClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	clientID uuid.UUID,
) error {
	err := r.db.InsertAnnouncementClient(ctx, db.InsertAnnouncementClientParams{
		AnnouncementID: announcementID,
		ClientID:       clientID,
	})
	if err != nil {
		r.logger.Error("failed to add client audience", err)
		return fmt.Errorf("add client audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ReplaceClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	clientIDs []uuid.UUID,
) error {
	err := r.execTx(ctx, func(q db.Querier) error {
		if err := q.DeleteAnnouncementClients(ctx, announcementID); err != nil {
			return fmt.Errorf("delete existing client audience: %w", err)
		}

		for _, clientID := range clientIDs {
			if err := q.InsertAnnouncementClient(ctx, db.InsertAnnouncementClientParams{
				AnnouncementID: announcementID,
				ClientID:       clientID,
			}); err != nil {
				return fmt.Errorf("insert client audience [%s]: %w", clientID, err)
			}
		}

		return nil
	})
	if err != nil {
		r.logger.Error("failed to replace client audience", err)
		return fmt.Errorf("replace client audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ListClientAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	items, err := r.db.ListAnnouncementClients(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to list client audience", err)
		return nil, fmt.Errorf("list client audience: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) AddRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	roleName string,
) error {
	err := r.db.InsertAnnouncementRole(ctx, db.InsertAnnouncementRoleParams{
		AnnouncementID: announcementID,
		RoleName:       roleName,
	})
	if err != nil {
		r.logger.Error("failed to add role audience", err)
		return fmt.Errorf("add role audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ReplaceRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	roleNames []string,
) error {
	err := r.execTx(ctx, func(q db.Querier) error {
		if err := q.DeleteAnnouncementRoles(ctx, announcementID); err != nil {
			return fmt.Errorf("delete existing role audience: %w", err)
		}

		for _, roleName := range roleNames {
			if err := q.InsertAnnouncementRole(ctx, db.InsertAnnouncementRoleParams{
				AnnouncementID: announcementID,
				RoleName:       roleName,
			}); err != nil {
				return fmt.Errorf("insert role audience [%s]: %w", roleName, err)
			}
		}

		return nil
	})
	if err != nil {
		r.logger.Error("failed to replace role audience", err)
		return fmt.Errorf("replace role audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ListRoleAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]string, error) {
	items, err := r.db.ListAnnouncementRoles(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to list role audience", err)
		return nil, fmt.Errorf("list role audience: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) AddUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	userID uuid.UUID,
) error {
	err := r.db.InsertAnnouncementUser(ctx, db.InsertAnnouncementUserParams{
		AnnouncementID: announcementID,
		UserID:         userID,
	})
	if err != nil {
		r.logger.Error("failed to add user audience", err)
		return fmt.Errorf("add user audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ReplaceUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	userIDs []uuid.UUID,
) error {
	err := r.execTx(ctx, func(q db.Querier) error {
		if err := q.DeleteAnnouncementUsers(ctx, announcementID); err != nil {
			return fmt.Errorf("delete existing user audience: %w", err)
		}

		for _, userID := range userIDs {
			if err := q.InsertAnnouncementUser(ctx, db.InsertAnnouncementUserParams{
				AnnouncementID: announcementID,
				UserID:         userID,
			}); err != nil {
				return fmt.Errorf("insert user audience [%s]: %w", userID, err)
			}
		}

		return nil
	})
	if err != nil {
		r.logger.Error("failed to replace user audience", err)
		return fmt.Errorf("replace user audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ListUserAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	items, err := r.db.ListAnnouncementUsers(ctx, announcementID)
	if err != nil {
		r.logger.Error("failed to list user audience", err)
		return nil, fmt.Errorf("list user audience: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) AddGroupAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	groupID uuid.UUID,
) error {
	rawDB, err := r.rawDB()
	if err != nil {
		return err
	}

	_, err = rawDB.ExecContext(
		ctx,
		`INSERT INTO announcement_groups (announcement_id, group_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		announcementID,
		groupID,
	)
	if err != nil {
		r.logger.Error("failed to add group audience", err)
		return fmt.Errorf("add group audience: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ReplaceGroupAudience(
	ctx context.Context,
	announcementID uuid.UUID,
	groupIDs []uuid.UUID,
) error {
	rawDB, err := r.rawDB()
	if err != nil {
		return err
	}

	tx, err := rawDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin group audience transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM announcement_groups WHERE announcement_id = $1`,
		announcementID,
	); err != nil {
		return fmt.Errorf("delete existing group audience: %w", err)
	}

	for _, groupID := range groupIDs {
		if groupID == uuid.Nil {
			continue
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO announcement_groups (announcement_id, group_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			announcementID,
			groupID,
		); err != nil {
			return fmt.Errorf("insert group audience [%s]: %w", groupID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit group audience transaction: %w", err)
	}

	return nil
}

func (r *announcementsRepository) ListGroupAudience(
	ctx context.Context,
	announcementID uuid.UUID,
) ([]uuid.UUID, error) {
	rawDB, err := r.rawDB()
	if err != nil {
		return nil, err
	}

	rows, err := rawDB.QueryContext(
		ctx,
		`SELECT group_id
		 FROM announcement_groups
		 WHERE announcement_id = $1
		 ORDER BY created_at ASC`,
		announcementID,
	)
	if err != nil {
		r.logger.Error("failed to list group audience", err)
		return nil, fmt.Errorf("list group audience: %w", err)
	}
	defer rows.Close()

	items := make([]uuid.UUID, 0)
	for rows.Next() {
		var groupID uuid.UUID
		if err := rows.Scan(&groupID); err != nil {
			return nil, fmt.Errorf("scan group audience: %w", err)
		}
		items = append(items, groupID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group audience: %w", err)
	}

	return items, nil
}

// ---------------------------------
// End-user targeting queries
// ---------------------------------

func (r *announcementsRepository) ListForClient(
	ctx context.Context,
	params db.ListAnnouncementsForClientParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsForClient(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements for client", err)
		return nil, fmt.Errorf("list announcements for client: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) ListForRole(
	ctx context.Context,
	params db.ListAnnouncementsForRoleParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsForRole(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements for role", err)
		return nil, fmt.Errorf("list announcements for role: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) ListForUser(
	ctx context.Context,
	params db.ListAnnouncementsForUserParams,
) ([]db.Announcement, error) {
	items, err := r.db.ListAnnouncementsForUser(ctx, params)
	if err != nil {
		r.logger.Error("failed to list announcements for user", err)
		return nil, fmt.Errorf("list announcements for user: %w", err)
	}

	return items, nil
}

func (r *announcementsRepository) ListPublicAnnouncements(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]db.Announcement, error) {
	return r.db.ListPublicAnnouncements(ctx, db.ListPublicAnnouncementsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *announcementsRepository) MarkEmailNotificationSent(
	ctx context.Context,
	id uuid.UUID,
) (db.Announcement, error) {
	return r.db.MarkAnnouncementEmailNotificationSent(ctx, id)
}

// ---------------------------------
// Reporting
// ---------------------------------

func (r *announcementsRepository) GetStats(
	ctx context.Context,
) (db.GetAnnouncementStatsRow, error) {
	stats, err := r.db.GetAnnouncementStats(ctx)
	if err != nil {
		r.logger.Error("failed to get announcement stats", err)
		return db.GetAnnouncementStatsRow{}, fmt.Errorf("get announcement stats: %w", err)
	}

	return stats, nil
}
