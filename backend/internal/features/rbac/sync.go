package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type realmExportFile struct {
	Roles struct {
		Realm []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"realm"`
		Client map[string][]struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"client"`
	} `json:"roles"`
	Clients []struct {
		ClientID               string            `json:"clientId"`
		Name                   string            `json:"name"`
		Description            string            `json:"description"`
		BaseURL                string            `json:"baseUrl"`
		RootURL                string            `json:"rootUrl"`
		Enabled                bool              `json:"enabled"`
		ServiceAccountsEnabled bool              `json:"serviceAccountsEnabled"`
		StandardFlowEnabled    bool              `json:"standardFlowEnabled"`
		Attributes             map[string]string `json:"attributes"`
	} `json:"clients"`
}

type discoveredRBAC struct {
	Systems    []KeycloakDiscoveredSystem
	RealmRoles []string
}

func (s *Service) DriftFromDefaultRealmExport(ctx context.Context) (RbacDriftReport, error) {
	path, ok := defaultRealmExportPath()
	if !ok {
		return RbacDriftReport{
			Source:     "default-realm-export",
			Systems:    []RbacDriftSystem{},
			RealmRoles: []RbacDriftRealmRole{},
			Warnings:   []string{"No local keycloak/realm-export.json file was found for drift comparison."},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return RbacDriftReport{}, err
	}
	report, err := s.DriftFromRealmExport(ctx, data)
	if err != nil {
		return RbacDriftReport{}, err
	}
	report.Source = path
	return report, nil
}

func (s *Service) DriftFromRealmExport(ctx context.Context, payload []byte) (RbacDriftReport, error) {
	discovered, err := parseRealmExport(payload)
	if err != nil {
		return RbacDriftReport{}, err
	}
	return s.buildDriftReport(ctx, "realm-export", discovered)
}

func (s *Service) PreviewRealmExportSync(ctx context.Context, payload []byte) (SyncPreviewResponse, discoveredRBAC, error) {
	discovered, err := parseRealmExport(payload)
	if err != nil {
		return SyncPreviewResponse{}, discoveredRBAC{}, err
	}
	report, err := s.buildDriftReport(ctx, "realm-export", discovered)
	if err != nil {
		return SyncPreviewResponse{}, discoveredRBAC{}, err
	}

	preview := SyncPreviewResponse{Report: report}
	for _, system := range report.Systems {
		if system.RBACStatus == "missing" {
			preview.SystemsToCreate++
		} else if system.KeycloakStatus == "present" && system.RBACStatus == "present" {
			preview.SystemsToUpdate++
		}
		preview.RolesToCreate += len(system.MissingRolesInRBAC)
		preview.AccessRolesToAdd += len(system.MissingAccessRoles)
	}
	for _, role := range report.RealmRoles {
		if role.RBACStatus == "missing" {
			preview.RealmRolesToTrack++
		}
	}
	return preview, discovered, nil
}

func (s *Service) ApplyRealmExportSync(ctx context.Context, payload []byte) (SyncApplyResponse, error) {
	preview, discovered, err := s.PreviewRealmExportSync(ctx, payload)
	if err != nil {
		return SyncApplyResponse{}, err
	}

	response := SyncApplyResponse{Preview: preview}
	for _, system := range discovered.Systems {
		enabled := system.Enabled
		if _, err := s.UpsertSystem(ctx, UpsertSystemInput{
			ClientID:    system.ClientID,
			DisplayName: system.DisplayName,
			Description: system.Description,
			Icon:        system.Icon,
			LaunchURL:   system.LaunchURL,
			Category:    system.Category,
			Navigation:  system.Navigation,
			Enabled:     &enabled,
		}); err != nil {
			return SyncApplyResponse{}, err
		}
		response.SystemsSynced++

		for _, role := range system.Roles {
			systemRole, err := s.repository.UpsertSystemRole(ctx, system.ClientID, RoleInput{
				Name:        role.Name,
				Description: role.Description,
				Enabled:     &enabled,
			})
			if err != nil {
				return SyncApplyResponse{}, err
			}
			if err := s.recordAudit(ctx, "system_role.synced", "system-role", systemRole.ID, system.ClientID, systemRole.Name, "", nil); err != nil {
				return SyncApplyResponse{}, err
			}
			response.RolesSynced++
			if err := s.repository.AddSystemAccessRole(ctx, system.ClientID, role.Name); err != nil {
				return SyncApplyResponse{}, err
			}
			if err := s.recordAudit(ctx, "system.access_role_synced", "system", system.ClientID, system.ClientID, role.Name, "", nil); err != nil {
				return SyncApplyResponse{}, err
			}
			response.AccessRoles++
		}
	}

	if err := s.recordAudit(ctx, "rbac.sync_applied", "sync", "realm-export", "", "", "", map[string]any{"systemsSynced": response.SystemsSynced, "rolesSynced": response.RolesSynced, "accessRolesSynced": response.AccessRoles}); err != nil {
		return SyncApplyResponse{}, err
	}
	return response, nil
}

func (s *Service) buildDriftReport(ctx context.Context, source string, discovered discoveredRBAC) (RbacDriftReport, error) {
	currentSystems, err := s.ListSystems(ctx)
	if err != nil {
		return RbacDriftReport{}, err
	}
	currentRealmRoles, err := s.ListRealmRolePermissions(ctx)
	if err != nil {
		return RbacDriftReport{}, err
	}

	currentByClient := make(map[string]System, len(currentSystems))
	for _, system := range currentSystems {
		currentByClient[system.ClientID] = system
	}
	discoveredByClient := make(map[string]KeycloakDiscoveredSystem, len(discovered.Systems))
	for _, system := range discovered.Systems {
		discoveredByClient[system.ClientID] = system
	}

	report := RbacDriftReport{
		Source:     source,
		Systems:    make([]RbacDriftSystem, 0, len(discovered.Systems)+len(currentSystems)),
		RealmRoles: make([]RbacDriftRealmRole, 0, len(discovered.RealmRoles)),
	}

	seenSystems := map[string]bool{}
	for _, discoveredSystem := range discovered.Systems {
		seenSystems[discoveredSystem.ClientID] = true
		current, exists := currentByClient[discoveredSystem.ClientID]
		drift := RbacDriftSystem{
			ClientID:       discoveredSystem.ClientID,
			DisplayName:    firstNonEmpty(discoveredSystem.DisplayName, discoveredSystem.ClientID),
			KeycloakStatus: "present",
			RBACStatus:     statusForExists(exists),
		}
		if exists {
			drift.DisplayName = firstNonEmpty(current.DisplayName, drift.DisplayName)
			detail, err := s.GetSystem(ctx, discoveredSystem.ClientID)
			if err != nil {
				return RbacDriftReport{}, err
			}
			drift.MissingRolesInRBAC = missingRoleNames(discoveredSystem.Roles, detail.Roles)
			drift.StaleRolesInRBAC = staleRoleNames(discoveredSystem.Roles, detail.Roles)
			drift.MissingAccessRoles = missingAccessRoleNames(discoveredSystem.Roles, detail.AccessRoles)
			drift.DiscoveredRoleDetails = roleDriftDetails(discoveredSystem.Roles, detail.Roles)
		} else {
			drift.MissingRolesInRBAC = discoveredRoleNames(discoveredSystem.Roles)
			drift.MissingAccessRoles = discoveredRoleNames(discoveredSystem.Roles)
			drift.DiscoveredRoleDetails = roleDriftDetails(discoveredSystem.Roles, nil)
		}
		updateSummary(&report.Summary, drift)
		report.Systems = append(report.Systems, drift)
	}

	for _, current := range currentSystems {
		if seenSystems[current.ClientID] {
			continue
		}
		drift := RbacDriftSystem{
			ClientID:       current.ClientID,
			DisplayName:    current.DisplayName,
			KeycloakStatus: "missing",
			RBACStatus:     "present",
		}
		updateSummary(&report.Summary, drift)
		report.Systems = append(report.Systems, drift)
	}

	currentRealmRoleSet := map[string]bool{}
	for _, group := range currentRealmRoles {
		currentRealmRoleSet[normalize(group.RealmRole)] = true
	}
	for _, role := range discovered.RealmRoles {
		normalized := normalize(role)
		exists := currentRealmRoleSet[normalized]
		if !exists {
			report.Summary.RealmRolesMissing++
		}
		report.RealmRoles = append(report.RealmRoles, RbacDriftRealmRole{
			Name:       normalized,
			Status:     "present",
			RBACStatus: statusForExists(exists),
		})
	}

	sort.Slice(report.Systems, func(i, j int) bool { return report.Systems[i].ClientID < report.Systems[j].ClientID })
	sort.Slice(report.RealmRoles, func(i, j int) bool { return report.RealmRoles[i].Name < report.RealmRoles[j].Name })
	return report, nil
}

func parseRealmExport(payload []byte) (discoveredRBAC, error) {
	var export realmExportFile
	if err := json.Unmarshal(payload, &export); err != nil {
		return discoveredRBAC{}, fmt.Errorf("%w: invalid realm export JSON", ErrInvalidInput)
	}

	discovered := discoveredRBAC{
		Systems:    make([]KeycloakDiscoveredSystem, 0, len(export.Clients)),
		RealmRoles: make([]string, 0, len(export.Roles.Realm)),
	}

	for _, role := range export.Roles.Realm {
		name := normalize(role.Name)
		if shouldSkipRealmRole(name) {
			continue
		}
		discovered.RealmRoles = append(discovered.RealmRoles, name)
	}

	for _, client := range export.Clients {
		clientID := strings.TrimSpace(client.ClientID)
		if shouldSkipClientID(clientID) {
			continue
		}
		roles := export.Roles.Client[clientID]
		if len(roles) == 0 && client.ServiceAccountsEnabled && !client.StandardFlowEnabled {
			continue
		}

		system := KeycloakDiscoveredSystem{
			ClientID:    clientID,
			DisplayName: firstNonEmpty(client.Name, client.Description, clientID),
			Description: client.Description,
			Icon:        client.Attributes["ui.icon"],
			LaunchURL:   firstNonEmpty(client.Attributes["ui.launchUrl"], client.Attributes["ui.home"], client.BaseURL, client.RootURL),
			Category:    client.Attributes["ui.category"],
			Navigation:  firstNonEmpty(client.Attributes["ui.navigation"], client.Attributes["ui.sidenav"]),
			Enabled:     client.Enabled,
			Roles:       make([]KeycloakDiscoveredRole, 0, len(roles)),
		}
		for _, role := range roles {
			roleName := normalize(role.Name)
			if roleName == "" {
				continue
			}
			system.Roles = append(system.Roles, KeycloakDiscoveredRole{Name: roleName, Description: role.Description})
		}
		discovered.Systems = append(discovered.Systems, system)
	}

	return discovered, nil
}

func defaultRealmExportPath() (string, bool) {
	for _, path := range []string{
		"../keycloak/realm-export.json",
		"keycloak/realm-export.json",
		"../../keycloak/realm-export.json",
	} {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func shouldSkipClientID(clientID string) bool {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return true
	}
	for _, prefix := range []string{"account", "realm-management", "security-admin-console", "admin-cli"} {
		if strings.HasPrefix(clientID, prefix) {
			return true
		}
	}
	return false
}

func shouldSkipRealmRole(roleName string) bool {
	roleName = strings.TrimSpace(roleName)
	return roleName == "" || strings.HasPrefix(roleName, "default-roles-") || strings.HasPrefix(roleName, "uma_")
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func statusForExists(exists bool) string {
	if exists {
		return "present"
	}
	return "missing"
}

func discoveredRoleNames(roles []KeycloakDiscoveredRole) []string {
	values := make([]string, 0, len(roles))
	for _, role := range roles {
		values = append(values, role.Name)
	}
	sort.Strings(values)
	return values
}

func missingRoleNames(discovered []KeycloakDiscoveredRole, current []SystemRole) []string {
	currentSet := map[string]bool{}
	for _, role := range current {
		currentSet[normalize(role.Name)] = true
	}
	missing := make([]string, 0)
	for _, role := range discovered {
		if !currentSet[normalize(role.Name)] {
			missing = append(missing, role.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

func staleRoleNames(discovered []KeycloakDiscoveredRole, current []SystemRole) []string {
	discoveredSet := map[string]bool{}
	for _, role := range discovered {
		discoveredSet[normalize(role.Name)] = true
	}
	stale := make([]string, 0)
	for _, role := range current {
		if !discoveredSet[normalize(role.Name)] {
			stale = append(stale, role.Name)
		}
	}
	sort.Strings(stale)
	return stale
}

func missingAccessRoleNames(discovered []KeycloakDiscoveredRole, accessRoles []string) []string {
	accessSet := map[string]bool{}
	for _, role := range accessRoles {
		accessSet[normalize(role)] = true
	}
	missing := make([]string, 0)
	for _, role := range discovered {
		if !accessSet[normalize(role.Name)] {
			missing = append(missing, role.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

func roleDriftDetails(discovered []KeycloakDiscoveredRole, current []SystemRole) []RbacDriftRole {
	currentSet := map[string]bool{}
	for _, role := range current {
		currentSet[normalize(role.Name)] = true
	}
	values := make([]RbacDriftRole, 0, len(discovered))
	for _, role := range discovered {
		values = append(values, RbacDriftRole{Name: role.Name, Status: statusForExists(currentSet[normalize(role.Name)])})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
	return values
}

func updateSummary(summary *RbacDriftSummary, drift RbacDriftSystem) {
	if drift.KeycloakStatus == "present" && drift.RBACStatus == "present" &&
		len(drift.MissingRolesInRBAC) == 0 && len(drift.StaleRolesInRBAC) == 0 && len(drift.MissingAccessRoles) == 0 {
		summary.SystemsInSync++
	}
	if drift.RBACStatus == "missing" {
		summary.SystemsMissingInRBAC++
	}
	if drift.KeycloakStatus == "missing" {
		summary.SystemsStaleInRBAC++
	}
	summary.RolesMissingInRBAC += len(drift.MissingRolesInRBAC)
	summary.RolesStaleInRBAC += len(drift.StaleRolesInRBAC)
}
