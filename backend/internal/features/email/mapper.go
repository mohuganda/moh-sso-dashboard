package email

import (
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/model"
)

func mapSendEmailRequest(req SendEmailRequest) (model.Message, error) {
	var scheduledAt *time.Time

	if req.ScheduledAt != nil && strings.TrimSpace(*req.ScheduledAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ScheduledAt))
		if err != nil {
			return model.Message{}, err
		}

		scheduledAt = &t
	}

	msg := model.Message{
		ID:           strings.TrimSpace(req.ID),
		To:           mapAddresses(req.To),
		Cc:           mapAddresses(req.Cc),
		Bcc:          mapAddresses(req.Bcc),
		ReplyTo:      mapAddresses(req.ReplyTo),
		Subject:      strings.TrimSpace(req.Subject),
		TextBody:     req.TextBody,
		HTMLBody:     req.HTMLBody,
		TemplateName: strings.TrimSpace(req.TemplateName),
		TemplateData: req.TemplateData,
		Attachments:  mapAttachments(req.Attachments),
		Headers:      req.Headers,
		Metadata:     req.Metadata,
		ScheduledAt:  scheduledAt,
	}

	if req.From != nil {
		msg.From = &model.Address{
			Name:  strings.TrimSpace(req.From.Name),
			Email: strings.TrimSpace(req.From.Email),
		}
	}

	return msg, nil
}

func mapAddresses(in []EmailAddressRequest) []model.Address {
	out := make([]model.Address, 0, len(in))

	for _, a := range in {
		out = append(out, model.Address{
			Name:  strings.TrimSpace(a.Name),
			Email: strings.TrimSpace(a.Email),
		})
	}

	return out
}

func mapAttachments(in []EmailAttachmentRequest) []model.Attachment {
	out := make([]model.Attachment, 0, len(in))

	for _, a := range in {
		out = append(out, model.Attachment{
			FileName:    strings.TrimSpace(a.FileName),
			ContentType: strings.TrimSpace(a.ContentType),
			Path:        strings.TrimSpace(a.Path),
			ContentID:   strings.TrimSpace(a.ContentID),
			Inline:      a.Inline,
		})
	}

	return out
}
