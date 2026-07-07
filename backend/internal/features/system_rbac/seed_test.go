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
			seed.RealmRoles[index].SystemRoles[authz.SystemDataStatistics] = []string{"missing-role"}
		}
	}
	if err := ValidateSeed(seed); err == nil {
		t.Fatal("expected unknown default system role validation error")
	}
}

func TestValidateSystemBehaviorSupportsAllLaunchModes(t *testing.T) {
	trueValue := true
	falseValue := false
	cases := []SeedSystem{
		{
			SystemType: "platform", LaunchMode: "internal", LaunchURL: "/portal/apps/dwh",
			DisplayInLauncher: &trueValue, DisplayInSideNav: &trueValue,
			Navigation: `[{"id":"dashboards","label":"Dashboards","path":"/apps/dwh/dashboards"}]`,
		},
		{
			SystemType: "external", LaunchMode: "new_tab", LaunchURL: "https://example.org/health",
			DisplayInLauncher: &trueValue, DisplayInSideNav: &falseValue,
		},
		{
			SystemType: "external", LaunchMode: "same_tab", LaunchURL: "https://example.org/support",
			DisplayInLauncher: &trueValue, DisplayInSideNav: &falseValue,
		},
	}
	for _, system := range cases {
		if err := ValidateSystemBehavior(system); err != nil {
			t.Fatalf("expected valid behavior %+v: %v", system, err)
		}
	}
}

func TestValidateSystemBehaviorRejectsUnsafeOrInconsistentConfiguration(t *testing.T) {
	trueValue := true
	cases := []SeedSystem{
		{SystemType: "external", LaunchMode: "new_tab", LaunchURL: "javascript://example.org"},
		{SystemType: "external", LaunchMode: "same_tab", LaunchURL: "https://user:pass@example.org"},
		{SystemType: "platform", LaunchMode: "new_tab", LaunchURL: "/portal/apps/dwh"},
		{SystemType: "external", LaunchMode: "new_tab", LaunchURL: "https://example.org", DisplayInSideNav: &trueValue},
		{SystemType: "platform", LaunchMode: "internal", LaunchURL: "/portal/apps/dwh", DisplayInSideNav: &trueValue, Navigation: "[]"},
	}
	for _, system := range cases {
		if err := ValidateSystemBehavior(system); err == nil {
			t.Fatalf("expected invalid behavior %+v", system)
		}
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
