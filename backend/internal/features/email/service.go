package email

import (
	"context"
	"strings"

	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

type Service struct {
	email sharedservice.EmailService
	repo  repository.EmailRepository
}

func NewService(email sharedservice.EmailService, repo repository.EmailRepository) *Service {
	return &Service{
		email: email,
		repo:  repo,
	}
}

func (s *Service) Send(ctx context.Context, msg model.Message) error {
	return s.email.Send(ctx, msg)
}

func (s *Service) Queue(ctx context.Context, msg model.Message) error {
	return s.email.Queue(ctx, msg)
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
