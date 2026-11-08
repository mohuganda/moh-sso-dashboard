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
					if group == "observatory" {
						// Existing natural keys can have local spelling, parent and
						// metadata edits. Reference seeds must not undo those edits.
						_, err := connection.Exec(`
						 INSERT INTO regions(id,name,code) VALUES ('10000000-0000-0000-0000-000000000001','aChOlI','local-code');
						 INSERT INTO districts(id,name,region_id,code) VALUES ('10000000-0000-0000-0000-000000000002','ABIM','10000000-0000-0000-0000-000000000001','local-code');
						 INSERT INTO sub_counties(id,name,district_id,code) VALUES ('10000000-0000-0000-0000-000000000003','ABIM DLG','10000000-0000-0000-0000-000000000002','local-code');
						 INSERT INTO facilities(id,external_id,name,district_id,sub_county_id,region_id,facility_type) VALUES
						 ('10000000-0000-0000-0000-000000000004','344357494','Locally renamed hospital','10000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000003','10000000-0000-0000-0000-000000000001','local-type'),
						 ('10000000-0000-0000-0000-000000000005','local-only','Local-only facility','10000000-0000-0000-0000-000000000002',NULL,'10000000-0000-0000-0000-000000000001','local-type'),
						 ('10000000-0000-0000-0000-000000000006',NULL,'abim obolokome health centre ii','10000000-0000-0000-0000-000000000002',NULL,'10000000-0000-0000-0000-000000000001','local-type');
						 INSERT INTO facility_weekly_metrics(facility_id,disease_id,epi_week_id,metric_value,source_name)
						 SELECT '10000000-0000-0000-0000-000000000004',d.id,w.id,42,'reference-upgrade-fixture'
						 FROM diseases d CROSS JOIN epi_weeks w WHERE d.name='Measles' ORDER BY w.epi_year,w.epi_week LIMIT 1;`)
						if err != nil {
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
				if group == "observatory" {
					for table, expected := range map[string]int{"regions": 17, "districts": 146, "sub_counties": 177, "facilities": 6150, "diseases": 14, "indicators": 9, "weekly_status": 0, "alerts": 0, "facility_weekly_metrics": 0, "surveillance_import_raw_rows": 0} {
						if upgrade && table == "facilities" {
							expected++
						}
						if upgrade && table == "facility_weekly_metrics" {
							expected = 1
						}
						var count int
						if err := connection.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != expected {
							t.Fatalf("%s reference rows=%d want=%d err=%v", table, count, expected, err)
						}
					}
					if upgrade {
						var preserved bool
						if err := connection.QueryRow(`SELECT EXISTS (
						 SELECT 1 FROM facilities f JOIN sub_counties s ON s.id=f.sub_county_id JOIN districts d ON d.id=s.district_id JOIN regions r ON r.id=d.region_id
						 WHERE f.id='10000000-0000-0000-0000-000000000004' AND f.external_id='344357494' AND f.name='Locally renamed hospital' AND f.facility_type='local-type'
						 AND s.id='10000000-0000-0000-0000-000000000003' AND s.name='ABIM DLG' AND s.code='local-code'
						 AND d.id='10000000-0000-0000-0000-000000000002' AND d.name='ABIM' AND d.code='local-code'
						 AND r.id='10000000-0000-0000-0000-000000000001' AND r.name='aChOlI' AND r.code='local-code'
						)`).Scan(&preserved); err != nil || !preserved {
							t.Fatalf("reference migration changed local metadata/IDs: preserved=%t err=%v", preserved, err)
						}
					}
					// Applying the data statements twice must be harmless, and a
					// rollback must retain reference rows already used by reports.
					fingerprints := map[string]string{}
					for _, table := range []string{"regions", "districts", "sub_counties", "facilities"} {
						var fingerprint string
						if err := connection.QueryRow("SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM " + table).Scan(&fingerprint); err != nil {
							t.Fatal(err)
						}
						fingerprints[table] = fingerprint
					}
					seed, err := os.ReadFile("../db/surveillance_migrations/000003_reference_geography.up.sql")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := connection.Exec(string(seed)); err != nil {
						t.Fatal(err)
					}
					if err := m.Steps(-1); err != nil {
						t.Fatal(err)
					}
					var count int
					want := 6150
					if upgrade {
						want++
					}
					if err := connection.QueryRow("SELECT count(*) FROM facilities").Scan(&count); err != nil || count != want {
						t.Fatalf("reference rollback lost facilities: count=%d err=%v", count, err)
					}
					if err := m.Up(); err != nil {
						t.Fatal(err)
					}
					for table, original := range fingerprints {
						var fingerprint string
						if err := connection.QueryRow("SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM " + table).Scan(&fingerprint); err != nil || fingerprint != original {
							t.Fatalf("repeated seed/rollback changed %s IDs: err=%v", table, err)
						}
					}
					if upgrade {
						var preserved bool
						if err := connection.QueryRow(`SELECT EXISTS (
						 SELECT 1 FROM facility_weekly_metrics WHERE facility_id='10000000-0000-0000-0000-000000000004' AND metric_value=42 AND source_name='reference-upgrade-fixture'
						)`).Scan(&preserved); err != nil || !preserved {
							t.Fatalf("reference migration/rollback changed an existing report: %v", err)
						}
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
