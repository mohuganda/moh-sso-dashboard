package email

import (
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/model"
	emailrepo "github.com/moh-sso-dashboard/internal/repository/email"
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
			DataBase64:  strings.TrimSpace(a.DataBase64),
			ContentID:   strings.TrimSpace(a.ContentID),
			Inline:      a.Inline,
		})
	}

	return out
}

func toOutboxMessageResponses(in []emailrepo.OutboxMessage) []OutboxMessageResponse {
	out := make([]OutboxMessageResponse, 0, len(in))
	for _, item := range in {
		out = append(out, toOutboxMessageResponse(item))
	}
	return out
}

func toOutboxMessageResponse(item emailrepo.OutboxMessage) OutboxMessageResponse {
	return OutboxMessageResponse{
		ID:          item.ID,
		TenantID:    item.TenantID,
		MessageID:   item.MessageID,
		Message:     toEmailMessageResponse(item.Message),
		Status:      item.Status,
		Attempts:    item.Attempts,
		MaxAttempts: item.MaxAttempts,
		LastError:   item.LastError,
		ScheduledAt: item.ScheduledAt,
		LockedAt:    item.LockedAt,
		SentAt:      item.SentAt,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toEmailMessageResponse(msg model.Message) EmailMessageResponse {
	out := EmailMessageResponse{
		ID:              msg.ID,
		To:              toAddressResponses(msg.To),
		Cc:              toAddressResponses(msg.Cc),
		Bcc:             toAddressResponses(msg.Bcc),
		ReplyTo:         toAddressResponses(msg.ReplyTo),
		Subject:         msg.Subject,
		TextBody:        msg.TextBody,
		HTMLBody:        msg.HTMLBody,
		TemplateName:    msg.TemplateName,
		TemplateData:    msg.TemplateData,
		Attachments:     toAttachmentResponses(msg.Attachments),
		AttachmentCount: len(msg.Attachments),
		Headers:         msg.Headers,
		Metadata:        msg.Metadata,
		ScheduledAt:     msg.ScheduledAt,
	}
	if msg.From != nil {
		out.From = &EmailAddressResponse{
			Name:  msg.From.Name,
			Email: msg.From.Email,
		}
	}
	return out
}

func toAddressResponses(in []model.Address) []EmailAddressResponse {
	out := make([]EmailAddressResponse, 0, len(in))
	for _, address := range in {
		out = append(out, EmailAddressResponse{
			Name:  address.Name,
			Email: address.Email,
		})
	}
	return out
}

func toAttachmentResponses(in []model.Attachment) []EmailAttachmentResponse {
	out := make([]EmailAttachmentResponse, 0, len(in))
	for _, attachment := range in {
		out = append(out, EmailAttachmentResponse{
			FileName:    attachment.FileName,
			ContentType: attachment.ContentType,
			Path:        attachment.Path,
			ContentID:   attachment.ContentID,
			Inline:      attachment.Inline,
		})
	}
	return out
}
