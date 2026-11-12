package health_context

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestGeographySyncAcrossDatabases(t *testing.T) {
	dsn := os.Getenv("MOH_HEALTH_CONTEXT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL fixture")
	}
	ctx := context.Background()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	openDatabase := func(prefix string) *sql.DB {
		name := prefix + uuid.NewString()[:8]
		if _, err := admin.ExecContext(ctx, `CREATE DATABASE "`+name+`"`); err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		u.Path = "/" + name
		db, err := sql.Open("postgres", u.String())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			db.Close()
			if _, err := admin.ExecContext(ctx, `DROP DATABASE "`+name+`" WITH (FORCE)`); err != nil {
				t.Error(err)
			}
		})
		return db
	}
	primary := openDatabase("health_primary_")
	observatory := openDatabase("health_observatory_")
	if _, err := primary.ExecContext(ctx, `CREATE TABLE ihp_rbac_groups (id uuid PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"000054_add_health_contexts.up.sql", "000056_bridge_surveillance_health_contexts.up.sql"} {
		contents, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := primary.ExecContext(ctx, string(contents)); err != nil {
			t.Fatal(err)
		}
	}
	_, err = observatory.ExecContext(ctx, `
	 CREATE TABLE regions (id uuid PRIMARY KEY, name text);
	 CREATE TABLE districts (id uuid PRIMARY KEY, name text, region_id uuid);
	 CREATE TABLE sub_counties (id uuid PRIMARY KEY, name text, district_id uuid);
	 CREATE TABLE facilities (id uuid PRIMARY KEY, name text, region_id uuid, district_id uuid, sub_county_id uuid)`)
	if err != nil {
		t.Fatal(err)
	}
	region, district, facility := uuid.New(), uuid.New(), uuid.New()
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO regions VALUES ($1, 'Region')`, []any{region}},
		{`INSERT INTO districts VALUES ($1, 'District', $2)`, []any{district, region}},
		{`INSERT INTO facilities VALUES ($1, 'Facility', $2, $3, NULL)`, []any{facility, region, district}},
	} {
		if _, err := observatory.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(primary, observatory).(*postgresRepository)
	node, err := repo.GetNodeByAlias(ctx, "surveillance-facility", facility.String())
	if err != nil {
		t.Fatal(err)
	}
	ancestors, err := repo.ListAncestors(ctx, node.ID)
	if err != nil || len(ancestors) != 4 {
		t.Fatalf("facility ancestry: %d, %v", len(ancestors), err)
	}
	if _, err := observatory.ExecContext(ctx, `UPDATE facilities SET name='Renamed facility' WHERE id=$1`, facility); err != nil {
		t.Fatal(err)
	}
	repo.geographySyncedAt = time.Time{}
	updated, err := repo.GetNodeByAlias(ctx, "surveillance-facility", facility.String())
	if err != nil || updated.ID != node.ID || updated.Name != "Renamed facility" {
		t.Fatalf("alias update lost identity: %+v %v", updated, err)
	}
	available := false
	isolated := NewRepositoryWithGeographyAvailability(primary, observatory, func() bool { return available }).(*postgresRepository)
	newRegion := uuid.New()
	if _, err := observatory.ExecContext(ctx, `INSERT INTO regions VALUES ($1, 'Recovered region')`, newRegion); err != nil {
		t.Fatal(err)
	}
	if _, err := isolated.ListNodes(ctx, nil, "", true); err != nil {
		t.Fatalf("primary contexts unavailable during Observatory outage: %v", err)
	}
	if !isolated.geographySyncedAt.IsZero() {
		t.Fatal("offline geography was queried")
	}
	available = true
	if _, err := isolated.GetNodeByAlias(ctx, "surveillance-region", newRegion.String()); err != nil {
		t.Fatalf("recovered geography did not synchronize: %v", err)
	}
	if _, err := observatory.ExecContext(ctx, `DROP TABLE facilities`); err != nil {
		t.Fatal(err)
	}
	repo.geographySyncedAt = time.Time{}
	if _, err := repo.ListAliasesInScope(ctx, node.ID, true); err == nil {
		t.Fatal("geography source failure must not fall back to unscoped access")
	}
}
