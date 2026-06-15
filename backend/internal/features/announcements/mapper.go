package announcements

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
)

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func toAnnouncementResponse(a db.Announcement) AnnouncementResponse {
	return AnnouncementResponse{
		ID:                      a.ID.String(),
		Title:                   a.Title,
		Message:                 a.Message,
		Summary:                 nullStringPtr(a.Summary),
		Level:                   normalizeLevel(model.AnnouncementLevel(interfaceToString(a.Level))),
		Tag:                     nullStringPtr(a.Tag),
		LinkURL:                 nullStringPtr(a.LinkUrl),
		Priority:                a.Priority,
		IsPinned:                a.IsPinned,
		Status:                  normalizeAnnouncementStatus(model.AnnouncementStatus(interfaceToString(a.Status))),
		PublishAt:               nullTimePtr(a.PublishAt),
		ExpiresAt:               nullTimePtr(a.ExpiresAt),
		AudienceType:            normalizeAudienceType(model.AnnouncementAudienceType(interfaceToString(a.AudienceType))),
		NotifyByEmail:           a.NotifyByEmail,
		EmailNotificationSentAt: nullTimePtr(a.EmailNotificationSentAt),
		CreatedBy:               a.CreatedBy.String(),
		UpdatedBy:               nullUUIDString(a.UpdatedBy),
		PublishedBy:             nullableUUID(a.PublishedBy),
		ArchivedBy:              nullableUUID(a.ArchivedBy),
		CreatedAt:               a.CreatedAt,
		UpdatedAt:               a.UpdatedAt,
		PublishedAt:             nullTimePtr(a.PublishedAt),
		ArchivedAt:              nullTimePtr(a.ArchivedAt),
		DeletedAt:               nullTimePtr(a.DeletedAt),
		DeletedBy:               nullableUUID(a.DeletedBy),
		Version:                 a.Version,
	}
}

func mapAnnouncementAttachments(in []announcementAttachmentRequest) []model.Attachment {
	out := make([]model.Attachment, 0, len(in))
	for _, attachment := range in {
		out = append(out, model.Attachment{
			FileName:    strings.TrimSpace(attachment.FileName),
			ContentType: strings.TrimSpace(attachment.ContentType),
			Path:        strings.TrimSpace(attachment.Path),
			DataBase64:  strings.TrimSpace(attachment.DataBase64),
			ContentID:   strings.TrimSpace(attachment.ContentID),
			Inline:      attachment.Inline,
		})
	}
	return out
}

func announcementEmailOptionsFromPublishRequest(req publishAnnouncementRequest) AnnouncementEmailOptions {
	return AnnouncementEmailOptions{
		Attachments:               mapAnnouncementAttachments(req.Attachments),
		IncludeAttachmentsInEmail: includeAnnouncementAttachments(req.IncludeAttachmentsInEmail, len(req.Attachments) > 0),
	}
}

func announcementEmailOptionsFromScheduleRequest(req scheduleAnnouncementRequest, publishAt time.Time) AnnouncementEmailOptions {
	return AnnouncementEmailOptions{
		Attachments:               mapAnnouncementAttachments(req.Attachments),
		IncludeAttachmentsInEmail: includeAnnouncementAttachments(req.IncludeAttachmentsInEmail, len(req.Attachments) > 0),
		ScheduledAt:               &publishAt,
	}
}

func includeAnnouncementAttachments(value *bool, hasAttachments bool) bool {
	if value != nil {
		return *value
	}
	return hasAttachments
}

func nullableString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: trimmed,
		Valid:  true,
	}
}

func nullableTime(s *string) (sql.NullTime, error) {
	if s == nil {
		return sql.NullTime{}, nil
	}

	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return sql.NullTime{}, nil
	}

	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return sql.NullTime{}, err
	}

	return sql.NullTime{
		Time:  t,
		Valid: true,
	}, nil
}

func nullableUUID(u uuid.NullUUID) *string {
	if !u.Valid {
		return nil
	}

	s := u.UUID.String()
	return &s
}

func parseUUIDList(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}

	out := make([]uuid.UUID, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		id, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, err
		}

		out = append(out, id)
	}

	return out, nil
}

func nullUUIDString(v uuid.NullUUID) string {
	if !v.Valid {
		return ""
	}

	return v.UUID.String()
}

func normalizeLevel(v model.AnnouncementLevel) string {
	return strings.ToLower(string(v))
}

func normalizeAnnouncementStatus(v model.AnnouncementStatus) string {
	return strings.ToLower(string(v))
}

func normalizeAudienceType(v model.AnnouncementAudienceType) string {
	return strings.ToLower(string(v))
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return ""
	}
}
