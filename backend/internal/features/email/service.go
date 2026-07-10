package email

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	log.Printf(
		"[GROUP EMAIL RECIPIENTS] started: group_ids_count=%d group_paths_count=%d group_ids=%v group_paths=%v",
		len(groupIDs),
		len(groupPaths),
		groupIDs,
		groupPaths,
	)

	if len(groupIDs) == 0 && len(groupPaths) == 0 {
		log.Printf(
			"[GROUP EMAIL RECIPIENTS] skipped: no group IDs or group paths provided",
		)

		return nil, nil
	}

	if s == nil {
		log.Printf(
			"[GROUP EMAIL RECIPIENTS] failed: email service is nil",
		)

		return nil, errors.New("email service is nil")
	}

	if s.rbacRepo == nil {
		log.Printf(
			"[GROUP EMAIL RECIPIENTS] failed: RBAC repository is nil",
		)

		return nil, errors.New(
			"rbac repository is required for group recipient resolution",
		)
	}

	resolvedGroupIDs, err := s.resolveGroupIDs(
		ctx,
		groupIDs,
		groupPaths,
	)
	if err != nil {
		log.Printf(
			"[GROUP EMAIL RECIPIENTS] failed to resolve groups: group_ids=%v group_paths=%v error=%v",
			groupIDs,
			groupPaths,
			err,
		)

		return nil, fmt.Errorf(
			"resolve recipient group IDs: %w",
			err,
		)
	}

	log.Printf(
		"[GROUP EMAIL RECIPIENTS] groups resolved: requested_group_ids=%d requested_group_paths=%d resolved_group_ids_count=%d resolved_group_ids=%v",
		len(groupIDs),
		len(groupPaths),
		len(resolvedGroupIDs),
		resolvedGroupIDs,
	)

	if len(resolvedGroupIDs) == 0 {
		log.Printf(
			"[GROUP EMAIL RECIPIENTS] no groups resolved: group_ids=%v group_paths=%v",
			groupIDs,
			groupPaths,
		)

		return nil, nil
	}

	seen := make(map[string]model.Address)

	totalMembers := 0
	validAddresses := 0
	skippedMembers := 0
	duplicateAddresses := 0

	for _, groupID := range resolvedGroupIDs {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" {
			log.Printf(
				"[GROUP EMAIL RECIPIENTS] skipped empty resolved group ID",
			)
			continue
		}

		log.Printf(
			"[GROUP EMAIL RECIPIENTS] listing group members: group_id=%s",
			groupID,
		)

		members, err := s.rbacRepo.ListGroupMembers(ctx, groupID)
		if err != nil {
			log.Printf(
				"[GROUP EMAIL RECIPIENTS] failed to list group members: group_id=%s error=%v",
				groupID,
				err,
			)

			return nil, fmt.Errorf(
				"list group members for %s: %w",
				groupID,
				err,
			)
		}

		totalMembers += len(members)

		log.Printf(
			"[GROUP EMAIL RECIPIENTS] group members loaded: group_id=%s members_count=%d",
			groupID,
			len(members),
		)

		if len(members) == 0 {
			log.Printf(
				"[GROUP EMAIL RECIPIENTS] group has no direct members: group_id=%s",
				groupID,
			)
		}

		for memberIndex, member := range members {
			address, ok, err := s.emailAddressFromGroupMember(
				ctx,
				member,
			)
			if err != nil {
				log.Printf(
					"[GROUP EMAIL RECIPIENTS] failed to extract member email: group_id=%s member_index=%d user_id=%q username=%q error=%v",
					groupID,
					memberIndex,
					member.UserID,
					member.Username,
					err,
				)

				return nil, fmt.Errorf(
					"extract email address for member %d in group %s: %w",
					memberIndex,
					groupID,
					err,
				)
			}

			if !ok {
				skippedMembers++

				log.Printf(
					"[GROUP EMAIL RECIPIENTS] member skipped: group_id=%s member_index=%d user_id=%q username=%q reason=no_valid_email",
					groupID,
					memberIndex,
					member.UserID,
					member.Username,
				)

				continue
			}

			email := strings.ToLower(
				strings.TrimSpace(address.Email),
			)
			if email == "" {
				skippedMembers++

				log.Printf(
					"[GROUP EMAIL RECIPIENTS] member skipped: group_id=%s member_index=%d user_id=%q username=%q reason=empty_email",
					groupID,
					memberIndex,
					member.UserID,
					member.Username,
				)

				continue
			}

			address.Email = email

			if _, exists := seen[email]; exists {
				duplicateAddresses++

				log.Printf(
					"[GROUP EMAIL RECIPIENTS] duplicate recipient skipped: group_id=%s member_index=%d user_id=%q username=%q",
					groupID,
					memberIndex,
					member.UserID,
					member.Username,
				)

				continue
			}

			seen[email] = address
			validAddresses++

			log.Printf(
				"[GROUP EMAIL RECIPIENTS] recipient added: group_id=%s member_index=%d user_id=%q username=%q total_unique_recipients=%d",
				groupID,
				memberIndex,
				member.UserID,
				member.Username,
				len(seen),
			)
		}
	}

	out := make([]model.Address, 0, len(seen))
	for _, address := range seen {
		out = append(out, address)
	}

	log.Printf(
		"[GROUP EMAIL RECIPIENTS] completed: resolved_groups=%d total_members=%d valid_addresses=%d skipped_members=%d duplicate_addresses=%d unique_recipients=%d",
		len(resolvedGroupIDs),
		totalMembers,
		validAddresses,
		skippedMembers,
		duplicateAddresses,
		len(out),
	)

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
	ctx context.Context,
	member rbacfeature.GroupMember,
) (model.Address, bool, error) {
	rawUserID := strings.TrimSpace(member.UserID)
	username := strings.TrimSpace(member.Username)
	memberEmail := strings.TrimSpace(member.Email)

	log.Printf(
		"[GROUP MEMBER EMAIL] started: user_id=%q username=%q member_email_present=%t user_repo_configured=%t",
		rawUserID,
		username,
		memberEmail != "",
		s.userRepo != nil,
	)

	if s.userRepo != nil {
		var user *model.User
		var err error

		userID, parseErr := uuid.Parse(rawUserID)
		if parseErr == nil && userID != uuid.Nil {
			log.Printf(
				"[GROUP MEMBER EMAIL] resolving user by UUID: user_id=%s",
				userID,
			)

			user, err = s.userRepo.GetUserByID(userID)
			if err != nil {
				return model.Address{}, false, fmt.Errorf(
					"get group member user by ID %s: %w",
					userID,
					err,
				)
			}
		} else {
			lookupUsername := username
			if lookupUsername == "" {
				lookupUsername = rawUserID
			}

			log.Printf(
				"[GROUP MEMBER EMAIL] resolving user by username: username=%q",
				lookupUsername,
			)

			if lookupUsername != "" {
				user, err = s.userRepo.GetUserByUsername(
					ctx,
					lookupUsername,
				)
				if err != nil {
					return model.Address{}, false, fmt.Errorf(
						"get group member user by username %q: %w",
						lookupUsername,
						err,
					)
				}
			}
		}

		if user != nil {
			userEmail := strings.TrimSpace(user.Email)

			log.Printf(
				"[GROUP MEMBER EMAIL] repository user loaded: user_id=%q username=%q enabled=%t email_present=%t",
				user.ID,
				user.Username,
				user.Enabled,
				userEmail != "",
			)

			if !user.Enabled {
				log.Printf(
					"[GROUP MEMBER EMAIL] member skipped: username=%q reason=user_disabled",
					user.Username,
				)

				return model.Address{}, false, nil
			}

			if userEmail != "" {
				name := strings.TrimSpace(user.FullName)

				if name == "" {
					name = strings.TrimSpace(
						strings.Join(
							[]string{
								strings.TrimSpace(user.FirstName),
								strings.TrimSpace(user.LastName),
							},
							" ",
						),
					)
				}

				if name == "" {
					name = strings.TrimSpace(user.Username)
				}

				return model.Address{
					Name:  name,
					Email: userEmail,
				}, true, nil
			}
		}
	}

	if memberEmail == "" {
		log.Printf(
			"[GROUP MEMBER EMAIL] member skipped: user_id=%q username=%q reason=no_repository_email_and_no_member_email",
			rawUserID,
			username,
		)

		return model.Address{}, false, nil
	}

	name := username
	if name == "" {
		name = memberEmail
	}

	return model.Address{
		Name:  name,
		Email: memberEmail,
	}, true, nil
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
