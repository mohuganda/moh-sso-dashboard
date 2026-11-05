import { Button, InlineNotification, SkeletonText, Tag } from "@carbon/react";
import {
  Checkmark,
  CheckmarkFilled,
  Information,
  TrashCan,
  WarningAltFilled,
} from "@carbon/react/icons";
import { useEffect, useId, useRef, useState } from "react";

import type { Notification } from "@moh-sso/types";
import { getSeverityLabel, getSeverityTagType, normalizeSeverity } from "../utils/severity";
import "./notifications-panel.scss";

type Props = {
  notifications: Notification[];
  unreadCount?: number;
  isLoading?: boolean;
  isError?: boolean;
  showHeading?: boolean;
  onRetry?: () => void;
  onMarkRead?: (id: string) => void | Promise<void>;
  onDelete?: (id: string) => void | Promise<void>;
  onView?: (notification: Notification) => void;
};

function notificationDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "Date unavailable"
    : new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}

export function NotificationsPanel({
  notifications,
  unreadCount,
  isLoading = false,
  isError = false,
  showHeading = true,
  onRetry,
  onMarkRead,
  onDelete,
  onView,
}: Props) {
  const [pending, setPending] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [error, setError] = useState("");
  const id = useId();
  const root = useRef<HTMLDivElement>(null);
  const actionFocus = useRef<HTMLElement | null>(null);
  const keepButton = useRef<HTMLButtonElement>(null);
  const unread = notifications.filter((item) => !item.read);
  const items = notifications.slice(0, 10);
  const count = unreadCount ?? unread.length;
  const loading = isLoading;
  const failed = isError;

  useEffect(() => {
    if (confirmDelete) keepButton.current?.focus();
  }, [confirmDelete]);

  useEffect(() => {
    if (actionFocus.current && !actionFocus.current.isConnected) {
      if (document.activeElement === document.body) {
        root.current?.querySelector<HTMLButtonElement>(".notification-card__open")?.focus();
      }
      actionFocus.current = null;
    }
  });

  const act = async (notificationId: string, action: "read" | "delete") => {
    actionFocus.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;
    setError("");
    setPending(notificationId);
    try {
      await (action === "read" ? onMarkRead?.(notificationId) : onDelete?.(notificationId));
      setConfirmDelete(null);
    } catch {
      setError(
        action === "read"
          ? "Could not mark the notification as read. Please try again."
          : "Could not delete the notification. Please try again.",
      );
    } finally {
      setPending(null);
    }
  };

  return (
    <div ref={root} className="notifications-panel" aria-busy={loading}>
      <div className="notifications-panel__toolbar">
        <div className="notifications-panel__header">
          <div>
            {showHeading && <h4 className="notifications-panel__title">Notifications</h4>}
            <p className="notifications-panel__subtitle" role="status" aria-live="polite">
              {loading
                ? "Loading activity…"
                : failed
                  ? "Activity is unavailable"
                  : "Most recent notifications"}
            </p>
          </div>
          {count > 0 && (
            <Tag type="blue" size="sm">
              {count} unread
            </Tag>
          )}
        </div>
      </div>
      {error && (
        <InlineNotification
          kind="error"
          title="Action failed"
          subtitle={error}
          lowContrast
          hideCloseButton
        />
      )}
      {failed && (
        <div className="notifications-panel__state" role="alert">
          <Information size={24} />
          <h5>Unable to load notifications</h5>
          <p>Please try again. Your notifications have not been changed.</p>
          {onRetry && (
            <Button kind="tertiary" size="sm" onClick={onRetry}>
              Try again
            </Button>
          )}
        </div>
      )}
      {!failed && loading && items.length === 0 ? (
        <div aria-label="Loading notifications">
          {[0, 1, 2].map((key) => (
            <div className="notification-card" key={key}>
              <SkeletonText paragraph lineCount={2} />
            </div>
          ))}
        </div>
      ) : !failed && items.length === 0 ? (
        <div className="notifications-panel__state">
          <CheckmarkFilled size={32} className="notifications-panel__empty-icon" />
          <h5>No recent notifications</h5>
          <p>New alerts and updates will appear here.</p>
        </div>
      ) : (
        !failed && (
          <ul className="notifications-panel__list" aria-label="Recent notifications">
            {items.map((item) => {
              const severity = normalizeSeverity(item.severity);
              const Icon =
                severity === "warning" || severity === "critical" || severity === "danger"
                  ? WarningAltFilled
                  : Information;
              const content = (
                <>
                  <span
                    className={`notification-card__icon notification-card__icon--${severity}`}
                    aria-hidden="true"
                  >
                    <Icon size={20} />
                  </span>
                  <span className="notification-card__content">
                    <span className="notification-card__top">
                      <span className="notification-card__title">{item.title}</span>
                      <Tag size="sm" type={getSeverityTagType(severity)}>
                        {getSeverityLabel(severity)}
                      </Tag>
                    </span>
                    <span className="notification-card__message">{item.message}</span>
                    <span className="notification-card__meta">
                      <span className={item.read ? "" : "notification-card__unread"}>
                        {item.read ? "Read" : "Unread"}
                      </span>
                      <time>{notificationDate(item.created_at)}</time>
                    </span>
                  </span>
                </>
              );
              return (
                <li key={item.id}>
                  <article
                    className={`notification-card ${item.read ? "" : "notification-card--unread"}`}
                    aria-label={item.title}
                  >
                    {onView ? (
                      <button
                        type="button"
                        className="notification-card__open"
                        onClick={() => onView(item)}
                        aria-label={`View ${item.title}`}
                      >
                        {content}
                      </button>
                    ) : (
                      <div className="notification-card__open">{content}</div>
                    )}
                    {(onMarkRead || onDelete) && (
                      <div className="notification-card__actions">
                        {confirmDelete === item.id ? (
                          <>
                            <span id={`${id}-confirm`}>Delete this notification?</span>
                            <Button
                              ref={keepButton}
                              kind="ghost"
                              size="sm"
                              disabled={pending !== null}
                              onClick={() => {
                                setConfirmDelete(null);
                                requestAnimationFrame(() =>
                                  document.getElementById(`${id}-delete-${item.id}`)?.focus(),
                                );
                              }}
                            >
                              Keep notification
                            </Button>
                            <Button
                              kind="danger--tertiary"
                              size="sm"
                              disabled={pending !== null}
                              aria-describedby={`${id}-confirm`}
                              onClick={() => void act(item.id, "delete")}
                            >
                              {pending === item.id ? "Deleting…" : "Confirm delete"}
                            </Button>
                          </>
                        ) : (
                          <>
                            {!item.read && onMarkRead && (
                              <Button
                                kind="ghost"
                                size="sm"
                                renderIcon={Checkmark}
                                disabled={pending !== null}
                                hasIconOnly
                                iconDescription={`Mark ${item.title} as read`}
                                aria-label={`Mark ${item.title} as read`}
                                onClick={() => void act(item.id, "read")}
                              />
                            )}
                            {onDelete && (
                              <Button
                                kind="ghost"
                                size="sm"
                                hasIconOnly
                                renderIcon={TrashCan}
                                disabled={pending !== null}
                                iconDescription={`Delete ${item.title}`}
                                id={`${id}-delete-${item.id}`}
                                tooltipAlignment="end"
                                onClick={() => setConfirmDelete(item.id)}
                              />
                            )}
                          </>
                        )}
                      </div>
                    )}
                  </article>
                </li>
              );
            })}
          </ul>
        )
      )}
      {!failed && items.length > 0 && (
        <p className="notifications-panel__limit">Showing {items.length} latest updates.</p>
      )}
    </div>
  );
}
