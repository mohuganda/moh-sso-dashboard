package rbac

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/model"
)

type KeycloakSyncSource interface {
	ListClients() ([]keycloak.ClientInfo, error)
	ListRealmRoles(ctx context.Context) ([]keycloak.RoleRep, error)
	ListClientRoles(ctx context.Context, clientID string) ([]keycloak.ClientRoleRep, error)
	CreateClient(opts keycloak.CreateClientParams) (string, error)
	CreateRealmRole(ctx context.Context, roleName string, description string) error
	CreateClientRole(ctx context.Context, clientID string, req *model.CreateClientRoleRequest) error
}

type KeycloakPushResult struct {
	ClientsCreated     int      `json:"clientsCreated"`
	RealmRolesCreated  int      `json:"realmRolesCreated"`
	ClientRolesCreated int      `json:"clientRolesCreated"`
	Warnings           []string `json:"warnings,omitempty"`
}

func (s *Service) ApplyRealmExportFileSync(ctx context.Context, path string) (SyncApplyResponse, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		discoveredPath, ok := defaultRealmExportPath()
		if !ok {
			return SyncApplyResponse{
				Preview: SyncPreviewResponse{
					Report: RbacDriftReport{
						Source:   "default-realm-export",
						Warnings: []string{"No local keycloak/realm-export.json file was found for startup sync."},
					},
				},
			}, nil
		}
		path = discoveredPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return SyncApplyResponse{}, fmt.Errorf("read realm export %q: %w", path, err)
	}

	response, err := s.ApplyRealmExportSync(ctx, data)
	if err != nil {
		return SyncApplyResponse{}, err
	}
	response.Preview.Report.Source = path
	return response, nil
}

func (s *Service) ApplyLiveKeycloakSync(ctx context.Context, source KeycloakSyncSource) (SyncApplyResponse, error) {
	if source == nil {
		return SyncApplyResponse{}, fmt.Errorf("%w: keycloak source is required", ErrInvalidInput)
	}

	discovered, warnings, err := discoverLiveKeycloakRBAC(ctx, source)
	if err != nil {
		return SyncApplyResponse{}, err
	}

	preview, err := s.previewDiscoveredSync(ctx, "live-keycloak", discovered)
	if err != nil {
		return SyncApplyResponse{}, err
	}
	preview.Report.Warnings = append(preview.Report.Warnings, warnings...)

	response, err := s.applyDiscoveredSync(ctx, preview, discovered, "live-keycloak")
	if err != nil {
		return SyncApplyResponse{}, err
	}
	response.Preview.Report.Warnings = append(response.Preview.Report.Warnings, warnings...)
	return response, nil
}

func (s *Service) PushMissingRBACRolesToKeycloak(ctx context.Context, source KeycloakSyncSource) (KeycloakPushResult, error) {
	if source == nil {
		return KeycloakPushResult{}, fmt.Errorf("%w: keycloak source is required", ErrInvalidInput)
	}

	discovered, warnings, err := discoverLiveKeycloakRBAC(ctx, source)
	if err != nil {
		return KeycloakPushResult{}, err
	}

	result := KeycloakPushResult{Warnings: warnings}
	discoveredSystems := map[string]KeycloakDiscoveredSystem{}
	for _, system := range discovered.Systems {
		discoveredSystems[system.ClientID] = system
	}

	systems, err := s.repository.ListSystems(ctx)
	if err != nil {
		return KeycloakPushResult{}, err
	}

	for _, system := range systems {
		clientID := strings.TrimSpace(system.ClientID)
		if clientID == "" {
			continue
		}

		discoveredSystem, ok := discoveredSystems[clientID]
		if !ok {
			if _, err := source.CreateClient(keycloak.CreateClientParams{
				ClientID:    clientID,
				Name:        system.DisplayName,
				Description: system.Description,
				BaseURL:     system.LaunchURL,
				RootURL:     system.LaunchURL,
				Enabled:     system.Enabled,
				Attributes: map[string]string{
					"ui.icon":       system.Icon,
					"ui.home":       system.LaunchURL,
					"ui.category":   system.Category,
					"ui.navigation": system.Navigation,
					"ui.sidenav":    system.Navigation,
					"portal.system": "true",
				},
			}); err != nil {
				return KeycloakPushResult{}, fmt.Errorf("create keycloak client %q: %w", clientID, err)
			}
			result.ClientsCreated++
			discoveredSystem = KeycloakDiscoveredSystem{ClientID: clientID}
		}

		detail, err := s.repository.GetSystem(ctx, clientID)
		if err != nil {
			return KeycloakPushResult{}, err
		}

		keycloakRoles := map[string]bool{}
		for _, role := range discoveredSystem.Roles {
			keycloakRoles[normalize(role.Name)] = true
		}

		for _, role := range detail.Roles {
			roleName := normalize(role.Name)
			if roleName == "" || keycloakRoles[roleName] {
				continue
			}

			if err := source.CreateClientRole(ctx, clientID, &model.CreateClientRoleRequest{
				Role:        roleName,
				Description: role.Description,
			}); err != nil {
				return KeycloakPushResult{}, fmt.Errorf("create keycloak client role %q for %q: %w", roleName, clientID, err)
			}
			result.ClientRolesCreated++
		}
	}

	discoveredRealmRoles := map[string]bool{}
	for _, role := range discovered.RealmRoles {
		discoveredRealmRoles[normalize(role)] = true
	}

	realmGroups, err := s.repository.ListRealmRolePermissions(ctx)
	if err != nil {
		return KeycloakPushResult{}, err
	}
	for _, group := range realmGroups {
		roleName := normalize(group.RealmRole)
		if roleName == "" || discoveredRealmRoles[roleName] {
			continue
		}
		if err := source.CreateRealmRole(ctx, roleName, "Synced from MOH Integrated Health Portal"); err != nil {
			return KeycloakPushResult{}, fmt.Errorf("create keycloak realm role %q: %w", roleName, err)
		}
		result.RealmRolesCreated++
	}

	sort.Strings(result.Warnings)
	if err := s.recordAudit(ctx, "rbac.keycloak_push_applied", "sync", "live-keycloak", "", "", "", map[string]any{
		"clients_created":      result.ClientsCreated,
		"realm_roles_created":  result.RealmRolesCreated,
		"client_roles_created": result.ClientRolesCreated,
		"warnings":             result.Warnings,
	}); err != nil {
		return KeycloakPushResult{}, err
	}

	return result, nil
}

