package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
	emailutil "github.com/moh-sso-dashboard/internal/email"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
)

var (
	ErrServiceNil         = errors.New("email service is nil")
	ErrSMTPSenderNil      = errors.New("smtp sender is nil")
	ErrQueueWriterNil     = errors.New("queue writer is nil")
	ErrTemplateManagerNil = errors.New("template manager is nil")
	ErrTemplateNameEmpty  = errors.New("template name is required")
)

type smtpSender interface {
	Send(ctx context.Context, msg model.Message) error
}

type queueWriter interface {
	Enqueue(ctx context.Context, msg model.Message) error
}

type EmailService interface {
	Send(ctx context.Context, msg model.Message) error
	SendTemplate(ctx context.Context, msg model.Message) error
	Queue(ctx context.Context, msg model.Message) error
	SendBulk(ctx context.Context, messages []model.Message) error
}

type service struct {
	sender smtpSender
	queue  queueWriter
	tm     *TemplateManager
	cfg    *config.Config
}

func NewEmailService(
	sender smtpSender,
	queue queueWriter,
	tm *TemplateManager,
	cfg ...*config.Config,
) (EmailService, error) {
	if sender == nil {
		return nil, ErrSMTPSenderNil
	}

	if queue == nil {
		return nil, ErrQueueWriterNil
	}

	if tm == nil {
		return nil, ErrTemplateManagerNil
	}

	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &service{
		sender: sender,
		queue:  queue,
		tm:     tm,
		cfg:    appConfig,
	}, nil
}

func (s *service) Send(ctx context.Context, msg model.Message) error {
	if err := s.requireSender(); err != nil {
		return err
	}

	if strings.TrimSpace(msg.TemplateName) != "" {
		return s.SendTemplate(ctx, msg)
	}

	var err error
	msg.Attachments, err = emailutil.NormalizeAttachments(s.cfg, msg.Attachments)
	if err != nil {
		return fmt.Errorf("validate attachments: %w", err)
	}

	if err := utils.ValidateMessage(msg); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}

	if err := s.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *service) SendTemplate(ctx context.Context, msg model.Message) error {
	if err := s.requireSender(); err != nil {
		return err
	}

	if err := s.requireTemplateManager(); err != nil {
		return err
	}

	renderedMsg, err := s.applyTemplate(msg)
	if err != nil {
		return fmt.Errorf("apply template: %w", err)
	}

	if err := utils.ValidateMessage(renderedMsg); err != nil {
		return fmt.Errorf("validate templated message: %w", err)
	}

	renderedMsg.Attachments, err = emailutil.NormalizeAttachments(s.cfg, renderedMsg.Attachments)
	if err != nil {
		return fmt.Errorf("validate attachments: %w", err)
	}

	if err := s.sender.Send(ctx, renderedMsg); err != nil {
		return fmt.Errorf("send templated email: %w", err)
	}

	return nil
}

func (s *service) Queue(ctx context.Context, msg model.Message) error {
	if err := s.requireQueue(); err != nil {
		return err
	}

	if strings.TrimSpace(msg.TemplateName) != "" {
		renderedMsg, err := s.applyTemplate(msg)
		if err != nil {
			return fmt.Errorf("apply template before queue: %w", err)
		}

		msg = renderedMsg
	}

	var err error
	msg.Attachments, err = emailutil.NormalizeAttachments(s.cfg, msg.Attachments)
	if err != nil {
		return fmt.Errorf("validate attachments: %w", err)
	}

	if err := utils.ValidateMessage(msg); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}

	if err := s.queue.Enqueue(ctx, msg); err != nil {
		return fmt.Errorf("queue email: %w", err)
	}

	return nil
}

func (s *service) SendBulk(ctx context.Context, messages []model.Message) error {
	if err := s.requireSender(); err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	for i, msg := range messages {
		if strings.TrimSpace(msg.TemplateName) != "" {
			renderedMsg, err := s.applyTemplate(msg)
			if err != nil {
				return fmt.Errorf("apply template for bulk message at index %d: %w", i, err)
			}

			msg = renderedMsg
		}

		if err := utils.ValidateMessage(msg); err != nil {
			return fmt.Errorf("validate bulk message at index %d: %w", i, err)
		}

		var err error
		msg.Attachments, err = emailutil.NormalizeAttachments(s.cfg, msg.Attachments)
		if err != nil {
			return fmt.Errorf("validate attachments for bulk message at index %d: %w", i, err)
		}

		if err := s.sender.Send(ctx, msg); err != nil {
			return fmt.Errorf("send bulk message at index %d: %w", i, err)
		}
	}

	return nil
}

func (s *service) applyTemplate(msg model.Message) (model.Message, error) {
	if err := s.requireTemplateManager(); err != nil {
		return msg, err
	}

	msg.TemplateName = strings.TrimSpace(msg.TemplateName)
	if msg.TemplateName == "" {
		return msg, ErrTemplateNameEmpty
	}

	rendered, err := s.tm.Render(msg.TemplateName, msg.TemplateData)
	if err != nil {
		return msg, fmt.Errorf("render template %q: %w", msg.TemplateName, err)
	}

	msg.HTMLBody = rendered

	if strings.TrimSpace(msg.TextBody) == "" {
		msg.TextBody = "Please use an email client that supports HTML content."
	}

	return msg, nil
}

func (s *service) requireSender() error {
	if s == nil {
		return ErrServiceNil
	}

	if s.sender == nil {
		return ErrSMTPSenderNil
	}

	return nil
}

func (s *service) requireQueue() error {
	if s == nil {
		return ErrServiceNil
	}

	if s.queue == nil {
		return ErrQueueWriterNil
	}

	return nil
}

func (s *service) requireTemplateManager() error {
	if s == nil {
		return ErrServiceNil
	}

	if s.tm == nil {
		return ErrTemplateManagerNil
	}

	return nil
}
