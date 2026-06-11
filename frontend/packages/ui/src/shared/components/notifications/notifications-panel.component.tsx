import { Button, SkeletonText, Tag } from "@carbon/react";
import { Checkmark, Information, Time } from "@carbon/react/icons";

import type { Notification } from "@moh-sso/types";
import { getSeverityTagType } from "@moh-sso/ui";

import "./notifications-panel.css";

type Props = {
  notifications: Notification[];
  isLoading?: boolean;
  onMarkRead?: (id: string) => void;
  onView?: (notification: Notification) => void;
};

function formatNotificationDate(dateString?: string | null): string {
  if (!dateString) {
    return "Recently";
  }

  const date = new Date(dateString);

  if (Number.isNaN(date.getTime())) {
    return "Recently";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatRelativeTime(dateString?: string | null): string {
  if (!dateString) {
    return "Recently";
  }

  const date = new Date(dateString);

  if (Number.isNaN(date.getTime())) {
    return "Recently";
  }

  const diffMs = Date.now() - date.getTime();

  if (diffMs < 0) {
    return formatNotificationDate(dateString);
  }

  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) {
    return "Just now";
  }

  if (diffMinutes < 60) {
    return `${diffMinutes} minute${diffMinutes === 1 ? "" : "s"} ago`;
  }

  if (diffHours < 24) {
    return `${diffHours} hour${diffHours === 1 ? "" : "s"} ago`;
  }

  if (diffDays < 7) {
    return `${diffDays} day${diffDays === 1 ? "" : "s"} ago`;
  }

  return formatNotificationDate(dateString);
}

function normalizeSeverity(severity?: string | null): string {
  return severity?.trim() || "INFO";
}

function LoadingNotifications() {
  return (
    <div className="notifications-panel__list" aria-label="Loading notifications">
      {Array.from({ length: 4 }).map((_, index) => (
        <div className="notification-card notification-card--loading" key={index}>
          <SkeletonText width="45%" />
          <SkeletonText paragraph lineCount={2} />
          <SkeletonText width="30%" />
        </div>
      ))}
    </div>
  );
}

export function NotificationsPanel({
  notifications,
  isLoading = false,
  onMarkRead,
  onView,
}: Props) {
  const unreadCount = notifications.filter((item) => !item.read).length;

  if (isLoading && notifications.length === 0) {
    return (
      <div className="notifications-panel">
        <div className="notifications-panel__header">
          <div>
            <h4 className="notifications-panel__title">Notifications</h4>
            <p className="notifications-panel__subtitle">Loading recent activity...</p>
          </div>
        </div>

        <LoadingNotifications />
      </div>
    );
  }

  if (notifications.length === 0) {
    return (
      <div className="notifications-panel notifications-panel--empty">
        <div className="notifications-panel__empty-icon">
          <Checkmark size={24} />
        </div>

        <h4 className="notifications-panel__empty-title">No notifications</h4>

        <p className="notifications-panel__empty">
          You are all caught up. New alerts and system updates will appear here.
        </p>
      </div>
    );
  }

  return (
    <div className="notifications-panel">
      <div className="notifications-panel__header">
        <div>
          <h4 className="notifications-panel__title">Notifications</h4>
          <p className="notifications-panel__subtitle">
            {unreadCount > 0
              ? `${unreadCount} unread notification${unreadCount === 1 ? "" : "s"}`
              : "You are all caught up"}
          </p>
        </div>

        {unreadCount > 0 && (
          <Tag type="blue" size="sm">
            {unreadCount} unread
          </Tag>
        )}
      </div>

      <div className="notifications-panel__list">
        {notifications.map((notification) => {
          const severity = normalizeSeverity(notification.severity);
          const isUnread = !notification.read;

          return (
            <article
              key={notification.id}
              className={["notification-card", isUnread ? "notification-card--unread" : ""]
                .filter(Boolean)
                .join(" ")}
              role={onView ? "button" : "article"}
              tabIndex={onView ? 0 : undefined}
              onClick={() => onView?.(notification)}
              onKeyDown={(event) => {
                if (!onView) {
                  return;
                }

                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onView(notification);
                }
              }}
            >
              <div className="notification-card__indicator" aria-hidden="true">
                <Information size={16} />
              </div>

              <div className="notification-card__content">
                <div className="notification-card__top">
                  <div className="notification-card__title-wrap">
                    <h5 className="notification-title">{notification.title}</h5>

                    {isUnread && (
                      <span className="notification-unread-dot" aria-label="Unread notification" />
                    )}
                  </div>

                  <Tag size="sm" type={getSeverityTagType(severity)}>
                    {severity}
                  </Tag>
                </div>

                <p className="notification-message">{notification.message}</p>

                <div
                  className="notification-meta"
                  title={formatNotificationDate(notification.created_at)}
                >
                  <Time size={14} />
                  <span>{formatRelativeTime(notification.created_at)}</span>
                </div>
              </div>

              {isUnread && onMarkRead && (
                <div className="notification-actions">
                  <Button
                    size="sm"
                    kind="ghost"
                    renderIcon={Checkmark}
                    iconDescription="Mark notification as read"
                    onClick={(event) => {
                      event.stopPropagation();
                      onMarkRead(notification.id);
                    }}
                  >
                    Mark read
                  </Button>
                </div>
              )}
            </article>
          );
        })}
      </div>
    </div>
  );
}
