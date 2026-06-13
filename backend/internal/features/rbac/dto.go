package rbac

type System struct {
	ID          string `json:"id"`
	ClientID    string `json:"clientId"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	LaunchURL   string `json:"launchUrl,omitempty"`
	Category    string `json:"category,omitempty"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int32  `json:"sortOrder"`
}

type SystemDetail struct {
	System
	AccessRoles []string     `json:"accessRoles"`
	Roles       []SystemRole `json:"roles"`
}

type Permission struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

type SystemRole struct {
	ID          string       `json:"id"`
	SystemID    string       `json:"systemId"`
	ClientID    string       `json:"clientId"`
	Name        string       `json:"name"`
	DisplayName string       `json:"displayName,omitempty"`
	Description string       `json:"description,omitempty"`
	Enabled     bool         `json:"enabled"`
	Permissions []Permission `json:"permissions"`
}

type RealmRolePermissionGroup struct {
	RealmRole   string       `json:"realmRole"`
	Permissions []Permission `json:"permissions"`
}

type UpsertSystemInput struct {
	ClientID    string `json:"clientId"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	LaunchURL   string `json:"launchUrl"`
	Category    string `json:"category"`
	Enabled     *bool  `json:"enabled"`
	SortOrder   int32  `json:"sortOrder"`
}

type RoleInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type PermissionInput struct {
	PermissionKey string `json:"permissionKey" binding:"required"`
}

type AccessRoleInput struct {
	RoleName string `json:"roleName" binding:"required"`
}
