package clients

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
)

type ClientRepository interface {
	CreateClient(app *models.Client) (string, error)
	GetClientByID(id uuid.UUID) (*models.Client, error)
	GetClientByClientID(clientID string) (*models.Client, error)
	ListClients() ([]models.Client, error)
	UpdateClient(app *models.Client) error
	DeleteClient(id uuid.UUID) error
	ToggleClientEnabled(
		ctx context.Context,
		clientID uuid.UUID,
		enabled bool,
	) error
	CreateClientRole(
		ctx context.Context,
		clientID uuid.UUID,
		payload *models.CreateClientRoleRequest,
	) error

	ListClientRoles(
		ctx context.Context,
		clientID uuid.UUID,
	) ([]keycloak.ClientRoleRep, error)

	GetClientRoleByName(
		ctx context.Context,
		clientID uuid.UUID,
		clientUUID uuid.UUID,
		role string,
	) (*keycloak.ClientRoleRep, error)

	DeleteClientRole(
		ctx context.Context,
		clientID uuid.UUID,
		role string,
	) error
}

type Repository = ClientRepository
