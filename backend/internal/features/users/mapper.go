package users

import (
	"strings"

	models "github.com/moh-sso-dashboard/internal/model"
)

func toUserResponse(u *models.User) models.UserResponse {
	fullName := strings.TrimSpace(
		strings.Join([]string{u.FirstName, u.LastName}, " "),
	)

	return models.UserResponse{
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
