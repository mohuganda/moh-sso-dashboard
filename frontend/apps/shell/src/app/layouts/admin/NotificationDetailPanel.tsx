import { Button, InlineLoading, InlineNotification, Tag } from "@carbon/react";
import { ArrowLeft, Checkmark, Renew } from "@carbon/react/icons";
import { useEffect, useRef, useState } from "react";
import {
  useGetNotificationQuery,
  useGetNotificationDeliveriesQuery,
  useRetryNotificationDeliveryMutation,
  useMarkNotificationAsReadMutation,
} from "../../api";
import type { Notification } from "@moh-sso/types";
import { getSeverityTagType, getSeverityLabel } from "@moh-sso/ui";

function formatDate(value?: string | null): string {
  const date = value ? new Date(value) : null;
  return !date || Number.isNaN(date.getTime())
    ? "Not available"
    : new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}

function statusTag(status: string): "green" | "red" | "warm-gray" | "blue" | "gray" {
  switch (status.toUpperCase()) {
    case "SENT":
      return "green";
    case "FAILED":
      return "red";
    case "RETRY":
      return "warm-gray";
    case "PROCESSING":
      return "blue";
    default:
      return "gray";
  }
}
function label(value: string): string {
  return value
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/^./, (first) => first.toUpperCase());
}

export function NotificationDetailPanel({
  notification: initial,
  onBack,
  canRetryDelivery = false,
}: {
  notification: Notification;
  onBack: () => void;
  canRetryDelivery?: boolean;
}) {
  const { data: notification = initial, isError: detailError } = useGetNotificationQuery(
    initial.id,
  );
  const {
    data: deliveries = [],
    isLoading,
    isError,
    refetch,
  } = useGetNotificationDeliveriesQuery(initial.id);
  const [retryDelivery] = useRetryNotificationDeliveryMutation();
  const [markRead, { isLoading: marking }] = useMarkNotificationAsReadMutation();
  const [retrying, setRetrying] = useState<string | null>(null);
  const [retryError, setRetryError] = useState<string | null>(null);
  const [error, setError] = useState("");
  const back = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    back.current?.focus();
  }, []);

  return (
    <div className="notification-detail-panel">
      <Button ref={back} kind="ghost" size="sm" renderIcon={ArrowLeft} onClick={onBack}>
        Back to notifications
      </Button>
      {detailError && (
        <InlineNotification
          kind="warning"
          title="Unable to refresh this notification"
          subtitle="Showing the last available details."
          lowContrast
          hideCloseButton
        />
      )}
      {error && (
        <InlineNotification
          kind="error"
          title="Action failed"
          subtitle={error}
          lowContrast
          hideCloseButton
        />
      )}
      <section className="notification-detail-panel__summary" aria-label="Notification summary">
        <div className="notification-detail-panel__title-row">
          <h3>{notification.title}</h3>
          <Tag size="sm" type={getSeverityTagType(notification.severity)}>
            {getSeverityLabel(notification.severity)}
          </Tag>
        </div>
        <p>{notification.message}</p>
        <dl className="notification-detail-panel__meta">
          <div>
            <dt>Status</dt>
            <dd>{notification.read ? "Read" : "Unread"}</dd>
          </div>
          <div>
            <dt>Received</dt>
            <dd>{formatDate(notification.created_at)}</dd>
          </div>
        </dl>
        {!notification.read && (
          <Button
            kind="tertiary"
            size="sm"
            renderIcon={Checkmark}
            disabled={marking}
            onClick={async () => {
              setError("");
              try {
                await markRead(notification.id).unwrap();
                back.current?.focus();
              } catch {
                setError("Could not mark this notification as read. Please try again.");
              }
            }}
          >
            {marking ? "Marking read…" : "Mark as read"}
          </Button>
        )}
      </section>
      <section aria-labelledby="notification-delivery-heading">
        <div className="notification-detail-panel__section-header">
          <h4 id="notification-delivery-heading">Delivery history</h4>
          <p>Status and attempts for each delivery channel.</p>
        </div>
        {isLoading && <InlineLoading description="Loading delivery history…" />}
        {isError ? (
          <div role="alert" className="notification-detail-panel__empty">
            <p>Unable to load delivery history.</p>
            <Button kind="tertiary" size="sm" onClick={() => void refetch()}>
              Try again
            </Button>
          </div>
        ) : !isLoading && deliveries.length === 0 ? (
          <p className="notification-detail-panel__empty">No delivery records yet.</p>
        ) : (
          <div className="notification-detail-panel__deliveries">
            {deliveries.map((delivery) => (
              <article
                key={delivery.id}
                className="notification-detail-panel__delivery"
                aria-label={`${label(delivery.channel)} delivery`}
              >
                <div className="notification-detail-panel__delivery-top">
                  <div className="notification-detail-panel__channel">
                    <strong>{label(delivery.channel)}</strong>
                    <span>
                      {delivery.attempts} of {delivery.max_attempts} attempts
                    </span>
                  </div>
                  <Tag size="sm" type={statusTag(delivery.status)}>
                    {label(delivery.status)}
                  </Tag>
                </div>
                <dl className="notification-detail-panel__meta">
                  <div>
                    <dt>Scheduled</dt>
                    <dd>{formatDate(delivery.scheduled_at)}</dd>
                  </div>
                  <div>
                    <dt>Sent</dt>
                    <dd>{formatDate(delivery.sent_at)}</dd>
                  </div>
                  <div>
                    <dt>Updated</dt>
                    <dd>{formatDate(delivery.updated_at)}</dd>
                  </div>
                </dl>
                {delivery.last_error && (
                  <p className="notification-detail-panel__error">{delivery.last_error}</p>
                )}
                {canRetryDelivery &&
                  ["FAILED", "RETRY", "CANCELLED"].includes(delivery.status.toUpperCase()) && (
                    <Button
                      kind="tertiary"
                      size="sm"
                      renderIcon={Renew}
                      disabled={retrying !== null}
                      aria-label={`Retry ${label(delivery.channel)} delivery`}
                      onClick={async () => {
                        setRetrying(delivery.id);
                        setRetryError(null);
                        try {
                          await retryDelivery({
                            deliveryId: delivery.id,
                            notificationId: notification.id,
                          }).unwrap();
                        } catch {
                          setRetryError(delivery.id);
                        } finally {
                          setRetrying(null);
                        }
                      }}
                    >
                      {retrying === delivery.id ? "Retrying…" : "Retry delivery"}
                    </Button>
                  )}
                {retryError === delivery.id && (
                  <p role="alert" className="notification-detail-panel__error">
                    Could not retry the delivery. Please try again.
                  </p>
                )}
              </article>
            ))}
          </div>
        )}
      </section>
      {(notification.type ||
        notification.target_role ||
        Object.keys(notification.metadata ?? {}).length > 0) && (
        <details className="notification-detail-panel__technical">
          <summary>Technical details</summary>
          <dl className="notification-detail-panel__meta">
            {notification.type && (
              <div>
                <dt>Event type</dt>
                <dd>{notification.type}</dd>
              </div>
            )}
            {notification.target_role && (
              <div>
                <dt>Target role</dt>
                <dd>{notification.target_role}</dd>
              </div>
            )}
          </dl>
          {Object.keys(notification.metadata ?? {}).length > 0 && (
            <pre className="notification-detail-panel__json">
              {JSON.stringify(notification.metadata, null, 2)}
            </pre>
          )}
        </details>
      )}
    </div>
  );
}
