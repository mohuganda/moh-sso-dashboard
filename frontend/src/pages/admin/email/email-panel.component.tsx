import React, { useMemo, useState } from "react";
import {
  Button,
  Form,
  InlineNotification,
  Loading,
  Stack,
  TextArea,
  TextInput,
  Tile,
} from "@carbon/react";

import { useQueueEmailMutation, useSendEmailMutation } from "../../../store/api/email.api";
import "./email.scss";

type DeliveryMode = "send" | "queue";

const EmailPanelComponent: React.FC = () => {
  const [recipient, setRecipient] = useState("");
  const [recipientName, setRecipientName] = useState("");
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [deliveryMode, setDeliveryMode] = useState<DeliveryMode>("queue");
  const [status, setStatus] = useState<{
    kind: "success" | "error" | "info";
    title: string;
    subtitle?: string;
  } | null>(null);

  const [sendEmail, sendState] = useSendEmailMutation();
  const [queueEmail, queueState] = useQueueEmailMutation();

  const isSubmitting = sendState.isLoading || queueState.isLoading;

  const canSubmit = useMemo(() => {
    return (
      recipient.trim().length > 0 &&
      subject.trim().length > 0 &&
      message.trim().length > 0 &&
      !isSubmitting
    );
  }, [recipient, subject, message, isSubmitting]);

  const resetForm = () => {
    setRecipient("");
    setRecipientName("");
    setSubject("");
    setMessage("");
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!canSubmit) {
      setStatus({
        kind: "error",
        title: "Missing required fields",
        subtitle: "Please provide recipient, subject, and message.",
      });
      return;
    }

    setStatus({
      kind: "info",
      title: deliveryMode === "send" ? "Sending email..." : "Queueing email...",
    });

    const payload = {
      to: [
        {
          name: recipientName.trim() || undefined,
          email: recipient.trim(),
        },
      ],
      subject: subject.trim(),
      text_body: message.trim(),
      metadata: {
        source: "admin-email-page",
      },
    };

    try {
      if (deliveryMode === "send") {
        await sendEmail(payload).unwrap();

        setStatus({
          kind: "success",
          title: "Email sent successfully",
          subtitle: "The email was delivered through the SMTP service.",
        });
      } else {
        await queueEmail(payload).unwrap();

        setStatus({
          kind: "success",
          title: "Email queued successfully",
          subtitle: "The background email worker will process it shortly.",
        });
      }

      resetForm();
    } catch (error) {
      console.error(error);

      setStatus({
        kind: "error",
        title: deliveryMode === "send" ? "Failed to send email" : "Failed to queue email",
        subtitle: "Please check the email service logs and try again.",
      });
    }
  };

  return (
    <div className="email-page">
      <Tile className="email-page__tile">
        <Stack gap={6}>
          <div>
            <h1 className="email-page__title">Send Admin Email</h1>
            <p className="email-page__description">
              Send an email immediately or queue it for background delivery.
            </p>
          </div>

          {status && (
            <InlineNotification
              kind={status.kind}
              title={status.title}
              subtitle={status.subtitle}
              lowContrast
              onCloseButtonClick={() => setStatus(null)}
            />
          )}

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
                invalid={recipient.trim() === ""}
                invalidText="Recipient email is required"
                disabled={isSubmitting}
                required
              />

              <TextInput
                id="subject"
                labelText="Subject"
                value={subject}
                onChange={(event) => setSubject(event.target.value)}
                placeholder="Email subject"
                invalid={subject.trim() === ""}
                invalidText="Subject is required"
                disabled={isSubmitting}
                required
              />

              <TextArea
                id="message"
                labelText="Message"
                value={message}
                onChange={(event) => setMessage(event.target.value)}
                placeholder="Write your message here"
                rows={8}
                invalid={message.trim() === ""}
                invalidText="Message is required"
                disabled={isSubmitting}
                required
              />

              <div className="email-form__actions">
                <Button type="submit" disabled={!canSubmit}>
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
