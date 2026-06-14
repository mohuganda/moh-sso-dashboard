package clients

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	models "github.com/moh-sso-dashboard/internal/model"
)

type sqlcClientRepository struct {
	keycloakClient *keycloak.KeyAdminClient
	config         *config.Config
	db             db.Store
	logger         *logger.Logger
}

func NewClientRepository(
	keycloakClient *keycloak.KeyAdminClient,
	config *config.Config,
	db db.Store,
	log logger.Logger,
) ClientRepository {

	return &sqlcClientRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             db,
		logger:         &log,
	}
}

func (r *sqlcClientRepository) CreateClient(client *models.Client) (string, error) {
	ctx := context.Background()

	kcID, err := r.keycloakClient.CreateClient(keycloak.CreateClientParams{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		RootURL:      client.RootURL,
		Protocol:     "openid-connect",
		PublicClient: client.PublicClient,
		RedirectURIs: client.RedirectUris,
		WebOrigins:   client.WebOrigins,
		Enabled:      client.Enabled,
		Attributes: map[string]string{
			// UI
			"ui.icon":    "applications",
			"ui.home":    "/admin",
			"ui.sidenav": client.Attributes["ui.sidenav"],
		},
	})

	if err != nil {
		return "", fmt.Errorf("keycloak create failed: %w", err)
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(client.ClientID)
	if err != nil {
		return "", fmt.Errorf("keycloak read-back failed: %w", err)
	}

	id, _ := uuid.Parse(kcID)

	var icon sql.NullString
	if v := kcClient.Attributes["ui.icon"]; v != "" {
		icon = sql.NullString{String: v, Valid: true}
	}

	attrsJSON, err := json.Marshal(kcClient.Attributes)
	if err != nil {
		return "", fmt.Errorf("marshal client attributes failed: %w", err)
	}

	if err := r.db.UpsertClient(ctx, db.UpsertClientParams{
		ID:       id,
		ClientID: kcClient.ClientID,
		Name:     kcClient.Name,
		Description: sql.NullString{
			String: kcClient.Description,
			Valid:  kcClient.Description != "",
		},
		Icon: icon,
		BaseUrl: sql.NullString{
			String: kcClient.BaseURL,
			Valid:  kcClient.BaseURL != "",
		},
		RootUrl: sql.NullString{
			String: kcClient.RootURL,
			Valid:  kcClient.RootURL != "",
		},
		AdminUrl: sql.NullString{
			String: kcClient.AdminURL,
			Valid:  kcClient.AdminURL != "",
		},
		PublicClient: kcClient.PublicClient,
		RedirectUris: kcClient.RedirectURIs,
		WebOrigins:   kcClient.WebOrigins,
		Enabled:      kcClient.Enabled,
		Attributes:   attrsJSON,
	}); err != nil {
		r.logger.Warn(
			"client created in keycloak but db sync failed",
			"clientId", kcClient.ClientID,
			"error", err,
		)
	}

	return kcID, nil
}

func (r *sqlcClientRepository) GetClientByID(id uuid.UUID) (*models.Client, error) {
	ctx := context.Background()

	dbRow, err := r.db.GetClientByID(ctx, id)
	if err != nil {
		return nil, err
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(dbRow.ClientID)
	if err != nil {
		return nil, err
	}

	return &models.Client{
		ID:           dbRow.ID.String(),
		ClientID:     kcClient.ClientID,
		Name:         kcClient.Name,
		Description:  kcClient.Description,
		BaseURL:      kcClient.BaseURL,
		RootURL:      kcClient.RootURL,
		AdminURL:     kcClient.AdminURL,
		RedirectUris: kcClient.RedirectURIs,
		WebOrigins:   kcClient.WebOrigins,
		PublicClient: kcClient.PublicClient,
		Enabled:      kcClient.Enabled,
		Attributes:   kcClient.Attributes,
	}, nil
}

func (r *sqlcClientRepository) GetClientByClientID(clientID string) (*models.Client, error) {
	kc, err := r.keycloakClient.GetClientByClientID(clientID)
	if err != nil || kc == nil {
		return nil, err
	}

	client := &models.Client{
		ID:           kc.ID,
		ClientID:     kc.ClientID,
		Name:         kc.Name,
		Description:  kc.Description,
		BaseURL:      kc.BaseURL,
		RootURL:      kc.RootURL,
		AdminURL:     kc.AdminURL,
		RedirectUris: kc.RedirectURIs,
		WebOrigins:   kc.WebOrigins,
		PublicClient: kc.PublicClient,
		Enabled:      kc.Enabled,
		Attributes:   kc.Attributes,
	}

	return client, nil
}

func (r *sqlcClientRepository) ListClients() ([]models.Client, error) {
	kcClients, err := r.keycloakClient.ListClients()
	if err != nil {
		return nil, err
	}

	out := make([]models.Client, 0, len(kcClients))

	for _, kc := range kcClients {
		client := models.Client{
			ID:           kc.ID,
			ClientID:     kc.ClientID,
			Name:         kc.Name,
			Description:  kc.Description,
			BaseURL:      kc.BaseURL,
			RootURL:      kc.RootURL,
			AdminURL:     kc.AdminURL,
			RedirectUris: kc.RedirectURIs,
			WebOrigins:   kc.WebOrigins,
			PublicClient: kc.PublicClient,
			Enabled:      kc.Enabled,
			Attributes:   kc.Attributes,
		}

		out = append(out, client)
	}

	return out, nil
}

// UPDATE (KC first, DB best-effort)
func (r *sqlcClientRepository) UpdateClient(client *models.Client) error {
	ctx := context.Background()

	if err := r.keycloakClient.UpdateClient(client.ID, client); err != nil {
		return err
	}

	id, err := uuid.Parse(client.ID)
	if err != nil {
		return fmt.Errorf("invalid client UUID after keycloak update: %w", err)
	}

	var icon sql.NullString
	if client.Attributes["ui.icon"] != "" {
		icon = sql.NullString{String: client.Attributes["ui.icon"], Valid: true}
	}

	attrsJSON, err := json.Marshal(client.Attributes)
	if err != nil {
		return fmt.Errorf("marshal client attributes failed: %w", err)
	}

	if err := r.db.UpsertClient(ctx, db.UpsertClientParams{
		ID:       id,
		ClientID: client.ClientID,
		Name:     client.Name,
		Description: sql.NullString{
			String: client.Description,
			Valid:  client.Description != "",
		},
		Icon: icon,
		BaseUrl: sql.NullString{
			String: client.BaseURL,
			Valid:  client.BaseURL != "",
		},
		RootUrl: sql.NullString{
			String: client.RootURL,
			Valid:  client.RootURL != "",
		},
		AdminUrl: sql.NullString{
			String: client.AdminURL,
			Valid:  client.AdminURL != "",
		},
		PublicClient: client.PublicClient,
		RedirectUris: client.RedirectUris,
		WebOrigins:   client.WebOrigins,
		Enabled:      client.Enabled,
		Attributes:   attrsJSON,
	}); err != nil {
		r.logger.Warn(
			"client updated in keycloak but db sync failed",
			"clientId", client.ClientID,
			"error", err,
		)
	}

	return nil
}

// DELETE
func (r *sqlcClientRepository) DeleteClient(id uuid.UUID) error {
	ctx := context.Background()

	dbClient, err := r.db.GetClientByID(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if dbClient.ClientID != "" {
		if err := r.keycloakClient.DeleteClient(dbClient.ClientID); err != nil {
			return err
		}
	}

	return r.db.DeleteClient(ctx, id)
}

func (r *sqlcClientRepository) CreateClientRole(ctx context.Context, clientID uuid.UUID, payload *models.CreateClientRoleRequest) error {
	dbClient, err := r.db.GetClientByID(ctx, clientID)
	clientIdentifier := clientID.String()
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && dbClient.ClientID != "" {
		clientIdentifier = dbClient.ClientID
	}

	return r.keycloakClient.CreateClientRole(
		ctx,
		clientIdentifier,
		payload,
	)
}

func (r *sqlcClientRepository) ListClientRoles(
	ctx context.Context,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {
	dbClient, err := r.db.GetClientByID(ctx, clientID)
	clientIdentifier := clientID.String()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && dbClient.ClientID != "" {
		clientIdentifier = dbClient.ClientID
	}

	roles, err := r.keycloakClient.ListClientRoles(
		ctx,
		clientIdentifier,
	)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *sqlcClientRepository) DeleteClientRole(ctx context.Context, clientID uuid.UUID, role string) error {
	dbClient, err := r.db.GetClientByID(ctx, clientID)
	clientIdentifier := clientID.String()
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && dbClient.ClientID != "" {
		clientIdentifier = dbClient.ClientID
	}

	return r.keycloakClient.DeleteClientRole(
		ctx,
		clientIdentifier,
		role,
	)
}

func (r *sqlcClientRepository) ToggleClientEnabled(
	ctx context.Context,
	clientID uuid.UUID,
	enabled bool,
) error {

	dbClient, err := r.db.GetClientByID(ctx, clientID)
	if err != nil {
		return err
	}

	if err := r.keycloakClient.UpdateClientEnabled(ctx,
		dbClient.ClientID,
		enabled,
	); err != nil {
		return fmt.Errorf("keycloak toggle client failed: %w", err)
	}

	if err := r.db.UpdateClientEnabled(ctx, db.UpdateClientEnabledParams{
		ID:      clientID,
		Enabled: enabled,
	}); err != nil {
		r.logger.Warn(
			"client enabled toggled in keycloak but db sync failed",
			"clientId", dbClient.ClientID,
			"enabled", enabled,
			"error", err,
		)
	}

	return nil
}

func (r *sqlcClientRepository) GetClientRoleByName(
	ctx context.Context,
	clientID uuid.UUID,
	clientUuid uuid.UUID,
	role string,
) (*keycloak.ClientRoleRep, error) {

	return r.keycloakClient.GetClientRoleByName(
		ctx,
		clientID.String(),
		clientUuid.String(),
		role,
	)
}
