package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	healthcontextfeature "github.com/moh-sso-dashboard/internal/features/health_context"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userRepository "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

type Service struct {
	email          sharedservice.EmailService
	repo           repository.EmailRepository
	rbacRepo       rbacfeature.Repository
	userRepo       userRepository.UserRepository
	healthContexts *healthcontextfeature.Service
}

func NewService(email sharedservice.EmailService, repo repository.EmailRepository) *Service {
	return &Service{
		email: email,
		repo:  repo,
	}
}

func (s *Service) SetRBACRepository(repo rbacfeature.Repository) {
	if s == nil {
		return
	}
	s.rbacRepo = repo
}

func (s *Service) SetUserRepository(repo userRepository.UserRepository) {
	if s == nil {
		return
	}
	s.userRepo = repo
}

func (s *Service) SetHealthContextService(service *healthcontextfeature.Service) {
	if s == nil {
		return
	}
	s.healthContexts = service
}

func (s *Service) Send(ctx context.Context, msg model.Message) error {
	return s.email.Send(ctx, msg)
}

func (s *Service) Queue(ctx context.Context, msg model.Message) error {
	return s.email.Queue(ctx, msg)
}

func (s *Service) ExpandGroupRecipients(
	ctx context.Context,
	msg model.Message,
	groupIDs []string,
	groupPaths []string,
) (model.Message, int, error) {
	recipients, err := s.ResolveGroupEmailRecipients(ctx, groupIDs, groupPaths)
	if err != nil {
		return model.Message{}, 0, err
	}

	seen := make(map[string]struct{}, len(msg.To)+len(recipients))
	for _, recipient := range msg.To {
		email := strings.ToLower(strings.TrimSpace(recipient.Email))
		if email != "" {
			seen[email] = struct{}{}
		}
	}

	added := 0
	for _, recipient := range recipients {
		email := strings.ToLower(strings.TrimSpace(recipient.Email))
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		msg.To = append(msg.To, recipient)
		seen[email] = struct{}{}
		added++
	}

	return msg, added, nil
}

func (s *Service) ResolveHealthContextEmailRecipients(
	ctx context.Context,
	actorID string,
	rawContextIDs []string,
	includeDescendants bool,
) ([]model.Address, error) {
	if len(rawContextIDs) == 0 {
		return nil, nil
	}
	if s == nil || s.healthContexts == nil || s.userRepo == nil {
		return nil, errors.New("health context recipient resolution is unavailable")
	}
	contextIDs := make([]uuid.UUID, 0, len(rawContextIDs))
	for _, rawID := range rawContextIDs {
		contextID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil || contextID == uuid.Nil {
			return nil, fmt.Errorf("invalid health context id %q", rawID)
		}
		contextIDs = append(contextIDs, contextID)
	}
	userIDs, err := s.healthContexts.ResolveAudienceUserIDs(
		ctx,
		actorID,
		contextIDs,
		includeDescendants,
	)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		wanted[strings.TrimSpace(userID)] = struct{}{}
	}
	users, err := s.userRepo.ListUsers()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]model.Address)
	for _, user := range users {
		if !user.Enabled {
			continue
		}
		if _, ok := wanted[strings.TrimSpace(user.ID)]; !ok {
			continue
		}
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if email == "" {
			continue
		}
		name := strings.TrimSpace(user.FullName)
		if name == "" {
			name = strings.TrimSpace(user.Username)
		}
		seen[email] = model.Address{Name: name, Email: email}
	}
	result := make([]model.Address, 0, len(seen))
	for _, address := range seen {
		result = append(result, address)
	}
	return result, nil
}

func addUniqueAddresses(message model.Message, recipients []model.Address) model.Message {
	seen := make(map[string]struct{}, len(message.To)+len(recipients))
	for _, recipient := range message.To {
		seen[strings.ToLower(strings.TrimSpace(recipient.Email))] = struct{}{}
	}
	for _, recipient := range recipients {
		key := strings.ToLower(strings.TrimSpace(recipient.Email))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		message.To = append(message.To, recipient)
		seen[key] = struct{}{}
	}
	return message
}

