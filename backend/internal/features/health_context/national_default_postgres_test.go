package health_context

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNationalDefaultPostgres(t *testing.T) {
	dsn := os.Getenv("MOH_HEALTH_CONTEXT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL fixture")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	schema := "national_default_" + uuid.NewString()[:8]
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE SCHEMA "` + schema + `"`)
	defer db.ExecContext(ctx, `DROP SCHEMA "`+schema+`" CASCADE`)
	exec(`SET search_path TO "` + schema + `", public`)
	exec(`CREATE TABLE ihp_rbac_groups (id uuid PRIMARY KEY, path text, enabled boolean);
 CREATE TABLE ihp_rbac_group_members (group_id uuid, user_id text);
 CREATE TABLE users (id uuid PRIMARY KEY, enabled boolean)`)
	contents, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "000054_add_health_contexts.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	exec(string(contents))
	repo := NewRepository(db)
	service := NewService(repo)
	user := uuid.NewString()
	exec(`INSERT INTO users VALUES ($1, true)`, user)

	contexts, active, err := service.ResolveAuthAccess(ctx, user)
	if err != nil || len(contexts) != 1 || active == nil || active.Code != "UG" || !active.IsDefault || !active.IsActive || active.Source != "DEFAULT" || active.ScopeMode != "NODE_AND_DESCENDANTS" {
		t.Fatalf("unassigned user needs active National baseline: %+v %+v %v", contexts, active, err)
	}
	national := uuid.MustParse(active.ID)
	// Newly authenticated users need the baseline even before the user cache is synchronized.
	fresh, err := service.ListEffectiveContexts(ctx, uuid.NewString())
	if err != nil || len(fresh) != 1 || fresh[0].ID != national {
		t.Fatalf("new user: %+v %v", fresh, err)
	}
	district, err := repo.CreateNode(ctx, CreateNodeInput{Code: "TEST-DISTRICT", Name: "District", ContextType: ContextDistrict, ParentID: &national, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ReplaceUserAssignments(ctx, user, user, []AssignmentInput{{ContextNodeID: district.ID, ScopeMode: ScopeNodeAndDescendants, IsDefault: true}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetActiveContext(ctx, user, district.ID); err != nil {
		t.Fatal(err)
	}
	contexts, active, err = service.ResolveAuthAccess(ctx, user)
	if err != nil || len(contexts) != 2 || active == nil || active.ID != district.ID.String() || !active.IsDefault {
		t.Fatalf("preserve selected/default district: %+v %+v %v", contexts, active, err)
	}
	explanation, err := service.ExplainAccess(ctx, user, national)
	if err != nil || !explanation.Allowed || explanation.AssignmentType != "DEFAULT" {
		t.Fatalf("baseline explanation: %+v %v", explanation, err)
	}
	// Explicit National grants and inherited grants must not duplicate the selector entry.
	group := uuid.New()
	exec(`INSERT INTO ihp_rbac_groups VALUES ($1, '/test', true)`, group)
	exec(`INSERT INTO ihp_rbac_group_members VALUES ($1, $2)`, group, user)
	exec(`INSERT INTO group_health_context_assignments (group_id, context_node_id) VALUES ($1, $2)`, group, national)
	if err := service.ReplaceUserAssignments(ctx, user, user, []AssignmentInput{{ContextNodeID: national, ScopeMode: ScopeNodeOnly, IsDefault: true}}); err != nil {
		t.Fatal(err)
	}
	contexts, active, err = service.ResolveAuthAccess(ctx, user)
	if err != nil || len(contexts) != 1 || active == nil || active.ID != national.String() || active.ScopeMode != "NODE_AND_DESCENDANTS" {
		t.Fatalf("deduplicate National grants: %+v %+v %v", contexts, active, err)
	}
	if err := service.ReplaceUserAssignments(ctx, user, user, nil); err != nil {
		t.Fatal(err)
	}
	remaining, err := service.ListEffectiveContexts(ctx, user)
	if err != nil || len(remaining) != 1 || remaining[0].ID != national {
		t.Fatalf("assignment edits must retain baseline: %+v %v", remaining, err)
	}
	disabled := uuid.NewString()
	exec(`INSERT INTO users VALUES ($1, false)`, disabled)
	audience, err := repo.ListAudienceUserIDs(ctx, []uuid.UUID{district.ID}, false)
	if err != nil || len(audience) != 1 || audience[0] != user {
		t.Fatalf("National audience includes enabled baseline user: %+v %v", audience, err)
	}
	// Unrelated custom hierarchies are not covered by Uganda's National baseline.
	other, err := repo.CreateNode(ctx, CreateNodeInput{Code: "OTHER", Name: "Other", ContextType: ContextCustom, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	audience, err = repo.ListAudienceUserIDs(ctx, []uuid.UUID{other.ID}, true)
	if err != nil || len(audience) != 0 {
		t.Fatalf("unrelated audience: %+v %v", audience, err)
	}
	exec(`UPDATE health_context_nodes SET enabled=false WHERE id=$1`, national)
	remaining, err = service.ListEffectiveContexts(ctx, uuid.NewString())
	if err != nil || len(remaining) != 0 {
		t.Fatalf("disabled National must grant no access: %+v %v", remaining, err)
	}
}
