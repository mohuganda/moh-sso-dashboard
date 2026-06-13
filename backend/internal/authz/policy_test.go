package authz

import "testing"

func TestNewContextNormalizesRolesAndAssignsAdminPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{" Admin ", "admin", "USER"}, nil)

	if !ctx.IsAdmin {
		t.Fatal("expected admin role to set IsAdmin")
	}
	if !ctx.IsUser {
		t.Fatal("expected user role to set IsUser")
	}
	if !ctx.HasPermission(PermissionUsersWrite) {
		t.Fatal("expected admin to have users:write permission")
	}
	if len(ctx.RealmRoles) != 2 {
		t.Fatalf("expected duplicate roles to be removed, got %v", ctx.RealmRoles)
	}
}

func TestUserRoleGetsReadOnlyPortalPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, nil)

	if !ctx.HasPermission(PermissionDocumentsRead) {
		t.Fatal("expected user to have documents:read permission")
	}
	if ctx.HasPermission(PermissionDocumentsWrite) {
		t.Fatal("did not expect user to have documents:write permission")
	}
}

func TestSpecializedRolesGetScopedPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleSurveillanceImporter}, nil)

	if !ctx.HasPermission(PermissionSurveillanceImport) {
		t.Fatal("expected surveillance importer to have import permission")
	}
	if ctx.HasPermission(PermissionSurveillanceManageLocations) {
		t.Fatal("did not expect surveillance importer to manage locations")
	}
}
