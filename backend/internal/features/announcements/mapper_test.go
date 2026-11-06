package announcements

import (
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"testing"
)

func TestAnnouncementResponsePreservesGeneratedEnums(t *testing.T) {
	response := toAnnouncementResponse(db.Announcement{
		Level:        db.AnnouncementLevelWARNING,
		Status:       db.AnnouncementStatusPUBLISHED,
		AudienceType: db.AnnouncementAudienceTypeSPECIFICGROUPS,
	})
	if response.Level != "WARNING" || response.Status != "PUBLISHED" || response.AudienceType != "SPECIFIC_GROUPS" {
		t.Fatalf("generated enum values were lost in the API response: %+v", response)
	}
}
