package repository

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	models "github.com/moh-sso-dashboard/internal/model"
)

type userRepository struct {
	keycloakClient *keycloak.KeyAdminClient
	config         *config.Config
	db             db.Store
	logger         *logger.Logger
}

func NewUserRepository(
	keycloakClient *keycloak.KeyAdminClient,
	config *config.Config,
	store db.Store,
	log logger.Logger,
) UserRepository {

	return &userRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             store,
		logger:         &log,
	}
}

func (r *userRepository) CreateUser(user *models.User) (string, error) {
	ctx := context.Background()

	kcID, err := r.keycloakClient.CreateUser(user)
	if err != nil {
		r.logger.Error("failed creating user in keycloak", "error", err)
		return "", fmt.Errorf("keycloak user creation failed: %w", err)
	}

	uid, err := uuid.Parse(kcID)
	if err != nil {
		return "", fmt.Errorf("invalid keycloak user id: %w", err)
	}

	if err := r.db.UpsertUser(ctx, db.UpsertUserParams{
		ID:       uid,
		Username: user.Username,
		Email:    user.Email,
		FirstName: sql.NullString{
			String: user.FirstName,
			Valid:  user.FirstName != "",
		},
		LastName: sql.NullString{
			String: user.LastName,
			Valid:  user.LastName != "",
		},
		Enabled: sql.NullBool{
			Bool:  user.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"user created in keycloak but db sync failed",
			"userId", kcID,
			"error", err,
		)
	}

	return kcID, nil
}

func (r *userRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	ctx := context.Background()

	kcUser, err := r.keycloakClient.GetUser(id.String())
	if err != nil {
		return nil, fmt.Errorf("keycloak user not found: %w", err)
	}

	row, err := r.db.GetUserByID(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var createdAt time.Time
	var updatedAt time.Time

	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &models.User{
		ID:            kcUser.ID,
		Username:      kcUser.Username,
		Email:         kcUser.Email,
		FirstName:     kcUser.FirstName,
		LastName:      kcUser.LastName,
		FullName:      strings.TrimSpace(kcUser.FirstName + " " + kcUser.LastName),
		RealmRoles:    kcUser.RealmRoles,
		ClientRoles:   kcUser.ClientRoles,
		IsAdmin:       slices.Contains(kcUser.RealmRoles, "admin"),
		Enabled:       kcUser.Enabled,
		EmailVerified: kcUser.EmailVerified,
		LastLoginAt:   kcUser.LastLoginAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (r *userRepository) ListUsers() ([]models.User, error) {
	ctx := context.Background()

	kcUsers, err := r.keycloakClient.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keycloak users: %w", err)
	}

	dbUsers, err := r.db.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	dbMap := map[string]db.User{}
	for _, u := range dbUsers {
		dbMap[u.ID.String()] = u
	}

	var result []models.User
	for _, kc := range kcUsers {
		if entry, ok := dbMap[kc.ID]; ok {
			result = append(result, models.User{
				ID:            entry.ID.String(),
				Username:      kc.Username,
				Email:         kc.Email,
				FirstName:     kc.FirstName,
				LastName:      kc.LastName,
				FullName:      strings.TrimSpace(kc.FirstName + " " + kc.LastName),
				RealmRoles:    kc.RealmRoles,
				ClientRoles:   kc.ClientRoles,
				IsAdmin:       slices.Contains(kc.RealmRoles, "admin"),
				Enabled:       kc.Enabled,
				EmailVerified: kc.EmailVerified,
				LastLoginAt:   pickTime(entry.LastLoginAt, kc.LastLoginAt),
				CreatedAt:     kc.CreatedAt,
			})
		} else {
			result = append(result, models.User{
				ID:            kc.ID,
				Username:      kc.Username,
				Email:         kc.Email,
				FirstName:     kc.FirstName,
				LastName:      kc.LastName,
				FullName:      strings.TrimSpace(kc.FirstName + " " + kc.LastName),
				RealmRoles:    kc.RealmRoles,
				ClientRoles:   kc.ClientRoles,
				IsAdmin:       slices.Contains(kc.RealmRoles, "admin"),
				Enabled:       kc.Enabled,
				EmailVerified: kc.EmailVerified,
				LastLoginAt:   kc.LastLoginAt,
				CreatedAt:     kc.CreatedAt,
			})
		}
	}

	return result, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	ctx := context.Background()

	if err := r.keycloakClient.UpdateUser(&models.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Enabled:   user.Enabled,
	}); err != nil {
		return fmt.Errorf("keycloak update failed: %w", err)
	}

	if err := r.db.UpsertUser(ctx, db.UpsertUserParams{
		ID:       uuid.MustParse(user.ID),
		Username: user.Username,
		Email:    user.Email,
		FirstName: sql.NullString{
			String: user.FirstName,
			Valid:  user.FirstName != "",
		},
		LastName: sql.NullString{
			String: user.LastName,
			Valid:  user.LastName != "",
		},
		Enabled: sql.NullBool{
			Bool:  user.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"user updated in keycloak but db sync failed",
			"userId", user.ID,
			"error", err,
		)
	}

	return nil
}

func (r *userRepository) DeleteUser(id string) error {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	if err := r.keycloakClient.DeleteUser(id); err != nil {
		return fmt.Errorf("keycloak delete failed: %w", err)
	}

	if err := r.db.DeleteUser(ctx, uid); err != nil {
		r.logger.Warn(
			"user deleted in keycloak but db delete failed",
			"userId", id,
			"error", err,
		)
	}

	return nil
}

func pickTime(db sql.NullTime, kc *time.Time) *time.Time {
	if db.Valid {
		return &db.Time
	}
	return kc
}

func (r *userRepository) ToggleUserEnabled(
	ctx context.Context,
	userID string,
	enabled bool,
) error {

	if err := r.keycloakClient.UpdateUser(&models.User{
		ID:      userID,
		Enabled: enabled,
	}); err != nil {
		return fmt.Errorf("keycloak toggle user failed: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	if err := r.db.UpdateUserEnabled(ctx, db.UpdateUserEnabledParams{
		ID:      uid,
		Enabled: sql.NullBool{Bool: enabled, Valid: true},
	}); err != nil {
		r.logger.Warn(
			"user enabled toggled in keycloak but db sync failed",
			"userId", userID,
			"enabled", enabled,
			"error", err,
		)
	}

	return nil
}

func (r *userRepository) ResetUserPassword(
	ctx context.Context,
	userID string,
) error {

	if err := r.keycloakClient.ResetUserPassword(ctx, userID); err != nil {
		return fmt.Errorf("reset password failed: %w", err)
	}

	r.logger.Info(
		"user password reset",
		"userId", userID,
	)

	return nil
}

func (r *userRepository) GetUserClientRoles(
	ctx context.Context,
	userID string,
) ([]keycloak.UserClientRoleAssignment, error) {

	clients, err := r.keycloakClient.ListClients()
	if err != nil {
		return nil, err
	}

	out := make([]keycloak.UserClientRoleAssignment, 0)

	for _, client := range clients {
		roles, err := r.keycloakClient.GetUserClientRoles(
			ctx,
			userID,
			"",
			client.ID,
		)
		if err != nil || len(roles) == 0 {
			continue
		}

		names := make([]string, 0, len(roles))
		for _, r := range roles {
			names = append(names, r.Name)
		}

		out = append(out, keycloak.UserClientRoleAssignment{
			Id:         client.ID,
			ClientID:   client.ClientID,
			ClientName: client.Name,
			Roles:      names,
		})
	}

	return out, nil
}

func (r *userRepository) AddUserClientRoles(
	ctx context.Context,
	userID string,
	clientID string,
	clientUUID string,
	roles []string,
) error {

	return r.keycloakClient.AssignClientRolesToUser(
		ctx,
		userID,
		clientID,
		clientUUID,
		roles,
	)
}

func (r *userRepository) RemoveUserClientRoles(
	ctx context.Context,
	userID string,
	clientID string,
	clientUUID string,
	roles []string,
) error {

	return r.keycloakClient.RemoveClientRolesFromUser(
		ctx,
		userID,
		clientID,
		clientUUID,
		roles,
	)
}

func (r *userRepository) GetUserClientRolesForClient(
	ctx context.Context,
	userID string,
	clientID string,
	clientUUID string,
) ([]keycloak.ClientRoleRep, error) {

	return r.keycloakClient.GetUserClientRoles(
		ctx,
		userID,
		clientID,
		clientUUID,
	)
}
