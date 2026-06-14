package audit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func TestAuditRoutesRegisterStaticRoutesBeforeIDRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/admin"), &Handler{}, ratelimit.New(nil))

	order := map[string]int{}
	for index, route := range router.Routes() {
		if route.Method == "GET" {
			order[route.Path] = index
		}
	}

	idIndex, ok := order["/admin/audit-logs/:id"]
	if !ok {
		t.Fatal("expected audit detail route to be registered")
	}

	for _, path := range []string{
		"/admin/audit-logs/metrics/overview",
		"/admin/audit-logs/metrics/failed-logins-by-day",
		"/admin/audit-logs/metrics/top-failure-ips",
		"/admin/audit-logs/export",
	} {
		staticIndex, ok := order[path]
		if !ok {
			t.Fatalf("expected %s to be registered", path)
		}
		if staticIndex > idIndex {
			t.Fatalf("expected %s to be registered before /:id", path)
		}
	}
}

func TestBuildAuditManifestUsesRecordCount(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	manifest := buildAuditManifest(from, to, []AuditLogResponse{{ID: "one"}, {ID: "two"}}, 2, nil)

	if got := manifest["record_count"]; got != 2 {
		t.Fatalf("expected record_count 2, got %v", got)
	}
	if manifest["payload_sha256"] == "" {
		t.Fatal("expected payload hash to be set")
	}
}

func TestAuditLogResponseNormalizesSQLCInternals(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	createdAt := time.Date(2026, 6, 1, 12, 30, 0, 0, time.UTC)

	response := toAuditLogResponse(db.ListAuditLogsRow{
		ID:        id,
		CreatedAt: sql.NullTime{Time: createdAt, Valid: true},
		UserID:    uuid.NullUUID{UUID: userID, Valid: true},
		Username:  "testuser",
		Action:    "login",
		Metadata: pqtype.NullRawMessage{
			RawMessage: []byte(`{"ip":"127.0.0.1","client_id":"dashboard-web","success":true}`),
			Valid:      true,
		},
	})

	if response.ID != id.String() {
		t.Fatalf("expected id %s, got %s", id, response.ID)
	}
	if response.CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected createdAt %s, got %s", createdAt.Format(time.RFC3339), response.CreatedAt)
	}
	if response.UserID != userID.String() {
		t.Fatalf("expected userId %s, got %s", userID, response.UserID)
	}
	if response.IP != "127.0.0.1" || response.ClientID != "dashboard-web" {
		t.Fatalf("expected extracted metadata fields, got ip=%s client=%s", response.IP, response.ClientID)
	}
	if response.Success == nil || !*response.Success {
		t.Fatalf("expected success true, got %v", response.Success)
	}
}
