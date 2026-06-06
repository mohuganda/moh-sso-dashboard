import React, { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  Form,
  Loading,
  Select,
  SelectItem,
  Stack,
  TextArea,
  TextInput,
  Tile,
} from "@carbon/react";

import { useQueueEmailMutation, useSendEmailMutation } from "@moh-sso/api";
import { useToast } from "@moh-sso/ui";
import "./email.scss";

type DeliveryMode = "send" | "queue";

type DefaultTemplate =
  | ""
  | "welcome"
  | "password-reset"
  | "verify-email"
  | "notification"
  | "document-processed"
  | "document-failed"
  | "weekly-summary"
  | "admin-alert";

type EmailPanelComponentProps = {
  onSuccess?: () => void;
};

const DEFAULT_PLATFORM = "MOH Integrated Health Portal";
const DEFAULT_DASHBOARD_URL = "http://localhost:3000/admin/home";
const DEFAULT_LOGIN_URL = "http://localhost:3000/login";

const TEMPLATE_OPTIONS: Array<{ value: DefaultTemplate; label: string }> = [
  { value: "", label: "Choose a template" },
  { value: "welcome", label: "Welcome" },
  { value: "notification", label: "Notification" },
  { value: "weekly-summary", label: "Weekly summary" },
  { value: "admin-alert", label: "Admin alert" },
  { value: "document-processed", label: "Document processed" },
  { value: "document-failed", label: "Document failed" },
  { value: "password-reset", label: "Password reset" },
  { value: "verify-email", label: "Verify email" },
];

function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

function parseEmailList(value: string) {
  return value
    .split(",")
    .map((email) => email.trim())
    .filter(Boolean)
    .map((email) => ({ email }));
}

function hasInvalidEmailList(value: string) {
  const emails = value
    .split(",")
    .map((email) => email.trim())
    .filter(Boolean);

  return emails.some((email) => !isValidEmail(email));
}

function getRecipientDisplayName(name: string, email: string) {
  return name.trim() || email.trim();
}

function buildTemplateData(args: {
  templateName: DefaultTemplate;
  recipientName: string;
  recipient: string;
  subject: string;
  message: string;
}) {
  const recipientDisplayName = getRecipientDisplayName(args.recipientName, args.recipient);

  switch (args.templateName) {
    case "welcome":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        LoginURL: DEFAULT_LOGIN_URL,
      };

    case "password-reset":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        ResetURL: `${DEFAULT_LOGIN_URL}?reset=true`,
        ExpiresIn: "30 minutes",
      };

    case "verify-email":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        VerifyURL: `${DEFAULT_LOGIN_URL}?verify=true`,
      };

    case "notification":
      return {
        Subject: args.subject.trim(),
        Heading: args.subject.trim(),
        Message: args.message.trim(),
        ActionURL: DEFAULT_DASHBOARD_URL,
        ActionLabel: "Open Dashboard",
        Platform: DEFAULT_PLATFORM,
      };

    case "document-processed":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        DocumentName: args.subject.trim() || "Uploaded document",
        DocumentType: "Document",
        ActionURL: DEFAULT_DASHBOARD_URL,
      };

    case "document-failed":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        DocumentName: args.subject.trim() || "Uploaded document",
        Reason: args.message.trim(),
        ActionURL: DEFAULT_DASHBOARD_URL,
      };

    case "weekly-summary":
      return {
        Name: recipientDisplayName,
        Platform: DEFAULT_PLATFORM,
        Period: "Current week",
        Summary: args.message.trim(),
        ActionURL: DEFAULT_DASHBOARD_URL,
      };

    case "admin-alert":
      return {
        Platform: DEFAULT_PLATFORM,
        Message: args.message.trim(),
        Details: args.subject.trim(),
        ActionURL: DEFAULT_DASHBOARD_URL,
      };

    default:
      return undefined;
  }
}

