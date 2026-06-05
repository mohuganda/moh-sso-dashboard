package email

type EmailAddressRequest struct {
	Name  string `json:"name"`
	Email string `json:"email" binding:"required,email"`
}

type EmailAttachmentRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	Path        string `json:"path,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
	Inline      bool   `json:"inline,omitempty"`
}

type SendEmailRequest struct {
	ID           string                   `json:"id,omitempty"`
	From         *EmailAddressRequest     `json:"from,omitempty"`
	To           []EmailAddressRequest    `json:"to" binding:"required,min=1,dive"`
	Cc           []EmailAddressRequest    `json:"cc,omitempty"`
	Bcc          []EmailAddressRequest    `json:"bcc,omitempty"`
	ReplyTo      []EmailAddressRequest    `json:"reply_to,omitempty"`
	Subject      string                   `json:"subject" binding:"required"`
	TextBody     string                   `json:"text_body,omitempty"`
	HTMLBody     string                   `json:"html_body,omitempty"`
	TemplateName string                   `json:"template_name,omitempty"`
	TemplateData map[string]any           `json:"template_data,omitempty"`
	Attachments  []EmailAttachmentRequest `json:"attachments,omitempty"`
	Headers      map[string]string        `json:"headers,omitempty"`
	Metadata     map[string]string        `json:"metadata,omitempty"`
	ScheduledAt  *string                  `json:"scheduled_at,omitempty"`
}
