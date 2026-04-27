package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/moh-sso-dashboard/internal/email"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	logger    *logger.Logger
}

func NewTemplateManager(logger *logger.Logger) (*TemplateManager, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &TemplateManager{
		templates: make(map[string]*template.Template),
		logger:    logger,
	}, nil
}

func (tm *TemplateManager) Register(name, content string) error {
	return tm.RegisterWithOptions(name, content, false)
}

func (tm *TemplateManager) RegisterWithOptions(name, content string, overwrite bool) error {
	if tm == nil {
		return errors.New("template manager is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("template name is required")
	}

	if strings.TrimSpace(content) == "" {
		return errors.New("template content is empty")
	}

	tpl, err := template.New(name).Parse(content)
	if err != nil {
		tm.logger.Error(
			"failed to parse email template",
			"template_name", name,
			"error", err,
		)

		return fmt.Errorf("parse template %q: %w", name, err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !overwrite {
		if _, exists := tm.templates[name]; exists {
			err := fmt.Errorf("template already exists: %s", name)

			tm.logger.Error(
				"email template already exists",
				"template_name", name,
				"error", err,
			)

			return err
		}
	}

	tm.templates[name] = tpl

	tm.logger.Info(
		"email template registered",
		"template_name", name,
		"overwrite", overwrite,
	)

	return nil
}

func (tm *TemplateManager) MustRegister(name, content string) {
	if err := tm.Register(name, content); err != nil {
		panic(err)
	}
}

func (tm *TemplateManager) RegisterMany(templates map[string]string, overwrite bool) error {
	if tm == nil {
		return errors.New("template manager is nil")
	}

	if len(templates) == 0 {
		tm.logger.Info("no email templates provided for registration")
		return nil
	}

	for name, content := range templates {
		if err := tm.RegisterWithOptions(name, content, overwrite); err != nil {
			tm.logger.Error(
				"failed to register email template",
				"template_name", name,
				"error", err,
			)

			return err
		}
	}

	tm.logger.Info(
		"email templates registered",
		"count", len(templates),
		"overwrite", overwrite,
	)

	return nil
}

func (tm *TemplateManager) Exists(name string) bool {
	if tm == nil {
		return false
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, ok := tm.templates[name]
	return ok
}

func (tm *TemplateManager) Render(name string, data any) (string, error) {
	b, err := tm.RenderToBytes(name, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func (tm *TemplateManager) RenderToBytes(name string, data any) (*bytes.Buffer, error) {
	if tm == nil {
		return nil, errors.New("template manager is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("template name is required")
	}

	tm.mu.RLock()
	tpl, ok := tm.templates[name]
	tm.mu.RUnlock()

	if !ok {
		err := fmt.Errorf("template not found: %s", name)

		tm.logger.Error(
			"email template not found",
			"template_name", name,
			"error", err,
		)

		return nil, err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		tm.logger.Error(
			"failed to execute email template",
			"template_name", name,
			"error", err,
		)

		return nil, fmt.Errorf("execute template %q: %w", name, err)
	}

	tm.logger.Info(
		"email template rendered",
		"template_name", name,
	)

	return &buf, nil
}

func RegisterDefaultTemplates(tm *TemplateManager) error {
	if tm == nil {
		return errors.New("template manager is nil")
	}

	return tm.RegisterMany(email.DefaultTemplates(), false)
}
