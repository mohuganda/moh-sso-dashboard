package system_rbac

import (
	"testing"

	"github.com/moh-sso-dashboard/internal/authz"
)

func TestDefaultSeedGivesUserDefaultSystems(t *testing.T) {
	seed := DefaultSeed()
	if err := ValidateSeed(seed); err != nil {
		t.Fatalf("default seed is invalid: %v", err)
	}

	for _, realmRole := range seed.RealmRoles {
		if realmRole.Name != authz.RoleUser {
			continue
		}
		expected := map[string]string{
			authz.SystemDataStatistics: authz.DataStatisticsAccess,
			authz.SystemUtilities:      authz.UtilitiesAccess,
			authz.SystemSettings:       authz.SettingsAccess,
		}
		for clientID, roleName := range expected {
			if !contains(realmRole.SystemRoles[clientID], roleName) {
				t.Fatalf("user defaults missing %s/%s: %+v", clientID, roleName, realmRole.SystemRoles)
			}
		}
		return
	}
	t.Fatal("default seed does not define user realm role")
}

func TestValidateSeedRejectsUnknownDefaultSystemRole(t *testing.T) {
	seed := DefaultSeed()
	for index := range seed.RealmRoles {
		if seed.RealmRoles[index].Name == authz.RoleUser {
			seed.RealmRoles[index].SystemRoles[authz.SystemUtilities] = []string{"missing-role"}
		}
	}
	if err := ValidateSeed(seed); err == nil {
		t.Fatal("expected unknown default system role validation error")
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
