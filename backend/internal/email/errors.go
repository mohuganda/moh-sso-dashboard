package email

import "errors"

var (
	ErrNoRecipients     = errors.New("email: no recipients")
	ErrNoSubject        = errors.New("email: missing subject")
	ErrNoBody           = errors.New("email: missing body")
	ErrTemplateNotFound = errors.New("email: template not found")
)
