package email

import (
	"context"
	"errors"
	"fmt"

	"github.com/moh-sso-dashboard/internal/model"
)

type smtpSender interface {
	Send(ctx context.Context, msg model.Message) error
}

type queueWriter interface {
	Enqueue(ctx context.Context, msg model.Message) error
}

type service struct {
	sender smtpSender
	queue  queueWriter
	tm     *TemplateManager
}

type Service interface {
	Send(ctx context.Context, msg model.Message) error
	SendTemplate(ctx context.Context, msg model.Message) error
	Queue(ctx context.Context, msg model.Message) error
	SendBulk(ctx context.Context, messages []model.Message) error
}

func NewService(sender smtpSender, queue queueWriter, tm *TemplateManager) Service {
	return &service{
		sender: sender,
		queue:  queue,
		tm:     tm,
	}
}

func (s *service) Send(ctx context.Context, msg model.Message) error {
	if s == nil {
		return errors.New("email service is nil")
	}
	if s.sender == nil {
		return errors.New("smtp sender is nil")
	}

	if err := validateMessage(msg); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}

	if err := s.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *service) SendTemplate(ctx context.Context, msg model.Message) error {
	if s == nil {
		return errors.New("email service is nil")
	}
	if s.sender == nil {
		return errors.New("smtp sender is nil")
	}
	if s.tm == nil {
		return errors.New("template manager is nil")
	}

	renderedMsg, err := s.applyTemplate(msg)
	if err != nil {
		return fmt.Errorf("apply template: %w", err)
	}

	if err := validateMessage(renderedMsg); err != nil {
		return fmt.Errorf("validate templated message: %w", err)
	}

	if err := s.sender.Send(ctx, renderedMsg); err != nil {
		return fmt.Errorf("send templated email: %w", err)
	}

	return nil
}

func (s *service) Queue(ctx context.Context, msg model.Message) error {
	if s == nil {
		return errors.New("email service is nil")
	}
	if s.queue == nil {
		return errors.New("queue writer is nil")
	}

	if err := validateMessage(msg); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}

	if err := s.queue.Enqueue(ctx, msg); err != nil {
		return fmt.Errorf("queue email: %w", err)
	}

	return nil
}

func (s *service) SendBulk(ctx context.Context, messages []model.Message) error {
	if s == nil {
		return errors.New("email service is nil")
	}
	if s.sender == nil {
		return errors.New("smtp sender is nil")
	}
	if len(messages) == 0 {
		return nil
	}

	for i, msg := range messages {
		if err := validateMessage(msg); err != nil {
			return fmt.Errorf("validate bulk message at index %d: %w", i, err)
		}

		if err := s.sender.Send(ctx, msg); err != nil {
			return fmt.Errorf("send bulk message at index %d: %w", i, err)
		}
	}

	return nil
}

func (s *service) applyTemplate(msg model.Message) (model.Message, error) {
	// Adjust this to match your actual model.Message fields.
	// Example assumptions:
	// - msg.TemplateName string
	// - msg.TemplateData map[string]any
	// - msg.Subject string
	// - msg.Body string

	if s.tm == nil {
		return msg, errors.New("template manager is nil")
	}

	if msg.TemplateName == "" {
		return msg, errors.New("template name is required")
	}

	rendered, err := s.tm.Render(msg.TemplateName, msg.TemplateData)
	if err != nil {
		return msg, err
	}

	msg.TextBody = rendered
	return msg, nil
}
