package clients

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
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

type Service struct {
	repo          Repository
	notifications sharedservice.NotificationsService
	cfg           *config.Config
}

func NewService(
	repo Repository,
	notifications sharedservice.NotificationsService,
	cfg ...*config.Config,
) *Service {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &Service{
		repo:          repo,
		notifications: notifications,
		cfg:           appConfig,
	}
}

func (s *Service) CreateClient(
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

	// In-app only.
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

func (s *Service) GetClient(id uuid.UUID) (*models.Client, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	if id == uuid.Nil {
		return nil, errors.New("client id is required")
	}

	return s.repo.GetClientByID(id)
}

func (s *Service) ListClients() ([]models.Client, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	return s.repo.ListClients()
}

func (s *Service) DeleteClient(
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

	if id == uuid.Nil {
		return errors.New("client id is required")
	}

	client, err := s.repo.GetClientByID(id)
	if err != nil {
		return fmt.Errorf("get client before delete: %w", err)
	}

	if err := s.repo.DeleteClient(id); err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	nt := models.ClientDeleted
	notification := models.Notification{
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
	}

	s.attachAdminEmailDelivery(
		&notification,
		"client-deleted",
		"Client application deleted",
		"A client application was deleted.",
		map[string]any{
			"Name":       s.systemAdminName(),
			"Platform":   s.platformName(),
			"ClientID":   client.ClientID,
			"ClientName": client.Name,
			"ActionURL":  s.adminDashboardURL(),
			"Details": fmt.Sprintf(
				"Client ID: %s\nClient Name: %s\nAdmin ID: %s",
				client.ClientID,
				client.Name,
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *Service) ToggleClientEnabled(
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

	if clientID == uuid.Nil {
		return errors.New("client id is required")
	}

	var client *models.Client
	if existing, err := s.repo.GetClientByID(clientID); err == nil {
		client = existing
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

	clientName := ""
	clientIdentifier := clientID.String()
	if client != nil {
		clientName = client.Name
		clientIdentifier = client.ClientID
	}

	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		ClientID:   clientIdentifier,
		Metadata: utils.MustJSON(map[string]any{
			"client_id":   clientID.String(),
			"client_name": clientName,
			"enabled":     enabled,
			"admin_id":    adminID.String(),
		}),
	}

	// Email only when disabled. Enabling remains in-app only.
	if !enabled {
		s.attachAdminEmailDelivery(
			&notification,
			"client-disabled",
			"Client application disabled",
			"A client application was disabled.",
			map[string]any{
				"Name":       s.systemAdminName(),
				"Platform":   s.platformName(),
				"ClientID":   clientIdentifier,
				"ClientName": clientName,
				"ActionURL":  s.adminDashboardURL(),
				"Details": fmt.Sprintf(
					"Client UUID: %s\nClient ID: %s\nClient Name: %s\nAdmin ID: %s",
					clientID.String(),
					clientIdentifier,
					clientName,
					adminID.String(),
				),
			},
		)
	}

	s.notify(ctx, notification)

	return nil
}

func (s *Service) CreateClientRole(
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

	if clientID == uuid.Nil {
		return errors.New("client id is required")
	}

	if payload == nil {
		return errors.New("client role payload is required")
	}

	role := strings.TrimSpace(payload.Role)
	if role == "" {
		return errors.New("role name is required")
	}

	payload.Role = role

	var client *models.Client
	if existing, err := s.repo.GetClientByID(clientID); err == nil {
		client = existing
	}

	if err := s.repo.CreateClientRole(ctx, clientID, payload); err != nil {
		return fmt.Errorf("create client role: %w", err)
	}

	clientName := ""
	clientIdentifier := clientID.String()
	if client != nil {
		clientName = client.Name
		clientIdentifier = client.ClientID
	}

	nt := models.ClientRoleCreated
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role created",
		TargetRole: "admin",
		ClientID:   clientIdentifier,
		Metadata: utils.MustJSON(map[string]any{
			"client_id":   clientID.String(),
			"client_name": clientName,
			"role":        role,
			"admin_id":    adminID.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"client-role-created",
		"Client role created",
		"A client role was created.",
		map[string]any{
			"Name":       s.systemAdminName(),
			"Platform":   s.platformName(),
			"ClientID":   clientIdentifier,
			"ClientName": clientName,
			"Role":       role,
			"ActionURL":  s.adminDashboardURL(),
			"Details": fmt.Sprintf(
				"Client UUID: %s\nClient ID: %s\nClient Name: %s\nRole: %s\nAdmin ID: %s",
				clientID.String(),
				clientIdentifier,
				clientName,
				role,
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *Service) ListClientRoles(
	ctx context.Context,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {
	if s == nil {
		return nil, errors.New("client service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("client repository is nil")
	}

	if clientID == uuid.Nil {
		return nil, errors.New("client id is required")
	}

	return s.repo.ListClientRoles(ctx, clientID)
}

func (s *Service) DeleteClientRole(
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

	if clientID == uuid.Nil {
		return errors.New("client id is required")
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return errors.New("role name is required")
	}

	var client *models.Client
	if existing, err := s.repo.GetClientByID(clientID); err == nil {
		client = existing
	}

	if err := s.repo.DeleteClientRole(ctx, clientID, role); err != nil {
		return fmt.Errorf("delete client role: %w", err)
	}

	clientName := ""
	clientIdentifier := clientID.String()
	if client != nil {
		clientName = client.Name
		clientIdentifier = client.ClientID
	}

	nt := models.ClientRoleDeleted
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role deleted",
		TargetRole: "admin",
		ClientID:   clientIdentifier,
		Metadata: utils.MustJSON(map[string]any{
			"client_id":   clientID.String(),
			"client_name": clientName,
			"role":        role,
			"admin_id":    adminID.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"client-role-deleted",
		"Client role deleted",
		"A client role was deleted.",
		map[string]any{
			"Name":       s.systemAdminName(),
			"Platform":   s.platformName(),
			"ClientID":   clientIdentifier,
			"ClientName": clientName,
			"Role":       role,
			"ActionURL":  s.adminDashboardURL(),
			"Details": fmt.Sprintf(
				"Client UUID: %s\nClient ID: %s\nClient Name: %s\nRole: %s\nAdmin ID: %s",
				clientID.String(),
				clientIdentifier,
				clientName,
				role,
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *Service) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	_, _ = s.notifications.Notify(ctx, notification)
}

func (s *Service) attachAdminEmailDelivery(
	notification *models.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminDashboardURL()
	}

	notification.Deliveries = []models.NotificationDeliveryRequest{
		{
			Channel: models.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: models.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *Service) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *Service) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *Service) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *Service) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}
