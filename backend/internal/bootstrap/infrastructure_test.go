package bootstrap

import (
	"context"
	"strings"
	"testing"

	"github.com/moh-sso-dashboard/internal/config"
)

func TestInitDatabasesRequiresObservatoryBeforeOpeningPrimary(t *testing.T) {
	cfg := &config.Config{ObservatoryDBPort: "5432", ObservatoryDBName: "observatory", ObservatoryDBUser: "surveillance", ObservatoryDBPassword: "fixture-only"}
	dbs, err := initDatabases(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "OBSERVATORY_DB_HOST") {
		t.Fatalf("missing Observatory must fail before connecting to primary: %v", err)
	}
	if dbs.Primary != nil || dbs.Observatory != nil {
		t.Fatal("missing Observatory must not open database pools")
	}
}
