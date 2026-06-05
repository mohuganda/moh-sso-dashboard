package announcements

import "time"

type createAnnouncementRequest struct {
	Title        string   `json:"title" binding:"required"`
	Message      string   `json:"message" binding:"required"`
	Summary      *string  `json:"summary"`
	Level        string   `json:"level" binding:"required"`
	Tag          *string  `json:"tag"`
	LinkURL      *string  `json:"link_url"`
	Priority     int32    `json:"priority"`
	IsPinned     bool     `json:"is_pinned"`
	Status       string   `json:"status"`
	PublishAt    *string  `json:"publish_at"`
	ExpiresAt    *string  `json:"expires_at"`
	AudienceType string   `json:"audience_type"`
	ClientIDs    []string `json:"client_ids"`
	RoleNames    []string `json:"role_names"`
	UserIDs      []string `json:"user_ids"`
}

type updateAnnouncementRequest struct {
	Title        string   `json:"title" binding:"required"`
	Message      string   `json:"message" binding:"required"`
	Summary      *string  `json:"summary"`
	Level        string   `json:"level" binding:"required"`
	Tag          *string  `json:"tag"`
	LinkURL      *string  `json:"link_url"`
	Priority     int32    `json:"priority"`
	IsPinned     bool     `json:"is_pinned"`
	PublishAt    *string  `json:"publish_at"`
	ExpiresAt    *string  `json:"expires_at"`
	AudienceType string   `json:"audience_type"`
	ClientIDs    []string `json:"client_ids"`
	RoleNames    []string `json:"role_names"`
	UserIDs      []string `json:"user_ids"`
}

type AnnouncementResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Message      string     `json:"message"`
	Summary      *string    `json:"summary,omitempty"`
	Level        string     `json:"level"`
	Tag          *string    `json:"tag,omitempty"`
	LinkURL      *string    `json:"link_url,omitempty"`
	Priority     int32      `json:"priority"`
	IsPinned     bool       `json:"is_pinned"`
	Status       string     `json:"status"`
	PublishAt    *time.Time `json:"publish_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	AudienceType string     `json:"audience_type"`
	CreatedBy    string     `json:"created_by"`
	UpdatedBy    string     `json:"updated_by"`
	PublishedBy  *string    `json:"published_by,omitempty"`
	ArchivedBy   *string    `json:"archived_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	DeletedBy    *string    `json:"deleted_by,omitempty"`
	Version      int32      `json:"version"`
}

type scheduleAnnouncementRequest struct {
	PublishAt string `json:"publish_at" binding:"required"`
}

type setPinnedRequest struct {
	IsPinned bool `json:"is_pinned"`
}

type setPriorityRequest struct {
	Priority int32 `json:"priority"`
}
