package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
)

type UserRepository interface {
	// ----------------------------------------------------
	// USERS
	// ----------------------------------------------------

	CreateUser(
		user *models.User,
	) (string, error)

	GetUserByID(
		id uuid.UUID,
	) (*models.User, error)

	GetUserByUsername(
		ctx context.Context,
		username string,
	) (*models.User, error)

	ListUsers() ([]models.User, error)

	SyncUsersFromKeycloak(
		ctx context.Context,
	) (int, error)

	UpdateUser(
		user *models.User,
	) error

	DeleteUser(
		id string,
	) error

	ToggleUserEnabled(
		ctx context.Context,
		userID string,
		enabled bool,
	) error

	// ResetUserPassword sends a Keycloak UPDATE_PASSWORD email.
	// It does not generate or store a temporary password.
	ResetUserPassword(
		ctx context.Context,
		userID string,
	) error

	// ----------------------------------------------------
	// USER EMAIL ACTIONS
	// ----------------------------------------------------

	SendUserOnboardingEmail(
		ctx context.Context,
		userID string,
	) error

	SendUserVerificationEmail(
		ctx context.Context,
		userID string,
	) error

	SendUserPasswordResetEmail(
		ctx context.Context,
		userID string,
	) error

	// ----------------------------------------------------
	// REALM ROLES
	// ----------------------------------------------------

	GetUsersByRealmRole(
		ctx context.Context,
		roleName string,
	) ([]models.User, error)

	AddUserRealmRoles(
		ctx context.Context,
		userID string,
		roles []string,
	) error

	RemoveUserRealmRoles(
		ctx context.Context,
		userID string,
		roles []string,
	) error

	// ----------------------------------------------------
	// CLIENT ROLES
	// ----------------------------------------------------

	GetUserClientRoles(
		ctx context.Context,
		userID string,
	) ([]keycloak.UserClientRoleAssignment, error)

	GetUserClientRolesForClient(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
	) ([]keycloak.ClientRoleRep, error)

	AddUserClientRoles(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
		roles []string,
	) error

	RemoveUserClientRoles(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
		roles []string,
	) error
}

type Repository = UserRepository
