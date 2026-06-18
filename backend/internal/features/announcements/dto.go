package announcements

import (
	"time"

	"github.com/google/uuid"
)

type createAnnouncementRequest struct {
	Title         string   `json:"title" binding:"required"`
	Message       string   `json:"message" binding:"required"`
	Summary       *string  `json:"summary"`
	Level         string   `json:"level" binding:"required"`
	Tag           *string  `json:"tag"`
	LinkURL       *string  `json:"link_url"`
	Priority      int32    `json:"priority"`
	IsPinned      bool     `json:"is_pinned"`
	Status        string   `json:"status"`
	PublishAt     *string  `json:"publish_at"`
	ExpiresAt     *string  `json:"expires_at"`
	AudienceType  string   `json:"audience_type"`
	ClientIDs     []string `json:"client_ids"`
	RoleNames     []string `json:"role_names"`
	UserIDs       []string `json:"user_ids"`
	NotifyByEmail bool     `json:"notify_by_email"`
}

type updateAnnouncementRequest struct {
	Title         string   `json:"title" binding:"required"`
	Message       string   `json:"message" binding:"required"`
	Summary       *string  `json:"summary"`
	Level         string   `json:"level" binding:"required"`
	Tag           *string  `json:"tag"`
	LinkURL       *string  `json:"link_url"`
	Priority      int32    `json:"priority"`
	IsPinned      bool     `json:"is_pinned"`
	PublishAt     *string  `json:"publish_at"`
	ExpiresAt     *string  `json:"expires_at"`
	AudienceType  string   `json:"audience_type"`
	ClientIDs     []string `json:"client_ids"`
	RoleNames     []string `json:"role_names"`
	UserIDs       []string `json:"user_ids"`
	NotifyByEmail bool     `json:"notify_by_email"`
}

type AnnouncementResponse struct {
	ID                      string                           `json:"id"`
	Title                   string                           `json:"title"`
	Message                 string                           `json:"message"`
	Summary                 *string                          `json:"summary,omitempty"`
	Level                   string                           `json:"level"`
	Tag                     *string                          `json:"tag,omitempty"`
	LinkURL                 *string                          `json:"link_url,omitempty"`
	Priority                int32                            `json:"priority"`
	IsPinned                bool                             `json:"is_pinned"`
	Status                  string                           `json:"status"`
	PublishAt               *time.Time                       `json:"publish_at,omitempty"`
	ExpiresAt               *time.Time                       `json:"expires_at,omitempty"`
	AudienceType            string                           `json:"audience_type"`
	NotifyByEmail           bool                             `json:"notify_by_email"`
	EmailNotificationSentAt *time.Time                       `json:"email_notification_sent_at,omitempty"`
	CreatedBy               string                           `json:"created_by"`
	UpdatedBy               string                           `json:"updated_by"`
	PublishedBy             *string                          `json:"published_by,omitempty"`
	ArchivedBy              *string                          `json:"archived_by,omitempty"`
	CreatedAt               time.Time                        `json:"created_at"`
	UpdatedAt               time.Time                        `json:"updated_at"`
	PublishedAt             *time.Time                       `json:"published_at,omitempty"`
	ArchivedAt              *time.Time                       `json:"archived_at,omitempty"`
	DeletedAt               *time.Time                       `json:"deleted_at,omitempty"`
	DeletedBy               *string                          `json:"deleted_by,omitempty"`
	Version                 int32                            `json:"version"`
	Attachments             []AnnouncementAttachmentResponse `json:"attachments,omitempty"`
	AttachmentCount         int                              `json:"attachment_count"`
}

type AnnouncementAttachmentResponse struct {
	ID               string     `json:"id"`
	AnnouncementID   string     `json:"announcement_id"`
	FileName         string     `json:"file_name"`
	OriginalFileName string     `json:"original_file_name"`
	ContentType      *string    `json:"content_type,omitempty"`
	FileSize         int64      `json:"file_size"`
	StorageProvider  string     `json:"storage_provider"`
	Checksum         *string    `json:"checksum,omitempty"`
	UploadedBy       *string    `json:"uploaded_by,omitempty"`
	IncludeInEmail   bool       `json:"include_in_email"`
	Inline           bool       `json:"inline"`
	ContentID        *string    `json:"content_id,omitempty"`
	SortOrder        int32      `json:"sort_order"`
	CreatedAt        time.Time  `json:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	DeletedBy        *string    `json:"deleted_by,omitempty"`
	DownloadURL      string     `json:"download_url,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AnnouncementStatsResponse struct {
	Total          int64 `json:"total"`
	DraftCount     int64 `json:"draft_count"`
	ScheduledCount int64 `json:"scheduled_count"`
	PublishedCount int64 `json:"published_count"`
	ArchivedCount  int64 `json:"archived_count"`
	ActiveCount    int64 `json:"active_count"`
}

type createAnnouncementAttachmentRequest struct {
	FileName       string `json:"file_name" binding:"required"`
	ContentType    string `json:"content_type,omitempty"`
	DataBase64     string `json:"data_base64" binding:"required"`
	IncludeInEmail *bool  `json:"include_in_email,omitempty"`
	Inline         bool   `json:"inline,omitempty"`
	ContentID      string `json:"content_id,omitempty"`
	SortOrder      int32  `json:"sort_order,omitempty"`
}

type updateAnnouncementAttachmentRequest struct {
	IncludeInEmail *bool   `json:"include_in_email,omitempty"`
	Inline         *bool   `json:"inline,omitempty"`
	ContentID      *string `json:"content_id,omitempty"`
	SortOrder      *int32  `json:"sort_order,omitempty"`
}

type scheduleAnnouncementRequest struct {
	PublishAt                 string                          `json:"publish_at" binding:"required"`
	Attachments               []announcementAttachmentRequest `json:"attachments,omitempty"`
	IncludeAttachmentsInEmail *bool                           `json:"include_attachments_in_email,omitempty"`
}

type publishAnnouncementRequest struct {
	Attachments               []announcementAttachmentRequest `json:"attachments,omitempty"`
	IncludeAttachmentsInEmail *bool                           `json:"include_attachments_in_email,omitempty"`
}

type announcementAttachmentRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	Path        string `json:"path,omitempty"`
	DataBase64  string `json:"data_base64,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
	Inline      bool   `json:"inline,omitempty"`
}

type setPinnedRequest struct {
	IsPinned bool `json:"is_pinned"`
}

type setPriorityRequest struct {
	Priority int32 `json:"priority"`
}

type AnnouncementEmailRecipient struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
}
