//go:build integration

package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
)

// The launcher supplies only a disposable PostgreSQL URL. These tests never
// load app.env or use the application's normal database configuration.
func TestMigrationPaths(t *testing.T) {
	dsn := os.Getenv("TEST_MIGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("run make migrations-check to provision disposable PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	for _, group := range []string{"primary", "remote", "observatory"} {
		directory, err := filepath.Abs("../db/migrations")
		if err != nil {
			t.Fatal(err)
		}
		if group == "remote" {
			directory = filepath.Join(directory, "remote")
		} else if group == "observatory" {
			directory, err = filepath.Abs("../db/surveillance_migrations")
			if err != nil {
				t.Fatal(err)
			}
		}
		entries, err := filepath.Glob(filepath.Join(directory, "*.up.sql"))
		if err != nil || len(entries) < 2 {
			t.Fatalf("expected ordered migration history in %s", directory)
		}
		sort.Strings(entries)
		version := func(path string) uint {
			value, e := strconv.Atoi(strings.Split(filepath.Base(path), "_")[0])
			if e != nil {
				t.Fatal(e)
			}
			return uint(value)
		}
		latest, previous := version(entries[len(entries)-1]), version(entries[len(entries)-2])
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%t", group, upgrade), func(t *testing.T) {
				name := fmt.Sprintf("migration_ci_%d", time.Now().UnixNano())
				if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if _, err := admin.Exec("DROP DATABASE " + name + " WITH (FORCE)"); err != nil {
						t.Error(err)
					}
				}()
				u, err := url.Parse(dsn)
				if err != nil {
					t.Fatal(err)
				}
				u.Path = "/" + name
				m, err := migrate.New("file://"+filepath.ToSlash(directory), u.String())
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					sourceErr, dbErr := m.Close()
					if sourceErr != nil {
						t.Error(sourceErr)
					}
					if dbErr != nil {
						t.Error(dbErr)
					}
				}()
				connection, err := sql.Open("postgres", u.String())
				if err != nil {
					t.Fatal(err)
				}
				defer connection.Close()
				if upgrade {
					if err := m.Migrate(previous); err != nil {
						t.Fatal(err)
					}
					if group == "primary" {
						// Simulate pre-cutover tables/data on an already migrated primary.
						legacySchema, err := os.ReadFile("../db/surveillance_migrations/000001_surveillance.up.sql")
						if err != nil {
							t.Fatal(err)
						}
						if _, err := connection.Exec(string(legacySchema)); err != nil {
							t.Fatal(err)
						}
						if _, err := connection.Exec("INSERT INTO regions(name) VALUES ('primary-data-to-transfer')"); err != nil {
							t.Fatal(err)
						}
					}
					if _, err := connection.Exec("CREATE TABLE ci_data_sentinel (value text); INSERT INTO ci_data_sentinel VALUES ('preserve-me')"); err != nil {
						t.Fatal(err)
					}
				}
				if err := m.Up(); err != nil {
					t.Fatal(err)
				}
				actual, dirty, err := m.Version()
				if err != nil || dirty || actual != latest {
					t.Fatalf("version=%d dirty=%t err=%v", actual, dirty, err)
				}
				if err := m.Up(); err != migrate.ErrNoChange {
					t.Fatalf("repeat migration should be a no-op: %v", err)
				}
				for _, table := range []string{"regions", "districts", "sub_counties", "facilities", "diseases", "indicators", "disease_indicators", "epi_weeks", "weekly_status", "alerts", "facility_weekly_metrics", "surveillance_import_batches", "surveillance_import_raw_rows"} {
					var exists bool
					if err := connection.QueryRow("SELECT to_regclass('public.' || $1) IS NOT NULL", table).Scan(&exists); err != nil {
						t.Fatal(err)
					}
					want := group == "observatory" || (group == "primary" && upgrade)
					if exists != want {
						t.Errorf("%s exists=%t, want %t", table, exists, want)
					}
				}
				if group == "primary" && upgrade {
					var count int
					if err := connection.QueryRow("SELECT count(*) FROM regions WHERE name='primary-data-to-transfer'").Scan(&count); err != nil || count != 1 {
						t.Fatalf("legacy primary data was changed: count=%d err=%v", count, err)
					}
				}

				if upgrade {
					var value string
					if err := connection.QueryRow("SELECT value FROM ci_data_sentinel").Scan(&value); err != nil || value != "preserve-me" {
						t.Fatalf("upgrade lost existing data: %q %v", value, err)
					}
				}
			})
		}
	}
}
