package bootstrap

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/migrate"
)

func TestObservatoryReconnectPostgres(t *testing.T) {
	dsn := os.Getenv("MOH_OBSERVATORY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL fixture")
	}
	ctx := context.Background()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := "observatory_retry_" + uuid.NewString()[:8]
	target, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	target.Path = "/" + name
	pool, err := db.OpenDB(db.DBConfig{Driver: "postgres", DSN: target.String(), MaxOpenConns: 1, MaxIdleConns: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { pool.Close(); admin.ExecContext(ctx, `DROP DATABASE IF EXISTS "`+name+`" WITH (FORCE)`) }()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "000001_marker.up.sql"), []byte("CREATE TABLE recovery_marker(id integer PRIMARY KEY)"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "000001_marker.down.sql"), []byte("DROP TABLE recovery_marker"), 0o600); err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	var initializations atomic.Int32
	monitor := startObservatoryMonitor(ctx, 20*time.Millisecond, func(ctx context.Context) error { attempts.Add(1); return pool.PingContext(ctx) }, func(ctx context.Context) error {
		initializations.Add(1)
		return db.MigrateDBContext(ctx, pool, "file://"+filepath.ToSlash(directory))
	})
	defer monitor.Close()
	waitForObservatory(t, func() bool { return attempts.Load() > 1 })
	if monitor.Ready() {
		t.Fatal("missing database must remain unavailable")
	}
	create := func() {
		t.Helper()
		if _, err := admin.ExecContext(ctx, `CREATE DATABASE "`+name+`"`); err != nil {
			t.Fatal(err)
		}
	}
	create()
	waitForObservatory(t, monitor.Ready)
	if _, err := pool.ExecContext(ctx, "INSERT INTO recovery_marker VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	if err := admin.PingContext(ctx); err != nil {
		t.Fatal("primary test connection was affected", err)
	}
	// The same pool must reconnect, initialize and serve requests after a database outage.
	if _, err := admin.ExecContext(ctx, `DROP DATABASE "`+name+`" WITH (FORCE)`); err != nil {
		t.Fatal(err)
	}
	waitForObservatory(t, func() bool { return !monitor.Ready() })
	create()
	waitForObservatory(t, monitor.Ready)
	if initializations.Load() < 2 {
		t.Fatal("reconnected before verifying migrations")
	}
	if _, err := pool.ExecContext(ctx, "INSERT INTO recovery_marker VALUES (2)"); err != nil {
		t.Fatal(err)
	}
	if pool.Stats().InUse != 0 {
		t.Fatal("migration retry leaked a connection")
	}
}
