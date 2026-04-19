package email

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/model"
	mail "github.com/wneessen/go-mail"
)

type SMTPService struct {
	cfg       config.Config
	templates *TemplateManager
}

func NewSMTPService(cfg config.Config, tm *TemplateManager) *SMTPService {
	return &SMTPService{
		cfg:       cfg,
		templates: tm,
	}
}

func (s *SMTPService) Send(ctx context.Context, msg model.Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}

	if strings.TrimSpace(msg.TemplateName) != "" {
		if s.templates == nil {
			return fmt.Errorf("template manager not configured")
		}
		rendered, err := s.templates.Render(msg.TemplateName, msg.TemplateData)
		if err != nil {
			return err
		}
		msg.HTMLBody = rendered
		if strings.TrimSpace(msg.TextBody) == "" {
			msg.TextBody = "Please use an email client that supports HTML."
		}
	}

	m := mail.NewMsg()

	from := msg.From
	if from == nil {
		from = &model.Address{
			Name:  s.cfg.SMTP.FromName,
			Email: s.cfg.SMTP.FromEmail,
		}
	}

	if from.Name != "" {
		if err := m.FromFormat(from.Name, from.Email); err != nil {
			return fmt.Errorf("set from: %w", err)
		}
	} else {
		if err := m.From(from.Email); err != nil {
			return fmt.Errorf("set from: %w", err)
		}
	}

	for _, to := range msg.To {
		if to.Name != "" {
			if err := m.AddToFormat(to.Name, to.Email); err != nil {
				return fmt.Errorf("add to: %w", err)
			}
		} else {
			if err := m.AddTo(to.Email); err != nil {
				return fmt.Errorf("add to: %w", err)
			}
		}
	}

	for _, cc := range msg.Cc {
		if cc.Name != "" {
			if err := m.AddCcFormat(cc.Name, cc.Email); err != nil {
				return fmt.Errorf("add cc: %w", err)
			}
		} else {
			if err := m.AddCc(cc.Email); err != nil {
				return fmt.Errorf("add cc: %w", err)
			}
		}
	}

	for _, bcc := range msg.Bcc {
		if bcc.Name != "" {
			if err := m.AddBccFormat(bcc.Name, bcc.Email); err != nil {
				return fmt.Errorf("add bcc: %w", err)
			}
		} else {
			if err := m.AddBcc(bcc.Email); err != nil {
				return fmt.Errorf("add bcc: %w", err)
			}
		}
	}

	for _, r := range msg.ReplyTo {
		if r.Name != "" {
			if err := m.ReplyToFormat(r.Name, r.Email); err != nil {
				return fmt.Errorf("set reply-to: %w", err)
			}
		} else {
			if err := m.ReplyTo(r.Email); err != nil {
				return fmt.Errorf("set reply-to: %w", err)
			}
		}
	}

	m.Subject(msg.Subject)

	if strings.TrimSpace(msg.TextBody) != "" {
		m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
	}
	if strings.TrimSpace(msg.HTMLBody) != "" {
		if strings.TrimSpace(msg.TextBody) == "" {
			m.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
		} else {
			m.AddAlternativeString(mail.TypeTextHTML, msg.HTMLBody)
		}
	}

	for k, v := range msg.Headers {
		m.SetGenHeader(mail.Header(k), v)
	}

	for _, att := range msg.Attachments {
		switch {
		case att.Path != "":
			if _, err := os.Stat(att.Path); err != nil {
				return fmt.Errorf("attachment path %q: %w", att.Path, err)
			}
			if att.FileName != "" {
				m.AttachFile(att.Path, mail.WithFileName(att.FileName))
			} else {
				m.AttachFile(att.Path)
			}

		case len(att.Data) > 0:
			name := att.FileName
			if name == "" {
				name = "attachment"
			}
			m.AttachReader(name, newReadSeeker(att.Data))

		default:
			return fmt.Errorf("attachment %q missing path or data", att.FileName)
		}
	}

	client, err := mail.NewClient(
		s.cfg.SMTP.Host,
		mail.WithPort(s.cfg.SMTP.Port),
		mail.WithUsername(s.cfg.SMTP.Username),
		mail.WithPassword(s.cfg.SMTP.Password),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
