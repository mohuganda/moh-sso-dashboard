package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateUserRequest struct {
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"first_name,omitempty"`
	LastName         string              `json:"last_name,omitempty"`
	FullName         string              `json:"full_name,omitempty"`
	Enabled          *bool               `json:"enabled,omitempty"`
	RequirePwdChange *bool               `json:"require_pwd_change,omitempty"`
	Password         string              `json:"password,omitempty"`
	SendInvite       bool                `json:"send_invite,omitempty"`
	RealmRoles       []string            `json:"realm_roles,omitempty"`
	ClientRoles      map[string][]string `json:"client_roles,omitempty"`
}

type UserService struct {
	repo          repository.UserRepository
	notifications NotificationsService
	cfg           *config.Config
}

func NewUserService(
	repo repository.UserRepository,
	notifications NotificationsService,
	cfg ...*config.Config,
) *UserService {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &UserService{
		repo:          repo,
		notifications: notifications,
		cfg:           appConfig,
	}
}

//
// ----------------------------------------------------
// USER LIFECYCLE
// ----------------------------------------------------
//

func (s *UserService) CreateUser(
	ctx context.Context,
	req CreateUserRequest,
	adminID uuid.UUID,
) (*models.User, error) {
	if s == nil {
		return nil, errors.New("user service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("user repository is nil")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   enabled,
	}

	if _, err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// In-app only.
	nt := models.UserCreated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New user account created",
		TargetRole: "admin",
		UserID:     user.ID,
		Metadata: utils.MustJSON(map[string]any{
			"user_id":            user.ID,
			"username":           user.Username,
			"email":              user.Email,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
			"enabled":            user.Enabled,
			"send_invite":        req.SendInvite,
			"require_pwd_change": req.RequirePwdChange != nil && *req.RequirePwdChange,
			"realm_roles":        req.RealmRoles,
			"client_roles":       req.ClientRoles,
			"admin_id":           adminID.String(),
		}),
	})

	return user, nil
}

func (s *UserService) SetUserEnabled(
	ctx context.Context,
	userID uuid.UUID,
	enabled bool,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("user service is nil")
	}

	if s.repo == nil {
		return errors.New("user repository is nil")
	}

	if userID == uuid.Nil {
		return errors.New("user id is required")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("get user before toggle enabled: %w", err)
	}

	if user.Enabled == enabled {
		return nil
	}

	if err := s.repo.ToggleUserEnabled(ctx, userID.String(), enabled); err != nil {
		return fmt.Errorf("toggle user enabled: %w", err)
	}

	var nt models.NotificationType
	msg := "User account updated"

	if enabled {
		nt = models.UserEnabled
		msg = "User account enabled"
	} else {
		nt = models.UserDisabled
		msg = "User account disabled"
	}

	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		UserID:     user.ID,
		Metadata: utils.MustJSON(map[string]any{
			"user_id":          user.ID,
			"username":         user.Username,
			"email":            user.Email,
			"previous_enabled": user.Enabled,
			"enabled":          enabled,
			"admin_id":         adminID.String(),
		}),
	}

	// Email only when disabled. Enabling remains in-app only.
	if !enabled {
		s.attachAdminEmailDelivery(
			&notification,
			"user-disabled",
			"User account disabled",
			fmt.Sprintf("User account %s was disabled.", user.Username),
			map[string]any{
				"Name":      s.systemAdminName(),
				"Platform":  s.platformName(),
				"Username":  user.Username,
				"Email":     user.Email,
				"ActionURL": s.adminUsersURL(),
				"Details": fmt.Sprintf(
					"User ID: %s\nUsername: %s\nEmail: %s\nPrevious Enabled: %v\nEnabled: %v\nAdmin ID: %s",
					user.ID,
					user.Username,
					user.Email,
					user.Enabled,
					enabled,
					adminID.String(),
				),
			},
		)
	}

	s.notify(ctx, notification)

	return nil
}

func (s *UserService) ResetUserPassword(
	ctx context.Context,
	userID uuid.UUID,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("user service is nil")
	}

	if s.repo == nil {
		return errors.New("user repository is nil")
	}

	if userID == uuid.Nil {
		return errors.New("user id is required")
	}

	user, _ := s.repo.GetUserByID(userID)

	if err := s.repo.ResetUserPassword(ctx, userID.String()); err != nil {
		return fmt.Errorf("reset user password: %w", err)
	}

	nt := models.UserPasswordReset
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User password reset",
		TargetRole: "admin",
		UserID:     userID.String(),
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  userID.String(),
			"username": user.Username,
			"email":    user.Email,
			"admin_id": adminID.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"user-password-reset",
		"User password reset",
		fmt.Sprintf("Password reset was triggered for user %s.", user.Username),
		map[string]any{
			"Name":      s.systemAdminName(),
			"Platform":  s.platformName(),
			"Username":  user.Username,
			"Email":     user.Email,
			"ActionURL": s.adminUsersURL(),
			"Details": fmt.Sprintf(
				"User ID: %s\nUsername: %s\nEmail: %s\nAdmin ID: %s",
				userID.String(),
				user.Username,
				user.Email,
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *UserService) GetUserClientRoles(
	ctx context.Context,
	userID uuid.UUID,
) ([]keycloak.UserClientRoleAssignment, error) {
	if s == nil {
		return nil, errors.New("user service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("user repository is nil")
	}

	if userID == uuid.Nil {
		return nil, errors.New("user id is required")
	}

	return s.repo.GetUserClientRoles(ctx, userID.String())
}

func (s *UserService) GetUserClientRolesForClient(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
	clientUUID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {
	if s == nil {
		return nil, errors.New("user service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("user repository is nil")
	}

	if userID == uuid.Nil {
		return nil, errors.New("user id is required")
	}

	if clientID == uuid.Nil {
		return nil, errors.New("client id is required")
	}

	if clientUUID == uuid.Nil {
		return nil, errors.New("client uuid is required")
	}

	return s.repo.GetUserClientRolesForClient(
		ctx,
		userID.String(),
		clientID.String(),
		clientUUID.String(),
	)
}

func (s *UserService) UpdateUserClientRoles(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
	clientUUID uuid.UUID,
	roles []string,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("user service is nil")
	}

	if s.repo == nil {
		return errors.New("user repository is nil")
	}

	if userID == uuid.Nil {
		return errors.New("user id is required")
	}

	if clientID == uuid.Nil {
		return errors.New("client id is required")
	}

	if clientUUID == uuid.Nil {
		return errors.New("client uuid is required")
	}

	roles = normalizeRoleNames(roles)

	current, err := s.repo.GetUserClientRolesForClient(
		ctx,
		userID.String(),
		clientID.String(),
		clientUUID.String(),
	)
	if err != nil {
		return fmt.Errorf("get current user client roles: %w", err)
	}

	currentSet := make(map[string]bool)
	for _, r := range current {
		roleName := strings.TrimSpace(r.Name)
		if roleName != "" {
			currentSet[roleName] = true
		}
	}

	desiredSet := make(map[string]bool)
	for _, r := range roles {
		desiredSet[r] = true
	}

	var toAdd []string
	var toRemove []string

	for r := range desiredSet {
		if !currentSet[r] {
			toAdd = append(toAdd, r)
		}
	}

	for r := range currentSet {
		if !desiredSet[r] {
			toRemove = append(toRemove, r)
		}
	}

	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	if len(toAdd) > 0 {
		if err := s.repo.AddUserClientRoles(
			ctx,
			userID.String(),
			clientID.String(),
			clientUUID.String(),
			toAdd,
		); err != nil {
			return fmt.Errorf("add user client roles: %w", err)
		}
	}

	if len(toRemove) > 0 {
		if err := s.repo.RemoveUserClientRoles(
			ctx,
			userID.String(),
			clientID.String(),
			clientUUID.String(),
			toRemove,
		); err != nil {
			return fmt.Errorf("remove user client roles: %w", err)
		}
	}

	user, _ := s.repo.GetUserByID(userID)

	nt := models.ClientRolesUpdated
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User client roles updated",
		TargetRole: "admin",
		UserID:     userID.String(),
		ClientID:   clientID.String(),
		Metadata: utils.MustJSON(map[string]any{
			"user_id":     userID.String(),
			"username":    user.Username,
			"email":       user.Email,
			"client_id":   clientID.String(),
			"client_uuid": clientUUID.String(),
			"added":       toAdd,
			"removed":     toRemove,
			"roles":       roles,
			"admin_id":    adminID.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"client-roles-updated",
		"User client roles updated",
		fmt.Sprintf("Client roles were updated for user %s.", user.Username),
		map[string]any{
			"Name":       s.systemAdminName(),
			"Platform":   s.platformName(),
			"Username":   user.Username,
			"Email":      user.Email,
			"ClientID":   clientID.String(),
			"ClientUUID": clientUUID.String(),
			"AddedRoles": strings.Join(toAdd, ", "),
			"RemovedRoles": strings.Join(
				toRemove,
				", ",
			),
			"ActionURL": s.adminUsersURL(),
			"Details": fmt.Sprintf(
				"User ID: %s\nUsername: %s\nEmail: %s\nClient ID: %s\nClient UUID: %s\nAdded Roles: %s\nRemoved Roles: %s\nAll Roles: %s\nAdmin ID: %s",
				userID.String(),
				user.Username,
				user.Email,
				clientID.String(),
				clientUUID.String(),
				strings.Join(toAdd, ", "),
				strings.Join(toRemove, ", "),
				strings.Join(roles, ", "),
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *UserService) GetUser(id uuid.UUID) (*models.User, error) {
	if s == nil {
		return nil, errors.New("user service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("user repository is nil")
	}

	if id == uuid.Nil {
		return nil, errors.New("user id is required")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *UserService) ListUsers() ([]models.User, error) {
	if s == nil {
		return nil, errors.New("user service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("user repository is nil")
	}

	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
	adminID uuid.UUID,
) error {
	if s == nil {
		return errors.New("user service is nil")
	}

	if s.repo == nil {
		return errors.New("user repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("user id is required")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("get user before delete: %w", err)
	}

	if err := s.repo.DeleteUser(id.String()); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	nt := models.UserDeleted
	notification := models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User account deleted",
		TargetRole: "admin",
		UserID:     user.ID,
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"admin_id": adminID.String(),
		}),
	}

	s.attachAdminEmailDelivery(
		&notification,
		"user-deleted",
		"User account deleted",
		fmt.Sprintf("User account %s was deleted.", user.Username),
		map[string]any{
			"Name":      s.systemAdminName(),
			"Platform":  s.platformName(),
			"Username":  user.Username,
			"Email":     user.Email,
			"ActionURL": s.adminUsersURL(),
			"Details": fmt.Sprintf(
				"User ID: %s\nUsername: %s\nEmail: %s\nAdmin ID: %s",
				user.ID,
				user.Username,
				user.Email,
				adminID.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *UserService) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		fmt.Printf("user notification failed type=%s error=%v\n", notification.Type, err)
	}
}

func (s *UserService) attachAdminEmailDelivery(
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
		templateData["ActionURL"] = s.adminUsersURL()
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

func (s *UserService) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *UserService) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *UserService) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *UserService) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *UserService) adminUsersURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/admin/users"
	}

	return base + "/users"
}

func normalizeRoleNames(roles []string) []string {
	if len(roles) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	out := make([]string, 0, len(roles))

	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}

		if seen[role] {
			continue
		}

		seen[role] = true
		out = append(out, role)
	}

	return out
}
