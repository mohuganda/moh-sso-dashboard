package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
)

type QueueService struct {
	repo repository.EmailRepository
}

func NewQueueService(repo repository.EmailRepository) (*QueueService, error) {
	if repo == nil {
		return nil, errors.New("email repository is required")
	}

	return &QueueService{repo: repo}, nil
}

func (q *QueueService) Enqueue(ctx context.Context, msg model.Message) error {
	if q == nil {
		return errors.New("queue service is nil")
	}

	if err := validateMessage(msg); err != nil {
		return fmt.Errorf("validate email message: %w", err)
	}

	if _, err := q.repo.Enqueue(ctx, msg); err != nil {
		return fmt.Errorf("enqueue email message: %w", err)
	}

	return nil
}

func validateMessage(msg model.Message) error {
	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	for _, to := range msg.To {
		if strings.TrimSpace(to.Email) == "" {
			return errors.New("recipient email cannot be empty")
		}
	}

	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("subject is required")
	}

	if strings.TrimSpace(msg.TextBody) == "" {
		return errors.New("body is required")
	}

	return nil
}
