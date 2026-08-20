package email

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/model"
	mail "github.com/wneessen/go-mail"
)

// SendEmailViaGM sends an email using dedicated SMTP_HOST_GM configuration (e.g. Gmail SMTP),
// falling back to standard SMTP configuration if SMTP_HOST_GM is not configured.
func SendEmailViaGM(ctx context.Context, msg model.Message, appCfg *config.Config) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST_GM"))
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT_GM"))
	user := strings.TrimSpace(os.Getenv("SMTP_USERNAME_GM"))
	pass := strings.TrimSpace(os.Getenv("SMTP_PASSWORD_GM"))
	fromEmail := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL_GM"))
	fromName := strings.TrimSpace(os.Getenv("SMTP_FROM_NAME_GM"))

	// Fallback to appCfg.SMTPGM if env vars not directly set
	if host == "" && appCfg != nil && appCfg.SMTPGM.Host != "" {
		host = appCfg.SMTPGM.Host
		if appCfg.SMTPGM.Port > 0 {
			portStr = strconv.Itoa(appCfg.SMTPGM.Port)
		}
		if user == "" {
			user = appCfg.SMTPGM.Username
		}
		if pass == "" {
			pass = appCfg.SMTPGM.Password
		}
		if fromEmail == "" {
			fromEmail = appCfg.SMTPGM.FromEmail
		}
		if fromName == "" {
			fromName = appCfg.SMTPGM.FromName
		}
	}

	// Fallback to standard SMTP config if SMTP_HOST_GM is still empty
	if host == "" && appCfg != nil {
		host = appCfg.SMTP.Host
		if appCfg.SMTP.Port > 0 {
			portStr = strconv.Itoa(appCfg.SMTP.Port)
		}
		if user == "" {
			user = appCfg.SMTP.Username
		}
		if pass == "" {
			pass = appCfg.SMTP.Password
		}
		if fromEmail == "" {
			fromEmail = appCfg.SMTP.FromEmail
		}
		if fromName == "" {
			fromName = appCfg.SMTP.FromName
		}
	}

	if host == "" {
		host = "smtp.gmail.com"
	}

	port := 587
	if portStr != "" {
		if parsed, err := strconv.Atoi(portStr); err == nil && parsed > 0 {
			port = parsed
		}
	}

	if fromEmail == "" {
		if user != "" {
			fromEmail = user
		} else {
			fromEmail = "noreply@moh.go.ug"
		}
	}
	if fromName == "" {
		fromName = "MOH Integrated Health Portal"
	}

	m := mail.NewMsg()
	if msg.From != nil && strings.TrimSpace(msg.From.Email) != "" {
		name := msg.From.Name
		if name == "" {
			name = fromName
		}
		if err := m.FromFormat(name, msg.From.Email); err != nil {
			_ = m.FromFormat(fromName, fromEmail)
		}
	} else {
		if err := m.FromFormat(fromName, fromEmail); err != nil {
			return fmt.Errorf("set GM from address: %w", err)
		}
	}

	for _, recipient := range msg.To {
		email := strings.TrimSpace(recipient.Email)
		if email == "" {
			continue
		}
		if recipient.Name != "" {
			if err := m.AddToFormat(recipient.Name, email); err != nil {
				_ = m.AddTo(email)
			}
		} else {
			if err := m.AddTo(email); err != nil {
				return fmt.Errorf("add GM recipient %s: %w", email, err)
			}
		}
	}

	m.Subject(strings.TrimSpace(msg.Subject))

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

	hostLower := strings.ToLower(host)
	isLocal := hostLower == "mailhog" || hostLower == "localhost" || hostLower == "127.0.0.1"

	clientOptions := []mail.Option{
		mail.WithPort(port),
		mail.WithTimeout(15 * time.Second),
	}

	if isLocal {
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.NoTLS))
	} else {
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	if user != "" {
		clientOptions = append(
			clientOptions,
			mail.WithUsername(user),
			mail.WithPassword(pass),
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		)
	}

	client, err := mail.NewClient(host, clientOptions...)
	if err != nil {
		return fmt.Errorf("create GM smtp client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("send email via GM smtp: %w", err)
	}

	return nil
}
