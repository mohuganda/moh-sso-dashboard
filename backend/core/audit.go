package core

import "context"

// AuditEvent records an action with its optional actor and event-specific metadata.
type AuditEvent struct {
	Subject  string
	Action   string
	Metadata map[string]any
}

type AuditRecorder interface {
	RecordAudit(context.Context, AuditEvent) error
}