const EmailPanelComponent: React.FC<EmailPanelComponentProps> = ({ onSuccess }) => {
  const toast = useToast();

  const [recipient, setRecipient] = useState("");
  const [recipientName, setRecipientName] = useState("");
  const [cc, setCc] = useState("");
  const [bcc, setBcc] = useState("");
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [htmlBody, setHtmlBody] = useState("");
  const [useHtmlBody, setUseHtmlBody] = useState(false);
  const [useTemplate, setUseTemplate] = useState(false);
  const [templateName, setTemplateName] = useState<DefaultTemplate>("");
  const [scheduledAt, setScheduledAt] = useState("");
  const [deliveryMode, setDeliveryMode] = useState<DeliveryMode>("queue");
  const [submitted, setSubmitted] = useState(false);

  const [sendEmail, sendState] = useSendEmailMutation();
  const [queueEmail, queueState] = useQueueEmailMutation();

  const isSubmitting = sendState.isLoading || queueState.isLoading;

  const recipientError = submitted && !isValidEmail(recipient);
  const subjectError = submitted && subject.trim().length === 0;
  const messageError = submitted && message.trim().length === 0;
  const templateError = submitted && useTemplate && templateName === "";
  const ccError = submitted && cc.trim().length > 0 && hasInvalidEmailList(cc);
  const bccError = submitted && bcc.trim().length > 0 && hasInvalidEmailList(bcc);

  const canSubmit = useMemo(() => {
    return (
      isValidEmail(recipient) &&
      subject.trim().length > 0 &&
      message.trim().length > 0 &&
      (!useTemplate || templateName !== "") &&
      !hasInvalidEmailList(cc) &&
      !hasInvalidEmailList(bcc) &&
      !isSubmitting
    );
  }, [recipient, subject, message, cc, bcc, useTemplate, templateName, isSubmitting]);

  const resetForm = () => {
    setRecipient("");
    setRecipientName("");
    setCc("");
    setBcc("");
    setSubject("");
    setMessage("");
    setHtmlBody("");
    setUseHtmlBody(false);
    setUseTemplate(false);
    setTemplateName("");
    setScheduledAt("");
    setSubmitted(false);
  };

  const buildPayload = () => {
    const templateData =
      useTemplate && templateName
        ? buildTemplateData({
            templateName,
            recipientName,
            recipient,
            subject,
            message,
          })
        : undefined;

    return {
      to: [
        {
          name: recipientName.trim() || undefined,
          email: recipient.trim(),
        },
      ],
      cc: cc.trim() ? parseEmailList(cc) : undefined,
      bcc: bcc.trim() ? parseEmailList(bcc) : undefined,
      subject: subject.trim(),
      text_body: message.trim(),
      html_body: !useTemplate && useHtmlBody && htmlBody.trim() ? htmlBody.trim() : undefined,
      template_name: useTemplate && templateName ? templateName : undefined,
      template_data: templateData,
      scheduled_at:
        deliveryMode === "queue" && scheduledAt ? new Date(scheduledAt).toISOString() : undefined,
      metadata: {
        source: "admin-email-panel",
        delivery_mode: deliveryMode,
        template_name: useTemplate ? templateName : "",
      },
    };
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitted(true);

    if (!canSubmit) {
      toast.error({
        title: "Missing or invalid fields",
        subtitle:
          "Please provide a valid recipient email, subject, message, valid CC/BCC emails, and a template if enabled.",
      });
      return;
    }

    try {
      const payload = buildPayload();

      if (deliveryMode === "send") {
        await sendEmail(payload).unwrap();

        toast.success({
          title: "Email sent",
          subtitle: useTemplate
            ? `Sent using the ${templateName} template.`
            : "The email was delivered through the SMTP service.",
        });
      } else {
        await queueEmail(payload).unwrap();

        toast.success({
          title: "Email queued",
          subtitle: useTemplate
            ? `Queued using the ${templateName} template.`
            : "The background email worker will process it shortly.",
        });
      }

      resetForm();
      onSuccess?.();
    } catch (error) {
      console.error(error);

      toast.error({
        title: deliveryMode === "send" ? "Failed to send email" : "Failed to queue email",
        subtitle: "Please check the email service logs and try again.",
      });
    }
  };

  return (
    <div className="email-panel">
      <Tile className="email-panel__tile">
        <Stack gap={6}>
          <div className="email-panel__header">
            <h2 className="email-panel__title">Send email</h2>
            <p className="email-panel__description">
              Send immediately through SMTP or queue the message for background delivery and
              retries. You can also use one of the default email templates.
            </p>
          </div>

          <Form onSubmit={handleSubmit} className="email-form">
            <Stack gap={5}>
              <div className="email-form__mode">
                <Button
                  type="button"
                  kind={deliveryMode === "queue" ? "primary" : "tertiary"}
                  size="sm"
                  disabled={isSubmitting}
                  onClick={() => setDeliveryMode("queue")}
                >
                  Queue email
                </Button>

                <Button
                  type="button"
                  kind={deliveryMode === "send" ? "primary" : "tertiary"}
                  size="sm"
                  disabled={isSubmitting}
                  onClick={() => setDeliveryMode("send")}
                >
                  Send now
                </Button>
              </div>

              <div className="email-form__section">
                <h3 className="email-form__section-title">Recipient</h3>

                <Stack gap={4}>
                  <TextInput
                    id="recipient-name"
                    labelText="Recipient name"
                    value={recipientName}
                    onChange={(event) => setRecipientName(event.target.value)}
                    placeholder="Optional name"
                    disabled={isSubmitting}
                  />

                  <TextInput
                    id="recipient"
                    type="email"
                    labelText="Recipient email"
                    value={recipient}
                    onChange={(event) => setRecipient(event.target.value)}
                    placeholder="user@example.com"
                    invalid={recipientError}
                    invalidText="Enter a valid recipient email address"
                    disabled={isSubmitting}
                    required
                  />

                  <TextInput
                    id="cc"
                    type="text"
                    labelText="CC"
                    value={cc}
                    onChange={(event) => setCc(event.target.value)}
                    placeholder="Optional, comma-separated emails"
                    invalid={ccError}
                    invalidText="One or more CC email addresses are invalid"
                    disabled={isSubmitting}
                  />

                  <TextInput
                    id="bcc"
                    type="text"
                    labelText="BCC"
                    value={bcc}
                    onChange={(event) => setBcc(event.target.value)}
                    placeholder="Optional, comma-separated emails"
                    invalid={bccError}
                    invalidText="One or more BCC email addresses are invalid"
                    disabled={isSubmitting}
                  />
                </Stack>
              </div>

              <div className="email-form__section">
                <h3 className="email-form__section-title">Template</h3>

                <Stack gap={4}>
                  <Checkbox
                    id="use-template"
                    labelText="Use default template"
                    checked={useTemplate}
                    disabled={isSubmitting}
                    onChange={(_, data) => {
                      const checked = Boolean(data.checked);
                      setUseTemplate(checked);

                      if (checked) {
                        setUseHtmlBody(false);
                        setHtmlBody("");
                      } else {
                        setTemplateName("");
                      }
                    }}
                  />

                  {useTemplate && (
                    <Select
                      id="template-name"
                      labelText="Default template"
                      value={templateName}
                      invalid={templateError}
                      invalidText="Choose a template"
                      disabled={isSubmitting}
                      onChange={(event) => setTemplateName(event.target.value as DefaultTemplate)}
                    >
                      {TEMPLATE_OPTIONS.map((template) => (
                        <SelectItem
                          key={template.value || "empty"}
                          value={template.value}
                          text={template.label}
                        />
                      ))}
                    </Select>
                  )}
                </Stack>
              </div>

              <div className="email-form__section">
                <h3 className="email-form__section-title">Message</h3>

                <Stack gap={4}>
                  <TextInput
                    id="subject"
                    labelText="Subject"
                    value={subject}
                    onChange={(event) => setSubject(event.target.value)}
                    placeholder="Email subject"
                    invalid={subjectError}
                    invalidText="Subject is required"
                    disabled={isSubmitting}
                    required
                  />

                  <TextArea
                    id="message"
                    labelText={useTemplate ? "Template message / content" : "Plain text message"}
                    helperText={
                      useTemplate ? "This content is mapped into the selected template." : undefined
                    }
                    value={message}
                    onChange={(event) => setMessage(event.target.value)}
                    placeholder="Write your message here"
                    rows={7}
                    invalid={messageError}
                    invalidText="Message is required"
                    disabled={isSubmitting}
                    required
                  />

                  {!useTemplate && (
                    <>
                      <Checkbox
                        id="use-html-body"
                        labelText="Add HTML body"
                        checked={useHtmlBody}
                        disabled={isSubmitting}
                        onChange={(_, data) => setUseHtmlBody(Boolean(data.checked))}
                      />

                      {useHtmlBody && (
                        <TextArea
                          id="html-body"
                          labelText="HTML body"
                          value={htmlBody}
                          onChange={(event) => setHtmlBody(event.target.value)}
                          placeholder="<p>Hello...</p>"
                          rows={7}
                          disabled={isSubmitting}
                        />
                      )}
                    </>
                  )}
                </Stack>
              </div>

              {deliveryMode === "queue" && (
                <div className="email-form__section">
                  <h3 className="email-form__section-title">Scheduling</h3>

                  <TextInput
                    id="scheduled-at"
                    type="datetime-local"
                    labelText="Scheduled time"
                    helperText="Optional. Leave empty to queue immediately."
                    value={scheduledAt}
                    onChange={(event) => setScheduledAt(event.target.value)}
                    disabled={isSubmitting}
                  />
                </div>
              )}

              <div className="email-form__actions">
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting
                    ? deliveryMode === "send"
                      ? "Sending..."
                      : "Queueing..."
                    : deliveryMode === "send"
                      ? "Send email"
                      : "Queue email"}
                </Button>

                <Button type="button" kind="secondary" disabled={isSubmitting} onClick={resetForm}>
                  Clear
                </Button>
              </div>

              {isSubmitting && <Loading small withOverlay={false} />}
            </Stack>
          </Form>
        </Stack>
      </Tile>
    </div>
  );
};

export default EmailPanelComponent;
