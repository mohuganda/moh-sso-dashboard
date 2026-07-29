import React, { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  Form,
  Loading,
  MultiSelect,
  Select,
  SelectItem,
  Stack,
  Tag,
  TextArea,
  TextInput,
  Tile,
} from "@carbon/react";
import { TrashCan } from "@carbon/react/icons";

import { useQueueEmailMutation, useSendEmailMutation } from "../../api";
import type { EmailAttachment } from "../../types";
import { useToast } from "@moh-sso/ui";
import { useListRbacGroupsQuery } from "@moh-sso/rbac";
import { useAvailableHealthContexts } from "@moh-sso/auth";
import "./email-outbox-panel.scss";

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

type EmailAttachmentDraft = EmailAttachment & {
  id: string;
  source: "file" | "path";
  file_size?: number;
};

const DEFAULT_PLATFORM = "MOH Integrated Health Portal";
const DEFAULT_DASHBOARD_URL = "http://localhost:3000/admin/home";
const DEFAULT_LOGIN_URL = "http://localhost:9000/api/v1/auth/login";
const MAX_ATTACHMENT_SIZE = 10 * 1024 * 1024;

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

function getApiErrorMessage(error: unknown) {
  if (
    typeof error === "object" &&
    error !== null &&
    "data" in error &&
    typeof (error as { data?: unknown }).data === "object" &&
    (error as { data?: unknown }).data !== null
  ) {
    const data = (error as { data: { error?: { message?: unknown }; message?: unknown } }).data;
    if (typeof data.error?.message === "string" && data.error.message.trim()) {
      return data.error.message;
    }
    if (typeof data.message === "string" && data.message.trim()) {
      return data.message;
    }
  }

  return undefined;
}

function getRecipientDisplayName(name: string, email: string) {
  return name.trim() || email.trim();
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) {
    return "";
  }

  if (bytes < 1024) {
    return `${bytes} B`;
  }

  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }

  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onerror = () => reject(new Error("Unable to read attachment file."));
    reader.onload = () => {
      const value = String(reader.result ?? "");
      const [, base64 = value] = value.split(",");
      resolve(base64);
    };

    reader.readAsDataURL(file);
  });
}

function cleanAttachments(attachments: EmailAttachmentDraft[]): EmailAttachment[] {
  return attachments.map(({ file_size: _fileSize, source: _source, id: _id, ...attachment }) => ({
    ...attachment,
    content_id: attachment.content_id?.trim() || undefined,
    content_type: attachment.content_type?.trim() || undefined,
    path: attachment.path?.trim() || undefined,
    data_base64: attachment.data_base64?.trim() || undefined,
  }));
}

