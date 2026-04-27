package service

import (
	"context"
	"errors"
	"fmt"

	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/email"
	"github.com/moh-sso-dashboard/internal/utils"
)

type QueueService struct {
	repo   repository.EmailRepository
	logger *logger.Logger
}

func NewQueueService(
	repo repository.EmailRepository,
	logger *logger.Logger,
) (*QueueService, error) {
	if repo == nil {
		return nil, errors.New("email repository is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &QueueService{
		repo:   repo,
		logger: logger,
	}, nil
}

func (q *QueueService) Enqueue(ctx context.Context, msg model.Message) error {
	if q == nil {
		return errors.New("queue service is nil")
	}

	if err := utils.ValidateMessage(msg); err != nil {
		q.logger.Error(
			"email message validation failed",
			"error", err,
			"subject", msg.Subject,
		)

		return fmt.Errorf("validate email message: %w", err)
	}

	queued, err := q.repo.Enqueue(ctx, msg)
	if err != nil {
		q.logger.Error(
			"failed to enqueue email message",
			"error", err,
			"subject", msg.Subject,
		)

		return fmt.Errorf("enqueue email message: %w", err)
	}

	q.logger.Info(
		"email message queued successfully",
		"email_id", queued.ID,
		"subject", msg.Subject,
	)

	return nil
}
