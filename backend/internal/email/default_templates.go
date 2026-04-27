package email

func DefaultTemplates() map[string]string {
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

		"notification": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>{{.Subject}}</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">{{.Heading}}</h2>
              <p style="color:#525252; font-size:14px; line-height:1.7;">
                {{.Message}}
              </p>
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:#0f62fe; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  {{if .ActionLabel}}{{.ActionLabel}}{{else}}Open{{end}}
                </a>
              </p>
              {{end}}
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

		"document-processed": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Document Processed</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">Document processed successfully</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Your document <strong>{{.DocumentName}}</strong> has been processed successfully.
              </p>
              {{if .DocumentType}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Type: <strong>{{.DocumentType}}</strong>
              </p>
              {{end}}
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:#24a148; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  View Details
                </a>
              </p>
              {{end}}
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

		"document-failed": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Document Processing Failed</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#da1e28;">Document processing failed</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                We were unable to process <strong>{{.DocumentName}}</strong>.
              </p>
              {{if .Reason}}
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Reason: <strong>{{.Reason}}</strong>
              </p>
              {{end}}
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:#da1e28; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  Review Document
                </a>
              </p>
              {{end}}
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

		"weekly-summary": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Weekly Summary</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#161616;">Weekly Summary</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                Here is your summary for <strong>{{.Period}}</strong>.
              </p>
              {{if .Summary}}
              <div style="color:#525252; font-size:14px; line-height:1.7; white-space:pre-line;">
                {{.Summary}}
              </div>
              {{end}}
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:#0f62fe; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  View Dashboard
                </a>
              </p>
              {{end}}
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

		"admin-alert": `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Admin Alert</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f4; padding:24px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:32px;">
              <h2 style="margin-top:0; color:#da1e28;">Admin Alert</h2>
              <p style="color:#525252; font-size:14px; line-height:1.6;">
                {{.Message}}
              </p>
              {{if .Details}}
              <pre style="background:#f4f4f4; padding:16px; border-radius:4px; color:#161616; font-size:12px; white-space:pre-wrap;">{{.Details}}</pre>
              {{end}}
              {{if .ActionURL}}
              <p style="margin:24px 0;">
                <a href="{{.ActionURL}}" style="background:#da1e28; color:#ffffff; text-decoration:none; padding:12px 20px; border-radius:4px; display:inline-block;">
                  Review Alert
                </a>
              </p>
              {{end}}
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
	}
}