function getAttachmentValidationError(attachments: EmailAttachmentDraft[]) {
  for (const attachment of attachments) {
    const hasPath = Boolean(attachment.path?.trim());
    const hasData = Boolean(attachment.data_base64?.trim());

    if (!attachment.file_name.trim()) {
      return "Every attachment must have a file name.";
    }

    if (hasPath === hasData) {
      return "Each attachment must provide exactly one of path or data_base64.";
    }
  }

  return "";
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
  const [recipientGroupIds, setRecipientGroupIds] = useState<string[]>([]);
  const [recipientHealthContextIds, setRecipientHealthContextIds] = useState<string[]>([]);
  const [includeHealthContextDescendants, setIncludeHealthContextDescendants] = useState(false);
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
  const [attachments, setAttachments] = useState<EmailAttachmentDraft[]>([]);
  const [pathFileName, setPathFileName] = useState("");
  const [pathContentType, setPathContentType] = useState("");
  const [pathValue, setPathValue] = useState("");
  const [pathInline, setPathInline] = useState(false);
  const [pathContentId, setPathContentId] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const [sendEmail, sendState] = useSendEmailMutation();
  const [queueEmail, queueState] = useQueueEmailMutation();
  const {
    data: groups = [],
    isLoading: groupsLoading,
    isError: groupsError,
  } = useListRbacGroupsQuery();
  const healthContexts = useAvailableHealthContexts();

  const isSubmitting = sendState.isLoading || queueState.isLoading;
  const groupItems = useMemo(
    () =>
      groups
        .filter((group) => group.enabled)
        .map((group) => ({
          id: group.id,
          text: group.displayName || group.path || group.name,
        })),
    [groups],
  );
  const healthContextItems = useMemo(
    () =>
      healthContexts
        .filter((context) => context.enabled)
        .map((context) => ({
          id: context.id,
          text: `${context.name} (${context.contextType.toLowerCase()})`,
        })),
    [healthContexts],
  );
  const hasDirectRecipient = recipient.trim().length > 0;
  const hasGroupRecipients = recipientGroupIds.length > 0;
  const hasHealthContextRecipients = recipientHealthContextIds.length > 0;

  const recipientError = submitted && hasDirectRecipient && !isValidEmail(recipient);
  const recipientRequiredError =
    submitted && !hasDirectRecipient && !hasGroupRecipients && !hasHealthContextRecipients;
  const subjectError = submitted && subject.trim().length === 0;
  const messageError = submitted && message.trim().length === 0;
  const templateError = submitted && useTemplate && templateName === "";
  const ccError = submitted && cc.trim().length > 0 && hasInvalidEmailList(cc);
  const bccError = submitted && bcc.trim().length > 0 && hasInvalidEmailList(bcc);
  const attachmentValidationError = getAttachmentValidationError(attachments);

  const canSubmit = useMemo(() => {
    return (
      (hasDirectRecipient || hasGroupRecipients || hasHealthContextRecipients) &&
      (!hasDirectRecipient || isValidEmail(recipient)) &&
      subject.trim().length > 0 &&
      message.trim().length > 0 &&
      (!useTemplate || templateName !== "") &&
      !hasInvalidEmailList(cc) &&
      !hasInvalidEmailList(bcc) &&
      !attachmentValidationError &&
      !isSubmitting
    );
  }, [
    recipient,
    hasDirectRecipient,
    hasGroupRecipients,
    hasHealthContextRecipients,
    subject,
    message,
    cc,
    bcc,
    useTemplate,
    templateName,
    attachmentValidationError,
    isSubmitting,
  ]);

  const resetForm = () => {
    setRecipient("");
    setRecipientName("");
    setRecipientGroupIds([]);
    setRecipientHealthContextIds([]);
    setIncludeHealthContextDescendants(false);
    setCc("");
    setBcc("");
    setSubject("");
    setMessage("");
    setHtmlBody("");
    setUseHtmlBody(false);
    setUseTemplate(false);
    setTemplateName("");
    setScheduledAt("");
    setAttachments([]);
    setPathFileName("");
    setPathContentType("");
    setPathValue("");
    setPathInline(false);
    setPathContentId("");
    setSubmitted(false);
  };

  const handleFileAttachments = async (files: FileList | null) => {
    if (!files?.length) {
      return;
    }

    const next: EmailAttachmentDraft[] = [];

    for (const file of Array.from(files)) {
      if (file.size > MAX_ATTACHMENT_SIZE) {
        toast.error({
          title: "Attachment too large",
          subtitle: `${file.name} is larger than 10 MB.`,
        });
        continue;
      }

      try {
        next.push({
          id: `${file.name}-${file.lastModified}-${crypto.randomUUID()}`,
          source: "file",
          file_name: file.name,
          content_type: file.type || "application/octet-stream",
          data_base64: await fileToBase64(file),
          file_size: file.size,
          inline: false,
        });
      } catch {
        toast.error({
          title: "Attachment failed",
          subtitle: `Unable to read ${file.name}.`,
        });
      }
    }

    if (next.length > 0) {
      setAttachments((current) => [...current, ...next]);
    }
  };

  const addPathAttachment = () => {
    const fileName = pathFileName.trim();
    const path = pathValue.trim();

    if (!fileName || !path) {
      toast.error({
        title: "Missing path attachment details",
        subtitle: "Provide both a file name and a server-side path.",
      });
      return;
    }

    setAttachments((current) => [
      ...current,
      {
        id: `path-${Date.now()}`,
        source: "path",
        file_name: fileName,
        content_type: pathContentType.trim() || undefined,
        path,
        inline: pathInline,
        content_id: pathInline ? pathContentId.trim() || undefined : undefined,
      },
    ]);
    setPathFileName("");
    setPathContentType("");
    setPathValue("");
    setPathInline(false);
    setPathContentId("");
  };

  const removeAttachment = (id: string) => {
    setAttachments((current) => current.filter((attachment) => attachment.id !== id));
  };

  const updateAttachment = (
    id: string,
    patch: Partial<Pick<EmailAttachmentDraft, "inline" | "content_id">>,
  ) => {
    setAttachments((current) =>
      current.map((attachment) =>
        attachment.id === id
          ? {
              ...attachment,
              ...patch,
              content_id:
                patch.inline === false ? undefined : (patch.content_id ?? attachment.content_id),
            }
          : attachment,
      ),
    );
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
      to: hasDirectRecipient
        ? [
            {
              name: recipientName.trim() || undefined,
              email: recipient.trim(),
            },
          ]
        : undefined,
      to_groups: recipientGroupIds.length > 0 ? recipientGroupIds : undefined,
      to_health_contexts:
        recipientHealthContextIds.length > 0 ? recipientHealthContextIds : undefined,
      include_health_context_descendants:
        recipientHealthContextIds.length > 0 ? includeHealthContextDescendants : undefined,
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
      attachments: attachments.length > 0 ? cleanAttachments(attachments) : undefined,
    };
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitted(true);

    if (!canSubmit) {
      toast.error({
        title: "Missing or invalid fields",
        subtitle:
          attachmentValidationError ||
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
      toast.error({
        title: deliveryMode === "send" ? "Failed to send email" : "Failed to queue email",
        subtitle: getApiErrorMessage(error) ?? "Please check the email service logs and try again.",
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
                    invalid={recipientError || recipientRequiredError}
                    invalidText={
                      recipientRequiredError
                        ? "Enter a recipient email or select at least one group or health context."
                        : "Enter a valid recipient email address"
                    }
                    disabled={isSubmitting}
                  />

                  <MultiSelect
                    id="recipient-groups"
                    titleText="Recipient groups"
                    label={groupsError ? "Unable to load groups" : "Select groups"}
                    items={groupItems}
                    itemToString={(item) => item?.text ?? ""}
                    selectedItems={groupItems.filter((item) => recipientGroupIds.includes(item.id))}
                    disabled={isSubmitting || groupsLoading || groupsError}
                    invalid={recipientRequiredError}
                    invalidText="Enter a recipient email or select at least one group or health context."
                    onChange={({ selectedItems }) =>
                      setRecipientGroupIds((selectedItems ?? []).map((item) => item.id))
                    }
                  />

                  <MultiSelect
                    id="recipient-health-contexts"
                    titleText="Recipient health contexts"
                    label="Select health contexts"
                    items={healthContextItems}
                    itemToString={(item) => item?.text ?? ""}
                    selectedItems={healthContextItems.filter((item) =>
                      recipientHealthContextIds.includes(item.id),
                    )}
                    disabled={isSubmitting}
                    invalid={recipientRequiredError}
                    invalidText="Enter a recipient email or select at least one group or health context."
                    onChange={({ selectedItems }) =>
                      setRecipientHealthContextIds((selectedItems ?? []).map((item) => item.id))
                    }
                  />

                  {hasHealthContextRecipients ? (
                    <Checkbox
                      id="recipient-health-context-descendants"
                      labelText="Include users assigned to descendant health contexts"
                      checked={includeHealthContextDescendants}
                      onChange={(_, { checked }) =>
                        setIncludeHealthContextDescendants(Boolean(checked))
                      }
                      disabled={isSubmitting}
                    />
                  ) : null}

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

              <div className="email-form__section">
                <h3 className="email-form__section-title">Attachments</h3>

                <Stack gap={4}>
                  <div className="email-form__attachment-upload">
                    <TextInput
                      id="email-attachment-file"
                      type="file"
                      labelText="Upload files"
                      helperText="Files are sent as base64 payloads. Maximum 10 MB per file."
                      multiple
                      disabled={isSubmitting}
                      onChange={(event) => {
                        void handleFileAttachments(event.target.files);
                        event.currentTarget.value = "";
                      }}
                    />
                  </div>

                  <div className="email-form__path-attachment">
                    <TextInput
                      id="email-path-attachment-name"
                      labelText="Server path attachment file name"
                      value={pathFileName}
                      onChange={(event) => setPathFileName(event.target.value)}
                      placeholder="report.pdf"
                      disabled={isSubmitting}
                    />

                    <TextInput
                      id="email-path-attachment-content-type"
                      labelText="Content type"
                      value={pathContentType}
                      onChange={(event) => setPathContentType(event.target.value)}
                      placeholder="application/pdf"
                      disabled={isSubmitting}
                    />

                    <TextInput
                      id="email-path-attachment-path"
                      labelText="Server path"
                      value={pathValue}
                      onChange={(event) => setPathValue(event.target.value)}
                      placeholder="/data/uploads/report.pdf"
                      disabled={isSubmitting}
                    />

                    <Checkbox
                      id="email-path-attachment-inline"
                      labelText="Inline"
                      checked={pathInline}
                      disabled={isSubmitting}
                      onChange={(_, data) => setPathInline(Boolean(data.checked))}
                    />

                    {pathInline && (
                      <TextInput
                        id="email-path-attachment-content-id"
                        labelText="Content ID"
                        value={pathContentId}
                        onChange={(event) => setPathContentId(event.target.value)}
                        placeholder="logo"
                        disabled={isSubmitting}
                      />
                    )}

                    <Button
                      type="button"
                      kind="tertiary"
                      size="sm"
                      disabled={isSubmitting || !pathFileName.trim() || !pathValue.trim()}
                      onClick={addPathAttachment}
                    >
                      Add path attachment
                    </Button>
                  </div>

                  {attachments.length === 0 ? (
                    <p className="email-form__attachment-empty">No attachments added.</p>
                  ) : (
                    <ul className="email-form__attachments">
                      {attachments.map((attachment) => (
                        <li key={attachment.id} className="email-form__attachment">
                          <div className="email-form__attachment-main">
                            <strong>{attachment.file_name}</strong>
                            <span>
                              {attachment.content_type || "application/octet-stream"}
                              {attachment.file_size
                                ? ` · ${formatFileSize(attachment.file_size)}`
                                : ""}
                            </span>
                            <Tag type={attachment.source === "file" ? "blue" : "purple"} size="sm">
                              {attachment.source === "file" ? "File" : "Path"}
                            </Tag>
                          </div>

                          <div className="email-form__attachment-meta">
                            <Checkbox
                              id={`email-attachment-inline-${attachment.id}`}
                              labelText="Inline"
                              checked={Boolean(attachment.inline)}
                              disabled={isSubmitting}
                              onChange={(_, data) =>
                                updateAttachment(attachment.id, {
                                  inline: Boolean(data.checked),
                                })
                              }
                            />

                            {attachment.inline && (
                              <TextInput
                                id={`email-attachment-content-id-${attachment.id}`}
                                labelText="Content ID"
                                value={attachment.content_id ?? ""}
                                disabled={isSubmitting}
                                onChange={(event) =>
                                  updateAttachment(attachment.id, {
                                    content_id: event.target.value,
                                  })
                                }
                              />
                            )}

                            <Button
                              type="button"
                              kind="ghost"
                              size="sm"
                              renderIcon={TrashCan}
                              iconDescription="Remove attachment"
                              disabled={isSubmitting}
                              onClick={() => removeAttachment(attachment.id)}
                            >
                              Remove
                            </Button>
                          </div>
                        </li>
                      ))}
                    </ul>
                  )}
                </Stack>
              </div>

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
