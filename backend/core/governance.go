package core

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

// Application is the governed registry view. It does not grant caller access,
// provision identity-provider clients, or replace a reviewed manifest.
type Application struct {
	ID               string
	Access           SystemAccess
	Description      string
	OwnerTeam        string
	OwnerName        string
	OwnerEmail       string
	SupportURL       string
	DocumentationURL string
	Environment      string
	Criticality      string
	Enabled          bool
	SortOrder        int32
	LifecycleStatus  string
	IntegrationType  string
	ContractVersion  string
}

// ApplicationRegistry reads host-authoritative registrations, including inactive
// applications. Authentication and caller authorization remain host concerns.
type ApplicationRegistry interface {
	GetApplication(context.Context, string) (Application, error)
}

type LifecycleActor struct {
	Subject      string
	CanConfigure bool
	CanApprove   bool
}

var applicationTransitions = map[string][]string{
	"registered": {"configured", "retired"},
	"configured": {"validated", "retired"},
	"validated":  {"approved", "configured", "retired"},
	"approved":   {"active", "configured", "retired"},
	"active":     {"suspended", "retired"},
	"suspended":  {"configured", "retired"},
	"retired":    {},
}

func ApplicationLifecycleStatuses() []string {
	return []string{"registered", "configured", "validated", "approved", "active", "suspended", "retired"}
}
func IsApplicationLifecycleStatus(status string) bool {
	_, ok := applicationTransitions[status]
	return ok
}
func CanTransitionApplication(from, to string) bool {
	for _, next := range applicationTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// CanManageApplicationLifecycle consumes capabilities computed by the trusted
// host. Approval and activation require both configuration and review authority.
func CanManageApplicationLifecycle(actor LifecycleActor, target string) bool {
	return IsApplicationLifecycleStatus(target) && actor.Subject != "" && actor.CanConfigure &&
		((target != "approved" && target != "active") || actor.CanApprove)
}

// ApplicationTransitionRequiresValidation reports whether a target lifecycle
// status requires current metadata and host-specific launch/access checks.
func ApplicationTransitionRequiresValidation(target string) bool {
	switch target {
	case "validated", "approved", "active":
		return true
	default:
		return false
	}
}

var ErrInvalidApplicationMetadata = errors.New("invalid application metadata")

type ApplicationValidationError struct{ Message string }

func (e *ApplicationValidationError) Error() string { return e.Message }
func (e *ApplicationValidationError) Unwrap() error { return ErrInvalidApplicationMetadata }

// ValidateApplicationMetadata preserves the host's metadata checks. Launch and
// navigation behavior and enabled access-role checks must also pass in the host
// before a reviewed transition commits.
func ValidateApplicationMetadata(a Application) error {
	invalid := func(message string) error { return &ApplicationValidationError{Message: message} }
	required := []struct{ key, value string }{
		{"displayName", a.Access.DisplayName}, {"ownerTeam", a.OwnerTeam},
		{"ownerName", a.OwnerName}, {"ownerEmail", a.OwnerEmail},
		{"supportUrl", a.SupportURL}, {"environment", a.Environment},
		{"launchUrl", a.Access.LaunchURL},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return invalid(fmt.Sprintf("%s is required before validation", field.key))
		}
	}
	address, err := mail.ParseAddress(a.OwnerEmail)
	if err != nil || address.Address != a.OwnerEmail {
		return invalid("ownerEmail must be an email address")
	}
	for _, field := range []struct{ key, value string }{{"supportUrl", a.SupportURL}, {"documentationUrl", a.DocumentationURL}} {
		if field.value == "" {
			continue
		}
		if strings.HasPrefix(field.value, "/") && !strings.HasPrefix(field.value, "//") {
			continue
		}
		parsed, err := url.Parse(field.value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return invalid(fmt.Sprintf("%s must be an absolute URL or root-relative path", field.key))
		}
	}
	if a.Environment != "development" && a.Environment != "test" && a.Environment != "staging" && a.Environment != "production" {
		return invalid("environment must be development, test, staging, or production")
	}
	if a.ContractVersion != "1" {
		return invalid("unsupported contractVersion")
	}
	expected := "external"
	if a.Access.SystemType == "platform" {
		expected = "hosted_module"
	}
	if a.IntegrationType != expected {
		return invalid("integrationType must match systemType")
	}
	return nil
}
