package rbac

import "context"

type Repository interface {
	ListSystems(ctx context.Context) ([]System, error)
	GetSystem(ctx context.Context, clientID string) (SystemDetail, error)
	UpsertSystem(ctx context.Context, input UpsertSystemInput) (System, error)
	ListPermissions(ctx context.Context) ([]Permission, error)

	ListSystemRoles(ctx context.Context, clientID string) ([]SystemRole, error)
	CreateSystemRole(ctx context.Context, clientID string, input RoleInput) (SystemRole, error)
	UpdateSystemRole(ctx context.Context, roleID string, input RoleInput) (SystemRole, error)
	DeleteSystemRole(ctx context.Context, roleID string) error

	AssignSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error
	RemoveSystemRolePermission(ctx context.Context, roleID string, permissionKey string) error

	ListRealmRolePermissions(ctx context.Context) ([]RealmRolePermissionGroup, error)
	AssignRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error
	RemoveRealmRolePermission(ctx context.Context, realmRole string, permissionKey string) error

	AddSystemAccessRole(ctx context.Context, clientID string, roleName string) error
	RemoveSystemAccessRole(ctx context.Context, clientID string, roleName string) error
	CountSystemAccessRoles(ctx context.Context, clientID string) (int, error)
	PermissionExists(ctx context.Context, permissionKey string) (bool, error)
}
