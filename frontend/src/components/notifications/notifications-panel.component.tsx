import {
  StructuredListWrapper,
  StructuredListBody,
  StructuredListRow,
  StructuredListCell,
  Tag,
  Button,
  Stack,
} from "@carbon/react";
import type { Notification } from "../../store/types/notifications.types";

import "./notifications-panel.css";
import { getSeverityTagType } from "../../ui/severity";

type Props = {
  notifications: Notification[];
  onMarkRead?: (id: string) => void;
  onView?: (notification: Notification) => void;
};

export function NotificationsPanel({ notifications, onMarkRead, onView }: Props) {
  if (notifications.length === 0) {
    return <p className="notifications-panel__empty">No notifications</p>;
  }

  return (
    <div className="notifications-panel">
      <StructuredListWrapper>
        <StructuredListBody>
          {notifications.map((n) => (
            <StructuredListRow
              key={n.id}
              className={`notification-row ${!n.read ? "notification-row--unread" : ""}`}
              tabIndex={0}
              onClick={() => onView?.(n)}
            >
              {/* ================= Content ================= */}
              <StructuredListCell>
                <Stack gap={1}>
                  <span className="notification-title">{n.title}</span>

                  <span className="notification-message">{n.message}</span>

                  <span className="notification-meta">
                    {new Date(n.created_at).toLocaleString()}
                  </span>
                </Stack>
              </StructuredListCell>

              {/* ================= Actions ================= */}
              <StructuredListCell className="notification-actions">
                <Stack gap={2}>
                  <Tag size="sm" type={getSeverityTagType(n.severity)}>
                    {n.severity}
                  </Tag>

                  {!n.read && onMarkRead && (
                    <Button
                      size="sm"
                      kind="ghost"
                      onClick={(e) => {
                        e.stopPropagation();
                        onMarkRead(n.id);
                      }}
                    >
                      Mark read
                    </Button>
                  )}
                </Stack>
              </StructuredListCell>
            </StructuredListRow>
          ))}
        </StructuredListBody>
      </StructuredListWrapper>
    </div>
  );
}
