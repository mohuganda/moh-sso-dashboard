import { Modal, Stack, Tag } from "@carbon/react";

import type { EmailOutboxItem } from "../../../store/types/email.types";
import "./email-outbox-details.scss";

type Props = {
  email: EmailOutboxItem | null;
  open: boolean;
  onClose: () => void;
};

function formatDate(value?: string | null) {
  if (!value) return "—";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";

  return date.toLocaleString();
}

function getRecipients(email?: EmailOutboxItem | null) {
  return email?.message?.to?.map((recipient) => recipient.email).join(", ") || "—";
}

function getCcRecipients(email?: EmailOutboxItem | null) {
  return email?.message?.cc?.map((recipient) => recipient.email).join(", ") || "—";
}

function getBccRecipients(email?: EmailOutboxItem | null) {
  return email?.message?.bcc?.map((recipient) => recipient.email).join(", ") || "—";
}

function getSubject(email?: EmailOutboxItem | null) {
  return email?.message?.subject || "—";
}

function getBody(email?: EmailOutboxItem | null) {
  return email?.message?.html_body || email?.message?.text_body || "No message body available.";
}

function DetailItem({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="email-details__item">
      <dt className="email-details__label">{label}</dt>
      <dd className="email-details__value">{children}</dd>
    </div>
  );
}

export function EmailDetailsModal({ email, open, onClose }: Props) {
  return (
    <Modal open={open} modalHeading="Email details" passiveModal onRequestClose={onClose} size="lg">
      {email && (
        <Stack gap={5}>
          <dl className="email-details">
            <DetailItem label="Subject">{getSubject(email)}</DetailItem>

            <DetailItem label="To">{getRecipients(email)}</DetailItem>

            {email.message?.cc?.length ? (
              <DetailItem label="CC">{getCcRecipients(email)}</DetailItem>
            ) : null}

            {email.message?.bcc?.length ? (
              <DetailItem label="BCC">{getBccRecipients(email)}</DetailItem>
            ) : null}

            <DetailItem label="Status">
              <Tag type={getStatusTagType(email.status)}>{email.status}</Tag>
            </DetailItem>

            <DetailItem label="Attempts">
              {email.attempts ?? 0}/{email.max_attempts ?? "—"}
            </DetailItem>

            <DetailItem label="Scheduled at">{formatDate(email.scheduled_at)}</DetailItem>

            <DetailItem label="Locked at">{formatDate(email.locked_at)}</DetailItem>

            <DetailItem label="Sent at">{formatDate(email.sent_at)}</DetailItem>

            <DetailItem label="Created at">{formatDate(email.created_at)}</DetailItem>

            <DetailItem label="Updated at">{formatDate(email.updated_at)}</DetailItem>

            {email.message?.template_name ? (
              <DetailItem label="Template">{email.message.template_name}</DetailItem>
            ) : null}
          </dl>

          {email.last_error && (
            <section className="email-details__section">
              <h4 className="email-details__section-title">Last error</h4>
              <pre className="email-outbox-page__error">{email.last_error}</pre>
            </section>
          )}

          {email.message?.attachments?.length ? (
            <section className="email-details__section">
              <h4 className="email-details__section-title">Attachments</h4>

              <ul className="email-details__attachments">
                {email.message.attachments.map((attachment, index) => (
                  <li key={`${attachment.file_name}-${index}`}>
                    <strong>{attachment.file_name || "Attachment"}</strong>
                    {attachment.content_type ? <span> · {attachment.content_type}</span> : null}
                    {attachment.path ? <div>{attachment.path}</div> : null}
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          <section className="email-details__section">
            <h4 className="email-details__section-title">Message</h4>
            <pre className="email-outbox-page__body">{getBody(email)}</pre>
          </section>
        </Stack>
      )}
    </Modal>
  );
}

function getStatusTagType(status?: string) {
  switch (status) {
    case "SENT":
      return "green";
    case "FAILED":
      return "red";
    case "PROCESSING":
      return "blue";
    case "RETRY":
      return "magenta";
    case "PENDING":
      return "gray";
    default:
      return "cool-gray";
  }
}
