package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/client"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateClientRequest struct {
	Name                   string            `json:"name" validate:"required"`
	Description            string            `json:"description,omitempty"`
	ClientID               string            `json:"clientId" validate:"required"`
	ClientSecret           string            `json:"clientSecret"`
	RedirectURIs           []string          `json:"redirectUris" validate:"required,dive,uri"`
	WebOrigins             []string          `json:"webOrigins"`
	StandardFlowEnabled    bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled    bool              `json:"implicitFlowEnabled"`
	DirectAccessGrants     bool              `json:"directAccessGrants"`
	ServiceAccountsEnabled bool              `json:"serviceAccountsEnabled"`
	PublicClient           bool              `json:"publicClient"`
	RootURL                string            `json:"rootUrl"`
	BaseURL                string            `json:"baseUrl"`
	AdminURL               string            `json:"adminUrl"`
	Enabled                bool              `json:"enabled"`
	Protocol               string            `json:"protocol,omitempty"`
	LoginURI               string            `json:"loginUri,omitempty"`
	LogoutURI              string            `json:"logoutUri,omitempty"`
	Attributes             map[string]string `json:"attributes"`
	DefaultClientScopes    []string          `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes   []string          `json:"optionalClientScopes,omitempty"`
	Tags                   []string          `json:"tags,omitempty"`
}

type ClientService struct {
	repo          repository.ClientRepository
	notifications NotificationsService
}

func NewClientService(
	repo repository.ClientRepository,
	notifications NotificationsService,
) *ClientService {
	return &ClientService{
		repo:          repo,
		notifications: notifications,
	}
}

func (s *ClientService) CreateClient(
	ctx context.Context,
	req CreateClientRequest,
	adminID uuid.UUID,
) (*models.Client, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("client name is required")
	}

	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		clientID = utils.GenerateClientID()
	}

	client := &models.Client{
		ClientID:     clientID,
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		BaseURL:      strings.TrimSpace(req.BaseURL),
		RootURL:      strings.TrimSpace(req.RootURL),
		RedirectUris: req.RedirectURIs,
		WebOrigins:   req.WebOrigins,
		PublicClient: req.PublicClient,
		Enabled:      req.Enabled,
		Attributes:   req.Attributes,
	}

	if _, err := s.repo.CreateClient(client); err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	nt := models.ClientCreated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New client application created",
		TargetRole: "admin",
		ClientID:   client.ClientID,
		Metadata: utils.MustJSON(map[string]any{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return client, nil
}

func (s *ClientService) GetClient(id uuid.UUID) (*models.Client, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	return s.repo.GetClientByID(id)
}

func (s *ClientService) ListClients() ([]models.Client, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	return s.repo.ListClients()
}

func (s *ClientService) DeleteClient(
	ctx context.Context,
	id uuid.UUID,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("client service is nil")
	}

	if s.repo == nil {
		return errors.New("client repository is nil")
	}

	client, err := s.repo.GetClientByID(id)
	if err != nil {
		return fmt.Errorf("get client before delete: %w", err)
	}

	if err := s.repo.DeleteClient(id); err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	nt := models.ClientDeleted
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client application deleted",
		TargetRole: "admin",
		ClientID:   client.ClientID,
		Metadata: utils.MustJSON(map[string]any{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *ClientService) ToggleClientEnabled(
	ctx context.Context,
	clientID uuid.UUID,
	enabled bool,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("client service is nil")
	}

	if s.repo == nil {
		return errors.New("client repository is nil")
	}

	if err := s.repo.ToggleClientEnabled(ctx, clientID, enabled); err != nil {
		return fmt.Errorf("toggle client enabled: %w", err)
	}

	var nt models.NotificationType
	msg := "Client updated"

	if enabled {
		nt = models.ClientEnabled
		msg = "Client enabled"
	} else {
		nt = models.ClientDisabled
		msg = "Client disabled"
	}

	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		ClientID:   clientID.String(),
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"enabled":   enabled,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *ClientService) CreateClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	payload *models.CreateClientRoleRequest,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("client service is nil")
	}

	if s.repo == nil {
		return errors.New("client repository is nil")
	}

	if payload == nil {
		return errors.New("client role payload is required")
	}

	role := strings.TrimSpace(payload.Role)
	if role == "" {
		return errors.New("role name is required")
	}

	payload.Role = role

	if err := s.repo.CreateClientRole(ctx, clientID, payload); err != nil {
		return fmt.Errorf("create client role: %w", err)
	}

	nt := models.ClientRoleCreated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role created",
		TargetRole: "admin",
		ClientID:   clientID.String(),
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *ClientService) ListClientRoles(
	ctx context.Context,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	return s.repo.ListClientRoles(ctx, clientID)
}

func (s *ClientService) DeleteClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("client service is nil")
	}

	if s.repo == nil {
		return errors.New("client repository is nil")
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return errors.New("role name is required")
	}

	if err := s.repo.DeleteClientRole(ctx, clientID, role); err != nil {
		return fmt.Errorf("delete client role: %w", err)
	}

	nt := models.ClientRoleDeleted
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role deleted",
		TargetRole: "admin",
		ClientID:   clientID.String(),
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *ClientService) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		fmt.Printf("client notification failed type=%s error=%v\n", notification.Type, err)
	}
}