func (s *Service) ResolveGroupEmailRecipients(
	ctx context.Context,
	groupIDs []string,
	groupPaths []string,
) ([]model.Address, error) {
	if len(groupIDs) == 0 && len(groupPaths) == 0 {
		return nil, nil
	}

	if s == nil {
		return nil, errors.New("email service is nil")
	}

	if s.rbacRepo == nil {
		return nil, errors.New("rbac repository is required for group recipient resolution")
	}

	resolvedGroupIDs, err := s.resolveGroupIDs(ctx, groupIDs, groupPaths)
	if err != nil {
		return nil, fmt.Errorf("resolve recipient group IDs: %w", err)
	}

	seen := make(map[string]model.Address)
	for _, groupID := range resolvedGroupIDs {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" {
			continue
		}

		members, err := s.rbacRepo.ListGroupMembers(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("list group members for %s: %w", groupID, err)
		}

		for _, member := range members {
			address, ok := s.groupMemberEmailAddress(ctx, member)
			if !ok {
				continue
			}

			email := strings.ToLower(strings.TrimSpace(address.Email))
			if email == "" {
				continue
			}

			address.Email = email
			if _, exists := seen[email]; exists {
				continue
			}
			seen[email] = address
		}
	}

	out := make([]model.Address, 0, len(seen))
	for _, address := range seen {
		out = append(out, address)
	}

	return out, nil
}

func (s *Service) groupMemberEmailAddress(
	ctx context.Context,
	member rbacfeature.GroupMember,
) (model.Address, bool) {
	rawUserID := strings.TrimSpace(member.UserID)
	username := strings.TrimSpace(member.Username)
	memberEmail := strings.TrimSpace(member.Email)

	if s.userRepo != nil {
		var user *model.User

		if userID, err := uuid.Parse(rawUserID); err == nil && userID != uuid.Nil {
			user, _ = s.userRepo.GetUserByID(userID)
		} else {
			lookupUsername := username
			if lookupUsername == "" {
				lookupUsername = rawUserID
			}
			if lookupUsername != "" {
				user, _ = s.userRepo.GetUserByUsername(ctx, lookupUsername)
			}
		}

		if user != nil {
			if !user.Enabled {
				return model.Address{}, false
			}

			if email := strings.TrimSpace(user.Email); email != "" {
				name := strings.TrimSpace(user.FullName)
				if name == "" {
					name = strings.TrimSpace(strings.Join([]string{
						strings.TrimSpace(user.FirstName),
						strings.TrimSpace(user.LastName),
					}, " "))
				}
				if name == "" {
					name = strings.TrimSpace(user.Username)
				}
				return model.Address{Name: name, Email: email}, true
			}
		}
	}

	if memberEmail == "" {
		return model.Address{}, false
	}

	name := username
	if name == "" {
		name = memberEmail
	}

	return model.Address{Name: name, Email: memberEmail}, true
}

func (s *Service) resolveGroupIDs(
	ctx context.Context,
	groupIDs []string,
	groupPaths []string,
) ([]string, error) {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(groupIDs)+len(groupPaths))

	for _, groupID := range groupIDs {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" {
			continue
		}
		if _, err := uuid.Parse(groupID); err != nil {
			return nil, fmt.Errorf("invalid group id %q: %w", groupID, err)
		}
		if _, exists := seen[groupID]; exists {
			continue
		}
		seen[groupID] = struct{}{}
		out = append(out, groupID)
	}

	if len(groupPaths) == 0 {
		return out, nil
	}

	groups, err := s.rbacRepo.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}

	byPath := make(map[string]string, len(groups))
	for _, group := range groups {
		byPath[normalizeGroupPath(group.Path)] = strings.TrimSpace(group.ID)
	}

	for _, groupPath := range groupPaths {
		groupPath = normalizeGroupPath(groupPath)
		if groupPath == "" {
			continue
		}

		groupID, ok := byPath[groupPath]
		if !ok || groupID == "" {
			return nil, fmt.Errorf("group path %q was not found", groupPath)
		}

		if _, exists := seen[groupID]; exists {
			continue
		}
		seen[groupID] = struct{}{}
		out = append(out, groupID)
	}

	return out, nil
}

func normalizeGroupPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]repository.OutboxMessage, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, id string) (*repository.OutboxMessage, error) {
	return s.repo.GetByID(ctx, strings.TrimSpace(id))
}

func (s *Service) ListByStatus(ctx context.Context, status string, limit, offset int32) ([]repository.OutboxMessage, error) {
	return s.repo.ListByStatus(ctx, strings.TrimSpace(strings.ToUpper(status)), limit, offset)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteByID(ctx, strings.TrimSpace(id))
}
