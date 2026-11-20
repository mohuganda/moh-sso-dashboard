package core

// HealthScope carries the authenticated geographic claims for a request.
// Consumers apply their own resource-specific scope policy.
type HealthScope struct {
	District string
	Facility string
}

// RequestAccess is the trusted host's identity and access decision for a request.
type RequestAccess struct {
	Identity
	Permissions []string
	Systems     []string
	HealthScope HealthScope
	IsAdmin     bool
}

func (a RequestAccess) HasPermission(permission string) bool {
	return HasPermission(a.Permissions, permission)
}

func (a RequestAccess) HasSystem(system string) bool {
	return HasSystem(a.Systems, system)
}
