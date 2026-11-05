//go:build integration

package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

func TestStandardUserDefaultsMigration(t *testing.T) {
	dsn := os.Getenv("TEST_MIGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("run make migrations-check with disposable PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("defaults_ci_%d", time.Now().UnixNano())
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
	directory, err := filepath.Abs("../db/migrations")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.New("file://"+filepath.ToSlash(directory), u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Migrate(52); err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	fixture := `
INSERT INTO ihp_systems (client_id, display_name) VALUES
 ('dashboard-web','Portal'), ('data-statistics','Data'), ('settings','Settings'), ('utilities','Utilities');
INSERT INTO ihp_system_roles (system_id, role_name)
 SELECT s.id, d.role_name FROM (VALUES
 ('dashboard-web','dashboard-web_access'), ('dashboard-web','portal_user'),
 ('data-statistics','report-scheduler_access'), ('data-statistics','report_viewer'),
 ('data-statistics','data-statistics_access'), ('data-statistics','issue_tracker_viewer'),
 ('data-statistics','document_viewer'), ('data-statistics','document_manager'),
 ('settings','settings_access'), ('utilities','utilities_access')) d(client_id,role_name)
 JOIN ihp_systems s ON s.client_id = d.client_id;
INSERT INTO ihp_realm_role_system_roles (realm_role, system_role_id)
 SELECT 'user', id FROM ihp_system_roles WHERE role_name IN ('utilities_access','document_viewer');
INSERT INTO ihp_realm_role_system_roles (realm_role, system_role_id)
 SELECT elevated.realm_role, r.id FROM (VALUES ('admin'),('manager')) elevated(realm_role)
 CROSS JOIN ihp_system_roles r WHERE r.role_name='document_manager';
INSERT INTO ihp_rbac_groups (path,name) VALUES ('/Explicit','Explicit');
INSERT INTO ihp_rbac_group_system_roles (group_id,system_role_id)
 SELECT g.id,r.id FROM ihp_rbac_groups g CROSS JOIN ihp_system_roles r WHERE r.role_name='document_manager';
`
	if _, err := database.Exec(fixture); err != nil {
		t.Fatal(err)
	}
	read := func(query string) string {
		t.Helper()
		var value string
		if err := database.QueryRow(query).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	mappings := `SELECT COALESCE(string_agg(rr.realm_role || ':' || s.client_id || ':' || sr.role_name, ',' ORDER BY rr.realm_role,s.client_id,sr.role_name),'')
 FROM ihp_realm_role_system_roles rr JOIN ihp_system_roles sr ON sr.id=rr.system_role_id JOIN ihp_systems s ON s.id=sr.system_id`
	before := read(mappings)
	for attempt := 0; attempt < 2; attempt++ {
		if err := m.Migrate(53); err != nil {
			t.Fatal(err)
		}
		actual := read(mappings)
		if strings.Contains(actual, "user:utilities:") || strings.Count(actual, "user:") != 8 {
			t.Fatalf("unexpected standard baseline: %s", actual)
		}
		for _, role := range []string{"admin", "manager"} {
			if !strings.Contains(actual, role+":data-statistics:document_manager") {
				t.Fatalf("changed elevated %s mapping: %s", role, actual)
			}
		}
		if read(`SELECT COUNT(*)::text FROM ihp_rbac_group_system_roles`) != "1" {
			t.Fatal("explicit group grants changed")
		}
		if err := m.Migrate(52); err != nil {
			t.Fatal(err)
		}
		if actual := read(mappings); actual != before {
			t.Fatalf("rollback did not restore original mappings: %s", actual)
		}
		if read(`SELECT COUNT(*)::text FROM ihp_system_access_roles WHERE role_name='portal_user'`) != "0" {
			t.Fatal("rollback retained a new access mapping")
		}
	}
}
