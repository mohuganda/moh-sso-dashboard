package email

func DefaultTemplates() map[string]string {
	baseLayout := func(title, heading, headingColor, body, buttonLabel, buttonColor string) string {
		if headingColor == "" {
			headingColor = "#161616"
		}
		if buttonColor == "" {
			buttonColor = "#0f62fe"
		}

		buttonBlock := ""
		if buttonLabel != "" {
			buttonBlock = `
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:` + buttonColor + `; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
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
  <title>` + title + `</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:` + headingColor + `;">` + heading + `</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{if .Name}}{{.Name}}{{else}}Administrator{{end}},
              </p>
` + body + buttonBlock + `
              {{if .Details}}
              <pre style="background:#f4f4f4; padding:16px; border-radius:4px; color:#161616; font-size:12px; white-space:pre-wrap;">{{.Details}}</pre>
              {{end}}
              <p style="color:#6f6f6f; font-size:12px; line-height:1.5;">
                This is an automated message from {{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:16px 32px; background:#f4f4f4; color:#6f6f6f; font-size:12px;">
              {{if .Platform}}{{.Platform}}{{else}}MOH Integrated Health Portal{{end}}
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
		"welcome": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Welcome</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">Welcome {{.Name}}</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Your account for <strong>{{.Platform}}</strong> has been created successfully.
              </p>
              {{if .LoginURL}}
              <p style="margin:24px 0;">
                <a href="{{.LoginURL}}" style="background:#0f62fe; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  Sign in
                </a>
              </p>
              {{end}}
              <p style="color:#6f6f6f; font-size:12px; line-height:1.5;">
                If you did not expect this email, please contact your administrator.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:16px 32px; background:#f4f4f4; color:#6f6f6f; font-size:12px;">
              {{.Platform}}
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,

		"password-reset": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Password Reset</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">Reset your password</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                We received a request to reset your password for <strong>{{.Platform}}</strong>.
              </p>
              <p style="margin:24px 0;">
                <a href="{{.ResetURL}}" style="background:#0f62fe; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  Reset Password
                </a>
              </p>
              {{if .ExpiresIn}}
              <p style="color:#6f6f6f; font-size:12px;">
                This link expires in {{.ExpiresIn}}.
              </p>
              {{end}}
              <p style="color:#6f6f6f; font-size:12px;">
                If you did not request a password reset, you can ignore this email.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:16px 32px; background:#f4f4f4; color:#6f6f6f; font-size:12px;">
              {{.Platform}}
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,

		"verify-email": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Verify Email</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">Verify your email address</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Please confirm your email address for <strong>{{.Platform}}</strong>.
              </p>
              <p style="margin:24px 0;">
                <a href="{{.VerifyURL}}" style="background:#0f62fe; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  Verify Email
                </a>
              </p>
              <p style="color:#6f6f6f; font-size:12px;">
                If you did not create this account, you can ignore this email.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:16px 32px; background:#f4f4f4; color:#6f6f6f; font-size:12px;">
              {{.Platform}}
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,

		"notification": baseLayout(
			"Notification",
			"{{if .Heading}}{{.Heading}}{{else}}{{.Subject}}{{end}}",
			"#161616",
			`
              {{if .Message}}
              <p style="color:#525252; font-size:14px; line-height:1.7;">
                {{.Message}}
              </p>
              {{end}}`,
			"{{if .ActionLabel}}{{.ActionLabel}}{{else}}Open{{end}}",
			"#0f62fe",
		),

		"admin-alert": baseLayout(
			"Admin Alert",
			"Admin Alert",
			"#da1e28",
			`
              {{if .Message}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                {{.Message}}
              </p>
              {{end}}`,
			"Review Alert",
			"#da1e28",
		),

		"weekly-summary": baseLayout(
			"Weekly Summary",
			"Weekly Summary",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Here is your summary for <strong>{{.Period}}</strong>.
              </p>
              {{if .Summary}}
              <div style="color:#525252; font-size:14px; line-height:1.7; white-space:pre-line;">
                {{.Summary}}
              </div>
              {{end}}`,
			"View Dashboard",
			"#0f62fe",
		),

		// ----------------------------------------------------
		// Documents
		// ----------------------------------------------------

		"document-created": baseLayout(
			"Document Created",
			"Document created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A document named <strong>{{.DocumentName}}</strong> has been uploaded to <strong>{{.Platform}}</strong>.
              </p>
              {{if .DocumentType}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Type: <strong>{{.DocumentType}}</strong>
              </p>
              {{end}}
              {{if .Status}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"View Document",
			"#0f62fe",
		),

		"document-processed": baseLayout(
			"Document Processed",
			"Document processed successfully",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Document <strong>{{.DocumentName}}</strong> has been processed successfully.
              </p>
              {{if .DocumentType}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                The system was unable to process <strong>{{.DocumentName}}</strong>.
              </p>
              {{if .Reason}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Reason: <strong>{{.Reason}}</strong>
              </p>
              {{end}}`,
			"Review Document",
			"#da1e28",
		),

		"document-edited": baseLayout(
			"Document Edited",
			"Document edited",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                The document <strong>{{.DocumentName}}</strong> has been updated.
              </p>`,
			"View Document",
			"#0f62fe",
		),

		"document-deleted": baseLayout(
			"Document Deleted",
			"Document deleted",
			"#da1e28",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A new user account has been created.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Username: <strong>{{.Username}}</strong><br/>
                Email: <strong>{{.Email}}</strong>
              </p>`,
			"View User",
			"#0f62fe",
		),

		"user-enabled": baseLayout(
			"User Enabled",
			"User account enabled",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                The user account <strong>{{.Username}}</strong> has been enabled.
              </p>
              {{if .Email}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"View User",
			"#24a148",
		),

		"user-disabled": baseLayout(
			"User Disabled",
			"User account disabled",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                The user account <strong>{{.Username}}</strong> has been disabled.
              </p>
              {{if .Email}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"View User",
			"#8d6b00",
		),

		"user-password-reset": baseLayout(
			"User Password Reset",
			"User password reset",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A password reset was performed for user <strong>{{.Username}}</strong>.
              </p>
              {{if .Email}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                The user account <strong>{{.Username}}</strong> has been deleted.
              </p>
              {{if .Email}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}`,
			"Open Users",
			"#da1e28",
		),

		"client-roles-updated": baseLayout(
			"Client Roles Updated",
			"User client roles updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Client roles were updated for user <strong>{{.Username}}</strong>.
              </p>
              {{if .Email}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Email: <strong>{{.Email}}</strong>
              </p>
              {{end}}
              {{if .ClientID}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Client ID: <strong>{{.ClientID}}</strong>
              </p>
              {{end}}
              {{if .AddedRoles}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Added Roles: <strong>{{.AddedRoles}}</strong>
              </p>
              {{end}}
              {{if .RemovedRoles}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Removed Roles: <strong>{{.RemovedRoles}}</strong>
              </p>
              {{end}}`,
			"Review Roles",
			"#0f62fe",
		),

		// ----------------------------------------------------
		// Clients
		// ----------------------------------------------------

		"client-created": baseLayout(
			"Client Created",
			"Client application created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A new client application <strong>{{.ClientName}}</strong> has been created.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Client ID: <strong>{{.ClientID}}</strong>
              </p>`,
			"View Client",
			"#0f62fe",
		),

		"client-enabled": baseLayout(
			"Client Enabled",
			"Client application enabled",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Client application <strong>{{.ClientName}}</strong> has been enabled.
              </p>`,
			"View Client",
			"#24a148",
		),

		"client-disabled": baseLayout(
			"Client Disabled",
			"Client application disabled",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Client application <strong>{{.ClientName}}</strong> has been deleted.
              </p>`,
			"Open Clients",
			"#da1e28",
		),

		"client-role-created": baseLayout(
			"Client Role Created",
			"Client role created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Role <strong>{{.Role}}</strong> has been created for client <strong>{{.ClientID}}</strong>.
              </p>`,
			"View Client",
			"#0f62fe",
		),

		"client-role-deleted": baseLayout(
			"Client Role Deleted",
			"Client role deleted",
			"#da1e28",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Role <strong>{{.Role}}</strong> has been deleted from client <strong>{{.ClientID}}</strong>.
              </p>`,
			"View Client",
			"#da1e28",
		),

		// ----------------------------------------------------
		// Announcements
		// ----------------------------------------------------

		"announcement-published": baseLayout(
			"Announcement Published",
			"Announcement published",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Announcement <strong>{{.Title}}</strong> has been published.
              </p>
              {{if .Status}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"Open Announcements",
			"#24a148",
		),

		"announcement-archived": baseLayout(
			"Announcement Archived",
			"Announcement archived",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Announcement <strong>{{.Title}}</strong> has been archived.
              </p>
              {{if .Status}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Status: <strong>{{.Status}}</strong>
              </p>
              {{end}}`,
			"Open Announcements",
			"#8d6b00",
		),

		"announcement-deleted": baseLayout(
			"Announcement Deleted",
			"Announcement deleted",
			"#da1e28",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Announcement <strong>{{.Title}}</strong> has been deleted.
              </p>
              {{if .Status}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Template <strong>{{.TemplateName}}</strong> has been created.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.TemplateCode}}</strong><br/>
                File Type: <strong>{{.FileType}}</strong>
              </p>`,
			"View Template",
			"#0f62fe",
		),

		"document-template-updated": baseLayout(
			"Document Template Updated",
			"Document template updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Template <strong>{{.TemplateName}}</strong> has been updated.
              </p>
              {{if .TemplateCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#0f62fe",
		),

		"document-template-published": baseLayout(
			"Document Template Published",
			"Document template published",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Template <strong>{{.TemplateName}}</strong> has been published and activated.
              </p>
              {{if .TemplateCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#24a148",
		),

		"document-template-archived": baseLayout(
			"Document Template Archived",
			"Document template archived",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Template <strong>{{.TemplateName}}</strong> has been archived.
              </p>
              {{if .TemplateCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Template <strong>{{.TemplateName}}</strong> and its versions have been deleted.
              </p>
              {{if .TemplateCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.TemplateCode}}</strong>
              </p>
              {{end}}`,
			"Open Templates",
			"#da1e28",
		),

		"document-template-structure-created": baseLayout(
			"Document Template Structure Created",
			"Template structure created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Structure for template <strong>{{.TemplateName}}</strong> has been created.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheets: <strong>{{.SheetsCount}}</strong><br/>
                Columns: <strong>{{.ColumnsCount}}</strong>
              </p>`,
			"View Structure",
			"#0f62fe",
		),

		"document-template-sheet-created": baseLayout(
			"Template Sheet Created",
			"Template sheet created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet <strong>{{.SheetName}}</strong> has been added.
              </p>
              {{if .SheetCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#0f62fe",
		),

		"document-template-sheet-updated": baseLayout(
			"Template Sheet Updated",
			"Template sheet updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet <strong>{{.SheetName}}</strong> has been updated.
              </p>
              {{if .SheetCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#0f62fe",
		),

		"document-template-sheet-archived": baseLayout(
			"Template Sheet Archived",
			"Template sheet archived",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet <strong>{{.SheetName}}</strong> has been archived.
              </p>
              {{if .SheetCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet <strong>{{.SheetName}}</strong> has been deleted.
              </p>
              {{if .SheetCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sheet Code: <strong>{{.SheetCode}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#da1e28",
		),

		"document-template-column-created": baseLayout(
			"Template Column Created",
			"Template column created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Column <strong>{{.ColumnName}}</strong> has been added.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Key: <strong>{{.ColumnKey}}</strong><br/>
                Type: <strong>{{.DataType}}</strong>
              </p>`,
			"View Template",
			"#0f62fe",
		),

		"document-template-column-updated": baseLayout(
			"Template Column Updated",
			"Template column updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Column <strong>{{.ColumnName}}</strong> has been updated.
              </p>
              {{if .ColumnKey}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Key: <strong>{{.ColumnKey}}</strong>
              </p>
              {{end}}`,
			"View Template",
			"#0f62fe",
		),

		"document-template-column-archived": baseLayout(
			"Template Column Archived",
			"Template column archived",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Column <strong>{{.ColumnName}}</strong> has been archived.
              </p>
              {{if .ColumnKey}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Column <strong>{{.ColumnName}}</strong> has been deleted.
              </p>
              {{if .ColumnKey}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been created.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.Code}}</strong><br/>
                Provider: <strong>{{.Provider}}</strong>
              </p>`,
			"View Storage",
			"#0f62fe",
		),

		"storage-location-updated": baseLayout(
			"Storage Location Updated",
			"Storage location updated",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been updated.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Storage location <strong>{{.StorageLocationName}}</strong> has been deleted.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Surveillance alert batch processing completed with failed rows.
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Surveillance alert batch processing failed.
              </p>
              {{if .BatchID}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Batch ID: <strong>{{.BatchID}}</strong>
              </p>
              {{end}}
              {{if .Error}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Error: <strong>{{.Error}}</strong>
              </p>
              {{end}}`,
			"Review Import",
			"#da1e28",
		),

		"surveillance-disease-created": baseLayout(
			"Disease Created",
			"Surveillance disease created",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Disease <strong>{{.DiseaseName}}</strong> has been created.
              </p>`,
			"View Diseases",
			"#0f62fe",
		),

		"surveillance-disease-upserted": baseLayout(
			"Disease Updated",
			"Surveillance disease updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Disease <strong>{{.DiseaseName}}</strong> has been created or updated.
              </p>`,
			"View Diseases",
			"#0f62fe",
		),

		"surveillance-disease-enabled": baseLayout(
			"Disease Enabled",
			"Surveillance disease enabled",
			"#24a148",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Disease <strong>{{.DiseaseName}}</strong> has been enabled.
              </p>`,
			"View Diseases",
			"#24a148",
		),

		"surveillance-disease-disabled": baseLayout(
			"Disease Disabled",
			"Surveillance disease disabled",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Disease <strong>{{.DiseaseName}}</strong> has been disabled.
              </p>
              {{if .DiseaseCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.DiseaseCode}}</strong>
              </p>
              {{end}}
              {{if .Category}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Disease <strong>{{.DiseaseName}}</strong> has been deleted.
              </p>
              {{if .DiseaseCode}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Code: <strong>{{.DiseaseCode}}</strong>
              </p>
              {{end}}
              {{if .Category}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Category: <strong>{{.Category}}</strong>
              </p>
              {{end}}`,
			"View Diseases",
			"#da1e28",
		),

		"surveillance-facility-upserted": baseLayout(
			"Facility Updated",
			"Surveillance facility updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Facility <strong>{{.FacilityName}}</strong> has been created or updated.
              </p>
              {{if .ExternalID}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                External ID: <strong>{{.ExternalID}}</strong>
              </p>
              {{end}}`,
			"View Facilities",
			"#0f62fe",
		),

		"surveillance-epi-week-upserted": baseLayout(
			"EPI Week Updated",
			"Surveillance EPI week updated",
			"#161616",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                EPI Week <strong>{{.EpiWeek}}</strong> for year <strong>{{.EpiYear}}</strong> has been created or updated.
              </p>`,
			"View EPI Weeks",
			"#0f62fe",
		),

		"surveillance-subcounty-deleted": baseLayout(
			"Sub-county Deleted",
			"Surveillance sub-county deleted",
			"#da1e28",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Sub-county <strong>{{.SubCountyName}}</strong> has been deleted from the surveillance location hierarchy.
              </p>
              {{if .DistrictID}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                {{.Platform}} has started successfully.
              </p>
              {{if .Environment}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Environment: <strong>{{.Environment}}</strong>
              </p>
              {{end}}`,
			"Open Dashboard",
			"#24a148",
		),

		"system-shutdown": baseLayout(
			"System Shutdown",
			"System shutdown",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                {{.Platform}} has been stopped.
              </p>
              {{if .Reason}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Backup <strong>{{.BackupID}}</strong> completed successfully.
              </p>
              {{if .SizeMB}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Backup <strong>{{.BackupID}}</strong> failed.
              </p>
              {{if .Error}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Error: <strong>{{.Error}}</strong>
              </p>
              {{end}}`,
			"Review Backup",
			"#da1e28",
		),

		"login-failed": baseLayout(
			"Login Failed",
			"Failed login attempt",
			"#f1c21b",
			`
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A failed login attempt was detected.
              </p>
              {{if .Username}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Username: <strong>{{.Username}}</strong>
              </p>
              {{end}}
              {{if .IP}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                A suspicious login event was detected.
              </p>
              {{if .IP}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                IP Address: <strong>{{.IP}}</strong>
              </p>
              {{end}}
              {{if .UserAgent}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
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
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                User account <strong>{{.Username}}</strong> has been locked.
              </p>`,
			"Review Account",
			"#da1e28",
		),
	}
}
