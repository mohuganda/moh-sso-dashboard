package authz

import "testing"

func TestUserDefaultsPreserveExplicitGrantsAndAdministratorAccess(t *testing.T) {
	user := NewContext("human", []string{RoleUser}, map[string][]string{
		SystemUtilities:      {UtilitiesAccess},
		SystemDataStatistics: {DocumentEditor},
	})
	if !user.HasSystem(SystemUtilities) || !user.HasPermission(PermissionDocumentsWrite) {
		t.Fatal("explicit assignments must remain additive to the baseline")
	}
	admin := NewContext("admin", []string{RoleAdmin, RoleUser}, nil)
	for _, permission := range AllPermissions {
		if !admin.HasPermission(permission) {
			t.Fatalf("administrator lost %s", permission)
		}
	}
	defaults := DefaultUserClientRoles()
	defaults[SystemDashboardWeb][0] = "changed"
	delete(defaults, SystemSettings)
	if DefaultUserClientRoles()[SystemDashboardWeb][0] != DashboardWebAccess || len(DefaultUserClientRoles()) != 3 {
		t.Fatal("a caller mutated the shared baseline")
	}
}
