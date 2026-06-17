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

	// SendUserOnboardingEmail sends VERIFY_EMAIL + UPDATE_PASSWORD.
	// Use this after creating a new user.
	SendUserOnboardingEmail(
		ctx context.Context,
		userID string,
	) error

	// SendUserVerificationEmail sends VERIFY_EMAIL only.
	// Use this when resending email verification.
	SendUserVerificationEmail(
		ctx context.Context,
		userID string,
	) error

	// SendUserPasswordResetEmail sends UPDATE_PASSWORD only.
	// Use this for admin-triggered password resets.
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