func (s *Service) previewDiscoveredSync(ctx context.Context, source string, discovered discoveredRBAC) (SyncPreviewResponse, error) {
	report, err := s.buildDriftReport(ctx, source, discovered)
	if err != nil {
		return SyncPreviewResponse{}, err
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
	return preview, nil
}

func (s *Service) applyDiscoveredSync(ctx context.Context, preview SyncPreviewResponse, discovered discoveredRBAC, source string) (SyncApplyResponse, error) {
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
			if err := s.recordAudit(ctx, "system_role.synced", "system-role", systemRole.ID, system.ClientID, systemRole.Name, "", map[string]any{"source": source}); err != nil {
				return SyncApplyResponse{}, err
			}
			response.RolesSynced++
			if err := s.repository.AddSystemAccessRole(ctx, system.ClientID, role.Name); err != nil {
				return SyncApplyResponse{}, err
			}
			if err := s.recordAudit(ctx, "system.access_role_synced", "system", system.ClientID, system.ClientID, role.Name, "", map[string]any{"source": source}); err != nil {
				return SyncApplyResponse{}, err
			}
			response.AccessRoles++
		}
	}

	if err := s.recordAudit(ctx, "rbac.sync_applied", "sync", source, "", "", "", map[string]any{"systemsSynced": response.SystemsSynced, "rolesSynced": response.RolesSynced, "accessRolesSynced": response.AccessRoles}); err != nil {
		return SyncApplyResponse{}, err
	}
	return response, nil
}

func discoverLiveKeycloakRBAC(ctx context.Context, source KeycloakSyncSource) (discoveredRBAC, []string, error) {
	clients, err := source.ListClients()
	if err != nil {
		return discoveredRBAC{}, nil, fmt.Errorf("list keycloak clients: %w", err)
	}

	realmRoles, err := source.ListRealmRoles(ctx)
	if err != nil {
		return discoveredRBAC{}, nil, fmt.Errorf("list keycloak realm roles: %w", err)
	}

	discovered := discoveredRBAC{
		Systems:    make([]KeycloakDiscoveredSystem, 0, len(clients)),
		RealmRoles: make([]string, 0, len(realmRoles)),
	}
	warnings := make([]string, 0)

	for _, role := range realmRoles {
		name := normalize(role.Name)
		if shouldSkipRealmRole(name) {
			continue
		}
		discovered.RealmRoles = append(discovered.RealmRoles, name)
	}

	for _, client := range clients {
		clientID := strings.TrimSpace(client.ClientID)
		if shouldSkipClientID(clientID) {
			continue
		}

		roles, err := source.ListClientRoles(ctx, clientID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to discover roles for client %q: %v", clientID, err))
			continue
		}

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

	sort.Strings(discovered.RealmRoles)
	sort.Slice(discovered.Systems, func(i, j int) bool {
		return discovered.Systems[i].ClientID < discovered.Systems[j].ClientID
	})
	for i := range discovered.Systems {
		sort.Slice(discovered.Systems[i].Roles, func(a, b int) bool {
			return discovered.Systems[i].Roles[a].Name < discovered.Systems[i].Roles[b].Name
		})
	}
	sort.Strings(warnings)
	return discovered, warnings, nil
}
