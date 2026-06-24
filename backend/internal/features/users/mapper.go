package users

import (
	"strings"

	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
)

func toUserResponse(u *models.User) UserResponse {
	fullName := strings.TrimSpace(
		strings.Join([]string{u.FirstName, u.LastName}, " "),
	)

	return UserResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		FullName:         fullName,
		IsAdmin:          u.IsAdmin,
		RealmRoles:       u.RealmRoles,
		ClientRoles:      u.ClientRoles,
		IsActive:         u.Enabled,
		EmailVerified:    u.EmailVerified,
		RequirePwdChange: u.RequirePwdChange,
		LastLoginAt:      u.LastLoginAt,
		CreatedAt:        &u.CreatedAt,
	}
}

func toUserClientRoleAssignmentResponse(item keycloak.UserClientRoleAssignment) UserClientRoleAssignmentResponse {
	return UserClientRoleAssignmentResponse{
		ID:         item.Id,
		ClientID:   item.ClientID,
		ClientName: item.ClientName,
		Roles:      item.Roles,
	}
}

func toUserClientRoleAssignmentResponses(items []keycloak.UserClientRoleAssignment) []UserClientRoleAssignmentResponse {
	out := make([]UserClientRoleAssignmentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toUserClientRoleAssignmentResponse(item))
	}
	return out
}

func toClientRoleResponse(role keycloak.ClientRoleRep) ClientRoleResponse {
	return ClientRoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
}

func toClientRoleResponses(roles []keycloak.ClientRoleRep) []ClientRoleResponse {
	out := make([]ClientRoleResponse, 0, len(roles))
	for _, role := range roles {
		out = append(out, toClientRoleResponse(role))
	}
	return out
}
