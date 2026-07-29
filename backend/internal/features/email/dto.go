package email

import "time"

type EmailAddressRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"required,email"`
}

type EmailAttachmentRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	Path        string `json:"path,omitempty"`
	DataBase64  string `json:"data_base64,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
	Inline      bool   `json:"inline,omitempty"`
}

type SendEmailRequest struct {
	ID                              string                   `json:"id,omitempty"`
	From                            *EmailAddressRequest     `json:"from,omitempty"`
	To                              []EmailAddressRequest    `json:"to,omitempty"`
	ToGroups                        []string                 `json:"to_groups,omitempty"`
	ToGroupPaths                    []string                 `json:"to_group_paths,omitempty"`
	ToHealthContexts                []string                 `json:"to_health_contexts,omitempty"`
	IncludeHealthContextDescendants bool                     `json:"include_health_context_descendants,omitempty"`
	Cc                              []EmailAddressRequest    `json:"cc,omitempty"`
	Bcc                             []EmailAddressRequest    `json:"bcc,omitempty"`
	ReplyTo                         []EmailAddressRequest    `json:"reply_to,omitempty"`
	Subject                         string                   `json:"subject" binding:"required"`
	TextBody                        string                   `json:"text_body,omitempty"`
	HTMLBody                        string                   `json:"html_body,omitempty"`
	TemplateName                    string                   `json:"template_name,omitempty"`
	TemplateData                    map[string]any           `json:"template_data,omitempty"`
	Attachments                     []EmailAttachmentRequest `json:"attachments,omitempty"`
	Headers                         map[string]string        `json:"headers,omitempty"`
	Metadata                        map[string]string        `json:"metadata,omitempty"`
	ScheduledAt                     *string                  `json:"scheduled_at,omitempty"`
}

type EmailRecipientPreviewRequest struct {
	ToGroups                        []string `json:"to_groups,omitempty"`
	ToGroupPaths                    []string `json:"to_group_paths,omitempty"`
	ToHealthContexts                []string `json:"to_health_contexts,omitempty"`
	IncludeHealthContextDescendants bool     `json:"include_health_context_descendants,omitempty"`
}

type EmailRecipientPreviewResponse struct {
	RecipientCount int                    `json:"recipient_count"`
	Recipients     []EmailAddressResponse `json:"recipients,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type EmailAddressResponse struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type EmailAttachmentResponse struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	Path        string `json:"path,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
	Inline      bool   `json:"inline,omitempty"`
}

type EmailMessageResponse struct {
	ID              string                    `json:"id,omitempty"`
	From            *EmailAddressResponse     `json:"from,omitempty"`
	To              []EmailAddressResponse    `json:"to"`
	Cc              []EmailAddressResponse    `json:"cc,omitempty"`
	Bcc             []EmailAddressResponse    `json:"bcc,omitempty"`
	ReplyTo         []EmailAddressResponse    `json:"reply_to,omitempty"`
	Subject         string                    `json:"subject"`
	TextBody        string                    `json:"text_body,omitempty"`
	HTMLBody        string                    `json:"html_body,omitempty"`
	TemplateName    string                    `json:"template_name,omitempty"`
	TemplateData    map[string]any            `json:"template_data,omitempty"`
	Attachments     []EmailAttachmentResponse `json:"attachments,omitempty"`
	AttachmentCount int                       `json:"attachment_count"`
	Headers         map[string]string         `json:"headers,omitempty"`
	Metadata        map[string]string         `json:"metadata,omitempty"`
	ScheduledAt     *time.Time                `json:"scheduled_at,omitempty"`
}

type OutboxMessageResponse struct {
	ID          string               `json:"id"`
	TenantID    *string              `json:"tenant_id,omitempty"`
	MessageID   *string              `json:"message_id,omitempty"`
	Message     EmailMessageResponse `json:"message"`
	Status      string               `json:"status"`
	Attempts    int32                `json:"attempts"`
	MaxAttempts int32                `json:"max_attempts"`
	LastError   *string              `json:"last_error,omitempty"`
	ScheduledAt *time.Time           `json:"scheduled_at,omitempty"`
	LockedAt    *time.Time           `json:"locked_at,omitempty"`
	SentAt      *time.Time           `json:"sent_at,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
