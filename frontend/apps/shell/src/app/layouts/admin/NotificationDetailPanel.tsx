import { Button, InlineLoading, Stack, Tag } from "@carbon/react";
import { Renew } from "@carbon/react/icons";

import {
  useGetNotificationDeliveriesQuery,
  useRetryNotificationDeliveryMutation,
} from "../../api";
import type { Notification, NotificationDelivery } from "@moh-sso/types";
import { getSeverityTagType } from "@moh-sso/ui";

type Props = {
  notification: Notification;
  canRetryDelivery?: boolean;
};

function formatDate(value?: string | null): string {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function deliveryStatusTag(status: string): "green" | "red" | "warm-gray" | "blue" | "gray" {
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

function canRetry(delivery: NotificationDelivery): boolean {
  const status = delivery.status.toUpperCase();
  return status === "FAILED" || status === "RETRY" || status === "CANCELLED";
}

function stringifyMetadata(value?: Record<string, unknown>): string {
  if (!value || Object.keys(value).length === 0) {
    return "";
  }

  return JSON.stringify(value, null, 2);
}

export function NotificationDetailPanel({ notification, canRetryDelivery = false }: Props) {
  const {
    data: deliveries = [],
    isLoading,
    isFetching,
    refetch,
  } = useGetNotificationDeliveriesQuery(notification.id);
  const [retryDelivery, { isLoading: isRetrying }] = useRetryNotificationDeliveryMutation();
  const metadata = stringifyMetadata(notification.metadata);

  return (
    <Stack gap={6} className="notification-detail-panel">
      <section className="notification-detail-panel__summary">
        <div className="notification-detail-panel__title-row">
          <h4>{notification.title}</h4>
          <Tag size="sm" type={getSeverityTagType(notification.severity)}>
            {notification.severity}
          </Tag>
        </div>

        <p>{notification.message}</p>

        <dl className="notification-detail-panel__meta">
          <div>
            <dt>Status</dt>
            <dd>{notification.read ? "Read" : "Unread"}</dd>
          </div>
          <div>
            <dt>Created</dt>
            <dd>{formatDate(notification.created_at)}</dd>
          </div>
          {notification.type && (
            <div>
              <dt>Type</dt>
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
      </section>

      {metadata && (
        <section>
          <h5 className="notification-detail-panel__section-title">Metadata</h5>
          <pre className="notification-detail-panel__json">{metadata}</pre>
        </section>
      )}

      <section>
        <div className="notification-detail-panel__section-header">
          <div>
            <h5 className="notification-detail-panel__section-title">Delivery history</h5>
            <p>Channel status, retry attempts, and last delivery error.</p>
          </div>
          {isFetching && <InlineLoading description="Refreshing..." />}
        </div>

        {isLoading && <InlineLoading description="Loading delivery history..." />}

        {!isLoading && deliveries.length === 0 && (
          <p className="notification-detail-panel__empty">No delivery records found.</p>
        )}

        {!isLoading && deliveries.length > 0 && (
          <div className="notification-detail-panel__deliveries">
            {deliveries.map((delivery) => (
              <article key={delivery.id} className="notification-detail-panel__delivery">
                <div className="notification-detail-panel__delivery-top">
                  <div>
                    <strong>{delivery.channel}</strong>
                    <span>
                      {delivery.attempts}/{delivery.max_attempts} attempts
                    </span>
                  </div>
                  <Tag size="sm" type={deliveryStatusTag(delivery.status)}>
                    {delivery.status}
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

                {canRetryDelivery && canRetry(delivery) && (
                  <Button
                    size="sm"
                    kind="tertiary"
                    renderIcon={Renew}
                    disabled={isRetrying}
                    onClick={() => {
                      void retryDelivery({
                        deliveryId: delivery.id,
                        notificationId: notification.id,
                      }).then(() => {
                        void refetch();
                      });
                    }}
                  >
                    Retry delivery
                  </Button>
                )}
              </article>
            ))}
          </div>
        )}
      </section>
    </Stack>
  );
}
