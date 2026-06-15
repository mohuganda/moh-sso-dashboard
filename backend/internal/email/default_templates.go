package email

const (
	TemplateWelcome               = "welcome"
	TemplatePasswordReset         = "password-reset"
	TemplateVerifyEmail           = "verify-email"
	TemplateNotification          = "notification"
	TemplateAdminAlert            = "admin-alert"
	TemplateWeeklySummary         = "weekly-summary"
	TemplateAnnouncement          = "announcement"
	TemplateAnnouncementPublished = "announcement-published"
	TemplateAnnouncementArchived  = "announcement-archived"
	TemplateAnnouncementDeleted   = "announcement-deleted"
)

func DefaultTemplates() map[string]string {
	baseLayout := func(title, heading, headingColor, body, buttonLabel, buttonColor string) string {
		if headingColor == "" {
			headingColor = "#145a92"
		}
		if buttonColor == "" {
			buttonColor = "#145a92"
		}

		buttonBlock := ""
		if buttonLabel != "" {
			buttonBlock = `
              {{if .ActionURL}}
              <p style="margin:24px 0 0 0;">
                <a href="{{.ActionURL}}" style="background:` + buttonColor + `; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block; font-size:14px; font-weight:700;">
                  ` + buttonLabel + `
                </a>
              </p>
              {{end}}`
		}

		return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>` + title + `</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f6f8; font-family:Arial, Helvetica, sans-serif; color:#161616;">
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8; padding:32px 12px;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" role="presentation" style="width:600px; max-width:100%; background-color:#ffffff; border-radius:8px; overflow:hidden; border:1px solid #e0e0e0;">
          <tr>
            <td style="background:` + headingColor + `; padding:24px 28px;">
              <h1 style="margin:0; color:#ffffff; font-size:22px; line-height:1.3; font-weight:700;">
                ` + heading + `
              </h1>
              <p style="margin:6px 0 0 0; color:#d0e2ff; font-size:13px; line-height:1.5;">
                {{if .Subtitle}}{{.Subtitle}}{{else}}{{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}{{end}}
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:28px;">
              <p style="margin:0 0 18px 0; color:#393939; font-size:14px; line-height:1.7;">
                Hello {{if .Name}}{{.Name}}{{else}}Administrator{{end}},
              </p>

` + body + buttonBlock + `

              {{if .Details}}
              <pre style="margin:20px 0 0 0; background:#f4f4f4; padding:16px; border-radius:6px; color:#161616; font-size:12px; line-height:1.6; white-space:pre-wrap; border:1px solid #e0e0e0;">{{.Details}}</pre>
              {{end}}

              {{if .ManualURL}}
              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#edf5ff; border-left:4px solid #0f62fe; border-radius:4px; margin:20px 0 0 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="margin:0; color:#044317; font-size:13px; line-height:1.6;">
                      <strong>Getting Started:</strong>
                      New to the platform? Review the
                      <a href="{{.ManualURL}}" style="color:#0f62fe; font-weight:700; text-decoration:underline;">
                        Platform Onboarding & Usage Manual
                      </a>
                      to get up to speed quickly.
                    </p>
                  </td>
                </tr>
              </table>
              {{end}}

              <p style="margin:22px 0 0 0; color:#6f6f6f; font-size:12px; line-height:1.6;">
                This is an automated message from {{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}.
              </p>
            </td>
          </tr>

          <tr>
            <td align="center" style="padding:16px 28px; background:#f4f4f4; color:#6f6f6f; font-size:12px; line-height:1.6;">
              This is an automated message. Please do not share sensitive account information with anyone.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
	}

	return map[string]string{
		TemplateWelcome: `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Welcome</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f6f8; font-family:Arial, Helvetica, sans-serif; color:#161616;">
  <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f6f8; padding:32px 12px;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" role="presentation" style="width:600px; max-width:100%; background-color:#ffffff; border-radius:8px; overflow:hidden; border:1px solid #e0e0e0;">
          <tr>
            <td style="background:#145a92; padding:24px 28px;">
              <h1 style="margin:0; color:#ffffff; font-size:22px; line-height:1.3; font-weight:700;">
                Welcome to {{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}
              </h1>
              <p style="margin:6px 0 0 0; color:#d0e2ff; font-size:13px; line-height:1.5;">
                Your account is ready
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:28px;">
              <p style="margin:0 0 18px 0; color:#393939; font-size:14px; line-height:1.7;">
                Hello {{if .Name}}{{.Name}}{{else}}User{{end}},
              </p>

              <p style="margin:0 0 18px 0; color:#393939; font-size:14px; line-height:1.7;">
                Your account has been successfully created. Please use the credentials below to sign in.
              </p>

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#f8f9fa; border:1px solid #e0e0e0; border-radius:6px; margin:18px 0;">
                <tr>
                  <td style="padding:18px;">
                    {{if .LoginURL}}
                    <p style="margin:0 0 10px 0; color:#393939; font-size:13px; line-height:1.6;">
                      <strong>Portal:</strong>
                      <a href="{{.LoginURL}}" style="color:#0f62fe; text-decoration:none;">{{.LoginURL}}</a>
                    </p>
                    {{end}}

                    {{if .Username}}
                    <p style="margin:0 0 10px 0; color:#393939; font-size:13px; line-height:1.6;">
                      <strong>Username:</strong>
                      <a href="mailto:{{.Username}}" style="color:#0f62fe; text-decoration:none;">{{.Username}}</a>
                    </p>
                    {{end}}

                    {{if .Email}}
                    <p style="margin:0 0 10px 0; color:#393939; font-size:13px; line-height:1.6;">
                      <strong>Email:</strong>
                      <a href="mailto:{{.Email}}" style="color:#0f62fe; text-decoration:none;">{{.Email}}</a>
                    </p>
                    {{end}}

                    {{if .TemporaryPassword}}
                    <p style="margin:0; color:#393939; font-size:13px; line-height:1.6;">
                      <strong>Temporary Password:</strong>
                      <span style="font-family:Consolas, Monaco, monospace; background:#ffffff; border:1px solid #e0e0e0; padding:3px 6px; border-radius:4px;">{{.TemporaryPassword}}</span>
                    </p>
                    {{else}}
                    {{if .Password}}
                    <p style="margin:0; color:#393939; font-size:13px; line-height:1.6;">
                      <strong>Temporary Password:</strong>
                      <span style="font-family:Consolas, Monaco, monospace; background:#ffffff; border:1px solid #e0e0e0; padding:3px 6px; border-radius:4px;">{{.Password}}</span>
                    </p>
                    {{end}}
                    {{end}}
                  </td>
                </tr>
              </table>

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#fff4e5; border-left:4px solid #ff832b; border-radius:4px; margin:18px 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="margin:0; color:#8a3800; font-size:13px; line-height:1.6;">
                      <strong>Security Notice:</strong>
                      For your safety, please sign in and change your password immediately.
                    </p>
                  </td>
                </tr>
              </table>

              {{if .ManualURL}}
              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#edf5ff; border-left:4px solid #0f62fe; border-radius:4px; margin:18px 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="margin:0; color:#044317; font-size:13px; line-height:1.6;">
                      <strong>Getting Started:</strong>
                      New to the platform? Review the
                      <a href="{{.ManualURL}}" style="color:#0f62fe; font-weight:700; text-decoration:underline;">
                        Platform Onboarding & Usage Manual
                      </a>
                      to get up to speed quickly.
                    </p>
                  </td>
                </tr>
              </table>
              {{end}}

              {{if .LoginURL}}
              <p style="margin:22px 0;">
                <a href="{{.LoginURL}}" style="background:#145a92; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block; font-size:14px; font-weight:700;">
                  Sign in to Portal
                </a>
              </p>
              {{end}}

              <p style="margin:20px 0 0 0; color:#393939; font-size:13px; line-height:1.7;">
                If you experience any issues accessing your account, please reply to this email.
              </p>

              <p style="margin:22px 0 0 0; color:#393939; font-size:13px; line-height:1.7;">
                Best regards,<br/>
                <strong>{{if .TeamName}}{{.TeamName}}{{else}}{{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}} Team{{end}}</strong>
              </p>
            </td>
          </tr>

          <tr>
            <td align="center" style="padding:16px 28px; background:#f4f4f4; color:#6f6f6f; font-size:12px; line-height:1.6;">
              This is an automated message. Please do not share your password with anyone.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,

		TemplatePasswordReset: baseLayout(
			"Password Reset",
			"Reset your password",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 18px 0;">
                We received a request to reset your password for <strong>{{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}</strong>.
              </p>

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#fff4e5; border-left:4px solid #ff832b; border-radius:4px; margin:18px 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="margin:0; color:#8a3800; font-size:13px; line-height:1.6;">
                      <strong>Security Notice:</strong>
                      Only continue if you requested this password reset.
                      {{if .ExpiresIn}} This link expires in <strong>{{.ExpiresIn}}</strong>.{{end}}
                    </p>
                  </td>
                </tr>
              </table>

              {{if .ResetURL}}
              <p style="margin:24px 0;">
                <a href="{{.ResetURL}}" style="background:#145a92; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block; font-size:14px; font-weight:700;">
                  Reset Password
                </a>
              </p>
              {{end}}

              <p style="color:#6f6f6f; font-size:12px; line-height:1.6; margin:0;">
                If you did not request a password reset, you can safely ignore this email.
              </p>`,
			"",
			"#145a92",
		),

		TemplateVerifyEmail: baseLayout(
			"Verify Email",
			"Verify your email address",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 18px 0;">
                Please confirm your email address for <strong>{{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}</strong>.
              </p>

              {{if .VerifyURL}}
              <p style="margin:24px 0;">
                <a href="{{.VerifyURL}}" style="background:#145a92; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block; font-size:14px; font-weight:700;">
                  Verify Email
                </a>
              </p>
              {{end}}

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#edf5ff; border-left:4px solid #0f62fe; border-radius:4px; margin:18px 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="margin:0; color:#044317; font-size:13px; line-height:1.6;">
                      <strong>Account Verification:</strong>
                      Verifying your email helps secure your account and enables account recovery.
                    </p>
                  </td>
                </tr>
              </table>

              <p style="color:#6f6f6f; font-size:12px; line-height:1.6; margin:0;">
                If you did not create this account, you can ignore this email.
              </p>`,
			"",
			"#145a92",
		),

		TemplateNotification: baseLayout(
			"Notification",
			"{{if .Heading}}{{.Heading}}{{else}}{{.Subject}}{{end}}",
			"#145a92",
			`
              {{if .Message}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                {{.Message}}
              </p>
              {{end}}`,
			"{{if .ActionLabel}}{{.ActionLabel}}{{else}}Open{{end}}",
			"#145a92",
		),

		TemplateAdminAlert: baseLayout(
			"Admin Alert",
			"Admin Alert",
			"#da1e28",
			`
              {{if .Message}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 16px 0;">
                {{.Message}}
              </p>
              {{end}}

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#fff1f1; border-left:4px solid #da1e28; border-radius:4px; margin:18px 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    {{if .Title}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Title:</strong> {{.Title}}
                    </p>
                    {{end}}

                    {{if .Status}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Status:</strong> {{.Status}}
                    </p>
                    {{end}}

                    {{if .AnnouncementID}}
                    <p style="color:#6f6f6f; font-size:12px; line-height:1.5; margin:0 0 8px 0;">
                      <strong>Announcement ID:</strong> {{.AnnouncementID}}
                    </p>
                    {{end}}

                    {{if .DocumentName}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0;">
                      <strong>Document:</strong> {{.DocumentName}}
                    </p>
                    {{end}}
                  </td>
                </tr>
              </table>`,
			"Review Alert",
			"#da1e28",
		),

		TemplateWeeklySummary: baseLayout(
			"Weekly Summary",
			"Weekly Summary",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 16px 0;">
                Here is your summary for <strong>{{.Period}}</strong>.
              </p>
              {{if .Summary}}
              <div style="color:#393939; font-size:14px; line-height:1.7; white-space:pre-line;">
                {{.Summary}}
              </div>
              {{end}}`,
			"View Dashboard",
			"#145a92",
		),

		// ----------------------------------------------------
		// Documents
		// ----------------------------------------------------

		"document-created": baseLayout(
			"Document Created",
			"Document created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                A document named <strong>{{.DocumentName}}</strong> has been uploaded to <strong>{{.Platform}}</strong>.
              </p>
              {{if .DocumentType}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                Type: <strong>{{.DocumentType}}</strong>
              </p>
              {{end}}
              {{if .Status}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"View Document",
			"#145a92",
		),

		"document-processed": baseLayout(
			"Document Processed",
			"Document processed successfully",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Document <strong>{{.DocumentName}}</strong> has been processed successfully.
              </p>
              {{if .DocumentType}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Type: <strong>{{.DocumentType}}</strong>
              </p>
              {{end}}`,
			"View Details",
			"#24a148",
		),

		"document-failed": baseLayout(
			"Document Processing Failed",
			"Document processing failed",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                The system was unable to process <strong>{{.DocumentName}}</strong>.
              </p>
              {{if .Reason}}
              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#fff1f1; border-left:4px solid #da1e28; border-radius:4px; margin:16px 0 0 0;">
                <tr>
                  <td style="padding:14px 16px;">
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0;">
                      <strong>Reason:</strong> {{.Reason}}
                    </p>
                  </td>
                </tr>
              </table>
              {{end}}`,
			"Review Document",
			"#da1e28",
		),

		"document-edited": baseLayout(
			"Document Edited",
			"Document edited",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                The document <strong>{{.DocumentName}}</strong> has been updated.
              </p>`,
			"View Document",
			"#145a92",
		),

		"document-deleted": baseLayout(
			"Document Deleted",
			"Document deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                The document <strong>{{.DocumentName}}</strong> has been deleted.
              </p>`,
			"Open Documents",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Users
		// ----------------------------------------------------

		"user-created": baseLayout(
			"User Created",
			"User account created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 14px 0;">
                A new user account has been created.
              </p>

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#f8f9fa; border:1px solid #e0e0e0; border-radius:6px; margin:18px 0;">
                <tr>
                  <td style="padding:16px;">
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Username:</strong> {{.Username}}
                    </p>
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0;">
                      <strong>Email:</strong> {{.Email}}
                    </p>
                  </td>
                </tr>
              </table>`,
			"View User",
			"#145a92",
		),

		"user-enabled": baseLayout(
			"User Enabled",
			"User account enabled",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                The user account <strong>{{.Username}}</strong> has been enabled.
              </p>
              {{if .Email}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"View User",
			"#24a148",
		),

		"user-disabled": baseLayout(
			"User Disabled",
			"User account disabled",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                The user account <strong>{{.Username}}</strong> has been disabled.
              </p>
              {{if .Email}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"View User",
			"#8d6b00",
		),

		"user-password-reset": baseLayout(
			"User Password Reset",
			"User password reset",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                A password reset was performed for user <strong>{{.Username}}</strong>.
              </p>
              {{if .Email}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"View User",
			"#8d6b00",
		),

		"user-deleted": baseLayout(
			"User Deleted",
			"User account deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                The user account <strong>{{.Username}}</strong> has been deleted.
              </p>
              {{if .Email}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"Open Users",
			"#da1e28",
		),

		"client-roles-updated": baseLayout(
			"Client Roles Updated",
			"User client roles updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 14px 0;">
                Client roles were updated for user <strong>{{.Username}}</strong>.
              </p>

              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#f8f9fa; border:1px solid #e0e0e0; border-radius:6px; margin:18px 0;">
                <tr>
                  <td style="padding:16px;">
                    {{if .Email}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Email:</strong> {{.Email}}
                    </p>
                    {{end}}
                    {{if .ClientID}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Client ID:</strong> {{.ClientID}}
                    </p>
                    {{end}}
                    {{if .AddedRoles}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                      <strong>Added Roles:</strong> {{.AddedRoles}}
                    </p>
                    {{end}}
                    {{if .RemovedRoles}}
                    <p style="color:#393939; font-size:13px; line-height:1.6; margin:0;">
                      <strong>Removed Roles:</strong> {{.RemovedRoles}}
                    </p>
                    {{end}}
                  </td>
                </tr>
              </table>`,
			"Review Roles",
			"#145a92",
		),

		// ----------------------------------------------------
		// Clients
		// ----------------------------------------------------

		"client-created": baseLayout(
			"Client Created",
			"Client application created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                A new client application <strong>{{.ClientName}}</strong> has been created.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Client ID: <strong>{{.ClientID}}</strong>
              </p>`,
			"View Client",
			"#145a92",
		),

		"client-enabled": baseLayout(
			"Client Enabled",
			"Client application enabled",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Client application <strong>{{.ClientName}}</strong> has been enabled.
              </p>`,
			"View Client",
			"#24a148",
		),

		"client-disabled": baseLayout(
			"Client Disabled",
			"Client application disabled",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Client application <strong>{{.ClientName}}</strong> has been disabled.
              </p>`,
			"View Client",
			"#8d6b00",
		),

		"client-deleted": baseLayout(
			"Client Deleted",
			"Client application deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Client application <strong>{{.ClientName}}</strong> has been deleted.
              </p>`,
			"Open Clients",
			"#da1e28",
		),

		"client-role-created": baseLayout(
			"Client Role Created",
			"Client role created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Role <strong>{{.Role}}</strong> has been created for client <strong>{{.ClientID}}</strong>.
              </p>`,
			"View Client",
			"#145a92",
		),

		"client-role-deleted": baseLayout(
			"Client Role Deleted",
			"Client role deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Role <strong>{{.Role}}</strong> has been deleted from client <strong>{{.ClientID}}</strong>.
              </p>`,
			"View Client",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Announcements
		// ----------------------------------------------------

		TemplateAnnouncement: baseLayout(
			"Announcement",
			"{{if .Title}}{{.Title}}{{else}}Announcement{{end}}",
			"#145a92",
			`
              {{if .Summary}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 16px 0;">
                {{.Summary}}
              </p>
              {{end}}

              {{if .Message}}
              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#f4f4f4; border-left:4px solid #145a92; border-radius:4px; margin:16px 0;">
                <tr>
                  <td style="padding:16px;">
                    <p style="margin:0; color:#161616; font-size:14px; line-height:1.7; white-space:pre-line;">
                      {{.Message}}
                    </p>
                  </td>
                </tr>
              </table>
              {{end}}

              {{if .Level}}
              <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                <strong>Level:</strong> {{.Level}}
              </p>
              {{end}}

              {{if .Status}}
              <p style="color:#393939; font-size:13px; line-height:1.6; margin:0 0 8px 0;">
                <strong>Status:</strong> {{.Status}}
              </p>
              {{end}}

              {{if .HasAttachments}}
              <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background:#edf5ff; border:1px solid #d0e2ff; border-radius:4px; margin:16px 0;">
                <tr>
                  <td style="padding:12px 14px;">
                    <p style="color:#0f62fe; font-size:13px; line-height:1.6; margin:0;">
                      This announcement includes {{.AttachmentCount}} attachment(s).
                    </p>
                    {{if .AttachmentNames}}
                    <p style="color:#393939; font-size:12px; line-height:1.6; margin:6px 0 0 0;">
                      {{range $index, $name := .AttachmentNames}}{{if $index}}, {{end}}{{$name}}{{end}}
                    </p>
                    {{end}}
                  </td>
                </tr>
              </table>
              {{end}}

              {{if .AnnouncementID}}
              <p style="color:#6f6f6f; font-size:12px; line-height:1.5; margin:0;">
                Reference: {{.AnnouncementID}}
              </p>
              {{end}}`,
			"View Announcement",
			"#145a92",
		),

		TemplateAnnouncementPublished: baseLayout(
			"Announcement Published",
			"Announcement published",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Announcement <strong>{{.Title}}</strong> has been published.
              </p>
              {{if .Message}}
              <p style="color:#393939; font-size:14px; line-height:1.7; white-space:pre-line; margin:0 0 12px 0;">
                {{.Message}}
              </p>
              {{end}}
              {{if .Status}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"Open Announcements",
			"#24a148",
		),

		TemplateAnnouncementArchived: baseLayout(
			"Announcement Archived",
			"Announcement archived",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Announcement <strong>{{.Title}}</strong> has been archived.
              </p>
              {{if .Status}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"Open Announcements",
			"#8d6b00",
		),

		TemplateAnnouncementDeleted: baseLayout(
			"Announcement Deleted",
			"Announcement deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Announcement <strong>{{.Title}}</strong> has been deleted.
              </p>
              {{if .Status}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Previous Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"Open Announcements",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Document template management
		// ----------------------------------------------------

		"document-template-created": baseLayout(
			"Document Template Created",
			"Document template created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Template <strong>{{.TemplateName}}</strong> has been created.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.TemplateCode}}</strong><br/>
                File Type: <strong>{{.FileType}}</strong>
              </p>`,
			"View Template",
			"#145a92",
		),

		"document-template-updated": baseLayout(
			"Document Template Updated",
			"Document template updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Template <strong>{{.TemplateName}}</strong> has been updated.
              </p>
              {{if .TemplateCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#145a92",
		),

		"document-template-published": baseLayout(
			"Document Template Published",
			"Document template published",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Template <strong>{{.TemplateName}}</strong> has been published and activated.
              </p>
              {{if .TemplateCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#24a148",
		),

		"document-template-archived": baseLayout(
			"Document Template Archived",
			"Document template archived",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Template <strong>{{.TemplateName}}</strong> has been archived.
              </p>
              {{if .TemplateCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"Open Templates",
			"#8d6b00",
		),

		"document-template-deleted": baseLayout(
			"Document Template Deleted",
			"Document template deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Template <strong>{{.TemplateName}}</strong> and its versions have been deleted.
              </p>
              {{if .TemplateCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"Open Templates",
			"#da1e28",
		),

		"document-template-structure-created": baseLayout(
			"Document Template Structure Created",
			"Template structure created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Structure for template <strong>{{.TemplateName}}</strong> has been created.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Sheets: <strong>{{.SheetsCount}}</strong><br/>
                Columns: <strong>{{.ColumnsCount}}</strong>
              </p>`,
			"View Structure",
			"#145a92",
		),

		"document-template-sheet-created": baseLayout(
			"Template Sheet Created",
			"Template sheet created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Sheet <strong>{{.SheetName}}</strong> has been added.
              </p>
              {{if .SheetCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#145a92",
		),

		"document-template-sheet-updated": baseLayout(
			"Template Sheet Updated",
			"Template sheet updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Sheet <strong>{{.SheetName}}</strong> has been updated.
              </p>
              {{if .SheetCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#145a92",
		),

		"document-template-sheet-archived": baseLayout(
			"Template Sheet Archived",
			"Template sheet archived",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Sheet <strong>{{.SheetName}}</strong> has been archived.
              </p>
              {{if .SheetCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#8d6b00",
		),

		"document-template-sheet-deleted": baseLayout(
			"Template Sheet Deleted",
			"Template sheet deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Sheet <strong>{{.SheetName}}</strong> has been deleted.
              </p>
              {{if .SheetCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#da1e28",
		),

		"document-template-column-created": baseLayout(
			"Template Column Created",
			"Template column created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Column <strong>{{.ColumnName}}</strong> has been added.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Key: <strong>{{.ColumnKey}}</strong><br/>
                Type: <strong>{{.DataType}}</strong>
              </p>`,
			"View Template",
			"#145a92",
		),

		"document-template-column-updated": baseLayout(
			"Template Column Updated",
			"Template column updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Column <strong>{{.ColumnName}}</strong> has been updated.
              </p>
              {{if .ColumnKey}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Key: <strong>{{.ColumnKey}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#145a92",
		),

		"document-template-column-archived": baseLayout(
			"Template Column Archived",
			"Template column archived",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Column <strong>{{.ColumnName}}</strong> has been archived.
              </p>
              {{if .ColumnKey}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Key: <strong>{{.ColumnKey}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#8d6b00",
		),

		"document-template-column-deleted": baseLayout(
			"Template Column Deleted",
			"Template column deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Column <strong>{{.ColumnName}}</strong> has been deleted.
              </p>
              {{if .ColumnKey}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Key: <strong>{{.ColumnKey}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Storage locations
		// ----------------------------------------------------

		"storage-location-created": baseLayout(
			"Storage Location Created",
			"Storage location created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been created.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.Code}}</strong><br/>
                Provider: <strong>{{.Provider}}</strong>
              </p>`,
			"View Storage",
			"#145a92",
		),

		"storage-location-updated": baseLayout(
			"Storage Location Updated",
			"Storage location updated",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been updated.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.Code}}</strong><br/>
                Provider: <strong>{{.Provider}}</strong>
              </p>`,
			"View Storage",
			"#8d6b00",
		),

		"storage-location-deleted": baseLayout(
			"Storage Location Deleted",
			"Storage location deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been deleted.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Code: <strong>{{.Code}}</strong><br/>
                Provider: <strong>{{.Provider}}</strong>
              </p>`,
			"Open Storage",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Surveillance
		// ----------------------------------------------------

		"surveillance-alert-import-completed": baseLayout(
			"Surveillance Alert Import Completed",
			"Alert import completed with errors",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Surveillance alert batch processing completed with failed rows.
              </p>
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Batch ID: <strong>{{.BatchID}}</strong><br/>
                Success rows: <strong>{{.SuccessRows}}</strong><br/>
                Failed rows: <strong>{{.FailedRows}}</strong>
              </p>`,
			"View Imports",
			"#8d6b00",
		),

		"surveillance-alert-import-failed": baseLayout(
			"Surveillance Alert Import Failed",
			"Alert import failed",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Surveillance alert batch processing failed.
              </p>
              {{if .BatchID}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                Batch ID: <strong>{{.BatchID}}</strong>
              </p>
              {{end}}
              {{if .Error}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Error: <strong>{{.Error}}</strong>
              </p>
              {{end}}`,
			"Review Import",
			"#da1e28",
		),

		"surveillance-disease-created": baseLayout(
			"Disease Created",
			"Surveillance disease created",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Disease <strong>{{.DiseaseName}}</strong> has been created.
              </p>`,
			"View Diseases",
			"#145a92",
		),

		"surveillance-disease-upserted": baseLayout(
			"Disease Updated",
			"Surveillance disease updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Disease <strong>{{.DiseaseName}}</strong> has been created or updated.
              </p>`,
			"View Diseases",
			"#145a92",
		),

		"surveillance-disease-enabled": baseLayout(
			"Disease Enabled",
			"Surveillance disease enabled",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Disease <strong>{{.DiseaseName}}</strong> has been enabled.
              </p>`,
			"View Diseases",
			"#24a148",
		),

		"surveillance-disease-disabled": baseLayout(
			"Disease Disabled",
			"Surveillance disease disabled",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Disease <strong>{{.DiseaseName}}</strong> has been disabled.
              </p>
              {{if .DiseaseCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                Code: <strong>{{.DiseaseCode}}</strong>
              </p>
              {{end}}
              {{if .Category}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Category: <strong>{{.Category}}</strong>
              </p>
              {{end}}`,
			"View Diseases",
			"#8d6b00",
		),

		"surveillance-disease-deleted": baseLayout(
			"Disease Deleted",
			"Surveillance disease deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Disease <strong>{{.DiseaseName}}</strong> has been deleted.
              </p>
              {{if .DiseaseCode}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                Code: <strong>{{.DiseaseCode}}</strong>
              </p>
              {{end}}
              {{if .Category}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Category: <strong>{{.Category}}</strong>
              </p>
              {{end}}`,
			"View Diseases",
			"#da1e28",
		),

		"surveillance-facility-upserted": baseLayout(
			"Facility Updated",
			"Surveillance facility updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Facility <strong>{{.FacilityName}}</strong> has been created or updated.
              </p>
              {{if .ExternalID}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                External ID: <strong>{{.ExternalID}}</strong>
              </p>
              {{end}}`,
			"View Facilities",
			"#145a92",
		),

		"surveillance-epi-week-upserted": baseLayout(
			"EPI Week Updated",
			"Surveillance EPI week updated",
			"#145a92",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                EPI Week <strong>{{.EpiWeek}}</strong> for year <strong>{{.EpiYear}}</strong> has been created or updated.
              </p>`,
			"View EPI Weeks",
			"#145a92",
		),

		"surveillance-subcounty-deleted": baseLayout(
			"Sub-county Deleted",
			"Surveillance sub-county deleted",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Sub-county <strong>{{.SubCountyName}}</strong> has been deleted from the surveillance location hierarchy.
              </p>
              {{if .DistrictID}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                District ID: <strong>{{.DistrictID}}</strong>
              </p>
              {{end}}`,
			"Open Surveillance",
			"#da1e28",
		),

		// ----------------------------------------------------
		// System / backup / security
		// ----------------------------------------------------

		"system-startup": baseLayout(
			"System Startup",
			"System started",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                {{.Platform}} has started successfully.
              </p>
              {{if .Environment}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Environment: <strong>{{.Environment}}</strong>
              </p>
              {{end}}`,
			"Open Dashboard",
			"#24a148",
		),

		"system-shutdown": baseLayout(
			"System Shutdown",
			"System shutdown",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                {{.Platform}} has been stopped.
              </p>
              {{if .Reason}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Reason: <strong>{{.Reason}}</strong>
              </p>
              {{end}}`,
			"Open Dashboard",
			"#8d6b00",
		),

		"backup-completed": baseLayout(
			"Backup Completed",
			"Backup completed",
			"#24a148",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Backup <strong>{{.BackupID}}</strong> completed successfully.
              </p>
              {{if .SizeMB}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Size: <strong>{{.SizeMB}} MB</strong>
              </p>
              {{end}}`,
			"View Backups",
			"#24a148",
		),

		"backup-failed": baseLayout(
			"Backup Failed",
			"Backup failed",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                Backup <strong>{{.BackupID}}</strong> failed.
              </p>
              {{if .Error}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                Error: <strong>{{.Error}}</strong>
              </p>
              {{end}}`,
			"Review Backup",
			"#da1e28",
		),

		"login-failed": baseLayout(
			"Login Failed",
			"Failed login attempt",
			"#8d6b00",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                A failed login attempt was detected.
              </p>
              {{if .Username}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                Username: <strong>{{.Username}}</strong>
              </p>
              {{end}}
              {{if .IP}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                IP Address: <strong>{{.IP}}</strong>
              </p>
              {{end}}`,
			"Review Security",
			"#8d6b00",
		),

		"suspicious-login": baseLayout(
			"Suspicious Login",
			"Suspicious login detected",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 12px 0;">
                A suspicious login event was detected.
              </p>
              {{if .IP}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0 0 8px 0;">
                IP Address: <strong>{{.IP}}</strong>
              </p>
              {{end}}
              {{if .UserAgent}}
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                User Agent: <strong>{{.UserAgent}}</strong>
              </p>
              {{end}}`,
			"Review Security",
			"#da1e28",
		),

		"account-locked": baseLayout(
			"Account Locked",
			"Account locked",
			"#da1e28",
			`
              <p style="color:#393939; font-size:14px; line-height:1.7; margin:0;">
                User account <strong>{{.Username}}</strong> has been locked.
              </p>`,
			"Review Account",
			"#da1e28",
		),
	}
}
