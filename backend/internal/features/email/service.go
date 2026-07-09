package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userRepository "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

type Service struct {
	email    sharedservice.EmailService
	repo     repository.EmailRepository
	rbacRepo rbacfeature.Repository
	userRepo userRepository.UserRepository
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
		return nil, err
	}

	seen := make(map[string]model.Address)
	for _, groupID := range resolvedGroupIDs {
		members, err := s.rbacRepo.ListGroupMembers(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("list group members for %s: %w", groupID, err)
		}

		for _, member := range members {
			address, ok, err := s.emailAddressFromGroupMember(member)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}

			seen[strings.ToLower(address.Email)] = address
		}
	}

	out := make([]model.Address, 0, len(seen))
	for _, address := range seen {
		out = append(out, address)
	}

	return out, nil
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

func (s *Service) emailAddressFromGroupMember(
	member rbacfeature.GroupMember,
) (model.Address, bool, error) {
	if s.userRepo != nil {
		userID, err := uuid.Parse(strings.TrimSpace(member.UserID))
		if err == nil && userID != uuid.Nil {
			user, err := s.userRepo.GetUserByID(userID)
			if err != nil {
				return model.Address{}, false, fmt.Errorf("get group member user %s: %w", userID, err)
			}

			if user != nil {
				if !user.Enabled || strings.TrimSpace(user.Email) == "" {
					return model.Address{}, false, nil
				}

				name := strings.TrimSpace(user.FullName)
				if name == "" {
					name = strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
				}
				if name == "" {
					name = strings.TrimSpace(user.Username)
				}

				return model.Address{
					Name:  name,
					Email: strings.TrimSpace(user.Email),
				}, true, nil
			}
		}
	}

	email := strings.TrimSpace(member.Email)
	if email == "" {
		return model.Address{}, false, nil
	}

	name := strings.TrimSpace(member.Username)
	if name == "" {
		name = email
	}

	return model.Address{Name: name, Email: email}, true, nil
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
