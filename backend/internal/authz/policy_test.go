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

	if !ctx.HasPermission(PermissionPortalAccess) {
		t.Fatal("expected user to have portal access")
	}
	if ctx.HasPermission(PermissionDocumentsWrite) {
		t.Fatal("did not expect user to have documents:write permission")
	}
	for _, system := range []string{SystemDataStatistics, SystemUtilities, SystemSettings} {
		if !ctx.HasSystem(system) {
			t.Fatalf("expected user realm role to have default access to %s", system)
		}
	}
	for _, permission := range []Permission{
		PermissionDataQualityRead,
		PermissionDocumentsRead,
		PermissionSurveillanceRead,
		PermissionReportBrowserRead,
	} {
		if !ctx.HasPermission(permission) {
			t.Fatalf("expected user realm role to have %s", permission)
		}
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

func TestDashboardWebAccessGetsPortalPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemDashboardWeb: {DashboardWebAccess},
	})

	if !ctx.HasPermission(PermissionPortalAccess) {
		t.Fatal("expected dashboard web access to grant portal access")
	}
	if !ctx.HasSystem(SystemDashboardWeb) {
		t.Fatal("expected dashboard web to be accessible")
	}
}

func TestOutbreakManagementViewerGetsReadOnlyPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemOutbreakManagement: {OutbreakManagementViewer},
	})

	if !ctx.HasPermission(PermissionSurveillanceRead) {
		t.Fatal("expected Outbreak Management viewer to have surveillance read")
	}
	if ctx.HasPermission(PermissionSurveillanceImport) {
		t.Fatal("did not expect Outbreak Management viewer to have surveillance import")
	}
	if !ctx.HasSystem(SystemOutbreakManagement) {
		t.Fatal("expected outbreak management to be accessible")
	}
}

func TestOutbreakManagementSuperAdminGetsElevatedPermissions(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemOutbreakManagement: {OutbreakManagementSuperAdmin},
	})

	for _, permission := range []Permission{
		PermissionSurveillanceImport,
		PermissionSurveillanceManageLocations,
		PermissionDataQualityResolve,
		PermissionDocumentsProcess,
	} {
		if !ctx.HasPermission(permission) {
			t.Fatalf("expected Outbreak Management super admin to have %s", permission)
		}
	}
}

func TestDataStatisticsAccessGetsReportBrowserPermission(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemDataStatistics: {DataStatisticsAccess},
	})

	if !ctx.HasPermission(PermissionReportBrowserRead) {
		t.Fatal("expected Data & Statistics access to grant report browser read")
	}
	if !ctx.HasSystem(SystemDataStatistics) {
		t.Fatal("expected Data & Statistics to be accessible")
	}
}

func TestDashboardRoleDoesNotExposeOtherSystems(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemDashboardWeb: {DashboardWebUser},
	})

	if !ctx.HasSystem(SystemDashboardWeb) {
		t.Fatal("expected dashboard web to be accessible")
	}
	if ctx.HasSystem(SystemOutbreakManagement) {
		t.Fatal("did not expect dashboard web role to expose outbreak management")
	}
}

func TestSystemClientRolesExposeEachSystemDifferently(t *testing.T) {
	ctx := NewContext("user-1", []string{RoleUser}, map[string][]string{
		SystemDashboardWeb:       {DashboardWebManager},
		SystemOutbreakManagement: {OutbreakManagementSurveillanceOfficer},
		SystemDataStatistics:     {ReportBrowserAnalyst},
	})

	for _, system := range []string{SystemDashboardWeb, SystemOutbreakManagement, SystemDataStatistics} {
		if !ctx.HasSystem(system) {
			t.Fatalf("expected %s to be accessible", system)
		}
	}
	if !ctx.HasPermission(PermissionSurveillanceImport) {
		t.Fatal("expected surveillance officer to have surveillance import")
	}
	if !ctx.HasPermission(PermissionReportBrowserRead) {
		t.Fatal("expected report analyst to have report browser read")
	}
	if ctx.HasPermission(PermissionRBACWrite) {
		t.Fatal("did not expect manager/analyst/scoped system roles to grant RBAC write")
	}
}
