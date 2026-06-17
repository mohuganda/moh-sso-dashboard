package email

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/sqlc-dev/pqtype"
)

type OutboxMessage struct {
	ID          string
	TenantID    *string
	MessageID   *string
	Message     model.Message
	Status      string
	Attempts    int32
	MaxAttempts int32
	LastError   *string
	ScheduledAt *time.Time
	LockedAt    *time.Time
	SentAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func toCreateEmailOutboxParams(msg model.Message) (db.CreateEmailOutboxParams, error) {
	var (
		fromJSON         []byte
		toJSON           []byte
		ccJSON           []byte
		bccJSON          []byte
		replyToJSON      []byte
		templateDataJSON []byte
		attachmentsJSON  []byte
		headersJSON      []byte
		metadataJSON     []byte
		err              error
	)

	if msg.From != nil {
		fromJSON, err = json.Marshal(msg.From)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal from_address: %w", err)
		}
	}

	toJSON, err = json.Marshal(msg.To)
	if err != nil {
		return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal to_addresses: %w", err)
	}

	if len(msg.Cc) > 0 {
		ccJSON, err = json.Marshal(msg.Cc)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal cc_addresses: %w", err)
		}
	}

	if len(msg.Bcc) > 0 {
		bccJSON, err = json.Marshal(msg.Bcc)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal bcc_addresses: %w", err)
		}
	}

	if len(msg.ReplyTo) > 0 {
		replyToJSON, err = json.Marshal(msg.ReplyTo)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal reply_to_addresses: %w", err)
		}
	}

	if msg.TemplateData != nil {
		templateDataJSON, err = json.Marshal(msg.TemplateData)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal template_data: %w", err)
		}
	}

	if len(msg.Attachments) > 0 {
		attachmentsJSON, err = json.Marshal(stripAttachmentBytes(msg.Attachments))
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal attachments: %w", err)
		}
	}

	if len(msg.Headers) > 0 {
		headersJSON, err = json.Marshal(msg.Headers)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal headers: %w", err)
		}
	}

	if msg.Metadata != nil {
		metadataJSON, err = json.Marshal(msg.Metadata)
		if err != nil {
			return db.CreateEmailOutboxParams{}, fmt.Errorf("marshal metadata: %w", err)
		}
	}

	var messageID sql.NullString
	if msg.ID != "" {
		messageID = sql.NullString{
			String: msg.ID,
			Valid:  true,
		}
	}

	var scheduledAt sql.NullTime
	if msg.ScheduledAt != nil {
		scheduledAt = sql.NullTime{
			Time:  *msg.ScheduledAt,
			Valid: true,
		}
	}

	return db.CreateEmailOutboxParams{
		MessageID:        messageID,
		FromAddress:      toNullRawMessage(fromJSON),
		ToAddresses:      toJSON,
		CcAddresses:      toNullRawMessage(ccJSON),
		BccAddresses:     toNullRawMessage(bccJSON),
		ReplyToAddresses: toNullRawMessage(replyToJSON),
		Subject:          msg.Subject,
		TextBody:         sql.NullString{String: msg.TextBody, Valid: msg.TextBody != ""},
		HtmlBody:         sql.NullString{String: msg.HTMLBody, Valid: msg.HTMLBody != ""},
		TemplateName:     sql.NullString{String: msg.TemplateName, Valid: msg.TemplateName != ""},
		TemplateData:     toNullRawMessage(templateDataJSON),
		Attachments:      toNullRawMessage(attachmentsJSON),
		Headers:          toNullRawMessage(headersJSON),
		Metadata:         toNullRawMessage(metadataJSON),
		Status:           "PENDING",
		Attempts:         int32(0),
		MaxAttempts:      int32(5),
		ScheduledAt:      scheduledAt,
	}, nil
}

func toNullRawMessage(b []byte) pqtype.NullRawMessage {
	if len(b) == 0 {
		return pqtype.NullRawMessage{}
	}
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}
}

func mapEmailOutboxRows(rows []db.EmailOutbox) ([]OutboxMessage, error) {
	out := make([]OutboxMessage, 0, len(rows))
	for _, row := range rows {
		item, err := mapEmailOutboxRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func mapEmailOutboxRow(row db.EmailOutbox) (OutboxMessage, error) {
	var msg model.Message

	if row.FromAddress.Valid {
		var from model.Address
		if err := json.Unmarshal(row.FromAddress.RawMessage, &from); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal from_address: %w", err)
		}
		msg.From = &from
	}

	if len(row.ToAddresses) > 0 {
		if err := json.Unmarshal(row.ToAddresses, &msg.To); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal to_addresses: %w", err)
		}
	}

	if row.CcAddresses.Valid {
		if err := json.Unmarshal(row.CcAddresses.RawMessage, &msg.Cc); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal cc_addresses: %w", err)
		}
	}

	if row.BccAddresses.Valid {
		if err := json.Unmarshal(row.BccAddresses.RawMessage, &msg.Bcc); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal bcc_addresses: %w", err)
		}
	}

	if row.ReplyToAddresses.Valid {
		if err := json.Unmarshal(row.ReplyToAddresses.RawMessage, &msg.ReplyTo); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal reply_to_addresses: %w", err)
		}
	}

	if row.TemplateData.Valid {
		if err := json.Unmarshal(row.TemplateData.RawMessage, &msg.TemplateData); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal template_data: %w", err)
		}
	}

	if row.Attachments.Valid {
		if err := json.Unmarshal(row.Attachments.RawMessage, &msg.Attachments); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal attachments: %w", err)
		}
	}

	if row.Headers.Valid {
		if err := json.Unmarshal(row.Headers.RawMessage, &msg.Headers); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal headers: %w", err)
		}
	}

	if row.Metadata.Valid {
		if err := json.Unmarshal(row.Metadata.RawMessage, &msg.Metadata); err != nil {
			return OutboxMessage{}, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	msg.Subject = row.Subject
	msg.TextBody = nullStringValue(row.TextBody)
	msg.HTMLBody = nullStringValue(row.HtmlBody)
	msg.TemplateName = nullStringValue(row.TemplateName)
	msg.ID = nullStringValue(row.MessageID)

	out := OutboxMessage{
		ID:          row.ID.String(),
		Message:     msg,
		Status:      row.Status,
		Attempts:    row.Attempts,
		MaxAttempts: row.MaxAttempts,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if row.MessageID.Valid {
		v := row.MessageID.String
		out.MessageID = &v
	}
	if row.LastError.Valid {
		v := row.LastError.String
		out.LastError = &v
	}
	if row.ScheduledAt.Valid {
		v := row.ScheduledAt.Time
		out.ScheduledAt = &v
	}
	if row.LockedAt.Valid {
		v := row.LockedAt.Time
		out.LockedAt = &v
	}
	if row.SentAt.Valid {
		v := row.SentAt.Time
		out.SentAt = &v
	}

	return out, nil
}

func stripAttachmentBytes(in []model.Attachment) []model.Attachment {
	out := make([]model.Attachment, 0, len(in))
	for _, a := range in {
		out = append(out, model.Attachment{
			FileName:    a.FileName,
			ContentType: a.ContentType,
			Path:        a.Path,
			DataBase64:  a.DataBase64,
			Inline:      a.Inline,
			ContentID:   a.ContentID,
		})
	}
	return out
}

func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
