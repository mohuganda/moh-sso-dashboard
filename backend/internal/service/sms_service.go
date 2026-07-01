package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	logger "github.com/moh-sso-dashboard/internal/log"
)

var nonPhoneCharacters = regexp.MustCompile(`[^\d+]`)

type SMSMessage struct {
	To       string            `json:"to"`
	Body     string            `json:"body"`
	From     string            `json:"from,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type SMSResult struct {
	Provider  string `json:"provider"`
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status"`
}

type SMSService interface {
	Send(ctx context.Context, message SMSMessage) (SMSResult, error)
	NormalizePhoneNumber(phone string) (string, error)
	Enabled() bool
}

type smsService struct {
	cfg      config.SMSConfig
	provider SMSProvider
	logger   *logger.Logger
}

type SMSProvider interface {
	Send(ctx context.Context, message SMSMessage) (SMSResult, error)
}

func NewSMSService(cfg *config.Config, log *logger.Logger) (SMSService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if log == nil {
		return nil, errors.New("logger is required")
	}

	var provider SMSProvider
	switch cfg.SMS.Provider {
	case "mock":
		provider = &mockSMSProvider{logger: log}
	case "noop":
		provider = noopSMSProvider{}
	case "africastalking", "twilio":
		// Real provider adapters can be added without changing callers or the worker.
		provider = &mockSMSProvider{logger: log, provider: cfg.SMS.Provider}
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", cfg.SMS.Provider)
	}

	return &smsService{
		cfg:      cfg.SMS,
		provider: provider,
		logger:   log,
	}, nil
}

func (s *smsService) Enabled() bool {
	return s != nil && s.cfg.Enabled
}

func (s *smsService) NormalizePhoneNumber(phone string) (string, error) {
	return NormalizePhoneNumber(phone, s.cfg.DefaultCountryCode)
}

func (s *smsService) Send(ctx context.Context, message SMSMessage) (SMSResult, error) {
	if s == nil {
		return SMSResult{}, errors.New("sms service is nil")
	}
	if !s.cfg.Enabled {
		return SMSResult{}, errors.New("sms delivery is disabled")
	}

	to, err := s.NormalizePhoneNumber(message.To)
	if err != nil {
		return SMSResult{}, err
	}

	body := strings.TrimSpace(message.Body)
	if body == "" {
		return SMSResult{}, errors.New("sms body is required")
	}
	if len([]rune(body)) > s.cfg.MaxLength {
		return SMSResult{}, fmt.Errorf("sms body exceeds max length of %d characters", s.cfg.MaxLength)
	}

	timeout := s.cfg.SendTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	sendCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	message.To = to
	if strings.TrimSpace(message.From) == "" {
		message.From = s.cfg.FromName
	}
	message.Body = body

	result, err := s.provider.Send(sendCtx, message)
	if err != nil {
		return SMSResult{}, err
	}

	return result, nil
}

func NormalizePhoneNumber(phone string, defaultCountryCode string) (string, error) {
	value := strings.TrimSpace(phone)
	if value == "" {
		return "", errors.New("phone number is required")
	}

	value = nonPhoneCharacters.ReplaceAllString(value, "")
	defaultCountryCode = strings.TrimSpace(defaultCountryCode)
	if defaultCountryCode != "" && !strings.HasPrefix(defaultCountryCode, "+") {
		defaultCountryCode = "+" + defaultCountryCode
	}

	switch {
	case strings.HasPrefix(value, "+"):
	case strings.HasPrefix(value, "00"):
		value = "+" + strings.TrimPrefix(value, "00")
	case strings.HasPrefix(value, "0") && defaultCountryCode != "":
		value = defaultCountryCode + strings.TrimPrefix(value, "0")
	case defaultCountryCode != "":
		value = defaultCountryCode + value
	default:
		return "", errors.New("phone number must include country code")
	}

	digits := strings.TrimPrefix(value, "+")
	if len(digits) < 8 || len(digits) > 15 {
		return "", errors.New("phone number must be 8 to 15 digits")
	}

	return value, nil
}

type mockSMSProvider struct {
	logger   *logger.Logger
	provider string
}

func (p *mockSMSProvider) Send(_ context.Context, message SMSMessage) (SMSResult, error) {
	provider := p.provider
	if provider == "" {
		provider = "mock"
	}
	if p.logger != nil {
		p.logger.Info(
			"mock sms sent",
			"provider", provider,
			"to", message.To,
			"body_length", len([]rune(message.Body)),
		)
	}

	return SMSResult{
		Provider:  provider,
		MessageID: fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		Status:    "SENT",
	}, nil
}

type noopSMSProvider struct{}

func (noopSMSProvider) Send(context.Context, SMSMessage) (SMSResult, error) {
	return SMSResult{}, errors.New("sms provider is disabled")
}
