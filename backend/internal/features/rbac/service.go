package rbac

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrPermissionMissing = errors.New("permission does not exist")
	ErrLastAccessRole    = errors.New("cannot remove last access role")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListSystems(ctx context.Context) ([]System, error) {
	return s.repository.ListSystems(ctx)
}

func (s *Service) GetSystem(ctx context.Context, clientID string) (SystemDetail, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return SystemDetail{}, fmt.Errorf("%w: clientId is required", ErrInvalidInput)
	}
	return s.repository.GetSystem(ctx, clientID)
}

func (s *Service) UpsertSystem(ctx context.Context, input UpsertSystemInput) (System, error) {
	input.ClientID = strings.TrimSpace(input.ClientID)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Description = strings.TrimSpace(input.Description)
	input.Icon = strings.TrimSpace(input.Icon)
	input.LaunchURL = strings.TrimSpace(input.LaunchURL)
	input.Category = strings.TrimSpace(input.Category)
	if input.ClientID == "" {
		return System{}, fmt.Errorf("%w: clientId is required", ErrInvalidInput)
	}
	if input.DisplayName == "" {
		return System{}, fmt.Errorf("%w: displayName is required", ErrInvalidInput)
	}
	if err := validateLaunchURL(input.LaunchURL); err != nil {
		return System{}, err
	}
	return s.repository.UpsertSystem(ctx, input)
}

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	return s.repository.ListPermissions(ctx)
}

func (s *Service) ListSystemRoles(ctx context.Context, clientID string) ([]SystemRole, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("%w: clientId is required", ErrInvalidInput)
	}
	return s.repository.ListSystemRoles(ctx, clientID)
}

func (s *Service) CreateSystemRole(ctx context.Context, clientID string, input RoleInput) (SystemRole, error) {
	clientID = strings.TrimSpace(clientID)
	input = normalizeRoleInput(input)
	if clientID == "" || input.Name == "" {
		return SystemRole{}, fmt.Errorf("%w: clientId and role name are required", ErrInvalidInput)
	}
	return s.repository.CreateSystemRole(ctx, clientID, input)
}

func (s *Service) UpdateSystemRole(ctx context.Context, roleID string, input RoleInput) (SystemRole, error) {
	roleID = strings.TrimSpace(roleID)
	input = normalizeRoleInput(input)
	if roleID == "" || input.Name == "" {
		return SystemRole{}, fmt.Errorf("%w: roleId and role name are required", ErrInvalidInput)
	}
	return s.repository.UpdateSystemRole(ctx, roleID, input)
}

func (s *Service) DeleteSystemRole(ctx context.Context, roleID string) error {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return fmt.Errorf("%w: roleId is required", ErrInvalidInput)
	}
	return s.repository.DeleteSystemRole(ctx, roleID)
}

func (s *Service) AssignSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error {
	roleID = strings.TrimSpace(roleID)
	permissionKey = strings.TrimSpace(permissionKey)
	if roleID == "" || permissionKey == "" {
		return fmt.Errorf("%w: roleId and permissionKey are required", ErrInvalidInput)
	}
	if err := s.requirePermission(ctx, permissionKey); err != nil {
		return err
	}
	return s.repository.AssignSystemRolePermission(ctx, roleID, permissionKey)
}

func (s *Service) RemoveSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error {
	roleID = strings.TrimSpace(roleID)
	permissionKey = strings.TrimSpace(permissionKey)
	if roleID == "" || permissionKey == "" {
		return fmt.Errorf("%w: roleId and permissionKey are required", ErrInvalidInput)
	}
	return s.repository.RemoveSystemRolePermission(ctx, roleID, permissionKey)
}

func (s *Service) ListRealmRolePermissions(ctx context.Context) ([]RealmRolePermissionGroup, error) {
	return s.repository.ListRealmRolePermissions(ctx)
}

func (s *Service) AssignRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error {
	realmRole = strings.ToLower(strings.TrimSpace(realmRole))
	permissionKey = strings.TrimSpace(permissionKey)
	if realmRole == "" || permissionKey == "" {
		return fmt.Errorf("%w: realmRole and permissionKey are required", ErrInvalidInput)
	}
	if err := s.requirePermission(ctx, permissionKey); err != nil {
		return err
	}
	return s.repository.AssignRealmRolePermission(ctx, realmRole, permissionKey)
}

func (s *Service) RemoveRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error {
	realmRole = strings.ToLower(strings.TrimSpace(realmRole))
	permissionKey = strings.TrimSpace(permissionKey)
	if realmRole == "" || permissionKey == "" {
		return fmt.Errorf("%w: realmRole and permissionKey are required", ErrInvalidInput)
	}
	return s.repository.RemoveRealmRolePermission(ctx, realmRole, permissionKey)
}

func (s *Service) AddSystemAccessRole(ctx context.Context, clientID string, roleName string) error {
	clientID = strings.TrimSpace(clientID)
	roleName = strings.ToLower(strings.TrimSpace(roleName))
	if clientID == "" || roleName == "" {
		return fmt.Errorf("%w: clientId and roleName are required", ErrInvalidInput)
	}
	return s.repository.AddSystemAccessRole(ctx, clientID, roleName)
}

func (s *Service) RemoveSystemAccessRole(ctx context.Context, clientID string, roleName string, force bool) error {
	clientID = strings.TrimSpace(clientID)
	roleName = strings.ToLower(strings.TrimSpace(roleName))
	if clientID == "" || roleName == "" {
		return fmt.Errorf("%w: clientId and roleName are required", ErrInvalidInput)
	}
	if !force {
		count, err := s.repository.CountSystemAccessRoles(ctx, clientID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAccessRole
		}
	}
	return s.repository.RemoveSystemAccessRole(ctx, clientID, roleName)
}

func (s *Service) requirePermission(ctx context.Context, permissionKey string) error {
	exists, err := s.repository.PermissionExists(ctx, permissionKey)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPermissionMissing
	}
	return nil
}

func normalizeRoleInput(input RoleInput) RoleInput {
	input.Name = strings.ToLower(strings.TrimSpace(input.Name))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Description = strings.TrimSpace(input.Description)
	return input
}

func validateLaunchURL(value string) error {
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "/portal") {
		return nil
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return nil
	}
	return fmt.Errorf("%w: launchUrl must be a /portal path or absolute URL", ErrInvalidInput)
}
