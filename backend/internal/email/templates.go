package email

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
)

type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
}

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*template.Template),
	}
}

func (tm *TemplateManager) Register(name, content string) error {
	tpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", name, err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.templates[name] = tpl
	return nil
}

func (tm *TemplateManager) Render(name string, data any) (string, error) {
	tm.mu.RLock()
	tpl, ok := tm.templates[name]
	tm.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.String(), nil
}
